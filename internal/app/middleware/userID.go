package middleware

import (
	"context"
	"errors"
	"net/http"
	"strconv"
)

var ErrNoUserID = errors.New("no user id")

var userIDContextKey = &contextKey{"user id context"}

type contextKey struct {
	name string
}

func (c *contextKey) String() string {
	return c.name
}

func UserID() func(http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			idParam := r.Header.Get("X-UserID")
			if idParam != "" {
				id, err := strconv.ParseInt(idParam, 10, 64)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				ctx := context.WithValue(r.Context(), userIDContextKey, id)
				r = r.WithContext(ctx)
			}
			handler.ServeHTTP(w, r)
		})
	}
}

func GetUserID(ctx context.Context) (int64, error) {
	if value, ok := ctx.Value(userIDContextKey).(int64); ok {
		return value, nil
	}
	return 0, ErrNoUserID
}
