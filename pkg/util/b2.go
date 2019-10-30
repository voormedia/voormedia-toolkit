package util

import (
	"context"
	"github.com/kurin/blazer/b2"
	"github.com/pkg/errors"
	"io"
	"os"
	"os/exec"
	"strings"
)

// B2Bucket instance for Backblaze
func B2Bucket(b2id string, b2key string, b2bucket string) (context.Context, *b2.Bucket, error) {
	ctx := context.Background()
	client, err := b2.NewClient(ctx, b2id, b2key)
	if err != nil {
		return nil, nil, errors.Errorf("Could not connect to Backblaze B2. Did you set up credentials? Error: %s", err.Error())
	}

	bucket, err := client.Bucket(ctx, b2bucket)
	if err != nil {
		return nil, nil, err
	}

	return ctx, bucket, nil
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
	cmd := exec.Command("openssl", "aes-256-cbc", "-d", "-in", target, "-out", strings.Replace(target, ".encrypted", "", 1), "-pass", "pass:"+encryptionKey)
	cmd.Run()

	return target, nil
}
