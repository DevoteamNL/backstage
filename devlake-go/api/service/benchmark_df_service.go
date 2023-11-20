package service

import (
	"github.com/devoteamnl/opendora/api/models"
	"github.com/devoteamnl/opendora/api/sql_client"
	"github.com/devoteamnl/opendora/api/sql_client/sql_queries"
)

type BenchmarkDfService struct {
	Client sql_client.ClientInterface
}

func (service BenchmarkDfService) ServeRequest(params ServiceParameters) (models.BenchmarkResponse, error) {
	typeQueryMap := map[string]string{
		"df": sql_queries.BenchmarkDfSql,
	}

	query := typeQueryMap[params.TypeQuery]

	key, err := service.Client.QueryBenchmark(query, sql_client.QueryParams{To: params.To, From: params.From, Project: params.Project})

	return models.BenchmarkResponse{Key: key}, err
}
