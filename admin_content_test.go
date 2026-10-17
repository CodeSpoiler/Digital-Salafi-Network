package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"
)

func TestAdminContentManagement(t *testing.T) {
	path := filepath.Join(t.TempDir(), "forum.db")
	c, e := openCommunity(path)
	if e != nil {
		t.Fatal(e)
	}
	defer func() { c.db.Close() }()
	if e = c.bootstrapAdmin("owner", "owner-password-12345"); e != nil {
		t.Fatal(e)
	}
	f := &forumClient{c: c, cookies: map[string]*http.Cookie{}}
	f.request("GET", "/community/login", nil)
	f.request("POST", "/community/login", url.Values{"username": {"owner"}, "password": {"owner-password-12345"}})
	save := url.Values{"action": {"content-save"}, "slug": {"new-article"}, "kind": {"article"}, "topic": {"study"}, "title_en": {"New article"}, "heading_en": {"Section"}, "body_en": {"<script>unsafe</script>"}}
	if w := f.request("POST", "/community/admin", save); w.Code != 303 {
		t.Fatal(w.Body.String())
	}
	a, e := newApp("/community")
	if e != nil {
		t.Fatal(e)
	}
	a.community = c
	w := httptest.NewRecorder()
	a.ServeHTTP(w, httptest.NewRequest("GET", "/en/articles/new-article", nil))
	if w.Code != 200 || strings.Contains(w.Body.String(), "<script>unsafe</script>") || !strings.Contains(w.Body.String(), "New article") {
		t.Fatal("public content rendering", w.Code)
	}
	save.Set("source", "javascript:alert(1)")
	save.Set("original", "new-article")
	if w = f.request("POST", "/community/admin", save); w.Code != 400 {
		t.Fatal("unsafe URL accepted")
	}
	save.Set("source", "")
	save.Set("title_en", "Updated article")
	if w = f.request("POST", "/community/admin", save); w.Code != 303 {
		t.Fatal(w.Body.String())
	}
	for _, kind := range []string{"book", "scholar", "path"} {
		save.Set("original", "")
		save.Set("slug", "new-"+kind)
		save.Set("kind", kind)
		if w = f.request("POST", "/community/admin", save); w.Code != 303 {
			t.Fatal(w.Body.String())
		}
	}
	del := url.Values{"action": {"content-delete"}, "slug": {"new-article"}}
	if w = f.request("POST", "/community/admin", del); w.Code != 400 {
		t.Fatal("confirmation bypass")
	}
	del.Set("confirm", "yes")
	if w = f.request("POST", "/community/admin", del); w.Code != 303 {
		t.Fatal(w.Body.String())
	}
	// Deletion of a seeded record must survive restart, including when every record is deleted.
	if _, e = c.db.Exec("DELETE FROM library"); e != nil {
		t.Fatal(e)
	}
	c.db.Close()
	c, e = openCommunity(path)
	if e != nil {
		t.Fatal(e)
	}
	items, e := c.catalogItems()
	if e != nil || len(items) != 0 {
		t.Fatal("seed resurrected deleted content", e)
	}
}
func TestAdminDeletionAndPermissions(t *testing.T) {
	c, e := openCommunity(filepath.Join(t.TempDir(), "forum.db"))
	if e != nil {
		t.Fatal(e)
	}
	defer c.db.Close()
	c.bootstrapAdmin("owner", "owner-password-12345")
	member := &forumClient{c: c, cookies: map[string]*http.Cookie{}}
	member.request("GET", "/community/register", nil)
	member.request("POST", "/community/register", url.Values{"username": {"reader"}, "password": {"reader-password-12345"}})
	for _, action := range []string{"content-save", "content-delete", "delete-user", "delete-topic", "delete-comment"} {
		w := member.request("POST", "/community/admin", url.Values{"action": {action}, "confirm": {"yes"}, "id": {"1"}})
		if w.Code != 400 {
			t.Fatal("non-admin allowed", action, w.Code)
		}
	}
	for _, q := range []string{"INSERT INTO topics(id,category_id,member_id,title,body,created) VALUES(1,1,2,'Topic','Body','today')", "INSERT INTO replies(id,topic_id,member_id,body,created) VALUES(1,1,2,'Parent','today')", "INSERT INTO replies(id,topic_id,member_id,body,created,parent_id) VALUES(2,1,2,'Child','today',1)"} {
		if _, e = c.db.Exec(q); e != nil {
			t.Fatal(e)
		}
	}
	f := &forumClient{c: c, cookies: map[string]*http.Cookie{}}
	f.request("GET", "/community/login", nil)
	f.request("POST", "/community/login", url.Values{"username": {"owner"}, "password": {"owner-password-12345"}})
	post := func(action, id string) int {
		return f.request("POST", "/community/admin", url.Values{"action": {action}, "id": {id}, "confirm": {"yes"}}).Code
	}
	if post("delete-user", "1") != 400 {
		t.Fatal("admin deleted")
	}
	if post("delete-comment", "1") != 303 {
		t.Fatal("delete parent failed")
	}
	var n int
	if e = c.db.QueryRow("SELECT count(*) FROM replies WHERE id=2 AND parent_id IS NULL").Scan(&n); e != nil || n != 1 {
		t.Fatal("child lost", e)
	}
	if post("delete-user", "2") != 303 {
		t.Fatal("delete user failed")
	}
	c.db.QueryRow("SELECT count(*) FROM members WHERE name='reader'").Scan(&n)
	if n != 0 {
		t.Fatal("member still present")
	}
	c.db.QueryRow("SELECT count(*) FROM sessions WHERE member_id=2").Scan(&n)
	if n != 0 {
		t.Fatal("session survived")
	}
	c.db.QueryRow("SELECT count(*) FROM topics").Scan(&n)
	if n != 1 {
		t.Fatal("conversation lost")
	}
	if post("delete-topic", "1") != 303 {
		t.Fatal("delete topic failed")
	}
	c.db.QueryRow("SELECT count(*) FROM replies").Scan(&n)
	if n != 0 {
		t.Fatal("comments survived topic deletion")
	}
	rows, e := c.db.Query("PRAGMA foreign_key_check")
	if e != nil {
		t.Fatal(e)
	}
	defer rows.Close()
	if rows.Next() {
		t.Fatal("broken foreign key")
	}
}

func TestContentMutationBoundary(t *testing.T) {
	c, err := openCommunity(filepath.Join(t.TempDir(), "forum.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer c.db.Close()
	c.bootstrapAdmin("owner", "owner-password-12345")
	f := &forumClient{c: c, cookies: map[string]*http.Cookie{}}
	f.request("GET", "/community/login", nil)
	f.request("POST", "/community/login", url.Values{"username": {"owner"}, "password": {"owner-password-12345"}})
	for _, remote := range []bool{false, true} {
		v := url.Values{"action": {"content-delete"}, "slug": {"a-steady-study-routine"}, "confirm": {"yes"}}
		if remote {
			v.Set("csrf", f.cookies["sn_csrf"].Value)
		}
		r := httptest.NewRequest("POST", "http://localhost/community/admin", strings.NewReader(v.Encode()))
		r.RemoteAddr = "127.0.0.1:1234"
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		if remote {
			r.Header.Set("CF-Connecting-IP", "203.0.113.1")
		}
		for _, cookie := range f.cookies {
			r.AddCookie(cookie)
		}
		w := httptest.NewRecorder()
		c.ServeHTTP(w, r)
		if w.Code != 403 {
			t.Fatal("mutation boundary bypass", remote, w.Code)
		}
	}
	var n int
	c.db.QueryRow("SELECT count(*) FROM library WHERE slug='a-steady-study-routine'").Scan(&n)
	if n != 1 {
		t.Fatal("blocked mutation changed data")
	}
}
