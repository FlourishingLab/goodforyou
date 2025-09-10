package main

import (
	"log"
	"net/http"

	"user-db/api"
)

// ---------- Main ----------
func main() {

	http.HandleFunc("/v1/userid/", api.WithCors(api.HandleUserId))
	http.HandleFunc("/v1/questions/", api.WithCors(api.GetQuestions))
	http.HandleFunc("/v1/responses", api.WithCors(api.SubmitResponses))
	http.HandleFunc("/v1/insights/llm/", api.WithCors(api.GetInsightsLLM))

	log.Println("Server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
