package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"
)

type Collection struct {
	Id        string
	Name      string
	Owner     string
	CreatedAt string
	UpdatedAt string
	Uid       string
	IsPublic  bool
}

type Collections struct {
	Collections []Collection
}

type Config struct {
	APIKey      string `yaml:"apikey"`
	WorkspaceID string `yaml:"workspaceid"`
}

func readConfig(filename string) (*Config, error) {
	// Read the YAML file
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	// Unmarshal the YAML data into the Config struct
	config := &Config{}
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, err
	}

	return config, nil
}

func main() {
	apikey := "your api key"
	workspaceid := "your workspace id"
	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	files, err := listFilesWithFullPath(currentDir)
	if err != nil {
		fmt.Println("Error in reading files from current dir", currentDir, err)
		return
	}

	if len(files) == 0 {
		fmt.Println("There is nothing to process, goodbye!")
		return
	}

	for _, file := range files {
		// Read the JSON file
		jsonData, err := os.ReadFile(file)
		if err != nil {
			fmt.Println("Error reading the JSON file:", err)
			return
		}

		var data map[string]interface{}

		err = json.Unmarshal(jsonData, &data)
		if err != nil {
			fmt.Println("Error unmarshaling JSON:", err)
			return
		}

		collectionMap := make(map[string]interface{})

		val := data["info"]
		var newInfoMap map[string]interface{}
		var name interface{}
		if infoMap, ok := val.(map[string]interface{}); ok {
			name = infoMap["name"]
			schema := infoMap["schema"]

			newInfoMap = make(map[string]interface{})
			newInfoMap["name"] = name
			newInfoMap["schema"] = schema
		}

		collectionMap["info"] = newInfoMap
		collectionMap["item"] = data["item"]

		finalReqBody := make(map[string]interface{})
		finalReqBody["collection"] = collectionMap

		jsonReqBody, err := json.Marshal(finalReqBody)
		if err != nil {
			fmt.Println("Error marshaling JSON:", err)
			return
		}

		cols, err := getCurrentCollections(apikey, workspaceid)
		if err != nil {
			fmt.Println("Could not get the current collections for the workspace")
			return
		}

		fmt.Println("Going to check if the name is present in the current collections", name)
		if str, ok := name.(string); ok {
			value, exists := cols[str]
			if exists {
				fmt.Println(str, "is a existing collection")
				fmt.Println("Proceeding to update the existing collection ...")
				colId := value.Id
				api := fmt.Sprintf("https://api.getpostman.com/collections/%s", colId)
				resp, err := makeApiReq(apikey, http.MethodPut, api, jsonReqBody)
				if err != nil {
					fmt.Println("Could not update the collection", str, err)
					continue
				}
				if resp.StatusCode == http.StatusOK {
					fmt.Println("Collection Updated successfully.")
				} else {
					fmt.Println("Error:", resp.Status)
				}
			} else {
				fmt.Println(str, "is not a existing collection")
				fmt.Println("Proceeding to create the collection ...")
				resp, err := makeApiReq(apikey, http.MethodPost, "https://api.getpostman.com/collections", jsonReqBody)
				if err != nil {
					fmt.Println("Could not update the collection", str, err)
					continue
				}
				if resp.StatusCode == http.StatusOK {
					fmt.Println("Collection Updated successfully.")
				} else {
					fmt.Println("Error:", resp.Status)
				}
			}
		} else {
			fmt.Println("The value is not a string")
		}

	}

	fmt.Println(files)
}

func listFilesWithFullPath(directoryPath string) ([]string, error) {
	files := make([]string, 0)

	fileInfos, err := os.ReadDir(directoryPath)
	if err != nil {
		return nil, err
	}

	for _, fileInfo := range fileInfos {
		if fileInfo.IsDir() {
			continue // Skip directories
		}
		if strings.HasSuffix(fileInfo.Name(), "postman_collection.json") {
			files = append(files, filepath.Join(directoryPath, fileInfo.Name()))
		}
	}

	return files, nil
}

func getCurrentCollections(apikey string, workspaceid string) (map[string]Collection, error) {
	api := fmt.Sprintf("https://api.getpostman.com/collections?workspace=%s", workspaceid)
	resp, err := makeApiReq(apikey, http.MethodGet, api, nil)
	if err != nil {
		fmt.Println("Unable to get the current collections")
		return nil, err
	}
	defer resp.Body.Close()

	collections := new(Collections)
	json.NewDecoder(resp.Body).Decode(collections)
	fmt.Println(collections)

	currColsMap := make(map[string]Collection)

	for _, c := range collections.Collections {
		currColsMap[c.Name] = c
	}
	return currColsMap, nil
}

func makeApiReq(apikey string, method string, url string, body []byte) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}
	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Api-Key", apikey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	return resp, nil
}
