package auth

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var secretKey []byte

func InitJWT() error {
	secret := os.Getenv("JWT_SECRET")

	if secret == "" {
		return errors.New("JWT_SECRET environment variable is not set")
	}

	secretKey = []byte(secret)

	return nil
}

type Payload struct {
	Id    int64
	Email string
	Role  string
	Type  string
	jwt.RegisteredClaims
}

func ParseToken(tokenString string) (*Payload, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&Payload{},
		func(t *jwt.Token) (any, error) {

			if t.Method != jwt.SigningMethodHS256 {
				return nil, errors.New("invalid signing method")
			}

			return secretKey, nil
		},
	)

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Payload)

	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

const (
	AccessToken  = "access"
	RefreshToken = "refresh"
)

// CreateAccessToken creates an access token that expires after 5 minutes.
func CreateAccessToken(id int64, email string, role string) (string, error) {
	return createToken(
		id,
		email,
		role,
		AccessToken,
		5*time.Minute,
	)
}

func CreateRefreshToken(id int64, email string, role string) (string, error) {
	return createToken(
		id,
		email,
		role,
		RefreshToken,
		7*24*time.Hour,
	)
}

func createToken(
	id int64,
	email string,
	role string,
	tokenType string,
	duration time.Duration,
) (string, error) {

	now := time.Now()

	claims := Payload{
		Id:    id,
		Email: email,
		Role:  role,
		Type:  tokenType,

		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(duration)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		claims,
	)

	return token.SignedString(secretKey)
}

// GetAccessTokenFromRefreshToken validates a refresh token and creates a new access token.
func GetAccessTokenFromRefreshToken(refreshToken string) (string, error) {

	claims, err := ParseToken(refreshToken)

	if err != nil {
		return "", err
	}

	// is actually a refresh token or not .
	if claims.Type != RefreshToken {
		return "", errors.New("token is not a refresh token")
	}

	return CreateAccessToken(
		claims.Id,
		claims.Email,
		claims.Role,
	)

}
