package middleware

import (
	"context"
	"fmt"
	"net/http"

	m "de.whatwapp/app/model"
	s "de.whatwapp/app/store"
)

type key string

type Middleware func(http.Handler) http.Handler

const (
	Role     key = "role"
	Username key = "username"
)

func BasicAuth(userStore *s.Store[m.User]) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// username := r.Header.Get("username")
			ctx := context.WithValue(r.Context(), Username, "SwagLord")
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func WithRole(roles []m.Role) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userRole := r.Context().Value(Role)
			fmt.Println(userRole)
			if len(roles) == 0 {
				next.ServeHTTP(w, r)
			}
			for _, acceptedRole := range roles {
				if userRole == acceptedRole {
					next.ServeHTTP(w, r)
					return
				}
			}
			http.Error(w, "you don't have the role to access this endpoint", http.StatusUnauthorized)
		})
	}
}

func GetMiddlewares(handler []Middleware) []func(http.Handler) http.Handler {
	var handlers []func(http.Handler) http.Handler
	for _, h := range handler {
		handlers = append(handlers, h)
	}
	return handlers
}
