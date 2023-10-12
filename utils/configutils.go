package utils

import (
	"bufio"
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	APIKey      string `yaml:"apikey"`
	WorkspaceID string `yaml:"workspaceid"`
}

func GetConfigValues(configFilePath string) (*Config, error) {
	_, err := os.Stat(configFilePath)
	if err == nil {
		fmt.Println(".letter.config already exists. Reading the config values from", configFilePath)
		config, err := readConfig(configFilePath)
		if err != nil {
			fmt.Println("Could not read the config values. Make sure the file is correct. Or delete the file and regenerate")
			return nil, err
		}
		return config, nil
	} else if os.IsNotExist(err) {
		fmt.Println(".letter.config does not exist here", configFilePath)
		config, err := getConfigFromUser()
		if err != nil {
			fmt.Println("Error creating .letter.config:", err)
			return nil, err
		} else {
			fmt.Println("Will create a config with the key", configFilePath)
			return config, nil
		}
	} else {
		fmt.Println("Error getting data from .letter.config:", err)
		return nil, err
	}

}

func CreateConfigFile(config *Config, filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := yaml.Marshal(&config)
	if err != nil {
		return err
	}

	_, err = file.Write(data)
	if err != nil {
		return err
	}
	return nil
}

func getConfigFromUser() (*Config, error) {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Print("Enter your API key: ")
	scanner.Scan()
	apikey := scanner.Text()

	if scanner.Err() != nil {
		return nil, scanner.Err()
	}

	config := new(Config)
	config.APIKey = apikey

	return config, nil
}

func readConfig(filename string) (*Config, error) {
	// Read the YAML file
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	config := &Config{}
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, err
	}

	return config, nil
}
