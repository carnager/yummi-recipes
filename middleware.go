package main

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"
)

type contextKey string

const userContextKey contextKey = "user"

func ctxUser(r *http.Request) *User {
	if u, ok := r.Context().Value(userContextKey).(*User); ok {
		return u
	}
	return nil
}

func (app *App) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var token string

		if strings.HasPrefix(r.URL.Path, "/api/") {
			auth := r.Header.Get("Authorization")
			if strings.HasPrefix(auth, "Bearer ") {
				token = strings.TrimPrefix(auth, "Bearer ")
			}
		} else {
			if c, err := r.Cookie("session"); err == nil {
				token = c.Value
			}
		}

		if token != "" {
			session, err := app.store.GetSessionByToken(token)
			if err == nil && session.ExpiresAt.After(time.Now()) {
				user, err := app.store.GetUserByID(session.UserID)
				if err == nil {
					ctx := context.WithValue(r.Context(), userContextKey, user)
					r = r.WithContext(ctx)
				}
			}
		}

		next.ServeHTTP(w, r)
	})
}

func requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if ctxUser(r) == nil {
			if strings.HasPrefix(r.URL.Path, "/api/") {
				http.Error(w, `{"error":"nicht autorisiert"}`, http.StatusUnauthorized)
			} else {
				http.Redirect(w, r, "/anmelden", http.StatusSeeOther)
			}
			return
		}
		next(w, r)
	}
}

func requireAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := ctxUser(r)
		if user == nil {
			http.Redirect(w, r, "/anmelden", http.StatusSeeOther)
			return
		}
		if !user.IsAdmin {
			http.Error(w, "Nicht berechtigt", http.StatusForbidden)
			return
		}
		next(w, r)
	}
}

func logMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start).Round(time.Millisecond))
	})
}
