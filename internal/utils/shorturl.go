package utils

import (
	"crypto/sha256"
	"encoding/base64"
	"regexp"
)

func GenerateShortURL(url string) string {
	hash := sha256.Sum256([]byte(url))
	shortURL := base64.URLEncoding.EncodeToString(hash[:])
	regExp := regexp.MustCompile("[^a-zA-Z0-9]+")
	shortURL = regExp.ReplaceAllString(shortURL, "")
	return shortURL[:8]
}
