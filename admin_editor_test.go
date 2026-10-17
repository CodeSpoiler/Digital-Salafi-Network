package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"
)

func TestSimplePublishingLanguages(t *testing.T) {
	c, e := openCommunity(filepath.Join(t.TempDir(), "db"))
	if e != nil {
		t.Fatal(e)
	}
	defer c.db.Close()
	c.bootstrapAdmin("owner", "owner-password-12345")
	f := &forumClient{c: c, cookies: map[string]*http.Cookie{}}
	f.request("GET", "/community/login", nil)
	f.request("POST", "/community/login", url.Values{"username": {"owner"}, "password": {"owner-password-12345"}})
	v := url.Values{"action": {"content-save"}, "tab": {"articles"}, "kind": {"article"}, "topic": {"aqeedah"}, "language": {"ar"}, "title_ar": {"علم"}, "heading_en": {""}, "body_ar": {"محتوى عربي فقط"}}
	w := f.request("POST", "/community/admin?tab=articles&new=1", v)
	if w.Code != 303 {
		t.Fatal(w.Body.String())
	}
	items, _ := c.catalogItems()
	var item catalogItem
	for _, x := range items {
		if x.Title.AR == "علم" {
			item = x
		}
	}
	if item.Slug == "" || item.Title.EN != "" {
		t.Fatal("Arabic-only publication failed")
	}
	a, _ := newApp("/community")
	a.community = c
	for _, tc := range []struct {
		url  string
		code int
		show bool
	}{{"/ar/articles", 200, true}, {"/en/articles", 200, false}, {"/ar/articles/" + item.Slug, 200, true}, {"/en/articles/" + item.Slug, 404, false}} {
		out := httptest.NewRecorder()
		a.ServeHTTP(out, httptest.NewRequest("GET", tc.url, nil))
		if out.Code != tc.code || strings.Contains(out.Body.String(), "علم") != tc.show {
			t.Fatal(tc.url, out.Code)
		}
	}
	// A translation can be added later without changing the address or Arabic content.
	v.Set("original", item.Slug)
	v.Set("title_en", "Knowledge")
	v.Set("language", "both")
	if w = f.request("POST", "/community/admin", v); w.Code != 303 {
		t.Fatal(w.Body.String())
	}
	var raw string
	c.db.QueryRow("SELECT payload FROM library WHERE slug=?", item.Slug).Scan(&raw)
	json.Unmarshal([]byte(raw), &item)
	if item.Title.AR != "علم" || item.Title.EN != "Knowledge" {
		t.Fatal("translation lost")
	}
	// Failed validation keeps the writer's text in the editor.
	v.Set("source", "javascript:bad")
	w = f.request("POST", "/community/admin?tab=articles", v)
	if w.Code != 400 || !strings.Contains(w.Body.String(), "Knowledge") || !strings.Contains(w.Body.String(), "محتوى عربي فقط") {
		t.Fatal("form lost on error")
	}
	// Arabic-only steps and custom categories.
	v = url.Values{"action": {"content-topic"}, "name_ar": {"السيرة"}, "tab": {"categories"}}
	if w = f.request("POST", "/community/admin", v); w.Code != 303 {
		t.Fatal(w.Body.String())
	}
	v = url.Values{"action": {"content-save"}, "kind": {"path"}, "topic": {"aqeedah"}, "title_ar": {"مسار"}, "step_en": {""}, "step_ar": {"الخطوة الأولى"}, "tab": {"learning"}}
	if w = f.request("POST", "/community/admin", v); w.Code != 303 {
		t.Fatal(w.Body.String())
	}
	items, _ = c.catalogItems()
	found := false
	for _, x := range items {
		if x.Title.AR == "مسار" {
			found = len(x.Steps) == 1
		}
	}
	if !found {
		t.Fatal("Arabic step missing")
	}
	for _, tab := range []string{"articles", "learning", "members", "categories", "discussions", "comments"} {
		w = f.request("GET", "/community/admin?tab="+tab, nil)
		if w.Code != 200 {
			t.Fatal(tab, w.Body.String())
		}
		if tab != "articles" && strings.Contains(w.Body.String(), "Add article") {
			t.Fatal("unrelated section shown")
		}
	}
	w = f.request("GET", "/community/admin?tab=articles&new=1", nil)
	if strings.Contains(w.Body.String(), `name="slug"`) {
		t.Fatal("manual slug exposed")
	}
	out := httptest.NewRecorder()
	a.ServeHTTP(out, httptest.NewRequest("GET", "/ar/learn", nil))
	if !strings.Contains(out.Body.String(), "العقيدة") {
		t.Fatal("Aqeedah missing")
	}
}
