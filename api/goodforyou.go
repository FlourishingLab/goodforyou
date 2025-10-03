package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
	"user-db/db"
	"user-db/llm"
	"user-db/questions"
	"user-db/shared"
)

const HOLISTIC string = "holistic"

func (s *Server) GetQuestions(w http.ResponseWriter, r *http.Request) {

	uid := getUid(r)

	// need answered questions
	userAnswer, err := db.GetUser(uid)
	if err != nil {
		log.Printf("error getting user (%s): %v", uid, err)
	}

	nextQuestions, err := questions.GetNextQuestions(userAnswer, 5)
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
