# v1.5.0

* **Backup downloads recover from interruptions and corruption automatically.** If a previous `vmt restore` was cancelled mid-download, the next run no longer fails with a confusing "decryption error" — it now detects the size mismatch against Backblaze, wipes the partial file, and re-downloads cleanly. Downloads also write to a `.partial` file and rename on success, so an interrupted transfer never leaves behind a file that looks complete.
* **Cloud SQL proxy binary download no longer needs `wget` or `curl`.** It's fetched through Go's standard HTTP client now, so the only requirement on the host is `vmt` itself. The download also shows a progress bar so you can see how long the first-run fetch will take.

# v1.4.3

* Download the Cloud SQL proxy binary matching the host architecture, so Apple Silicon and Linux arm64 machines no longer fetch the Intel build.

# v1.4.2

* Set the database port automatically based on the adapter type in the Rails database configuration file.

# v1.4.1

* Show a 💾 icon next to backups already downloaded to disk in the restore picker.

# v1.4.0

* Automatically pick the default Backblaze bucket based on the active GCP project.

# v1.3.0

* Add the option to specify a shard when interpreting Rails database configuration files.

# v1.2.0

* Change the default value of the `b2bucket` flag to `voormedia-eu-db-backups`.

# v1.1.0

* Fix decryption parameters for database backups.
