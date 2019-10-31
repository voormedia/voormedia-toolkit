package util

import (
	"bytes"
	"github.com/pkg/errors"
	"os/exec"
	"path/filepath"
	"strings"
)

// FindSQLInstances for Google Cloud
func FindSQLInstances() ([]string, error) {
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

// FindSQLDatabases for Google Cloud
func FindSQLDatabases(instance string) ([]string, error) {
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

// FindConnectionName for Google Cloud
func FindConnectionName(instance string) (string, error) {
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
