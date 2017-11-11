package common

import (
	"errors"
	"net/http"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/dgrijalva/jwt-go/request"
)

func AuthJwt(signingKey string, req *http.Request) error {
	if signingKey == "" {
		return nil
	}

	extractor := request.AuthorizationHeaderExtractor
	tokenString, err := extractor.ExtractToken(req)
	if err != nil {
		return err
	}

	token, err := jwt.ParseWithClaims(tokenString, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(signingKey), nil
	})

	if err != nil {
		return err
	}

	if _, ok := token.Claims.(*jwt.StandardClaims); ok && token.Valid {
		return nil
	}

	return errors.New("Invalid token")

}

func GetJwt(signingKey string, expiration int) (string, error) {
	now := time.Now().Unix()
	claims := &jwt.StandardClaims{
		ExpiresAt: time.Unix(now, 0).Add(time.Duration(expiration) * time.Second).Unix(),
		IssuedAt:  now,
		NotBefore: time.Unix(now, 0).Add(time.Duration(-1) * time.Minute).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString([]byte(signingKey))
	return ss, err
}
