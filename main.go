package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var headers = http.Header{
	"Content-Type": {"application/json"},
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
			if(err != nil){
				fmt.Printf("Error Occured while fetching Project %s's Files\n", project.name)
				fmt.Println(err)
				continue
			}

			projects[iidx].files = files
		}

		teams[idx].projects = projects
	}

	dataToString(teams)

	createBackup(teams, 5)
}

// https://design.penpot.app/api/rpc/command/export-binfile
// {"~:file-id":"~u<File_ID>","~:include-libraries":true,"~:embed-assets":false}
func downloadFile() {

}

type Team struct {
	id       string
	name     string
	projects []Project
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

func parseTeamList(data string) []Team {
	var teams []Team

	var id, name, parsedStr string
	isStringOpen := false
	bracketCount := 0
	for idx := 1; idx < len(data)-1; idx++ {

		// Looks for inner Arrays
		if data[idx] == '[' {
			bracketCount++
		} else if data[idx] == ']' {
			bracketCount--
			if bracketCount != 1 {
				continue
			}
		}

		if data[idx] == '"' {
			if isStringOpen {
				if strings.Contains(parsedStr, "~u") {
					id = strings.Replace(parsedStr, "~u", "", 1)
				} else if !strings.Contains(parsedStr, "~") && !strings.Contains(parsedStr, "^") {
					name = parsedStr
				}

				if len(id) > 0 && len(name) > 0 {
					teams = append(teams, Team{id, name, []Project{}})
					id = ""
					name = ""
				}
			}
			parsedStr = ""
			isStringOpen = !isStringOpen
			continue
		}

		if isStringOpen {
			parsedStr += string(data[idx])
		}
	}

	return teams
}

type Project struct {
	id           string
	name         string
	lastModified time.Time
	files        []ProjectFile
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

func parseProjectList(data string) []Project {
	var projects []Project

	var id, name, parsedStr string
	isStringOpen := false
	var bracketCount int
	var timeStamp1, timeStamp2 int64

	for idx := 1; idx < len(data)-1; idx++ {

		// Looks for inner Arrays
		if data[idx] == '[' {
			bracketCount++
		} else if data[idx] == ']' {
			bracketCount--
			if bracketCount <= 1 {
				continue
			}
		}

		if data[idx] == '"' {
			if isStringOpen {

				if strings.Contains(parsedStr, "~u") && len(id) == 0 {
					id = strings.Replace(parsedStr, "~u", "", 1)
				} else if strings.Contains(parsedStr, "~m") {
					timeStamp, err := strconv.ParseInt(strings.Replace(parsedStr, "~m", "", 1), 10, 64)
					if err != nil {
						parsedStr = ""
						continue
					}

					if timeStamp1 == 0 {
						timeStamp1 = timeStamp
					} else {
						timeStamp2 = timeStamp
					}
				} else if !strings.Contains(parsedStr, "~") && !strings.Contains(parsedStr, "^") {
					name = parsedStr
				}

				// fmt.Println(id, name, timeStamp1, timeStamp2)
				if len(id) > 0 && len(name) > 0 && timeStamp1 != 0 && timeStamp2 != 0 {
					if timeStamp1 < timeStamp2 {
						timeStamp1 = timeStamp2
					}
					projects = append(projects, Project{id, name, time.UnixMilli(timeStamp1), []ProjectFile{}})

					id = ""
					name = ""
					timeStamp1 = 0
					timeStamp2 = 0
				}
			}
			parsedStr = ""
			isStringOpen = !isStringOpen
			continue
		}

		if isStringOpen {
			parsedStr += string(data[idx])
		}
	}

	return projects
}

type ProjectFile struct {
	id           string
	name         string
	lastModified time.Time
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
	if(err != nil){
		return files, err
	}

	data := string(bytes)

	if res.StatusCode != 200 {
		return files, fmt.Errorf("%d : Failed to get data from server \n%s", res.StatusCode, data)
	}

	files = parseFileList(data)

	return files, nil
}

func parseFileList(data string) []ProjectFile {

	var files []ProjectFile

	var id, name, parsedStr string
	isStringOpen := false
	var bracketCount int
	var timeStamp1, timeStamp2 int64

	for idx := 1; idx < len(data)-1; idx++ {

		// Looks for inner Arrays
		if data[idx] == '[' {
			bracketCount++
		} else if data[idx] == ']' {
			bracketCount--
			if bracketCount <= 1 {
				continue
			}
		}

		if data[idx] == '"' {
			if isStringOpen {

				if strings.Contains(parsedStr, "~u") && len(id) == 0 {
					id = strings.Replace(parsedStr, "~u", "", 1)
				} else if strings.Contains(parsedStr, "~m") {
					timeStamp, err := strconv.ParseInt(strings.Replace(parsedStr, "~m", "", 1), 10, 64)
					if err != nil {
						parsedStr = ""
						continue
					}

					if timeStamp1 == 0 {
						timeStamp1 = timeStamp
					} else {
						timeStamp2 = timeStamp
					}
				} else if !(strings.Contains(parsedStr, "~") || strings.Contains(parsedStr, "^") || strings.Contains(parsedStr, "https://") ){
					name = parsedStr
				}

				// fmt.Println(id, name, timeStamp1, timeStamp2)
				if len(id) > 0 && len(name) > 0 && timeStamp1 != 0 && timeStamp2 != 0 {
					if timeStamp1 < timeStamp2 {
						timeStamp1 = timeStamp2
					}
					files = append(files, ProjectFile{id, name, time.UnixMilli(timeStamp1)})

					id = ""
					name = ""
					timeStamp1 = 0
					timeStamp2 = 0
				}
			}
			parsedStr = ""
			isStringOpen = !isStringOpen
			continue
		}

		if isStringOpen {
			parsedStr += string(data[idx])
		}
	}

	return files

}


func createBackup(teams []Team, backupCount int) error {
	backupsInfo, err := os.Stat("backups")
	if(os.IsNotExist(err)){
		err := os.Mkdir("backups", os.ModeDir)
		if(err != nil){
			return err
		}
		backupsInfo, err = os.Stat("backups")
		if(err != nil){
			return fmt.Errorf("Failed to get backup folder info")
		}
	}else {
		return err
	}

	if(!backupsInfo.IsDir()){
		return fmt.Errorf("Another entry already exists with \"backups\" name")
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
