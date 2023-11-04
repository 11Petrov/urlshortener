package auth

import (
	"context"
	"net/http"
	"time"

	"github.com/11Petrov/urlshortener/internal/logger"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

type Claims struct {
	jwt.RegisteredClaims
	UserID string
}

type KeyType string

const (
	UserIDKey  KeyType = "userID"
	TokenEXP           = time.Hour * 3
	SecretKEY          = "supersecretkey"
	CookieName         = "auth"
)

func AuthMiddleware(next http.Handler) http.Handler {
	authFn := func(rw http.ResponseWriter, r *http.Request) {
		var userID string
		var tokenString string

		log := logger.LoggerFromContext(r.Context())

		cookie, err := r.Cookie(CookieName)
		if err != nil {
			userID = uuid.New().String()
			tokenString, err = BuildJWTString(r.Context(), userID)
			if err != nil {
				log.Errorf("AuthMiddleware BuildJWTString err = ", err)
			}
		} else if userID, err = GetUserID(r.Context(), cookie.Value); err != nil {
			userID := uuid.New().String()
			tokenString, err = BuildJWTString(r.Context(), userID)
			if err != nil {
				log.Errorf("AuthMiddleware BuildJWTString err = ", err)
			}
		}

		if tokenString != "" {
			http.SetCookie(rw, &http.Cookie{
				Name:    CookieName,
				Value:   tokenString,
				Expires: time.Now().Add(TokenEXP),
			})
		}

		cxt := context.WithValue(r.Context(), UserIDKey, userID)
		next.ServeHTTP(rw, r.WithContext(cxt))
	}

	return http.HandlerFunc(authFn)
}

func BuildJWTString(ctx context.Context, userID string) (string, error) {
	log := logger.LoggerFromContext(ctx)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenEXP)),
		},
		UserID: userID,
	})

	tokenString, err := token.SignedString([]byte(SecretKEY))
	if err != nil {
		log.Errorf("error tokenString in BuildJWTString()... ", err)
		return "", err
	}
	return tokenString, nil
}

func GetUserID(ctx context.Context, tokenString string) (string, error) {
	log := logger.LoggerFromContext(ctx)
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (interface{}, error) {
			return []byte(SecretKEY), nil
		})
	if err != nil {
		log.Errorf("error in GetUserID, ", err)
		return "", err
	}

	if !token.Valid {
		log.Errorf("no valid token ...", err)
		return "", err
	}

	return claims.UserID, err
}
