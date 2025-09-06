package auth

import (
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/litongjava/hfile/config"
	"testing"
)

func TestIsJWTExpired(t *testing.T) {
	config.GetServerUrl(".")

	token, _, err := config.LoadToken()
	if err != nil {
		panic(err)
	}

	isExpired, _, err := IsJWTExpired(token)
	if err != nil {
		panic(err)
	}
	hlog.Info(isExpired)
}
