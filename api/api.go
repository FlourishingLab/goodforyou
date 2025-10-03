package api

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"slices"
	"time"
	"user-db/db"
)

const COOKIENAME string = "uid"

func (s *Server) ResetUser(w http.ResponseWriter, r *http.Request) {

	// Delete the cookie by setting MaxAge to -1 and Expires to a past date
	http.SetCookie(w, &http.Cookie{
		Name:     COOKIENAME,
		Value:    "",
		Path:     "/",
		MaxAge:   -1, // This expires the cookie immediately
		Expires:  time.Unix(0, 0),
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteNoneMode,
	})

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
		return ""
	} else {
		uid = c.Value
	}
	return uid
}
