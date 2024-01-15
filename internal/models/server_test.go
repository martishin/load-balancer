package models

import (
	"net/http/httputil"
	"net/url"
	"testing"
)

func TestServer_IsAlive(t *testing.T) {
	tests := []struct {
		name  string
		alive bool
		want  bool
	}{
		{"Server initially dead", false, false},
		{"Server set to alive", true, true},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				serverURL, _ := url.Parse("http://example.com")
				server := &Server{
					URL:   serverURL,
					Alive: tt.alive,
					Proxy: httputil.NewSingleHostReverseProxy(serverURL),
				}

				if got := server.IsAlive(); got != tt.want {
					t.Errorf("Server.IsAlive() = %v, want %v", got, tt.want)
				}
			},
		)
	}
}
func TestServer_SetAlive(t *testing.T) {
	tests := []struct {
		name     string
		setAlive bool
		want     bool
	}{
		{"Set server to alive", true, true},
		{"Set server to not alive", false, false},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				serverURL, _ := url.Parse("http://example.com")
				server := &Server{
					URL:   serverURL,
					Alive: false,
					Proxy: httputil.NewSingleHostReverseProxy(serverURL),
				}

				server.SetAlive(tt.setAlive)
				if got := server.IsAlive(); got != tt.want {
					t.Errorf("After SetAlive(%v), Server.IsAlive() = %v, want %v", tt.setAlive, got, tt.want)
				}
			},
		)
	}
}
