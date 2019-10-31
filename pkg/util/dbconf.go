package util

import (
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

// DatabaseConfig for Google Cloud
type DatabaseConfig struct {
	Development struct {
		Hostname    string
		Port        string
		Database    string
		Username    string
		Password    string
		Environment string
	}
	Acceptance struct {
		Hostname    string
		Port        string
		Database    string
		Username    string
		Password    string
		Environment string
	}
	Production struct {
		Hostname    string
		Port        string
		Database    string
		Username    string
		Password    string
		Environment string
	}
}

// TargetConfig for Google Cloud database
type TargetConfig struct {
	Hostname    string
	Port        string
	Database    string
	Username    string
	Password    string
	Environment string
}

// GetDatabaseConfig based on provided arguments
func GetDatabaseConfig(database string, environment string, user string, password string, host string, port string, configFile string) (TargetConfig, error) {
	target := TargetConfig{}
	if database == "" {
		yamlFile, err := ioutil.ReadFile(configFile)
		if err != nil {
			return target, err
		}

		dbConfig := DatabaseConfig{}
		err = yaml.Unmarshal(yamlFile, &dbConfig)
		if err != nil {
			return target, err
		}

		if environment == "development" {
			target = dbConfig.Development
		} else if environment == "acceptance" {
			target = dbConfig.Acceptance
		} else if environment == "production" {
			target = dbConfig.Production
		} else {
			return target, errors.Errorf("Invalid target specified: " + environment)
		}
	} else {
		target.Database = database
		target.Username = user
		target.Password = password
		environment = "custom"
	}

	target.Hostname = host
	target.Port = port
	target.Environment = environment

	if target.Database == "" {
		return target, errors.Errorf("Could not find a database belonging to the target")
	}

	return target, nil
}
