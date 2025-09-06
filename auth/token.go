package auth

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

type jwtPayload struct {
	Exp *int64 `json:"exp"`
}

var errInvalidToken = errors.New("invalid jwt format")

// 返回：是否已过期、过期时间（若有）、错误
func IsJWTExpired(token string) (bool, *time.Time, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return true, nil, errInvalidToken
	}

	// JWT 使用 base64url（无填充）
	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return true, nil, err
	}

	var pl jwtPayload
	if err := json.Unmarshal(payloadBytes, &pl); err != nil {
		return true, nil, err
	}

	// 没有 exp 就视为不受控或已过期（按你的安全策略决定）
	if pl.Exp == nil {
		return true, nil, errors.New("exp not present in jwt")
	}

	exp := time.Unix(*pl.Exp, 0)
	return false, &exp, nil
}
