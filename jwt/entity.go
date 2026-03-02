package jwt

import "github.com/golang-jwt/jwt/v5"

type CustomClaims[T any] struct {
	Identity T `json:"identity"`

	jwt.RegisteredClaims
}
