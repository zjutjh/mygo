package oauth

import (
	"os"
	"testing"

	"github.com/go-resty/resty/v2"
)

const (
	testUsernameEnv = "OAUTH_USERNAME"
	testPasswordEnv = "OAUTH_PASSWORD"
)

func testCredentials(t *testing.T) (string, string) {
	t.Helper()

	username := os.Getenv(testUsernameEnv)
	password := os.Getenv(testPasswordEnv)
	if username == "" || password == "" {
		t.Skipf("未设置%s或%s，跳过OAuth集成测试", testUsernameEnv, testPasswordEnv)
	}

	return username, password
}

func newTestOAuth() *OAuth {
	return &OAuth{
		conf:   DefaultConfig,
		client: resty.New(),
	}
}

func TestLogin(t *testing.T) {
	username, password := testCredentials(t)

	cookies, err := newTestOAuth().Login(username, password)
	if err != nil {
		t.Fatalf("OAuth登录失败: %v", err)
	}
	if len(cookies) == 0 {
		t.Fatal("OAuth登录成功但未获取到Cookie")
	}
}

func TestGetUserInfo(t *testing.T) {
	username, password := testCredentials(t)

	cookies, userInfo, err := newTestOAuth().GetUserInfo(username, password)
	if err != nil {
		t.Fatalf("获取OAuth用户信息失败: %v", err)
	}
	if len(cookies) == 0 {
		t.Fatal("获取OAuth用户信息成功但未获取到Cookie")
	}
	if userInfo.StudentID == "" {
		t.Fatal("OAuth用户信息中学号为空")
	}
}
