package util

import (
	"io/ioutil"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

// TargetConfig for Google Cloud database
type TargetConfig struct {
	Hostname    string
	Port        string
	Adapter     string
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

type ShardedTargetConfig = map[string]TargetConfig

type ShardedDatabaseConfig struct {
	Development ShardedTargetConfig
	Acceptance  ShardedTargetConfig
	Production  ShardedTargetConfig
}

// GetDatabaseConfig based on provided arguments
func GetDatabaseConfig(database string, environment string, shard string, user string, password string, host string, port string, configFile string) (TargetConfig, error) {
	target := TargetConfig{}
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

			shardedTarget := ShardedTargetConfig{}
			if environment == "development" {
				shardedTarget = dbConfig.Development
			} else if environment == "acceptance" {
				shardedTarget = dbConfig.Acceptance
			} else if environment == "production" {
				shardedTarget = dbConfig.Production
			} else {
				return target, errors.Errorf("Invalid target specified: " + environment)
			}

			shardConfig, keyFound := shardedTarget[shard]
			if keyFound {
				target = shardConfig
			} else {
				return target, errors.Errorf("Invalid shard specified: " + shard)
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
	target.Environment = environment

	if port == "" {
		switch target.Adapter {
		case "mysql2":
			{
				target.Port = "3306"
			}
		case "postgresql":
			{
				target.Port = "5432"
			}
		case "sqlserver":
			{
				target.Port = "1433"
			}
		case "oracle_enhanced":
			{
				target.Port = "1521"
			}
		default:
			{
				target.Port = "3306"
			}
		}
	} else {
		target.Port = port
	}

	if target.Database == "" {
		return target, errors.Errorf("Could not find a database belonging to the target")
	}

	return target, nil
}
