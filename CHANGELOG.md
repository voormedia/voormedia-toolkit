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
