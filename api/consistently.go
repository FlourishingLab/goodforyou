package api

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
	"user-db/db"
	"user-db/paragraph"
	"user-db/questions"
)

func (s *Server) Consistently(w http.ResponseWriter, r *http.Request) {

	uid := getUid(r)

	userData, err := db.GetUser(uid)
	if err != nil {
		log.Printf("error getting user (%s): %v", uid, err)
	}

	var response ConsistentlyResponse

	nextQuestions, err := questions.GetNextQuestions(userData, 3)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("Could not get questions for user (%s): %v", uid, err)
		return
	}

	if !isToday(userData.LastVisit) {
		nextParagraph := userData.IDOfNextParagraph()

		newStreak := 1
		if isYesterday(userData.LastVisit) {
			newStreak = userData.Streak + 1
		}

		db.NewDay(userData.UserID, newStreak, nextParagraph)

		response = ConsistentlyResponse{
			Streak:    newStreak,
			Paragraph: paragraph.GetParagraph(nextParagraph),
			Questions: nextQuestions,
		}
	} else {
		response = ConsistentlyResponse{
			Streak:    userData.Streak,
			Paragraph: paragraph.GetParagraph(userData.IDOfNextParagraph()),
			Questions: nextQuestions,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func isToday(t time.Time) bool {
	return ofSameDay(t, time.Now())
}

func isYesterday(t time.Time) bool {
	return ofSameDay(t, time.Now().AddDate(0, 0, -1))
}

func ofSameDay(one time.Time, two time.Time) bool {
	y1, m1, d1 := one.Date()
	y2, m2, d2 := two.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}
