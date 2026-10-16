package main

import (
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"testing"
)

func TestProfilesAndAdminPermissions(t *testing.T) {
	c, err := openCommunity(filepath.Join(t.TempDir(), "forum.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer c.db.Close()
	if err = c.bootstrapAdmin("owner", "owner-password-12345"); err != nil {
		t.Fatal(err)
	}
	f := &forumClient{c: c, cookies: map[string]*http.Cookie{}}
	f.request("GET", "/community/register", nil)
	if w := f.request("POST", "/community/register", url.Values{"username": {"reader"}, "password": {"reader-password-12345"}}); w.Code != 303 {
		t.Fatal(w.Body.String())
	}
	if w := f.request("GET", "/community/admin", nil); w.Code != 403 {
		t.Fatal("member saw admin page")
	}
	if w := f.request("POST", "/community/u/owner", url.Values{"bio": {"changed"}}); w.Code != 400 {
		t.Fatal("edited another user")
	}
	if w := f.request("POST", "/community/u/reader", url.Values{"display_name": {"A Reader"}, "bio": {"<script>unsafe</script>"}, "interests": {"Arabic"}, "goals": {"Read daily"}}); w.Code != 303 {
		t.Fatal(w.Body.String())
	}
	w := f.request("GET", "/community/u/reader", nil)
	if w.Code != 200 || !strings.Contains(w.Body.String(), "A Reader") || strings.Contains(w.Body.String(), "<script>unsafe</script>") {
		t.Fatal("profile rendering or escaping failed")
	}
	if w = f.request("POST", "/community/new", url.Values{"title": {"A visible topic"}, "body": {"Useful discussion"}, "category": {"1"}}); w.Code != 303 {
		t.Fatal(w.Body.String())
	}
	if w = f.request("POST", "/community/admin", url.Values{"action": {"hide"}, "id": {"1"}}); w.Code != 400 {
		t.Fatal("non-admin moderated")
	}
	f.request("POST", "/community/logout", nil)
	f.request("GET", "/community/login", nil)
	if w = f.request("POST", "/community/login", url.Values{"username": {"owner"}, "password": {"owner-password-12345"}}); w.Code != 303 {
		t.Fatal(w.Body.String())
	}
	if w = f.request("GET", "/community/admin", nil); w.Code != 200 || !strings.Contains(w.Body.String(), "Admin dashboard") {
		t.Fatal(w.Body.String())
	}
	if w = f.request("POST", "/community/admin", url.Values{"action": {"hide"}, "id": {"1"}}); w.Code != 303 {
		t.Fatal(w.Body.String())
	}
	if w = f.request("GET", "/community/u/reader", nil); strings.Contains(w.Body.String(), "A visible topic") {
		t.Fatal("hidden topic leaked into profile")
	}
	if w = f.request("POST", "/community/admin", url.Values{"action": {"restore"}, "id": {"1"}}); w.Code != 303 {
		t.Fatal(w.Body.String())
	}
	if w = f.request("GET", "/community/u/reader?tab=edit", nil); w.Code != 404 {
		t.Fatal("owner edit form exposed")
	}
	if w = f.request("POST", "/community/admin", url.Values{"action": {"category-edit"}, "id": {"1"}, "name": {"Useful discussion"}, "description": {"Updated"}}); w.Code != 303 {
		t.Fatal(w.Body.String())
	}
	for _, path := range []string{"/community/u/reader?tab=comments", "/community/u/reader?tab=about"} {
		if w = f.request("GET", path, nil); w.Code != 200 {
			t.Fatal(w.Body.String())
		}
	}
}
