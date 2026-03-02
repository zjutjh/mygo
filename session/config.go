package session

import (
	"net/http"
)

const (
	DriverRedis  = "redis"
	DriverMemory = "memory"
)

var DefaultConfig = Config{
	Driver: "memory",
	Name:   "session",
	Secret: "secret",
	Redis:  "",

	Path:     "/",
	Domain:   "",
	MaxAge:   86400 * 7,
	Secure:   false,
	HttpOnly: false,
	SameSite: 0,
}

type Config struct {
	Driver string `mapstructure:"driver"`
	Name   string `mapstructure:"name"`
	Secret string `mapstructure:"secret"`
	Redis  string `mapstructure:"redis"`

	// Cookie 相关选项
	Path     string        `mapstructure:"path"`
	Domain   string        `mapstructure:"domain"`
	MaxAge   int           `mapstructure:"max_age"`
	Secure   bool          `mapstructure:"secure"`
	HttpOnly bool          `mapstructure:"http_only"`
	SameSite http.SameSite `mapstructure:"same_site"`
}
