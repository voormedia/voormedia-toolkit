package backup

import (
	"fmt"
	"github.com/AlecAivazis/survey"
	"github.com/emielvanlankveld/voormedia-toolkit/pkg/util"
	"github.com/pkg/errors"
	"os"
	"os/exec"
	"strings"
	"time"
)

// Run backup creation and stores it in Backblaze B2.
func Run(log *util.Logger, port string, host string, b2id string, b2key string, b2encrypt string, b2bucketName string, configFile string) error {
	sqlInstances, err := util.FindSQLInstances()
	if err != nil {
		return err
	}

	q := []*survey.Question{
		{
			Name: "instance",
			Prompt: &survey.Select{
				Message: "Choose an instance:",
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

	q = []*survey.Question{
		{
			Name:     "user",
			Prompt:   &survey.Input{Message: "Username:", Default: "root"},
			Validate: survey.Required,
		},
	}

	response := struct{ User string }{}

	err = survey.Ask(q, &response)
	if err != nil {
		return err
	}

	target, err := util.GetDatabaseConfig(databaseSelection.Database, "custom", response.User, "", host, port, configFile)
	if err != nil {
		return err
	}

	file := ""
	if strings.Contains(instanceSelection.Instance, "mysql") {
		file, err = backupFromMySQL(target)
		if err != nil {
			return err
		}
	} else {
		file, err = backupFromPostgres(target)
		if err != nil {
			return err
		}
	}

	cmd := exec.Command("openssl", "aes-256-cbc", "-in", "/tmp/"+file, "-out", "/tmp/"+file+".encrypted", "-pass", "pass:"+b2encrypt)
	cmd.Run()

	err = util.B2Upload(b2Context, b2Bucket, databaseSelection.Database, file+".encrypted")
	if err != nil {
		return err
	}

	return nil
}

func backupFromMySQL(target util.TargetConfig) (string, error) {
	fmt.Printf("Backing up MySQL database " + target.Database + " (" + target.Hostname + ":" + target.Port + ")...\n" +
		"You may have to enter a password for user " + target.Username + "\n")

	fileName := target.Database + "_" + time.Now().Format("2006-01-02_15:04:05") + ".sql"

	cmd := exec.Command("mysqldump", "-u", target.Username, "--set-gtid-purged=OFF", "-h", target.Hostname, "-P", target.Port, "-p", target.Database)
	outfile, err := os.Create("/tmp/" + fileName)
	if err != nil {
		return "", err
	}
	defer outfile.Close()
	cmd.Stderr = os.Stderr
	cmd.Stdout = outfile

	err = cmd.Run()
	if err != nil {
		return "", errors.Errorf("Couldn't connect to the target database. Please check that the proxy is running on port " + target.Port + "\n")
	}

	return fileName, nil
}

func backupFromPostgres(target util.TargetConfig) (string, error) {
	fmt.Printf("Backing up Postgres database " + target.Database + " (" + target.Hostname + ":" + target.Port + ")...\n")

	cmd := exec.Command("pg_dump", "-U", target.Username, "-h", target.Hostname, "-p", target.Port, "--format=plain", "--no-owner", "--no-acl", "--clean", "-c", target.Database)
	cmd.Run()

	return "", nil
}
