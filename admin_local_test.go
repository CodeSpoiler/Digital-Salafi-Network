package main

import (
	"net/http/httptest"
	"path/filepath"
	"testing"
)

func TestLocalAdminBoundary(t *testing.T) {
	c, err := openCommunity(filepath.Join(t.TempDir(), "db"))
	if err != nil {
		t.Fatal(err)
	}
	defer c.db.Close()
	for _, tc := range []struct {
		host, peer, header string
		allowed            bool
	}{
		{"localhost:8095", "127.0.0.1:1234", "", true}, {"[::1]:8095", "[::1]:1234", "", true},
		{"salafinetwork.com", "127.0.0.1:1234", "", false}, {"localhost:8095", "192.168.1.5:1234", "", false},
		{"localhost:8095", "127.0.0.1:1234", "CF-Connecting-IP", false}, {"localhost:8095", "127.0.0.1:1234", "X-Forwarded-For", false},
		{"localhost:8095", "127.0.0.1:1234", "Forwarded", false},
	} {
		r := httptest.NewRequest("GET", "http://"+tc.host+"/community/admin", nil)
		r.RemoteAddr = tc.peer
		if tc.header != "" {
			r.Header.Set(tc.header, "203.0.113.2")
		}
		if localAdminRequest(r) != tc.allowed {
			t.Fatalf("boundary mismatch: %+v", tc)
		}
		if !tc.allowed {
			w := httptest.NewRecorder()
			c.ServeHTTP(w, r)
			if w.Code != 403 {
				t.Fatalf("expected blocked dashboard, got %d", w.Code)
			}
		}
	}
}
