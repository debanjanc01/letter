package utils

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

type Config struct {
	APIKey      string `yaml:"apikey"`
	WorkspaceID string `yaml:"workspaceid"`
}

func GetConfigValues(binaryDir string) (*Config, error) {
	configFilePath := filepath.Join(binaryDir, ".letter.config")

	// Check if the file exists
	_, err := os.Stat(configFilePath)
	if err == nil {
		fmt.Println(".letter.config already exists. Reading the config values from", binaryDir)
		config, err := readConfig(configFilePath)
		if err != nil {
			fmt.Println("Could not read the config values. Make sure the file is correct. Or delete the file and regenerate")
			return nil, err
		}
		return config, nil
	} else if os.IsNotExist(err) {
		fmt.Println(".letter.config does not exist in", binaryDir)
		config, err := createConfigFile(configFilePath)
		if err != nil {
			fmt.Println("Error creating .letter.config:", err)
			return nil, err
		} else {
			fmt.Println(".letter.config created in", binaryDir)
			return config, nil
		}
	} else {
		fmt.Println("Error getting data from .letter.config:", err)
		return nil, err
	}

}

func createConfigFile(filePath string) (*Config, error) {
	config, err := getConfigFromUser()
	if err != nil {
		return nil, err
	}
	file, err := os.Create(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data, err := yaml.Marshal(&config)
	if err != nil {
		return nil, err
	}

	_, err = file.Write(data)
	if err != nil {
		return nil, err
	}
	return config, nil
}

func getConfigFromUser() (*Config, error) {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Print("Enter your API key: ")
	scanner.Scan()
	apikey := scanner.Text()

	if scanner.Err() != nil {
		return nil, scanner.Err()
	}

	fmt.Print("Enter your workspace ID: ")
	scanner.Scan()
	workspaceid := scanner.Text()

	if scanner.Err() != nil {
		return nil, scanner.Err()
	}

	config := new(Config)
	config.APIKey = apikey
	config.WorkspaceID = workspaceid

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
