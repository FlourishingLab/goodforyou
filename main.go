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
	// Wrap handlers with CORS middleware and user middleware
	http.Handle("/v1/user/reset", api.WithCORS(config.CorsOrigins)(http.HandlerFunc(s.ResetUser)))
	http.Handle("/v1/user/id", api.WithCORS(config.CorsOrigins)(http.HandlerFunc(s.GetUserId)))
	http.Handle("/v1/questions", api.WithCORS(config.CorsOrigins)(http.HandlerFunc(s.GetQuestions)))
	http.Handle("/v1/responses", api.WithCORS(config.CorsOrigins)(http.HandlerFunc(s.SubmitResponses)))
	http.Handle("/v1/insights/llm/generate/holistic", api.WithCORS(config.CorsOrigins)(http.HandlerFunc(s.GenerateHolistic)))
	http.Handle("/v1/insights/llm", api.WithCORS(config.CorsOrigins)(http.HandlerFunc(s.GetInsightsLLM)))
	http.Handle("/v1/insights/stream", api.WithCORS(config.CorsOrigins)(http.HandlerFunc(s.InsightsStream)))

	log.Println("Server running")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
