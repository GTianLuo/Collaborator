package jwts

import (
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"log"
	"time"
)

type JwtToken struct {
	AccessToken  string
	RefreshToken string
	AccessExp    time.Duration
}

func CreateToken(str string, secret string, t time.Duration, rf time.Duration) *JwtToken {
	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"tokenKey": str,
		"exp":      time.Now().Add(t).Unix(),
	})
	token, err := claims.SignedString([]byte(secret))
	if err != nil {
		log.Println(err)
	}
	refreshClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"key": str,
		"exp": time.Now().Add(rf).Unix(),
	})
	refreshToken, err := refreshClaims.SignedString([]byte(secret))
	if err != nil {
		log.Println(err)
	}
	return &JwtToken{
		AccessToken:  token,
		RefreshToken: refreshToken,
		AccessExp:    t,
	}
}
func ParseToken(tokenString string, secret string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return "", err
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		v := claims["tokenKey"].(string)
		exp := int64(claims["exp"].(float64))
		if exp <= time.Now().Unix() {
			return "", errors.New("token过期")
		}
		return v, nil
	} else {
		return "", err
	}
}
