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
	fmt.Printf("Uploading backup to Backblaze B2...\n")
	file, err := os.Open("/tmp/" + fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	object := bucket.Object(database + "/" + fileName)
	w := object.NewWriter(ctx)
	if _, err := io.Copy(w, file); err != nil {
		w.Close()
		return err
	}
	w.Close()
	return nil
}

// B2Object download and decrypt an object from Backblaze
func B2Object(ctx context.Context, bucket *b2.Bucket, fileName string, encryptionKey string) (string, error) {
	splitFileName := strings.Split(fileName, "/")
	target := "/tmp/" + splitFileName[len(splitFileName)-1]
	r := bucket.Object(fileName).NewReader(ctx)
	defer r.Close()

	f, err := os.Create(target)
	if err != nil {
		return "", err
	}

	r.ConcurrentDownloads = 1
	if _, err := io.Copy(f, r); err != nil {
		f.Close()
		return "", err
	}

	f.Close()
	cmd := exec.Command("openssl", "aes-256-cbc", "-md", "md5", "-d", "-in", target, "-out", strings.Replace(target, ".encrypted", "", 1), "-pass", "pass:"+encryptionKey)
	cmd.Run()

	return target, nil
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
