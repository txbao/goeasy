package jwt

import (
	"errors"
	"time"

	jwtv5 "github.com/golang-jwt/jwt/v5"

	"github.com/txbao/goeasy/config"
)

// Token JWT 签发与校验（无 User/Role 业务模型）。
type Token struct {
	secret []byte
	issuer string
	expire time.Duration
}

type Claims struct {
	Subject string `json:"sub"`
	jwtv5.RegisteredClaims
}

func New(cfg config.JWTCfg) (*Token, error) {
	if !cfg.Enabled {
		return nil, nil
	}
	if cfg.Secret == "" {
		return nil, errors.New("jwt secret is required when enabled")
	}
	exp := time.Duration(cfg.ExpireMin) * time.Minute
	if exp <= 0 {
		exp = time.Hour
	}
	return &Token{
		secret: []byte(cfg.Secret),
		issuer: cfg.Issuer,
		expire: exp,
	}, nil
}

func (t *Token) Generate(subject string) (string, error) {
	if t == nil {
		return "", errors.New("jwt disabled")
	}
	now := time.Now()
	claims := Claims{
		Subject: subject,
		RegisteredClaims: jwtv5.RegisteredClaims{
			Issuer:    t.issuer,
			Subject:   subject,
			ExpiresAt: jwtv5.NewNumericDate(now.Add(t.expire)),
			IssuedAt:  jwtv5.NewNumericDate(now),
		},
	}
	token := jwtv5.NewWithClaims(jwtv5.SigningMethodHS256, claims)
	return token.SignedString(t.secret)
}

func (t *Token) Parse(tokenStr string) (*Claims, error) {
	if t == nil {
		return nil, errors.New("jwt disabled")
	}
	token, err := jwtv5.ParseWithClaims(tokenStr, &Claims{}, func(token *jwtv5.Token) (any, error) {
		return t.secret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}
