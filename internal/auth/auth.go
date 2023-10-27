package auth

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

type Claims struct {
	jwt.RegisteredClaims
	UserID string
}

type KeyType string

const UserIDKey KeyType = "userID"
const TokenEXP = time.Hour * 3
const SecretKEY = "supersecretkey"
const CookieName = "auth"

func AuthMiddleware(next http.Handler) http.Handler {
	authFn := func(rw http.ResponseWriter, r *http.Request) {
		var userID string
		var tokenString string

		cookie, err := r.Cookie(CookieName)
		if err != nil {
			userID = uuid.New().String()
			tokenString, err = BuildJWTString(userID)
			if err != nil {
				fmt.Println("AuthMiddleware BuildJWTString err = ", err)
			}
		} else if userID, err = GetUserID(cookie.Value); err != nil {
			userID := uuid.New().String()
			tokenString, err = BuildJWTString(userID)
			if err != nil {
				fmt.Println("AuthMiddleware BuildJWTString err = ", err)
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

func BuildJWTString(userID string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenEXP)),
		},
		UserID: userID,
	})

	tokenString, err := token.SignedString([]byte(SecretKEY))
	if err != nil {
		fmt.Println("error tokenString in BuildJWTString()... ", err)
		return "", err
	}
	return tokenString, nil
}

func GetUserID(tokenString string) (string, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (interface{}, error) {
			return []byte(SecretKEY), nil
		})
	if err != nil {
		fmt.Println("error in GetUserID, ", err)
		return "", err
	}

	if !token.Valid {
		fmt.Println("no valid token ...", err)
		return "", err
	}

	return claims.UserID, err
}
