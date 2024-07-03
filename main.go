package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	
	"strings"
	"time"
)

var headers = http.Header{
	"Content-Type": {"application/json+transit"},
}

var client = &http.Client{}

func main() {
	// Gettings the Penpot Access Token
	if len(os.Getenv("PENPOT_TOKEN")) == 0 {
		fmt.Println("ENV not set.")
		return
	}
	headers.Add("Authorization", "Token "+os.Getenv("PENPOT_TOKEN"))

	teams, err := getTeams()
	if err != nil {
		fmt.Println(err)
		return
	}

	for idx, team := range teams {
		projects, err := getProjects(team.id)
		if err != nil {
			fmt.Printf("Error Occured while fetching Team %s's Projects\n", team.name)
			fmt.Println(err)
			continue
		}

		for iidx, project := range projects {
			files, err := getProjectFiles(project.id)
			if err != nil {
				fmt.Printf("Error Occured while fetching Project %s's Files\n", project.name)
				fmt.Println(err)
				continue
			}

			projects[iidx].files = files
		}

		teams[idx].projects = projects
	}

	dataToString(teams)

	fmt.Println(createBackup(teams))
}


// {"~:file-id":"~u<File_ID>","~:include-libraries":true,"~:embed-assets":false}
func downloadPenpotFile(fileId string) ([]byte, error) {
	var data []byte

	reqBody, err := json.Marshal(map[string]interface {}{
		"~:embed-assets": false,
		"~:file-id": "~u"+fileId,
		"~:include-libraries": true,
	})
	if(err != nil){
		return data, err
	}

	req, err := http.NewRequest("POST", "https://design.penpot.app/api/rpc/command/export-binfile", bytes.NewBuffer(reqBody))
	if(err != nil){
		return data, err
	}

	req.Header = headers
	res, err := client.Do(req)
	if(err != nil){
		return data, err
	}

	bytes, err := io.ReadAll(res.Body)
	if(err != nil){
		return data, err
	}

	if(res.StatusCode != 200){
		return data, fmt.Errorf("%d: %s", res.StatusCode, string(bytes))
	}

	return bytes, nil
}


func getTeams() ([]Team, error) {
	var teams []Team

	req, err := http.NewRequest("GET", "https://design.penpot.app/api/rpc/command/get-teams", bytes.NewBuffer([]byte("")))
	if err != nil {
		return teams, err
	}
	req.Header = headers
	res, err := client.Do(req)
	if err != nil {
		return teams, err
	}

	bytes, err := io.ReadAll(res.Body)
	if err != nil {
		return teams, err
	}

	data := string(bytes)

	if res.StatusCode != 200 {
		return teams, fmt.Errorf("%d : Failed to get data from server \n%s", res.StatusCode, data)
	}

	teams = parseTeamList(data)

	return teams, nil
}

func getProjects(projectId string) ([]Project, error) {
	var projects []Project
	req, err := http.NewRequest("GET", fmt.Sprintf("https://design.penpot.app/api/rpc/command/get-projects?team-id=%s", projectId), bytes.NewBuffer([]byte("")))
	if err != nil {
		return projects, err
	}

	req.Header = headers
	res, err := client.Do(req)
	if err != nil {
		return projects, err
	}

	bytes, err := io.ReadAll(res.Body)
	if err != nil {
		return projects, err
	}

	data := string(bytes)

	if res.StatusCode != 200 {
		return projects, fmt.Errorf("%d : Failed to get data from server \n%s", res.StatusCode, data)
	}

	projects = parseProjectList(data)

	return projects, nil
}

func getProjectFiles(projectId string) ([]ProjectFile, error) {
	var files []ProjectFile

	req, err := http.NewRequest("GET", fmt.Sprintf("https://design.penpot.app/api/rpc/command/get-project-files?project-id=%s", projectId), bytes.NewBuffer([]byte("")))
	if err != nil {
		return files, err
	}

	req.Header = headers
	res, err := client.Do(req)
	if err != nil {
		return files, err
	}

	bytes, err := io.ReadAll(res.Body)
	if err != nil {
		return files, err
	}

	data := string(bytes)

	if res.StatusCode != 200 {
		return files, fmt.Errorf("%d : Failed to get data from server \n%s", res.StatusCode, data)
	}

	files = parseFileList(data)

	return files, nil
}


func createBackup(teams []Team) error {
	backupsInfo, err := os.Stat("backups")
	if os.IsNotExist(err) {
		err := os.Mkdir("backups", 0755)
		if err != nil {
			return err
		}
		backupsInfo, err = os.Stat("backups")
		if err != nil {
			return fmt.Errorf("failed to get backup folder info")
		}
	} else if (err != nil) {
		return err
	}

	if !backupsInfo.IsDir() {
		return fmt.Errorf("another entry already exists with \"backups\" name")
	}

	backupFolder := fmt.Sprintf("backups/backup-%s", strings.Join(strings.Split(time.Now().Local().String(), " ")[0:2], "-"))
	if os.MkdirAll(backupFolder, 0755) != nil {
		return fmt.Errorf("unable to create a backup at %s", time.Now().Local().String())
	}

	for _, team := range teams {
		if os.MkdirAll(fmt.Sprintf("%s/%s", backupFolder, team.name), 0755) != nil {
			fmt.Printf("unable to create a folder for team %s at %s\n", team.name, time.Now().Local().String())
			continue
		}

		for _, project := range team.projects {
			if os.MkdirAll(fmt.Sprintf("%s/%s/%s", backupFolder, team.name, project.name), 0755) != nil {
				fmt.Printf("unable to create a folder for team %s's %s project at %s\n", team.name, project.name, time.Now().Local().String())
				continue
			}

			for _, projectFile := range project.files {
				filePath := fmt.Sprintf("%s/%s/%s/%s", backupFolder, team.name, project.name, projectFile.name)
				if os.MkdirAll(filePath, 0755) != nil {
					fmt.Printf("unable to create a folder for team %s's %s file in %s project at %s\n", team.name, projectFile.name, project.name, time.Now().Local().String())
					continue
				}	

				penpotFile, err := os.Create(fmt.Sprintf("%s/%s.penpot",filePath, projectFile.name))
				if(err != nil){
					fmt.Printf("unable to create penpot file for team\n%s\n", err)
				}else {
					bytes, err := downloadPenpotFile(projectFile.id)
					if(err != nil){
						fmt.Printf("unable to download penpot file \n%s \n", err)
					}else {
						penpotFile.Write(bytes)
						penpotFile.Close()
					}
				}
			}

		}
	}
	return nil
}

func dataToString(teams []Team) {
	file, err := os.Create("data.json")
	if err != nil {
		fmt.Println(err)
		return
	}

	jsonData := "{"
	for idx, team := range teams {
		jsonData += fmt.Sprintf("\"%d\" : {\"Team Id\": \"%s\", \"Team Name\": \"%s\"", idx, team.id, team.name)

		if len(team.projects) != 0 {
			jsonData += ",\"projects\" : {"

			for indx, project := range team.projects {
				jsonData += fmt.Sprintf("\"%d\" : {\"Id\" : \"%s\", \"Name\" : \"%s\", \"lastModified\" : \"%s\", ", indx, project.id, project.name, project.lastModified.Local().String())

				if len(project.files) != 0 {
					jsonData += "\"files\" : {"

					for iindx, file := range project.files {
						jsonData += fmt.Sprintf("\"%d\" : { \"Id\" : \"%s\", \"Name\" : \"%s\", \"lastModified\" : \"%s\"},", iindx, file.id, file.name, file.lastModified.Local().String())

					}

					jsonData = strings.TrimSuffix(jsonData, ",")
					jsonData += "}"
				}
				jsonData += "},"
			}
			jsonData = strings.TrimSuffix(jsonData, ",")
			jsonData += "},"
		}
		jsonData = strings.TrimSuffix(jsonData, ",")

		jsonData += "},"
	}
	jsonData = strings.TrimSuffix(jsonData, ",")
	jsonData += "}"

	file.Write([]byte(jsonData))
}
