package util

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/AlecAivazis/survey"
	"github.com/kurin/blazer/b2"
	"github.com/pkg/errors"
)

// B2Bucket instance for Backblaze
func B2Bucket(b2id string, b2key string, b2bucket string, b2encrypt string, manual bool) (context.Context, *b2.Bucket, string, error) {
	ctx := context.Background()
	client, err := b2.NewClient(ctx, b2id, b2key)
	if err != nil {
		if manual {
			return nil, nil, "", errors.Errorf("Could not connect to Backblaze B2. Please set up credentials.\n\n"+
				"You should add the following environment variables (or pass their values in as arguments):\n"+
				"- B2_ACCOUNT_ID (Your personal App Key ID)\n"+
				"- B2_ACCOUNT_KEY (Your personal App Key secret) \n"+
				"- B2_ENCRYPTION_KEY (The password used to encrypt/decrypt Backblaze backups)\n\n"+
				"Error: %s", err.Error())
		}
		fmt.Printf("Could not connect to Backblaze B2 using environment variables or provided arguments. Please provide credentials.\n")
		return B2Setup(b2id, b2key, b2bucket, b2encrypt)
	}

	bucket, err := client.Bucket(ctx, b2bucket)
	if err != nil {
		return nil, nil, "", err
	}

	return ctx, bucket, b2encrypt, nil
}

// B2Upload an object to Backblaze
func B2Upload(ctx context.Context, bucket *b2.Bucket, database string, fileName string) error {
	file, err := os.Open("/tmp/" + fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return err
	}

	object := bucket.Object(database + "/" + fileName)
	w := object.NewWriter(ctx)
	pw := NewProgressWriter(w, "Uploading", info.Size())
	if _, err := io.Copy(pw, file); err != nil {
		pw.FinishFail()
		w.Close()
		return err
	}
	pw.Finish()
	w.Close()
	return nil
}

// B2Download ensures /tmp/<filename> contains the named Backblaze object.
// It compares the local file's size against the remote object's size; if
// they match the local copy is treated as valid cache and no transfer
// happens. Otherwise (file missing, partial, or corrupt) the file is
// re-downloaded — written first to a .partial path and renamed on
// success, so an interrupted transfer never leaves a "looks complete"
// file behind. The returned bool reports whether bytes were actually
// transferred. If header is non-empty, it is rendered as the bar's
// header line with the percentage right-anchored alongside it.
func B2Download(ctx context.Context, bucket *b2.Bucket, fileName, header string) (string, bool, error) {
	splitFileName := strings.Split(fileName, "/")
	target := "/tmp/" + splitFileName[len(splitFileName)-1]
	partial := target + ".partial"

	obj := bucket.Object(fileName)
	attrs, err := obj.Attrs(ctx)
	if err != nil {
		return "", false, err
	}

	if info, err := os.Stat(target); err == nil && info.Size() == attrs.Size {
		return target, false, nil
	}

	// Either missing or wrong size — wipe any stale copies and re-download.
	os.Remove(target)
	os.Remove(partial)

	r := obj.NewReader(ctx)
	defer r.Close()

	f, err := os.Create(partial)
	if err != nil {
		return "", false, err
	}

	r.ConcurrentDownloads = 1
	pw := NewProgressWriter(f, "", attrs.Size).SetHeader(header)
	if _, err := io.Copy(pw, r); err != nil {
		pw.FinishFail()
		f.Close()
		os.Remove(partial)
		return "", false, err
	}
	pw.Finish()
	f.Close()

	if err := os.Rename(partial, target); err != nil {
		os.Remove(partial)
		return "", false, err
	}

	return target, true, nil
}

// DecryptBackup decrypts an AES-256-CBC encrypted backup file to dst,
// showing a spinner while openssl runs. The destination is written
// atomically (to a .partial file first, renamed on success) for the same
// reason as B2Download.
func DecryptBackup(src, dst, key string) error {
	partial := dst + ".partial"
	os.Remove(partial)

	sp := StartSpinner("Decrypting backup")
	cmd := exec.Command("openssl", "aes-256-cbc", "-md", "md5", "-d", "-in", src, "-out", partial, "-pass", "pass:"+key)
	if err := cmd.Run(); err != nil {
		sp.StopFail()
		os.Remove(partial)
		return err
	}

	if err := os.Rename(partial, dst); err != nil {
		sp.StopFail()
		os.Remove(partial)
		return err
	}
	sp.Stop()
	return nil
}

// B2Setup credentials for Backblaze manually
func B2Setup(b2id string, b2key string, b2bucket string, b2encrypt string) (context.Context, *b2.Bucket, string, error) {
	var qs = []*survey.Question{
		{
			Name:     "b2id",
			Prompt:   &survey.Input{Message: "B2_ACCOUNT_ID (Your personal App Key ID)", Default: b2id},
			Validate: survey.Required,
		},
		{
			Name:     "b2key",
			Prompt:   &survey.Input{Message: "B2_ACCOUNT_KEY (Your personal App Key secret)", Default: b2key},
			Validate: survey.Required,
		},
		{
			Name:     "b2encrypt",
			Prompt:   &survey.Input{Message: "B2_ENCRYPTION_KEY (The password used to encrypt/decrypt Backblaze backups)", Default: b2encrypt},
			Validate: survey.Required,
		},
		{
			Name:     "b2bucket",
			Prompt:   &survey.Input{Message: "The name of the bucket backups are stored in", Default: b2bucket},
			Validate: survey.Required,
		},
	}

	credentials := struct {
		B2id      string
		B2key     string
		B2bucket  string
		B2encrypt string
	}{}

	err := survey.Ask(qs, &credentials)
	if err != nil {
		return nil, nil, "", err
	}

	return B2Bucket(credentials.B2id, credentials.B2key, credentials.B2bucket, credentials.B2encrypt, true)
}

// Returns the Backblaze B2 configuration for the current GCP project
func GetB2Config() (string, string, string, string) {
	gcloudProject, err := GetCurrentGCPProject()
	if err != nil {
		log.Fatal(err)
	}

	var b2bucketName, b2encrypt, b2id, b2key string
	switch gcloudProject {
	case "taxology-381314":
		b2bucketName = "taxology-eu-db-backups"
		b2encrypt = os.Getenv("B2_TAXOLOGY_ENCRYPTION_KEY")
		b2id = os.Getenv("B2_TAXOLOGY_ACCOUNT_ID")
		b2key = os.Getenv("B2_TAXOLOGY_ACCOUNT_KEY")
	default:
		b2bucketName = "voormedia-eu-db-backups"
		b2encrypt = os.Getenv("B2_ENCRYPTION_KEY")
		b2id = os.Getenv("B2_ACCOUNT_ID")
		b2key = os.Getenv("B2_ACCOUNT_KEY")
	}

	return b2bucketName, b2encrypt, b2id, b2key
}
