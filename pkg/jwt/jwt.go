package jwt

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/dgrijalva/jwt-go"
)

const (
	TokenExpireDuration        = time.Hour * 2
	RefreshTokenExpireDuration = time.Hour * 24 * 14
	TokenTypeAccess            = "access"
	TokenTypeRefresh           = "refresh"
)

var mySecret = []byte("hyanhsing0101")

type MyClaims struct {
	UserID    int64  `json:"user_id"`
	Username  string `json:"username"`
	TokenType string `json:"token_type"`
	jwt.StandardClaims
}

func GenAccessToken(userID int64, username string) (string, error) {
	c := MyClaims{
		UserID:   userID,
		Username: username,
		TokenType: TokenTypeAccess,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(TokenExpireDuration).Unix(),
			Issuer:    "hyanhsing0101",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	return token.SignedString(mySecret)
}

func GenRefreshToken(userID int64, username string) (string, string, error) {
	jti, err := genJTI()
	if err != nil {
		return "", "", err
	}
	c := MyClaims{
		UserID:   userID,
		Username: username,
		TokenType: TokenTypeRefresh,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(RefreshTokenExpireDuration).Unix(),
			Issuer:    "hyanhsing0101",
			Id:        jti,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	signedToken, err := token.SignedString(mySecret)
	if err != nil {
		return "", "", err
	}
	return signedToken, jti, nil
}

func ParseToken(tokenString string) (*MyClaims, error) {
	var mc = new(MyClaims)
	token, err := jwt.ParseWithClaims(tokenString, mc, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("unexpected signing method")
		}
		return mySecret, nil
	})
	if err != nil {
		return nil, err
	}
	if token.Valid {
		return mc, nil
	}
	return nil, errors.New("invalid token")
}

func genJTI() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}