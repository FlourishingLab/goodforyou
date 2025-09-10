package api

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"
	"user-db/db"
	"user-db/llm"
	"user-db/questions"
	"user-db/shared"
)

var HOLISTIC string = "holistic"

// ---------- Handlers ----------
func HandleUserId(w http.ResponseWriter, r *http.Request) {

	// get userID from path
	var userID string
	parts := strings.Split(r.URL.Path, "/")
	// it always seems to have at least 4 parts: "", "v1", "userid", ""
	log.Print(parts)
	if len(parts) == 4 {
		userID = parts[3]
		// check if userID exists
		if _, exists := db.GetUser(userID); !exists {
			log.Printf("userID not found: %s", userID)
			userID = generateUserID()
			db.NewUser(userID)
		}
	} else {
		http.Error(w, "invalid path", http.StatusBadRequest)
		log.Printf("Invalid path: %s", r.URL.Path)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"userId": userID})
}

func GetQuestions(w http.ResponseWriter, r *http.Request) {

	// get userID from path
	var userID string
	parts := strings.Split(r.URL.Path, "/")
	// path: v1/questions/USERID
	if len(parts) >= 4 {
		userID = parts[3]
	} else {
		http.Error(w, "invalid path", http.StatusBadRequest)
		log.Printf("Invalid path: %s", r.URL.Path)
		return
	}

	nextQuestions, err := questions.GetNextQuestions(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// TODO, add this to db

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(nextQuestions)
}

func (s *Server) SubmitResponses(w http.ResponseWriter, r *http.Request) {
	var payload ResponsePayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	for _, answer := range payload.Answers {
		kind, err := shared.ToAnswerKind(answer.Kind)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		// TODO Insert Many. This is not atomic
		db.UpsertAnswer(payload.UserID, answer.QuestionID, kind, answer.Value)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})

	ua, _ := db.GetUser(payload.UserID)

	completeDims := questions.GetCompleteDimensions(ua)

	for _, dimensionName := range completeDims {
		if !ua.HasInsight(dimensionName) {
			go func(userID, dimName string, uaCopy db.UserAnswers) {
				dimensionInsight := llm.DimensionPrompt(dimName, uaCopy.DimensionRatingsToString(dimName, questions.GetDimensions()))
				db.UpcertInsight(userID, dimName, dimensionInsight)

				s.Broker.Publish(InsightEvent{
					Name:   dimName,
					UserID: userID,
					Data:   dimensionInsight,
				})

			}(ua.UserID, dimensionName, ua)
		}
	}

}

func (s *Server) GetInsightsLLM(w http.ResponseWriter, r *http.Request) {

	log.Printf("Received request for insights with path: %s", r.URL.Path)

	// get userID from path
	var userID string
	parts := strings.Split(r.URL.Path, "/")
	// path: v1/insights/llm/USERID
	if len(parts) >= 6 {
		userID = parts[4]
		userAnswers, exists := db.GetUser(userID)
		if !exists {
			http.Error(w, "user not found", http.StatusNotFound)
			log.Printf("UserID not found: %s", userID)
			return
		}

		insights := map[string]string{}
		insightsType := parts[5]
		if insightsType == HOLISTIC {
			sortedDims, sortedFacets := userAnswers.GetSorted(questions.GetQuestions(), questions.GetDimensions())
			resp := llm.HolisticPrompt(sortedDims, sortedFacets)
			insights = map[string]string{"holistic": resp}
			db.UpcertInsight(userID, HOLISTIC, resp)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]bool{"success": true})
			s.Broker.Publish(InsightEvent{
				Name:   HOLISTIC,
				UserID: userID,
				Data:   resp,
			})
		} else {
			if userAnswers.HasInsight(HOLISTIC) {
				insights[HOLISTIC] = string(userAnswers.GetInsight(HOLISTIC))
			}
			for dimensionName := range questions.GetDimensions() {
				if userAnswers.HasInsight(dimensionName) {
					insights[dimensionName] = string(userAnswers.GetInsight(dimensionName))
				}
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(insights)
		}

	} else {
		http.Error(w, "invalid path, expecting format 'v1/insights/llm/USERID/(DIMENSION|HOLISTIC)'", http.StatusBadRequest)
		log.Printf("Invalid path: %s", r.URL.Path)
		return
	}
}

func (s *Server) InsightsStream(w http.ResponseWriter, r *http.Request) {

	// Set http headers required for SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// get userID from path
	var userID string
	parts := strings.Split(r.URL.Path, "/")
	// path: v1/insights/stream/USERID
	if len(parts) >= 5 {
		userID = parts[4]
	} else {
		http.Error(w, "invalid path", http.StatusBadRequest)
		log.Printf("Invalid path: %s", r.URL.Path)
		return
	}

	// Create a channel for client disconnection
	clientGone := r.Context().Done()

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	// Send a comment line immediately so the connection is considered "active".
	fmt.Fprintf(w, ": connected %s\n\n", time.Now().UTC().Format(time.RFC3339))
	flusher.Flush()

	heartbeat := time.NewTicker(15 * time.Second)
	defer heartbeat.Stop()

	sub := s.Broker.Subscribe(r.Context(), userID, 8)

	for {
		select {
		case <-clientGone:
			log.Println("Client disconnected")
			return
		case <-heartbeat.C:
			// Send an event to the client
			// Here we send only the "data" field, but there are few others
			_, err := fmt.Fprintf(w, ": ping %d\n\n", time.Now().Unix())
			if err != nil {
				return
			}
			flusher.Flush()
		case insightEvent, ok := <-sub.ch:
			if !ok {
				return
			}
			payload, _ := json.Marshal(insightEvent.Data)
			fmt.Fprintf(w, "event: %s\n", insightEvent.Name)
			fmt.Fprintf(w, "data: %s\n\n", payload)
			flusher.Flush()
		}
	}

}

// ---------- Utilities ----------

// TODO before production
func WithCors(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*") // Replace * with your frontend's origin in production
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
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
