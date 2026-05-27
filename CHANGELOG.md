# v2.1.0

* **`vmt restore` now reads Rails `database.yml` files that use ERB.** Database names written as `<%= ENV.fetch("PGDATABASE") { "myapp-dev" } %>` previously caused a confusing "Couldn't connect to the target database" failure, because vmt fed the raw template text to Postgres as the database name. vmt now renders the env-reading expressions these files use — `<%= ENV["X"] %>`, `<%= ENV.fetch("X") { "default" } %>`, `<%= ENV.fetch("X", "default") %>`, and the `<%= ENV["X"] || "default" %>` fallback form — honouring environment variables and their defaults exactly as Rails does, before parsing the YAML. Any output tag it can't evaluate (e.g. `<%= Rails.application.credentials… %>`) is left untouched and reported with a warning, so an unsupported construct surfaces clearly instead of as a cryptic connection error. Config files without ERB (non-Ruby projects) are unaffected, and there's no new runtime dependency. Pass `--database` to override the file as before.

# v2.0.1

* **CI release pipeline now produces downloadable binaries again.** The previous releases (v1.4.3, v1.5.0, v2.0.0) had their tags pushed but the release workflow failed before it could attach any artifacts, so their GitHub release pages are empty. This release is functionally identical to v2.0.0 — install it to get the v2.0.0 features with working install scripts.

# v2.0.0

* **Coordinated visual refresh across every interactive command.** Backups, restores, the proxy, and the shell now share a consistent look: the active GCP project is shown as a banner at the top, prompts and selections use a blue accent instead of survey's default cyan, and progress bars and spinners get a styled, full-terminal-width layout.
* **The restore command makes acceptance and production targets impossible to miss.** Those environment names render in bold red on the download header and on the cached-backup banner, so there's no quiet "development" default that slips past before you press Enter.
* **The backup picker is human-readable.** The list now shows `2026-05-19  04:00:09  💾` instead of the raw encrypted filename; the 💾 on the right means a cached copy is on disk and will be reused without re-downloading.
* **The Cloud SQL proxy tells you when it's actually ready.** Instead of leaving you to scan `cloud_sql_proxy`'s startup logs for "Ready for new connections," `vmt proxy` now prints `✓ Cloud SQL proxy listening on localhost:3307 (Ctrl+C to stop)` once the proxy is accepting connections. Connection logs and errors still pass through.

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
