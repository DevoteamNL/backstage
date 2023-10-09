package main

import (
	"devlake-go/group-sync/api/service"
	"devlake-go/group-sync/api/sql_client"
	"devlake-go/group-sync/api/validation"
	"encoding/json"
	"log"
	"net/http"
)

func metricHandler(client sql_client.ClientInterface) func(w http.ResponseWriter, r *http.Request) {
	serviceMap := map[string]service.Service{
		"df_count":   service.DfCountService{Client: client},
		"df_average": service.DfAverageService{Client: client},
	}

	return func(w http.ResponseWriter, r *http.Request) {
		queries := r.URL.Query()
		parameters, err := validation.ValidServiceParameters(queries)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		response, err := serviceMap[parameters.TypeQuery].ServeRequest(parameters)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		responseError := json.NewEncoder(w).Encode(response)
		if responseError != nil {
			http.Error(w, responseError.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func main() {
	http.HandleFunc("/dora/api/metric", metricHandler(sql_client.New()))

	log.Fatal(http.ListenAndServe(":10666", nil))
}
