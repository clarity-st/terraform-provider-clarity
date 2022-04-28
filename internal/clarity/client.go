package clarity

import (
	"net/http"
)

type Client struct {
	Host      string
	Token     string
	UserAgent string
	Client    *http.Client
}

type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
