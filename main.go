package main

import (
	"log"
	"net/http"

	"user-db/api"
)

// ---------- Main ----------
func main() {

	s := api.Server{
		Broker: api.NewBroker(1024),
	}

	http.HandleFunc("/v1/userid/", api.WithCors(api.HandleUserId))
	http.HandleFunc("/v1/questions/", api.WithCors(api.GetQuestions))
	http.HandleFunc("/v1/responses", api.WithCors(s.SubmitResponses))
	http.HandleFunc("/v1/insights/llm/", api.WithCors(s.GetInsightsLLM))
	http.HandleFunc("/v1/insights/stream/", api.WithCors(s.InsightsStream))

	log.Println("Server running")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
