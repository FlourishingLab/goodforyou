package api

import (
	"encoding/json"
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

func SubmitResponses(w http.ResponseWriter, r *http.Request) {
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
			}(ua.UserID, dimensionName, ua)
		}
	}

}

func GetInsightsLLM(w http.ResponseWriter, r *http.Request) {

	log.Printf("Received request for topics with path: %s", r.URL.Path)

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
			sortedDims, sortedFacets := userAnswers.GetSorted(questions.GetQuestions())
			resp := llm.HolisticPrompt(sortedDims, sortedFacets)
			insights = map[string]string{"holistic": resp}
			db.UpcertInsight(userID, HOLISTIC, resp)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]bool{"success": true})
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
