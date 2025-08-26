package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"strings"

	"sync"
	"time"
	"user-db/llm"
	"user-db/questions"
)

type ResponsePayload struct {
	UserID  string             `json:"userId"`
	Answers []questions.Answer `json:"answers"`
}

// In-memory storage (for prototyping)
var responses = struct {
	sync.Mutex
	Data map[string][]questions.Answer
}{Data: make(map[string][]questions.Answer)}

// ---------- Handlers ----------
func handleUserId(w http.ResponseWriter, r *http.Request) {

	// get userID from path
	var userID string
	parts := strings.Split(r.URL.Path, "/")
	// it always seems to have at least 4 parts: "", "v1", "userid", ""
	log.Print(parts)
	if len(parts) == 4 {
		userID = parts[3]
		// check if userID exists
		if _, exists := responses.Data[userID]; !exists {
			log.Printf("userID not found: %s", userID)
			userID = generateUserID()
			responses.Lock()
			responses.Data[userID] = []questions.Answer{}
			responses.Unlock()
		}
	} else {
		http.Error(w, "invalid path", http.StatusBadRequest)
		log.Printf("Invalid path: %s", r.URL.Path)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"userId": userID})
}

func getQuestions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(questions.GetQuestions())
}

func submitResponses(w http.ResponseWriter, r *http.Request) {
	var payload ResponsePayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	responses.Lock()
	defer responses.Unlock()
	responses.Data[payload.UserID] = payload.Answers

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

func getTopicsLLM(w http.ResponseWriter, r *http.Request) {

	log.Printf("Received request for topics with path: %s", r.URL.Path)

	// get userID from path
	var userID string
	parts := strings.Split(r.URL.Path, "/")
	// path: v1/topics/llm/USERID/x
	if len(parts) >= 5 {
		userID = parts[4]
	} else {
		http.Error(w, "invalid path", http.StatusBadRequest)
		log.Printf("Invalid path: %s", r.URL.Path)
		return
	}

	sortedDims, sortedFacets := questions.GetSorted(responses.Data[userID])

	llmResponse := llm.Prompt(sortedDims, sortedFacets)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(llmResponse)

}

// ---------- Utilities ----------

// TODO before production
func withCors(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// Handle preflight
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		h(w, r)
	}
}

func generateUserID() string {
	rand.New(rand.NewSource(time.Now().UnixNano()))
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 8)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}

	log.Printf("Generated UserID: %s", string(b))
	return string(b)
}

// ---------- Main ----------
func main() {

	http.HandleFunc("/v1/userid/", withCors(handleUserId))
	http.HandleFunc("/v1/questions", withCors(getQuestions))
	http.HandleFunc("/v1/responses", withCors(submitResponses))
	http.HandleFunc("/v1/topics/llm/", withCors(getTopicsLLM))

	log.Println("Server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
