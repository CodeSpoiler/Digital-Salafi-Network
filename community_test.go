package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"
)

type forumClient struct {
	c       *Community
	cookies map[string]*http.Cookie
}

func (f *forumClient) request(method, path string, values url.Values) *httptest.ResponseRecorder {
	if values == nil {
		values = url.Values{}
	}
	if cookie := f.cookies["sn_csrf"]; cookie != nil && method == "POST" {
		values.Set("csrf", cookie.Value)
	}
	r := httptest.NewRequest(method, "https://community.example.test"+path, strings.NewReader(values.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	for _, cookie := range f.cookies {
		r.AddCookie(cookie)
	}
	w := httptest.NewRecorder()
	f.c.ServeHTTP(w, r)
	for _, cookie := range w.Result().Cookies() {
		if cookie.MaxAge < 0 {
			delete(f.cookies, cookie.Name)
		} else {
			f.cookies[cookie.Name] = cookie
		}
	}
	return w
}
func TestCommunityFlowAndPersistence(t *testing.T) {
	path := filepath.Join(t.TempDir(), "community.db")
	c, err := openCommunity(path)
	if err != nil {
		t.Fatal(err)
	}
	if err = c.bootstrapAdmin("owner", "a-strong-admin-password"); err != nil {
		t.Fatal(err)
	}
	f := &forumClient{c: c, cookies: map[string]*http.Cookie{}}
	if w := f.request("GET", "/community/register", nil); w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	if w := f.request("POST", "/community/register", url.Values{"username": {"reader"}, "password": {"a-strong-reader-password"}}); w.Code != 303 {
		t.Fatal(w.Body.String())
	}
	session := f.cookies["sn_session"]
	if session == nil || !session.HttpOnly || !session.Secure {
		t.Fatal("missing secure session")
	}
	w := f.request("POST", "/community/new", url.Values{"title": {"A useful first discussion"}, "body": {"Hello <script>alert(1)</script>"}, "category": {"1"}})
	if w.Code != 303 {
		t.Fatal(w.Body.String())
	}
	topicURL := w.Header().Get("Location")
	if w = f.request("POST", topicURL, url.Values{"body": {"A thoughtful reply."}}); w.Code != 303 {
		t.Fatal(w.Body.String())
	}
	if w = f.request("GET", topicURL, nil); !strings.Contains(w.Body.String(), "A thoughtful reply.") || strings.Contains(w.Body.String(), "<script>alert") {
		t.Fatal("reply missing or unescaped content")
	}
	if w = f.request("POST", "/community/admin", url.Values{"name": {"Private category"}}); w.Code == 303 {
		t.Fatal("member created category")
	}
	if w = f.request("POST", topicURL, url.Values{"action": {"hide"}}); w.Code == 303 {
		t.Fatal("member moderated topic")
	}
	c.db.Close()
	c, err = openCommunity(path)
	if err != nil {
		t.Fatal(err)
	}
	defer c.db.Close()
	f.c = c
	if w = f.request("GET", topicURL, nil); w.Code != 200 || !strings.Contains(w.Body.String(), "A thoughtful reply.") {
		t.Fatal("topic not persisted")
	}
	f.request("POST", "/community/logout", nil)
	if w = f.request("POST", topicURL, url.Values{"body": {"Anonymous reply"}}); w.Code == 303 {
		t.Fatal("anonymous write accepted")
	}
	if w = f.request("POST", "/community/login", url.Values{"username": {"owner"}, "password": {"a-strong-admin-password"}}); w.Code != 303 {
		t.Fatal(w.Body.String())
	}
	if w = f.request("POST", "/community/admin", url.Values{"name": {"Books and learning"}, "description": {"Discuss reading."}}); w.Code != 303 {
		t.Fatal(w.Body.String())
	}
	if w = f.request("POST", topicURL, url.Values{"action": {"lock"}}); w.Code != 303 {
		t.Fatal(w.Body.String())
	}
	if w = f.request("POST", topicURL, url.Values{"body": {"Reply to closed topic"}}); w.Code == 303 {
		t.Fatal("closed topic accepted reply")
	}
	if w = f.request("POST", topicURL, url.Values{"action": {"hide"}}); w.Code != 303 {
		t.Fatal(w.Body.String())
	}
	if w = f.request("GET", topicURL, nil); w.Code != 404 {
		t.Fatal("hidden topic visible")
	}
}
func TestCommunityCSRFAndInvalidCredentials(t *testing.T) {
	c, err := openCommunity(filepath.Join(t.TempDir(), "db"))
	if err != nil {
		t.Fatal(err)
	}
	defer c.db.Close()
	r := httptest.NewRequest("POST", "/community/register", strings.NewReader("username=attacker&password=some-long-password"))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	c.ServeHTTP(w, r)
	if w.Code != 403 {
		t.Fatal("missing CSRF accepted")
	}
	f := &forumClient{c: c, cookies: map[string]*http.Cookie{}}
	f.request("GET", "/community/register", nil)
	w = f.request("POST", "/community/register", url.Values{"username": {"reader"}, "password": {"short"}})
	if w.Code == 303 {
		t.Fatal("short password accepted")
	}
	f.request("POST", "/community/register", url.Values{"username": {"reader"}, "password": {"a-strong-reader-password"}})
	f.request("POST", "/community/logout", nil)
	w = f.request("POST", "/community/login", url.Values{"username": {"reader"}, "password": {"incorrect-password"}})
	if w.Code == 303 || f.cookies["sn_session"] != nil {
		t.Fatal("bad password authenticated")
	}
	w = f.request("POST", "/community/register", url.Values{"username": {"READER"}, "password": {"another-long-password"}})
	if w.Code == 303 {
		t.Fatal("duplicate username accepted")
	}
	var hash string
	c.db.QueryRow("SELECT password FROM members WHERE name='reader'").Scan(&hash)
	if hash == "a-strong-reader-password" || !strings.HasPrefix(hash, "$2") {
		t.Fatal("password not hashed")
	}
}

func TestCommentReplies(t *testing.T) {
	c, err := openCommunity(filepath.Join(t.TempDir(), "db"))
	if err != nil {
		t.Fatal(err)
	}
	defer c.db.Close()
	f := &forumClient{c: c, cookies: map[string]*http.Cookie{}}
	f.request("GET", "/community/register", nil)
	f.request("POST", "/community/register", url.Values{"username": {"reader"}, "password": {"a-strong-reader-password"}})
	w := f.request("POST", "/community/new", url.Values{"title": {"First thread"}, "body": {"Original message"}, "category": {"1"}})
	path := w.Header().Get("Location")
	f.request("POST", path, url.Values{"body": {"First comment"}})
	w = f.request("POST", path, url.Values{"body": {"A direct response"}, "parent_id": {"1"}})
	if w.Code != 303 || !strings.HasSuffix(w.Header().Get("Location"), "#reply-2") {
		t.Fatal(w.Body.String())
	}
	var parent int
	c.db.QueryRow("SELECT parent_id FROM replies WHERE id=2").Scan(&parent)
	if parent != 1 {
		t.Fatal("parent not saved")
	}
	w = f.request("GET", path+"?reply_to=2", nil)
	for _, want := range []string{"ORIGINAL POST", "comment-response", "Replying to @reader", `name="parent_id" value="2"`} {
		if !strings.Contains(w.Body.String(), want) {
			t.Fatalf("missing %s", want)
		}
	}
	// Interleave another root, a grandchild, and a sibling: display must follow ancestry, not creation order.
	for _, values := range []url.Values{
		{"body": {"Second root"}},
		{"body": {"Grandchild"}, "parent_id": {"2"}},
		{"body": {"Sibling response"}, "parent_id": {"1"}},
	} {
		if result := f.request("POST", path, values); result.Code != 303 {
			t.Fatal(result.Body.String())
		}
	}
	page := &communityPage{View: strings.TrimPrefix(path, "/community")}
	if err := c.load(httptest.NewRequest("GET", path, nil), page); err != nil {
		t.Fatal(err)
	}
	if len(page.CommentTree) != 2 || len(page.CommentTree[0].Children) != 2 || len(page.CommentTree[0].Children[0].Children) != 1 {
		t.Fatal("incorrect reply tree")
	}
	if page.CommentTree[0].Children[0].Children[0].ID != 4 || page.CommentTree[1].ID != 3 {
		t.Fatal("incorrect parent association")
	}
	rendered := f.request("GET", path, nil).Body.String()
	previous := -1
	for _, marker := range []string{`id="reply-1"`, `id="reply-2"`, `id="reply-4"`, `id="reply-5"`, `id="reply-3"`} {
		position := strings.Index(rendered, marker)
		if position <= previous {
			t.Fatal("comments not grouped under parents")
		}
		previous = position
	}
	if !strings.Contains(rendered, `class="comment-children"`) {
		t.Fatal("missing nested containers")
	}
	w = f.request("POST", "/community/new", url.Values{"title": {"Other thread"}, "body": {"Another message"}, "category": {"1"}})
	other := w.Header().Get("Location")
	for _, id := range []string{"1", "999", "-1", "bad"} {
		if w = f.request("POST", other, url.Values{"body": {"Invalid response"}, "parent_id": {id}}); w.Code != 400 {
			t.Fatalf("accepted invalid parent %s", id)
		}
	}
}
