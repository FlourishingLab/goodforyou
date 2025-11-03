package api

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"slices"
	"time"
	"user-db/db"
	"user-db/llm"
	"user-db/questions"
	"user-db/shared"
)

const HOLISTIC string = "holistic"
const COOKIENAME string = "uid"

func (s *Server) ResetUser(w http.ResponseWriter, r *http.Request) {
	uid := getUid(r)
	db.ResetUser(uid)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

func (s *Server) GetUserId(w http.ResponseWriter, r *http.Request) {

	uid := getUid(r)
	if uid == "" {
		var err error
		uid, err = randomID(16) // 128-bit
		if err != nil {
			log.Printf("Error generating new user uid: %s", uid)
			return
		}
		err = db.NewUser(uid)
		if err != nil {
			log.Printf("Not able to create user with ID: %s", uid)
			return
		}

		log.Printf("Created new user with ID: %s", uid)
		http.SetCookie(w, &http.Cookie{
			Name:     COOKIENAME,
			Value:    uid,
			Path:     "/",
			MaxAge:   31536000, // 1 year
			Secure:   true,     // set true in HTTPS
			HttpOnly: true,
			SameSite: http.SameSiteNoneMode,
		})
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{COOKIENAME: uid})
}

func (s *Server) GetQuestions(w http.ResponseWriter, r *http.Request) {

	uid := getUid(r)

	// get answers from DB
	userAnswer, err := db.GetUser(uid)
	if err != nil {
		log.Printf("error getting user (%s): %v", uid, err)
	}

	prioDimension := r.URL.Query().Get("dimension")

	// Expected path structure: /v1/questions or /v1/questions/{dimension}
	if prioDimension != "" {
		if !questions.IsValidDimension(prioDimension) {
			http.Error(w, fmt.Sprintf("Invalid dimension: %s", prioDimension), http.StatusInternalServerError)
			log.Printf("error, unknown dimension: %s", prioDimension)
			return
		}
		log.Printf("Prioritised dimension: %s", prioDimension)
	} else {
		log.Printf("No prioritised dimension specified")
	}

	nextQuestions, err := questions.GetNextQuestions(userAnswer, prioDimension)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("Could not get questions for user (%s): %v", uid, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(nextQuestions)
}

func (s *Server) SubmitResponses(w http.ResponseWriter, r *http.Request) {
	var payload ResponsePayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Printf("Failed trying to decode paypload: %s", err)
		return
	}

	uid := getUid(r)

	for _, answer := range payload.Answers {
		kind, err := shared.ToAnswerKind(answer.Kind)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			log.Printf("Failed trying to convert answer kind: %s", err)
			return
		}
		// TODO Insert Many. This is not atomic
		err = db.UpsertAnswer(uid, answer.QuestionID, kind, answer.Value)
		if err != nil {
			log.Printf("Failed trying to upsert answer %d with error: %s", answer.QuestionID, err)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})

	ua, err := db.GetUser(uid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("Failed trying to get user: %s", err)
		return
	}

	completeDims := questions.GetCompleteDimensions(ua)

	for _, dimensionName := range completeDims {
		if ua.NeedsInsight(dimensionName) {
			go func(userID, dimName string, uaCopy db.UserAnswers) {
				err := db.UpsertInsight(userID, dimName, "", db.GENERATING)
				if err != nil {
					log.Printf("Failed trying to upsert insight: %s", err)
					return
				}
				dimensionInsight := llm.DimensionPrompt(dimName, uaCopy.DimensionRatingsToString(dimName, questions.GetDimensions()))
				err = db.UpsertInsight(userID, dimName, dimensionInsight, db.DONE)
				if err != nil {
					log.Printf("Failed trying to upsert insight: %s", err)
					return
				}

				s.Broker.Publish(InsightEvent{
					Name:   dimName,
					UserID: userID,
					Data:   dimensionInsight,
				})

			}(uid, dimensionName, ua)
		}
	}
}

func (s *Server) GenerateHolistic(w http.ResponseWriter, r *http.Request) {

	uid := getUid(r)

	userAnswers, err := db.GetUser(uid)
	if err != nil {
		log.Printf("error getting user (%s): %v", uid, err)
		return
	}

	sortedDims, sortedFacets := userAnswers.GetSorted(questions.GetQuestions(), questions.GetDimensions())
	resp := llm.HolisticPrompt(sortedDims, sortedFacets)

	err = db.UpsertInsight(uid, HOLISTIC, resp, db.DONE)
	if err != nil {
		log.Printf("Failed trying to upsert insight: %s", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
	s.Broker.Publish(InsightEvent{
		Name:   HOLISTIC,
		UserID: uid,
		Data:   resp,
	})

}

func (s *Server) GetInsightsLLM(w http.ResponseWriter, r *http.Request) {

	uid := getUid(r)

	userAnswers, err := db.GetUser(uid)
	if err != nil {
		log.Printf("error getting user (%s): %v", uid, err)
		return
	}

	insights := map[string]string{}

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

func (s *Server) InsightsStream(w http.ResponseWriter, r *http.Request) {

	// Set http headers required for SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	uid := getUid(r)
	// Create a channel for client disconnection
	clientGone := r.Context().Done()

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		log.Printf("Streaming unsupported for uid(%s), user-agent (%s)", uid, r.UserAgent())
		return
	}

	// Send a comment line immediately so the connection is considered "active".
	fmt.Fprintf(w, ": connected %s\n\n", time.Now().UTC().Format(time.RFC3339))
	flusher.Flush()

	heartbeat := time.NewTicker(15 * time.Second)
	defer heartbeat.Stop()

	sub := s.Broker.Subscribe(r.Context(), uid, 8)

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
				log.Printf("Error writing heartbeat: %s", err)
				return
			}
			flusher.Flush()
		case insightEvent, ok := <-sub.ch:
			if !ok {
				return
			}
			fmt.Fprintf(w, "event: %s\n", insightEvent.Name)
			flusher.Flush()
		}
	}

}

// ---------- Utilities ----------
func WithCORS(allowedOrigins []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// log.Printf("CORS: Request from origin: %s, allowed: %v", origin, allowedOrigins)

			// Always vary on these so caches behave
			w.Header().Add("Vary", "Origin")
			w.Header().Add("Vary", "Access-Control-Request-Method")
			w.Header().Add("Vary", "Access-Control-Request-Headers")

			// Reflect a permitted origin (must NOT be *)
			if origin != "" && slices.Contains(allowedOrigins, origin) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
				// Optional: cache preflight
				w.Header().Set("Access-Control-Max-Age", "86400")
			} else {
				log.Printf("CORS: Origin %s NOT allowed", origin)
			}

			if r.Method == http.MethodOptions {
				// log.Printf("CORS: Handling preflight request")
				// Preflight response
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func randomID(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func getUid(r *http.Request) string {
	var uid string
	if c, err := r.Cookie(COOKIENAME); err != nil {
		// TODO check for cookie not found error otherwise handle error differently
		if err == http.ErrNoCookie {
			// Handle the case where the cookie is not found
			log.Printf("Cookie '%s' not found", COOKIENAME)
			return ""
		}
		// Handle other errors (e.g., malformed cookies)
		log.Printf("Error getting cookie '%s': %s", COOKIENAME, err)
		return ""
	} else {
		uid = c.Value
	}
	return uid
}
