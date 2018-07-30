package shell

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/AlecAivazis/survey"
	"github.com/emielvanlankveld/gsql/pkg/util"
	"github.com/pkg/errors"
)

// Run a shell of a Google Cloud SQL database of choice.
func Run(log *util.Logger) error {
	sqlInstances, err := findSQLInstances()
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

	sqlDatabases, err := findSQLDatabases(instanceSelection.Instance)
	if err != nil {
		return err
	}

	q = []*survey.Question{
		{
			Name: "database",
			Prompt: &survey.Select{
				Message: "Choose a database:",
				Options: sqlDatabases,
			},
		},
		{
			Name:     "username",
			Prompt:   &survey.Input{Message: "Username:"},
			Validate: survey.Required,
		},
	}

	selection := struct {
		Database string
		Username string
	}{}

	err = survey.Ask(q, &selection)
	if err != nil {
		return err
	}

	cmd := exec.Command("sh", "-c", fmt.Sprintf("psql %s -U %s -h localhost -p 3307",
		selection.Database,
		selection.Username,
	))
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Run()
	return nil
}

func findSQLInstances() ([]string, error) {
	cmd := exec.Command("gcloud", "sql", "instances", "list", "--uri")
	var out, errOut bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errOut
	if err := cmd.Run(); err != nil {
		return nil, errors.Errorf("Failed to get SQL instances: %s", errOut.String())
	}

	instances := strings.Split(out.String(), "\n")
	for i, instance := range instances[:len(instances)-1] {
		instances[i] = filepath.Base(instance)
	}

	return instances, nil
}

func findSQLDatabases(instance string) ([]string, error) {
	cmd := exec.Command("gcloud", "sql", "databases", "list", "-i", instance, "--uri")
	var out, errOut bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errOut
	if err := cmd.Run(); err != nil {
		return nil, errors.Errorf("Failed to get SQL databases: %s", errOut.String())
	}

	databases := strings.Split(out.String(), "\n")
	for i, database := range databases[:len(databases)-1] {
		databases[i] = filepath.Base(database)
	}

	return databases, nil
}
