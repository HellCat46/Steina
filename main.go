package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

var headers = http.Header{
	"Content-Type": {"application/json"},
}

var client = &http.Client{
}

func main(){
	// Gettings the Penpot Access Token
	if(len(os.Getenv("PENPOT_TOKEN")) == 0){
		fmt.Println("ENV not set.")
		return;
	}
	headers.Add("Authorization","Token " +os.Getenv("PENPOT_TOKEN"))
	
	getTeams()
}

// https://design.penpot.app/api/rpc/command/get-projects?team-id=<TEAM_ID>
func getProjects(){

}

// https://design.penpot.app/api/rpc/command/get-project-files?project-id=<PROJECT_ID>
func getProjectFiles(){}


// https://design.penpot.app/api/rpc/command/export-binfile
// {"~:file-id":"~u<File_ID>","~:include-libraries":true,"~:embed-assets":false}
func downloadFile(){

}


type Team struct {
	id   string
	name string
}

func getTeams(){
	
	req, err := http.NewRequest("GET", "https://design.penpot.app/api/rpc/command/get-teams", bytes.NewBuffer([]byte("")))
	if(err != nil){
		fmt.Println(err)
		return
	}
	req.Header = headers
	res, err := client.Do(req)
	if(err != nil){
		fmt.Println(err)
		return
	}

	data, err := io.ReadAll(res.Body)
	if(err != nil){
		fmt.Println(err)
		return
	}
	
	if(res.StatusCode == 200){
		teams := parseTeamList(string(data))
		for _, team := range teams {
			fmt.Printf("ID: %s\nName: %s\n\n", team.id, team.name)
		}
	}
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
					teams = append(teams, Team{id, name})
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