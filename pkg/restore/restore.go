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

	sqlBackups, err := findSQLBackups(b2Context, databaseSelection.Database, b2Bucket)
	if err != nil {
		return err
	}

	q = []*survey.Question{
		{
			Name: "backup",
			Prompt: &survey.Select{
				Message: "Choose a backup to restore:",
				Options: sqlBackups,
			},
		},
	}

	backupSelection := struct{ Backup string }{}

	err = survey.Ask(q, &backupSelection)
	if err != nil {
		return err
	}

	target, err := util.GetDatabaseConfig(targetDatabase, targetEnvironment, targetShard, targetUsername, targetPassword, targetHost, targetPort, configFile)
	if err != nil {
		return err
	}

	file := ""
	splitFileName := strings.Split(backupSelection.Backup, "/")
	if _, err := os.Stat("/tmp/" + splitFileName[len(splitFileName)-1]); err == nil {
		fmt.Printf("Selected Backblaze backup has already been downloaded. Using file on disk to restore on the " + target.Environment + " environment...\n")
	} else {
		fmt.Printf("Downloading Backblaze backup to restore it on the " + target.Environment + " environment...\n")
		file, err = downloadBackup(b2Context, backupSelection.Backup, b2Bucket, b2encrypt)
		if err != nil {
			return err
		}
	}

	file = strings.Replace("/tmp/"+splitFileName[len(splitFileName)-1], ".encrypted", "", 1)

	if strings.Contains(instanceSelection.Instance, "mysql") {
		err = restoreBackupToMySQL(target, file)
		if err != nil {
			return err
		}
	} else {
		err = restoreBackupToPostgres(target, file)
		if err != nil {
			return err
		}
	}

	return nil
}

func findSQLBackups(ctx context.Context, database string, bucket *b2.Bucket) ([]string, error) {
	var results []string

	backups := bucket.List(ctx, b2.ListPrefix(database))
	for backups.Next() {
		results = append(results, backups.Object().Name())
	}

	if len(results) == 0 {
		return nil, errors.Errorf("Could not find any backups for the selected database")
	}

	// Show the most recent backups at the top of the selection list
	var reversedResults []string
	for i := len(results) - 1; i >= 0; i-- {
		reversedResults = append(reversedResults, results[i])
	}

	return reversedResults, nil
}

func downloadBackup(ctx context.Context, file string, bucket *b2.Bucket, encryptionKey string) (string, error) {
	localFile, err := util.B2Object(ctx, bucket, file, encryptionKey)
	if err != nil {
		return "", err
	}

	return localFile, nil
}

func restoreBackupToMySQL(target util.TargetConfig, backup string) error {
	fmt.Printf("Restoring to MySQL database " + target.Database + " (" + target.Hostname + ":" + target.Port + ")...\n")

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
			return errors.Errorf("Couldn't connect to the target database. Please check that the proxy is running on port " + target.Port + "\n\n" + stderr.String())
		}
		return errors.Errorf("Couldn't connect to the target database. Please check that your database server running on port " + target.Port + "\n\n" + stderr.String())
	}

	return nil
}

func restoreBackupToPostgres(target util.TargetConfig, backup string) error {
	fmt.Printf("Restoring to Postgres database " + target.Database + " (" + target.Hostname + ":" + target.Port + ")...\n")

	if target.Environment != "development" {
		cmd := exec.Command("psql", "-d", target.Database, "-h", target.Hostname, "-p", target.Port, "-U", target.Username, "-f", backup)
		cmd.Env = os.Environ()
		cmd.Env = append(cmd.Env, "PGPASSWORD="+target.Password)
		err := cmd.Run()
		if err != nil {
			return errors.Errorf("Couldn't connect to the target database. Please check that the proxy is running on port " + target.Port + "\n\n" + err.Error())
		}
	} else {
		// Attempt to create the database in case it doesn't exist
		cmd := exec.Command("createdb", target.Database, "-h", target.Hostname, "-p", target.Port)
		cmd.Run()

		cmd = exec.Command("psql", "-d", target.Database, "-h", target.Hostname, "-p", target.Port, "-f", backup)
		err := cmd.Run()
		if err != nil {
			return errors.Errorf("Couldn't connect to the target database. Please check that your database server running on port " + target.Port + "\n\n" + err.Error())
		}
	}
	return nil
}
