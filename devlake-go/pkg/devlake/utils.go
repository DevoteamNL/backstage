package devlake

import (
	"devlake-go/group-sync/pkg/config"
	"strconv"
	"strings"
)

func teamsApiUrlFromEnv() string {
	return config.LookupEnvDefault("DEVLAKE_URL", "http://localhost:4000/") + "api/plugins/org/teams.csv"
}

func TeamNamePredicate(teamName string) func(devLakeTeam []string) bool {
	return func(devLakeTeam []string) bool {
		return strings.EqualFold(devLakeTeam[TeamNameColumn], teamName)
	}
}

func LargestTeamId(devLakeTeams [][]string) int {
	latestId := 0
	for _, devLakeTeam := range devLakeTeams {
		idAsInt, err := strconv.Atoi(devLakeTeam[TeamIdColumn])
		if err == nil && latestId < idAsInt {
			latestId = idAsInt
		}
	}
	return latestId
}
