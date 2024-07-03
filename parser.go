package main

import (
	"strings"
	"time"
	"strconv"
)


type Team struct {
	id       string
	name     string
	projects []Project
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
				} else if !(strings.Contains(parsedStr, "~") || strings.Contains(parsedStr, "^") || strings.Contains(parsedStr, "https://")) {
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