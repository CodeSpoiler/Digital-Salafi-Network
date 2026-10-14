package main

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLandingAndProgress(t *testing.T) {
	t.Setenv("FORUM_URL", "http://localhost")
	a, err := newApp("")
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"/en/", "/ar/", "/en/articles", "/en/learn", "/ar/articles", "/ar/learn", "/en/about"} {
		w := httptest.NewRecorder()
		a.ServeHTTP(w, httptest.NewRequest("GET", path, nil))
		if w.Code != 200 {
			t.Fatalf("%s: %d", path, w.Code)
		}
		if (path == "/en/articles" || path == "/en/learn") && !strings.Contains(w.Body.String(), "Still in progress.") {
			t.Fatal("Missing progress notice")
		}
	}
}
func TestForumRedirectsAndRetiredRoutes(t *testing.T) {
	t.Setenv("FORUM_URL", "http://localhost")
	a, _ := newApp("")
	cases := map[string]string{"/en/forum": "http://localhost/", "/ar/forum": "http://localhost/", "/en/forum/old-topic": "http://localhost/", "/en/search?q=books+%26+learning": "http://localhost?q=books+%26+learning", "/en/articles/welcome": "/en/articles", "/en/learn/books": "/en/learn"}
	for path, want := range cases {
		w := httptest.NewRecorder()
		a.ServeHTTP(w, httptest.NewRequest("GET", path, nil))
		if w.Code != 302 || w.Header().Get("Location") != want {
			t.Fatalf("%s: %d %s", path, w.Code, w.Header().Get("Location"))
		}
	}
	w := httptest.NewRecorder()
	a.ServeHTTP(w, httptest.NewRequest("POST", "/en/forum", strings.NewReader("title=old")))
	if w.Code != 405 {
		t.Fatal("Old forum writes must be disabled")
	}
}
func TestForumConfiguration(t *testing.T) {
	t.Setenv("FORUM_URL", "https://forum.example.com")
	a, err := newApp("")
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	a.ServeHTTP(w, httptest.NewRequest("GET", "/en/forum", nil))
	if w.Header().Get("Location") != "https://forum.example.com/" {
		t.Fatal("Configured destination ignored")
	}
	t.Setenv("FORUM_URL", "javascript:alert(1)")
	if _, err := newApp(""); err == nil {
		t.Fatal("Invalid forum URL accepted")
	}
}
