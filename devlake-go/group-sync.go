package main

import (
	"bytes"
	"context"
	"encoding/csv"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/tdabasinskas/go-backstage/v2/backstage"
)

func lookupEnvDefault(envKey string, envDefaultValue string) string {
	if val, ok := os.LookupEnv(envKey); ok {
		return val
	}
	return envDefaultValue
}

const DevLakeTeamIdColumn = 0
const DevLakeTeamNameColumn = 1
const DevLakeTeamParentIdColumn = 3

func retrieveBackstageTeams() []backstage.Entity {
	client, err := backstage.NewClient(lookupEnvDefault("BACKSTAGE_URL", "http://localhost:7007/"), "default", nil)
	backstageTeams, _, err := client.Catalog.Entities.List(context.Background(), &backstage.ListEntityOptions{
		Filters: []string{
			"kind=group",
		},
		Fields: []string{},
		Order:  []backstage.ListEntityOrder{{Direction: backstage.OrderDescending, Field: "metadata.name"}},
	})

	if err != nil {
		log.Fatal("Cannot retrieve Backstage teams: ", err)
	}

	return backstageTeams
}

func devLakeTeamsApiUrlFromEnv() string {
	return lookupEnvDefault("DEVLAKE_URL", "http://localhost:4000/") + "api/plugins/org/teams.csv"
}

func retrieveDevLakeTeams() [][]string {
	if _, ok := os.LookupEnv("REPLACE_DEVLAKE_TEAMS"); ok {
		return [][]string{{"Id", "Name", "Alias", "ParentId", "SortingIndex"}}
	}

	resp, err := http.Get(devLakeTeamsApiUrlFromEnv())
	if err != nil {
		log.Fatal("Cannot retrieve DevLake teams: ", err)
	}
	csvReader := csv.NewReader(resp.Body)
	devLakeTeams, err := csvReader.ReadAll()
	if err != nil {
		log.Fatal("Cannot read DevLake team CSV format: ", err)
	}

	return devLakeTeams
}

func devLakeTeamPredicate(teamName string) func(devLakeTeam []string) bool {
	return func(devLakeTeam []string) bool {
		return strings.EqualFold(devLakeTeam[DevLakeTeamNameColumn], teamName)
	}
}

func largestTeamId(devLakeTeams [][]string) int {
	latestId := 0
	for _, devLakeTeam := range devLakeTeams {
		idAsInt, err := strconv.Atoi(devLakeTeam[DevLakeTeamIdColumn])
		if err == nil && latestId < idAsInt {
			latestId = idAsInt
		}
	}
	return latestId
}

func appendNewTeams(backstageTeams []backstage.Entity, devLakeTeams [][]string) [][]string {
	lastId := largestTeamId(devLakeTeams)

	for _, backStageTeam := range backstageTeams {
		currentIndex := slices.IndexFunc(devLakeTeams, devLakeTeamPredicate(backStageTeam.Metadata.Name))

		if currentIndex != -1 {
			log.Printf("Team already exists in DevLake: %s\n", backStageTeam.Metadata.Name)
		} else {
			lastId += 1
			devLakeTeams = append(devLakeTeams, []string{strconv.Itoa(lastId), backStageTeam.Metadata.Name, "", "", ""})
			currentIndex = len(devLakeTeams) - 1
		}

		createRelationships(backStageTeam, devLakeTeams, currentIndex)
	}
	return devLakeTeams
}

func createRelationships(backStageTeam backstage.Entity, devLakeTeams [][]string, sourceIndex int) {
	for _, relation := range backStageTeam.Relations {
		targetIndex := slices.IndexFunc(devLakeTeams, devLakeTeamPredicate(relation.Target.Name))

		if targetIndex == -1 {
			continue
		}
		if relation.Type == "childOf" {
			devLakeTeams[sourceIndex][DevLakeTeamParentIdColumn] = devLakeTeams[targetIndex][DevLakeTeamIdColumn]
		} else if relation.Type == "parentOf" {
			devLakeTeams[targetIndex][DevLakeTeamParentIdColumn] = devLakeTeams[sourceIndex][DevLakeTeamIdColumn]
		}
	}
}

func updateDevLakeTeams(devLakeTeams [][]string) {
	buf := new(bytes.Buffer)
	csvWriter := csv.NewWriter(buf)

	if err := csvWriter.WriteAll(devLakeTeams); err != nil {
		log.Fatal("Cannot write DevLake teams to CSV format: ", err)
	}

	multipartBody := &bytes.Buffer{}
	writer := multipart.NewWriter(multipartBody)
	part, _ := writer.CreateFormFile("file", "teams.csv")

	if _, err := io.Copy(part, buf); err != nil {
		log.Fatal("Cannot copy CSV buffer to multipart file: ", err)
	}

	if err := writer.Close(); err != nil {
		log.Fatal("Cannot close CSV writer: ", err)
	}

	req, err := http.NewRequest("PUT", devLakeTeamsApiUrlFromEnv(), multipartBody)

	if err != nil {
		log.Fatal("Cannot create DevLake PUT request: ", err)
	}

	req.Header.Add("Content-Type", writer.FormDataContentType())

	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)

	if err != nil {
		log.Fatal("Cannot update DevLake teams CSV: ", err)
	}
	resBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("Cannot read response from DevLake teams update request: ", err)
	}

	log.Printf("Response: %s\n", resBody)
}

func main() {
	backstageTeams := retrieveBackstageTeams()
	devLakeTeams := retrieveDevLakeTeams()
	devLakeTeams = appendNewTeams(backstageTeams, devLakeTeams)

	updateDevLakeTeams(devLakeTeams)
}
