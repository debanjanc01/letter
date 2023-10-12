package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/debanjanc01/letter/utils"
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

type Workspace struct {
	Id         string
	Name       string
	Type       string
	Visibility string
}

type Workspaces struct {
	Workspaces []Workspace
}

type Letdown struct {
	Message string
}

func (e Letdown) Error() string {
	return e.Message
}

func main() {
	path, err := utils.GetCurrRunningPath()
	if err != nil {
		fmt.Println("Unable to get the current running path")
		return
	}
	configFilePath := filepath.Join(path, ".letter.config")
	config, err := utils.GetConfigValues(configFilePath)
	if err != nil {
		fmt.Println("Could not get the config values. Goodbye")
		return
	}
	apikey := config.APIKey
	workspaceid := config.WorkspaceID

	if workspaceid == "" {
		workspaceid, err = getWorkspaceId(apikey)
		if err != nil {
			fmt.Println("Could not get the workspace.")
			return
		}
		config.WorkspaceID = workspaceid
		utils.CreateConfigFile(config, configFilePath)
	}

	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	files, err := utils.ListFilesWithFullPath(currentDir)
	if err != nil {
		fmt.Println("Error in reading files from current dir", currentDir, err)
		return
	}

	if len(files) == 0 {
		fmt.Println("There is nothing to process, goodbye!")
		return
	}

	for _, file := range files {
		fmt.Println("Proceeding to read file", file)
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

		fmt.Printf("Going to check if %s is an existing collection\n", name)
		if str, ok := name.(string); ok {
			value, exists := cols[str]
			if exists {
				fmt.Println(str, "is a existing collection")
				fmt.Println("Proceeding to update the existing collection...")
				colId := value.Id
				api := fmt.Sprintf("https://api.getpostman.com/collections/%s", colId)
				resp, err := makeApiReq(apikey, http.MethodPut, api, jsonReqBody)
				if err != nil {
					fmt.Println("Could not update the collection", str, err)
					continue
				}
				if resp.StatusCode == http.StatusOK {
					fmt.Println("Collection Updated successfully!")
				} else {
					fmt.Println("Error:", resp.Status)
				}
			} else {
				fmt.Println(str, "is not a existing collection")
				fmt.Println("Proceeding to create the collection...")
				resp, err := makeApiReq(apikey, http.MethodPost, fmt.Sprintf("https://api.getpostman.com/collections?workspace=%s", workspaceid), jsonReqBody)
				if err != nil {
					fmt.Println("Could not update the collection", str, err)
					continue
				}
				if resp.StatusCode == http.StatusOK {
					fmt.Println("Collection Created successfully.")
				} else {
					fmt.Println("Error:", resp.Status)
				}
			}
		} else {
			fmt.Println("The value is not a string")
		}
	}
}

func getWorkspaceId(apikey string) (string, error) {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Print("Do you want to create a new Workspace? Enter y/Y/Yes\n")
	scanner.Scan()
	input := scanner.Text()

	if scanner.Err() != nil {
		return "", scanner.Err()
	}
	if utils.IsAffirmative(input) {
		workspaceid, err := createNewWorkspace(apikey)
		if err != nil {
			return "", nil
		}
		if workspaceid != "" {
			fmt.Println("New workspace created!")
			return workspaceid, nil
		}
	}
	workspaces, err := getExistingWorkspaces(apikey)
	if err != nil {
		return "", err
	}

	fmt.Println("Select a workspace from your existing workspaces")
	for key, value := range workspaces {
		fmt.Printf("Enter %s for %s\n", string(key), value.Name)
	}

	for {
		scanner.Scan()
		input = scanner.Text()

		if val, ok := workspaces[input]; ok {
			fmt.Printf("%s will be used\n", val.Name)
			return val.Id, nil
		} else {
			fmt.Printf("Invalid input. Do you want to retry? Enter y/Y/yes")
			scanner.Scan()
			input = scanner.Text()
			if !utils.IsAffirmative(input) {
				break
			}
		}
	}
	return "", Letdown{Message: "Something went wrong"}
}

func createNewWorkspace(apikey string) (string, error) {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Print("Provide a name for your new workspace:\n")
	scanner.Scan()
	name := scanner.Text()

	if scanner.Err() != nil {
		return "", scanner.Err()
	}

	reqJsonString := `
	{
		"workspace": {
			"name": "%s",
			"type": "personal",
			"visibility": "personal"
		}
	}`

	requestBody := []byte(fmt.Sprintf(reqJsonString, name))

	resp, err := makeApiReq(apikey, http.MethodPost, "https://api.getpostman.com/workspaces", requestBody)
	if err != nil {
		return "", err
	}

	if resp.StatusCode == http.StatusOK {
		defer resp.Body.Close()
		workspace := new(CreatedWorkspace)
		json.NewDecoder(resp.Body).Decode(workspace)
		return workspace.Workspace.Id, nil
	}
	return "", Letdown{}
}

type CreatedWorkspace struct {
	Workspace Workspace
}

func getExistingWorkspaces(apikey string) (map[string]Workspace, error) {
	api := "https://api.getpostman.com/workspaces"
	resp, err := makeApiReq(apikey, http.MethodGet, api, nil)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	workspaces := new(Workspaces)
	json.NewDecoder(resp.Body).Decode(workspaces)

	currWorkspaces := make(map[string]Workspace)

	for index, workspace := range workspaces.Workspaces {
		currWorkspaces[fmt.Sprint(index)] = workspace
	}
	return currWorkspaces, nil
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
