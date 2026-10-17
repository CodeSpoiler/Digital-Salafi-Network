package main

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"strings"
)

type memberProfile struct {
	ID                                               int64
	Name, DisplayName, Bio, Interests, Goals, Joined string
	Admin                                            bool
	Discussions, Comments                            int
}
type adminTopic struct {
	ID                      int64
	Title, Author, Category string
	Locked, Hidden          bool
}
type profileComment struct {
	ID, TopicID          int64
	Title, Body, Created string
}

const profileSelect = `SELECT m.id,m.name,m.admin,COALESCE(p.display_name,''),COALESCE(p.bio,''),COALESCE(p.interests,''),COALESCE(p.goals,''),COALESCE(p.joined,''),(SELECT count(*) FROM topics WHERE member_id=m.id AND hidden=0),(SELECT count(*) FROM replies r JOIN topics t ON t.id=r.topic_id WHERE r.member_id=m.id AND t.hidden=0) FROM members m LEFT JOIN profiles p ON p.member_id=m.id `

func scanProfile(row interface{ Scan(...any) error }, p *memberProfile) error {
	err := row.Scan(&p.ID, &p.Name, &p.Admin, &p.DisplayName, &p.Bio, &p.Interests, &p.Goals, &p.Joined, &p.Discussions, &p.Comments)
	if p.DisplayName == "" {
		p.DisplayName = p.Name
	}
	return err
}
func (c *Community) loadProfile(r *http.Request, p *communityPage) error {
	name := strings.TrimPrefix(p.View, "/u/")
	if err := scanProfile(c.db.QueryRow(profileSelect+"WHERE m.name=?", name), &p.Profile); err != nil {
		return err
	}
	p.Tab = r.URL.Query().Get("tab")
	if p.Tab != "comments" && p.Tab != "about" && p.Tab != "edit" {
		p.Tab = "discussions"
	}
	if p.Tab == "edit" && (p.User == nil || p.User.ID != p.Profile.ID) {
		return sql.ErrNoRows
	}
	rows, err := c.db.Query(`SELECT t.id,t.title,m.name,c.name,t.created,(SELECT count(*) FROM replies WHERE topic_id=t.id),t.locked FROM topics t JOIN members m ON m.id=t.member_id JOIN categories c ON c.id=t.category_id WHERE t.hidden=0 AND t.member_id=? ORDER BY t.id DESC LIMIT 100`, p.Profile.ID)
	if err != nil {
		return err
	}
	for rows.Next() {
		var t topic
		if err = rows.Scan(&t.ID, &t.Title, &t.Author, &t.Category, &t.Created, &t.Replies, &t.Locked); err != nil {
			rows.Close()
			return err
		}
		p.Topics = append(p.Topics, t)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return err
	}
	rows, err = c.db.Query(`SELECT r.id,t.id,t.title,r.body,r.created FROM replies r JOIN topics t ON t.id=r.topic_id WHERE r.member_id=? AND t.hidden=0 ORDER BY r.id DESC LIMIT 100`, p.Profile.ID)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var x profileComment
		if err = rows.Scan(&x.ID, &x.TopicID, &x.Title, &x.Body, &x.Created); err != nil {
			return err
		}
		p.ProfileComments = append(p.ProfileComments, x)
	}
	return rows.Err()
}
func (c *Community) saveProfile(w http.ResponseWriter, r *http.Request, p *communityPage) error {
	if p.User == nil || !strings.EqualFold(strings.TrimPrefix(p.View, "/u/"), p.User.Name) {
		return errors.New("You can only edit your own profile.")
	}
	name, bio, interests, goals := strings.TrimSpace(r.PostForm.Get("display_name")), strings.TrimSpace(r.PostForm.Get("bio")), strings.TrimSpace(r.PostForm.Get("interests")), strings.TrimSpace(r.PostForm.Get("goals"))
	if len([]rune(name)) > 60 || len([]rune(bio)) > 1200 || len([]rune(interests)) > 300 || len([]rune(goals)) > 600 {
		return errors.New("Please keep your profile within the field limits.")
	}
	_, err := c.db.Exec(`INSERT INTO profiles(member_id,display_name,bio,interests,goals) VALUES(?,?,?,?,?) ON CONFLICT(member_id) DO UPDATE SET display_name=excluded.display_name,bio=excluded.bio,interests=excluded.interests,goals=excluded.goals`, p.User.ID, name, bio, interests, goals)
	if err != nil {
		return err
	}
	http.Redirect(w, r, "/community/u/"+p.User.Name, http.StatusSeeOther)
	return nil
}
func (c *Community) loadAdmin(r *http.Request, p *communityPage) error {
	if p.User == nil || !p.User.Admin {
		return sql.ErrNoRows
	}
	if err := c.prepareEditor(r, p); err != nil {
		return err
	}
	var err error
	p.CatalogItems, err = c.catalogItems()
	if err != nil {
		return err
	}
	cr, err := c.db.Query("SELECT r.id,r.topic_id,t.title,r.body,r.created FROM replies r JOIN topics t ON t.id=r.topic_id ORDER BY r.id DESC LIMIT 200")
	if err != nil {
		return err
	}
	for cr.Next() {
		var x profileComment
		if err = cr.Scan(&x.ID, &x.TopicID, &x.Title, &x.Body, &x.Created); err != nil {
			cr.Close()
			return err
		}
		p.AdminComments = append(p.AdminComments, x)
	}
	err = cr.Err()
	cr.Close()
	if err != nil {
		return err
	}
	rows, err := c.db.Query(profileSelect + "WHERE m.name != 'deleted_member' ORDER BY m.id DESC LIMIT 200")
	if err != nil {
		return err
	}
	for rows.Next() {
		var m memberProfile
		if err = scanProfile(rows, &m); err != nil {
			rows.Close()
			return err
		}
		p.Members = append(p.Members, m)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return err
	}
	rows, err = c.db.Query(`SELECT t.id,t.title,m.name,c.name,t.locked,t.hidden FROM topics t JOIN members m ON m.id=t.member_id JOIN categories c ON c.id=t.category_id ORDER BY t.id DESC LIMIT 200`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var t adminTopic
		if err = rows.Scan(&t.ID, &t.Title, &t.Author, &t.Category, &t.Locked, &t.Hidden); err != nil {
			return err
		}
		p.AdminTopics = append(p.AdminTopics, t)
	}
	return rows.Err()
}
func (c *Community) adminAction(w http.ResponseWriter, r *http.Request, p *communityPage) error {
	if p.User == nil || !p.User.Admin {
		return errors.New("Administrator access required.")
	}
	if !localAdminRequest(r) {
		return errors.New("Local administrator access required.")
	}
	if strings.HasPrefix(r.PostForm.Get("action"), "content-") || strings.HasPrefix(r.PostForm.Get("action"), "delete-") {
		return c.manageContent(w, r, p)
	}
	id, err := strconv.ParseInt(r.PostForm.Get("id"), 10, 64)
	if err != nil || id <= 0 {
		return errors.New("Invalid item.")
	}
	var result sql.Result
	switch r.PostForm.Get("action") {
	case "close":
		result, err = c.db.Exec("UPDATE topics SET locked=1 WHERE id=?", id)
	case "reopen":
		result, err = c.db.Exec("UPDATE topics SET locked=0 WHERE id=?", id)
	case "hide":
		result, err = c.db.Exec("UPDATE topics SET hidden=1 WHERE id=?", id)
	case "restore":
		result, err = c.db.Exec("UPDATE topics SET hidden=0 WHERE id=?", id)
	case "category-edit":
		name, desc := strings.TrimSpace(r.PostForm.Get("name")), strings.TrimSpace(r.PostForm.Get("description"))
		if len([]rune(name)) < 1 || len([]rune(name)) > 60 || len(desc) > 500 {
			return errors.New("Enter a category name (up to 60 characters) and a shorter description.")
		}
		result, err = c.db.Exec("UPDATE categories SET name=?,description=? WHERE id=?", name, desc, id)
	default:
		return errors.New("Unknown admin action.")
	}
	if err != nil {
		return errors.New("Unable to save changes. Check for a duplicate category name.")
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return errors.New("Item not found.")
	}
	http.Redirect(w, r, "/community/admin?tab="+adminTab(r.FormValue("tab"))+"&saved=1", http.StatusSeeOther)
	return nil
}
