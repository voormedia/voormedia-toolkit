## Voormedia Toolkit (Google Cloud Utilities)

## Installing

### Option 1 – prebuilt

1. Ensure you have `~/.bin` directory or similar that is in your `$PATH`
2. Install Voormedia Toolkit

- When using an Apple M1+ Mac run: `rm -f ~/.bin/vmt && curl -L $(curl -s https://api.github.com/repos/voormedia/voormedia-toolkit/releases/latest | grep browser_download_url | grep darwin_arm64 | cut -d '"' -f 4) -o ~/.bin/vmt && chmod +x ~/.bin/vmt`
- When using an Intel Mac run: `rm -f ~/.bin/vmt && curl -L $(curl -s https://api.github.com/repos/voormedia/voormedia-toolkit/releases/latest | grep browser_download_url | grep darwin_amd64 | cut -d '"' -f 4) -o ~/.bin/vmt && chmod +x ~/.bin/vmt`

### Option 2 – from source

1. Make sure you have a working `go` installation
2. Build Voormedia Toolkit from source: `go install github.com/voormedia/voormedia-toolkit`

## Restore script

The default Backblaze bucket is set depending on your selected GCP project (change this with `gcloud init`).

To connect to the Voormedia Backblaze B2 bucket, you should create the following environment variables:

- B2_ACCOUNT_ID (your Backblaze B2 account ID)
- B2_ACCOUNT_KEY (your Backblaze B2 account key)
- B2_ENCRYPTION_KEY (the key used to encrypt the backups)

To connect to the Taxology Backblaze B2 bucket, you should create the following environment variables:

- B2_TAXOLOGY_ACCOUNT_ID
- B2_TAXOLOGY_ACCOUNT_KEY
- B2_TAXOLOGY_ENCRYPTION_KEY

Alternatively you can pass in each value separately, see `vmt restore --help` for more information.
