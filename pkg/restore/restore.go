package restore

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/AlecAivazis/survey"
	"github.com/kurin/blazer/b2"
	"github.com/pkg/errors"
	"github.com/voormedia/voormedia-toolkit/pkg/util"
)

// Run backup download (from Backblaze) and restore of a Google Cloud SQL database
func Run(log *util.Logger, targetEnvironment string, targetShard string, b2id string, b2key string, b2encrypt string, b2bucketName string,
	configFile string, targetPort string, targetHost string, targetUsername string, targetPassword string, targetDatabase string) error {

	sqlInstances, err := util.FindSQLInstances()
	if err != nil {
		return err
	}

	q := []*survey.Question{
		{
			Name: "instance",
			Prompt: &survey.Select{
				Message: "Choose a source instance:",
				Options: sqlInstances,
			},
		},
	}

	instanceSelection := struct{ Instance string }{}

	err = survey.Ask(q, &instanceSelection)
	if err != nil {
		return err
	}

	sqlDatabases, err := util.FindSQLDatabases(instanceSelection.Instance)
	if err != nil {
		return err
	}

	q = []*survey.Question{
		{
			Name: "database",
			Prompt: &survey.Select{
				Message: "Choose a source database:",
				Options: sqlDatabases,
			},
		},
	}

	databaseSelection := struct{ Database string }{}

	err = survey.Ask(q, &databaseSelection)
	if err != nil {
		return err
	}

	b2Context, b2Bucket, b2encrypt, err := util.B2Bucket(b2id, b2key, b2bucketName, b2encrypt, false)
	if err != nil {
		return err
	}

	sqlBackups, sqlDownloads, err := findSQLBackups(b2Context, databaseSelection.Database, b2Bucket)
	if err != nil {
		return err
	}

	var backupOptions []string
	for i, backup := range sqlBackups {
		displayText := backup
		if sqlDownloads[i] {
			displayText = "💾 " + backup
		}
		backupOptions = append(backupOptions, displayText)
	}

	q = []*survey.Question{
		{
			Name: "backup",
			Prompt: &survey.Select{
				Message: "Choose a backup to restore:",
				Options: backupOptions,
			},
		},
	}

	backupSelection := struct{ Backup string }{}

	err = survey.Ask(q, &backupSelection)
	if err != nil {
		return err
	}

	selectedBackup := strings.TrimPrefix(backupSelection.Backup, "💾 ")

	target, err := util.GetDatabaseConfig(targetDatabase, targetEnvironment, targetShard, targetUsername, targetPassword, targetHost, targetPort, configFile)
	if err != nil {
		return err
	}

	splitFileName := strings.Split(selectedBackup, "/")
	encryptedPath := "/tmp/" + splitFileName[len(splitFileName)-1]
	decryptedPath := strings.Replace(encryptedPath, ".encrypted", "", 1)

	header := fmt.Sprintf("Downloading Backblaze backup to restore it on the %s environment", target.Environment)
	_, downloaded, err := util.B2Download(b2Context, b2Bucket, selectedBackup, header)
	if err != nil {
		return err
	}

	if downloaded {
		os.Remove(decryptedPath)
	}
	if _, err := os.Stat(decryptedPath); err != nil {
		if err := util.DecryptBackup(encryptedPath, decryptedPath, b2encrypt); err != nil {
			return err
		}
	}

	if strings.Contains(instanceSelection.Instance, "mysql") {
		if err := restoreBackupToMySQL(target, decryptedPath); err != nil {
			return err
		}
	} else {
		if err := restoreBackupToPostgres(target, decryptedPath); err != nil {
			return err
		}
	}

	return nil
}

func findSQLBackups(ctx context.Context, database string, bucket *b2.Bucket) ([]string, []bool, error) {
	var results []string
	var downloaded []bool

	backups := bucket.List(ctx, b2.ListPrefix(database))
	for backups.Next() {
		backupName := backups.Object().Name()
		results = append(results, backupName)

		// Determine if the file is already downloaded
		splitFileName := strings.Split(backupName, "/")
		file := strings.Replace("/tmp/"+splitFileName[len(splitFileName)-1], ".encrypted", "", 1)

		_, err := os.Stat(file)
		downloaded = append(downloaded, err == nil)
	}

	if len(results) == 0 {
		return nil, nil, errors.Errorf("Could not find any backups for the selected database")
	}

	// Show the most recent backups at the top of the selection list
	var reversedResults []string
	var reversedDownloaded []bool
	for i := len(results) - 1; i >= 0; i-- {
		reversedResults = append(reversedResults, results[i])
		reversedDownloaded = append(reversedDownloaded, downloaded[i])
	}

	return reversedResults, reversedDownloaded, nil
}

func restoreBackupToMySQL(target util.TargetConfig, backup string) error {
	fmt.Printf("Restoring to MySQL database %s (%s:%s)...\n", target.Database, target.Hostname, target.Port)

	// Attempt to create the database in case it doesn't exist
	cmd := exec.Command("mysqladmin", "-u", target.Username, "-h", target.Hostname, "create", target.Database, "&>", "/dev/null")
	cmd.Run()

	cmd = exec.Command("mysql", "-u", target.Username, "-h", target.Hostname, "--password="+target.Password, "-P", target.Port, target.Database, "-e", "source "+backup)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		if target.Environment != "development" {
			return errors.Errorf("Couldn't connect to the target database. Please check that the proxy is running on port %s\n\n%s", target.Port, stderr.String())
		}
		return errors.Errorf("Couldn't connect to the target database. Please check that your database server running on port %s\n\n%s", target.Port, stderr.String())
	}

	return nil
}

func restoreBackupToPostgres(target util.TargetConfig, backup string) error {
	fmt.Printf("Restoring to Postgres database %s (%s:%s)...\n", target.Database, target.Hostname, target.Port)

	if target.Environment != "development" {
		cmd := exec.Command("psql", "-d", target.Database, "-h", target.Hostname, "-p", target.Port, "-U", target.Username, "-f", backup)
		cmd.Env = os.Environ()
		cmd.Env = append(cmd.Env, "PGPASSWORD="+target.Password)
		err := cmd.Run()
		if err != nil {
			return errors.Errorf("Couldn't connect to the target database. Please check that the proxy is running on port %s\n\n%s", target.Port, err.Error())
		}
	} else {
		// Attempt to create the database in case it doesn't exist
		createCmd := exec.Command("createdb", target.Database, "-h", target.Hostname, "-p", target.Port)
		createCmd.Run()

		cmd := exec.Command("psql", "-d", target.Database, "-h", target.Hostname, "-p", target.Port, "-f", backup)
		err := cmd.Run()
		if err != nil {
			return errors.Errorf("Couldn't connect to the target database. Please check that your database server running on port %s\n\n%s", target.Port, err.Error())
		}
	}
	return nil
}
