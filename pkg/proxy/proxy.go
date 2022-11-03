package proxy

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/AlecAivazis/survey"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	"github.com/voormedia/voormedia-toolkit/pkg/util"
)

// Run Google Cloud SQL proxy container
func Run(log *util.Logger, port string) error {
	sqlInstances, err := util.FindSQLInstances()
	if err != nil {
		return err
	}

	q := []*survey.Question{
		{
			Name: "instance",
			Prompt: &survey.Select{
				Message: "Choose a SQL instance:",
				Options: sqlInstances,
			},
		},
	}

	selection := struct{ Instance string }{}

	err = survey.Ask(q, &selection)
	if err != nil {
		return err
	}

	connectionName, err := util.FindConnectionName(selection.Instance)
	if err != nil {
		return err
	}

	proxyFile, err := findProxyFile()
	if err != nil {
		return err
	}

	args := []string{
		"-instances", connectionName + "=tcp:" + port,
	}

	cmd := exec.Command(proxyFile, args...)
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Run()
	return nil
}

func findProxyFile() (string, error) {
	proxyFile, err := homedir.Expand("~/cloud_sql_proxy")
	if err != nil {
		return "", err
	}

	if _, err := os.Stat(proxyFile); os.IsNotExist(err) {
		fmt.Printf("Proxy file not found in the home directory. Downloading it now.\n")
		cmd := exec.Command("")
		if runtime.GOOS == "linux" {
			cmd = exec.Command("wget", "https://dl.google.com/cloudsql/cloud_sql_proxy.linux.amd64", "-O", proxyFile)
		} else if runtime.GOOS == "darwin" {
			cmd = exec.Command("curl", "-o", proxyFile, "https://dl.google.com/cloudsql/cloud_sql_proxy.darwin.amd64")
		}

		var out, errOut bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &errOut
		if err := cmd.Run(); err != nil {
			return "", errors.Errorf("Failed to download proxy file: %s", errOut.String())
		}

		fmt.Printf("Downloaded the proxy file. Making it executable now.\n")
		cmd = exec.Command("chmod", "+x", proxyFile)
		if err := cmd.Run(); err != nil {
			return "", errors.Errorf("Failed to make the proxy file executable: %s", errOut.String())
		}
	}

	fmt.Printf("Starting Google Cloud SQL proxy...\n")
	return proxyFile, nil
}
