package util

import (
	"io/ioutil"

	"github.com/oleiade/reflections"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

// TargetConfig for Google Cloud database
type TargetConfig struct {
	Hostname    string
	Port        string
	Database    string
	Username    string
	Password    string
	Environment string
}

// DatabaseConfig for Google Cloud
type DatabaseConfig struct {
	Development TargetConfig
	Acceptance  TargetConfig
	Production  TargetConfig
}

type ShardedTargetConfig struct {
	Receptel TargetConfig
	Bouwens  TargetConfig
}

type ShardedDatabaseConfig struct {
	Development struct {
		Receptel TargetConfig
		Bouwens  TargetConfig
	}
	Acceptance struct {
		Receptel TargetConfig
		Bouwens  TargetConfig
	}
	Production struct {
		Receptel TargetConfig
		Bouwens  TargetConfig
	}
}

// GetDatabaseConfig based on provided arguments
func GetDatabaseConfig(database string, environment string, shard string, user string, password string, host string, port string, configFile string) (TargetConfig, error) {
	target := TargetConfig{}
	shardedTarget := ShardedTargetConfig{}
	if database == "" {
		yamlFile, err := ioutil.ReadFile(configFile)
		if err != nil {
			return target, err
		}

		if shard != "" {
			dbConfig := ShardedDatabaseConfig{}
			err = yaml.Unmarshal(yamlFile, &dbConfig)
			if err != nil {
				return target, err
			}

			if environment == "development" {
				shardedTarget = dbConfig.Development
			} else if environment == "acceptance" {
				shardedTarget = dbConfig.Acceptance
			} else if environment == "production" {
				shardedTarget = dbConfig.Production
			} else {
				return target, errors.Errorf("Invalid target specified: " + environment)
			}

			shardConfig, err := reflections.GetField(shardedTarget, "Receptel")
			if err != nil && shardConfig != nil {
				return target, errors.Errorf("Invalid shard specified: " + shard + "\n\n" + err.Error())
			} else {
				target = shardConfig.(TargetConfig)
			}
		} else {
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
