package shared

import (
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"strings"
	"time"
)

func IsTokenExpired(tokenString string) (bool, error) {
	claims, err := getTokenClaims(tokenString)
	if err != nil {
		return true, err
	}

	expiry := time.Unix(int64(claims["exp"].(float64)), 0)

	return time.Now().After(expiry), nil
}

func GetTokenEmail(tokenString string) (string, error) {
	claims, err := getTokenClaims(tokenString)
	if err != nil {
		return "", err
	}

	return claims["email"].(string), nil
}

func IsTokenValid(tokenString string, envName string) (bool, error) {
	claims, err := getTokenClaims(tokenString)
	if err != nil {
		return false, err
	}

	return strings.Contains(claims["iss"].(string), envName), nil
}

func getTokenClaims(tokenString string) (jwt.MapClaims, error) {
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token, err: %v", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("can't convert token's claims to standard claims")
	}

	return claims, nil
}
