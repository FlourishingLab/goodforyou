package main

import (
	"log"
	"net/http"

	"user-db/api"
	"user-db/shared"
)

// ---------- Main ----------
func main() {

	config, err := shared.LoadConfig()
	if err != nil {
		log.Fatalf("Error getting config: %s", err)
	}

	s := api.Server{
		Broker: api.NewBroker(1024),
	}

	http.HandleFunc("/v1/userid/", api.WithCors(s.HandleUserId, config.CorsOrigin))
	http.HandleFunc("/v1/questions/", api.WithCors(s.GetQuestions, config.CorsOrigin))
	http.HandleFunc("/v1/responses", api.WithCors(s.SubmitResponses, config.CorsOrigin))
	http.HandleFunc("/v1/insights/llm/", api.WithCors(s.GetInsightsLLM, config.CorsOrigin))
	http.HandleFunc("/v1/insights/stream/", api.WithCors(s.InsightsStream, config.CorsOrigin))

	log.Println("Server running")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
