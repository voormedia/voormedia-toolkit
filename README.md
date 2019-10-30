## Voormedia Toolkit (Google Cloud Utilities)

# Restore
To automatically connect to the correct Backblaze B2 bucket you should create the following environment variables:

- B2_ACCOUNT_ID (your Backblaze B2 account ID)
- B2_ACCOUNT_KEY (your Backblaze B2 account key)
- B2_ENCRYPTION_KEY (the key used to encrypt the backups)
- B2_BACKUP_BUCKET (name of the bucket containing the backups)

Alternatively you can pass in each value separately, see `vmt restore --help` for more information.
