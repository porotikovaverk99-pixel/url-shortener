package middleware

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type contextKey string

const UserIDKey contextKey = "userID"

type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

func Auth(secretKey string) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := []byte(secretKey)

			cookie, err := r.Cookie("user_id")
			var userID string
			var signedCookie string

			if err == nil && cookie != nil {
				signedCookie = cookie.Value
			}

			if signedCookie == "" || !verifyCookie(signedCookie, key) {
				userID = uuid.New().String()
				signedValue := signUserID(userID, key)
				http.SetCookie(w, &http.Cookie{
					Name:     "user_id",
					Value:    signedValue,
					Path:     "/",
					HttpOnly: true,
					MaxAge:   30 * 24 * 60 * 60,
				})
			} else {
				userID = extractUserID(cookie.Value)
			}

			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			h.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func ValidateToken(tokenString, secretKey string) (string, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})
	if err != nil {
		return "", err
	}
	if !token.Valid {
		return "", fmt.Errorf("invalid token")
	}
	return claims.UserID, nil
}

func verifyCookie(signed string, key []byte) bool {
	parts := strings.Split(signed, ".")
	if len(parts) != 2 {
		return false
	}
	userID := parts[0]
	expectedMAC, err := base64.URLEncoding.DecodeString(parts[1])
	if err != nil {
		return false
	}
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(userID))
	actualMAC := mac.Sum(nil)
	return hmac.Equal(expectedMAC, actualMAC)
}

func extractUserID(signed string) string {
	parts := strings.Split(signed, ".")
	if len(parts) != 2 {
		return ""
	}
	return parts[0]
}

func signUserID(userID string, key []byte) string {
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(userID))
	signature := mac.Sum(nil)
	return userID + "." + base64.URLEncoding.EncodeToString(signature)
}

func GetUserID(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(UserIDKey).(string)
	return userID, ok && userID != ""
}
