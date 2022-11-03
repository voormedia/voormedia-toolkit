package shell

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/AlecAivazis/survey"
	"github.com/voormedia/voormedia-toolkit/pkg/util"
)

// Run a shell of a Google Cloud SQL database of choice.
func Run(log *util.Logger) error {
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
