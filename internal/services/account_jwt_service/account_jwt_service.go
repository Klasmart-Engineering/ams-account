package account_jwt_service

import (
	"errors"
	"fmt"
	"os"
	"time"

	"bitbucket.org/calmisland/account-lambda-funcs/internal/globals"
	jwt "github.com/dgrijalva/jwt-go"
)

type TokenMapClaims struct {
	Email            string `json:"email"`
	PhoneNumber      string `json:"phoneNr"`
	Password         string `json:"pw"`
	VerificationCode string `json:"verificationCode"`
	Language         string `json:"lang"`
	ExpireAt         int64  `json:"expireAt"`
}

func (token *TokenMapClaims) Valid() error {
	if token.VerificationCode == "" {
		return errors.New("Claim does not contain verificationCode")
	}
	if token.ExpireAt < time.Now().Unix() {
		return errors.New("Token is expired.")
	}
	return nil
}

func EncryptHashedCode(code string) string {
	hashedPassword, _ := globals.PasswordHasher.GeneratePasswordHash(code, false)
	return hashedPassword
}

func GetSecret() []byte {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return []byte("C@almIsl@nd")
	}
	return []byte(jwtSecret)
}

func CreateToken(claims *TokenMapClaims) (string, error) {

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email":            claims.Email,
		"phoneNr":          claims.PhoneNumber,
		"pw":               claims.Password,
		"verificationCode": EncryptHashedCode(claims.VerificationCode),
		"expireAt":         time.Now().Add(time.Minute * 10).Unix(),
	})
	secret := GetSecret()
	tokenString, err := token.SignedString(secret)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func VerifyToken(tokenString string) (*TokenMapClaims, error) {
	// Parse the token
	token, err := jwt.ParseWithClaims(tokenString, &TokenMapClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return GetSecret(), nil
	})

	if err != nil {
		return nil, err
	}

	claims := token.Claims.(*TokenMapClaims)
	return claims, nil
}
