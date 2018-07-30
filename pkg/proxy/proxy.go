package proxy

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/AlecAivazis/survey"
	"github.com/emielvanlankveld/gsql/pkg/util"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
)

// Run Google Cloud SQL proxy container
func Run(log *util.Logger) error {
	sqlInstances, err := findSQLInstances()
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

	connectionName, err := findConnectionName(selection.Instance)
	if err != nil {
		return err
	}

	credentialFile, err := findCredentialFile()
	if err != nil {
		return err
	}

	proxyFile, err := findProxyFile()
	if err != nil {
		return err
	}

	args := []string{
		"-instances", connectionName + "=tcp:3307",
		"-credential_file", credentialFile,
	}

	cmd := exec.Command(proxyFile, args...)
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

func findConnectionName(instance string) (string, error) {
	cmd := exec.Command("gcloud", "sql", "instances", "describe", instance, "--flatten", "connectionName")
	var out, errOut bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errOut
	if err := cmd.Run(); err != nil {
		return "", errors.Errorf("Failed to get SQL connection name: %s", errOut.String())
	}

	result := strings.Split(out.String(), "\n")
	return strings.TrimSpace(result[:len(result)-1][1]), nil
}

func findCredentialFile() (string, error) {
	// TODO: Implement more flexible solution for storing/retrieving credentials
	credentialFile, err := homedir.Expand("~/gcloud_proxy.json")
	if err != nil {
		return "", err
	}

	if _, err := os.Stat(credentialFile); os.IsNotExist(err) {
		return "", errors.Errorf("Credentials file `gcloud_proxy.json` does not exist in the home directory.")
	}
	return credentialFile, nil
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
