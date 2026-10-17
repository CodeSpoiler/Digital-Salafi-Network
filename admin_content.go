package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

func initCatalog(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.Exec(`CREATE TABLE IF NOT EXISTS library(slug TEXT PRIMARY KEY,payload TEXT NOT NULL); CREATE TABLE IF NOT EXISTS library_migrations(version INTEGER PRIMARY KEY);`)
	if err != nil {
		return err
	}
	var n int
	if err = tx.QueryRow("SELECT count(*) FROM library_migrations WHERE version=1").Scan(&n); err != nil {
		return err
	}
	if n == 0 {
		var items []catalogItem
		if err = json.Unmarshal(catalogJSON, &items); err != nil {
			return err
		}
		for _, item := range items {
			b, _ := json.Marshal(item)
			if _, err = tx.Exec("INSERT INTO library(slug,payload) VALUES(?,?)", item.Slug, string(b)); err != nil {
				return err
			}
		}
		if _, err = tx.Exec("INSERT INTO library_migrations VALUES(1)"); err != nil {
			return err
		}
	}
	return tx.Commit()
}
func (c *Community) catalogItems() ([]catalogItem, error) {
	rows, err := c.db.Query("SELECT payload FROM library ORDER BY rowid DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []catalogItem
	for rows.Next() {
		var b string
		var item catalogItem
		if err = rows.Scan(&b); err != nil {
			return nil, err
		}
		if err = json.Unmarshal([]byte(b), &item); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
func (c *Community) manageContent(w http.ResponseWriter, r *http.Request, p *communityPage) error {
	if p.User == nil || !p.User.Admin || !localAdminRequest(r) {
		return errors.New("Local administrator access required.")
	}
	action := r.PostForm.Get("action")
	if action == "content-save" {
		item := catalogItem{Slug: strings.TrimSpace(r.FormValue("slug")), Kind: r.FormValue("kind"), Topic: r.FormValue("topic"), Author: strings.TrimSpace(r.FormValue("author")), Source: strings.TrimSpace(r.FormValue("source")), Image: "serene-courtyard.png", Title: localizedText{r.FormValue("title_en"), r.FormValue("title_ar")}, Summary: localizedText{r.FormValue("summary_en"), r.FormValue("summary_ar")}, Featured: r.FormValue("featured") == "on", Beginner: r.FormValue("beginner") == "on"}
		if !regexp.MustCompile(`^[a-z0-9][a-z0-9-]{1,79}$`).MatchString(item.Slug) {
			return errors.New("Use a unique URL name of 2–80 lowercase letters, numbers and hyphens.")
		}
		if item.Kind != "article" && item.Kind != "book" && item.Kind != "scholar" && item.Kind != "path" {
			return errors.New("Choose a content type.")
		}
		if strings.TrimSpace(item.Title.EN) == "" || len(item.Title.EN) > 300 || len(item.Title.AR) > 600 {
			return errors.New("An English title is required (up to 300 bytes).")
		}
		switch item.Topic {
		case "study", "hadith", "fiqh", "manners", "resources":
		default:
			return errors.New("Choose a topic.")
		}
		if item.Source != "" {
			u, e := url.Parse(item.Source)
			if e != nil || u.Scheme != "https" || u.Hostname() == "" || u.User != nil {
				return errors.New("Source must be a complete HTTPS URL.")
			}
		}
		for i, h := range r.PostForm["heading_en"] {
			value := func(k string) string {
				v := r.PostForm[k]
				if i < len(v) {
					return v[i]
				}
				return ""
			}
			if strings.TrimSpace(value("body_en")+value("body_ar")) != "" {
				item.Sections = append(item.Sections, catalogSection{localizedText{h, value("heading_ar")}, localizedText{value("body_en"), value("body_ar")}})
			}
		}
		for i, t := range r.PostForm["step_en"] {
			value := func(k string) string {
				v := r.PostForm[k]
				if i < len(v) {
					return v[i]
				}
				return ""
			}
			if strings.TrimSpace(t) == "" {
				continue
			}
			href := strings.TrimSpace(value("step_href"))
			if href != "" && !regexp.MustCompile(`^/(articles|learn)/[a-z0-9-]+$`).MatchString(href) {
				return errors.New("Step links must be /articles/url-name or /learn/url-name.")
			}
			item.Steps = append(item.Steps, catalogStep{localizedText{t, value("step_ar")}, localizedText{value("step_body_en"), value("step_body_ar")}, href})
		}
		if original := r.FormValue("original"); original != "" {
			var raw string
			if err := c.db.QueryRow("SELECT payload FROM library WHERE slug=?", original).Scan(&raw); err != nil {
				return errors.New("Content not found.")
			}
			var old catalogItem
			if err := json.Unmarshal([]byte(raw), &old); err != nil {
				return err
			}
			item.Image = old.Image
		}
		b, err := json.Marshal(item)
		if err != nil {
			return err
		}
		if len(b) > 28000 {
			return errors.New("Please keep content below 28 KB.")
		}
		original := r.FormValue("original")
		if original == "" {
			_, err = c.db.Exec("INSERT INTO library(slug,payload) VALUES(?,?)", item.Slug, string(b))
		} else {
			if original != item.Slug {
				return errors.New("Existing URL names cannot be changed.")
			}
			var result sql.Result
			result, err = c.db.Exec("UPDATE library SET payload=? WHERE slug=?", string(b), original)
			if err == nil {
				n, _ := result.RowsAffected()
				if n == 0 {
					return errors.New("Content not found.")
				}
			}
		}
		if err != nil {
			return errors.New("Unable to save. The URL name may already be used.")
		}
	} else {
		if r.FormValue("confirm") != "yes" {
			return errors.New("Select the confirmation checkbox before deleting.")
		}
		tx, err := c.db.Begin()
		if err != nil {
			return err
		}
		defer tx.Rollback()
		id, _ := strconv.ParseInt(r.FormValue("id"), 10, 64)
		switch action {
		case "content-delete":
			slug := r.FormValue("slug")
			rows, e := tx.Query("SELECT payload FROM library")
			if e != nil {
				return e
			}
			var changed []catalogItem
			for rows.Next() {
				var raw string
				if e = rows.Scan(&raw); e != nil {
					rows.Close()
					return e
				}
				var x catalogItem
				if e = json.Unmarshal([]byte(raw), &x); e != nil {
					rows.Close()
					return e
				}
				dirty := false
				for i := range x.Steps {
					if x.Steps[i].Href == "/articles/"+slug || x.Steps[i].Href == "/learn/"+slug {
						x.Steps[i].Href = ""
						dirty = true
					}
				}
				if dirty {
					changed = append(changed, x)
				}
			}
			e = rows.Err()
			rows.Close()
			if e != nil {
				return e
			}
			for _, x := range changed {
				b, e := json.Marshal(x)
				if e != nil {
					return e
				}
				if _, e = tx.Exec("UPDATE library SET payload=? WHERE slug=?", string(b), x.Slug); e != nil {
					return e
				}
			}
			_, err = tx.Exec("DELETE FROM library WHERE slug=?", slug)
		case "delete-topic":
			_, err = tx.Exec("UPDATE replies SET parent_id=NULL WHERE topic_id=?", id)
			if err == nil {
				_, err = tx.Exec("DELETE FROM replies WHERE topic_id=?", id)
			}
			if err == nil {
				_, err = tx.Exec("DELETE FROM topics WHERE id=?", id)
			}
		case "delete-comment":
			_, err = tx.Exec("UPDATE replies SET parent_id=(SELECT parent_id FROM replies WHERE id=?) WHERE parent_id=?", id, id)
			if err == nil {
				_, err = tx.Exec("DELETE FROM replies WHERE id=?", id)
			}
		case "delete-user":
			var admin bool
			var name string
			if err = tx.QueryRow("SELECT name,admin FROM members WHERE id=?", id).Scan(&name, &admin); err != nil {
				return errors.New("Member not found.")
			}
			if admin || id == p.User.ID || name == "deleted_member" {
				return errors.New("Administrator and system accounts cannot be deleted.")
			}
			var existingPassword string
			e := tx.QueryRow("SELECT password FROM members WHERE name='deleted_member'").Scan(&existingPassword)
			if e != sql.ErrNoRows && (e != nil || existingPassword != "!disabled") {
				return errors.New("The reserved deleted_member username is occupied. Rename it before deleting users.")
			}
			_, err = tx.Exec("INSERT OR IGNORE INTO members(name,password,admin) VALUES('deleted_member','!disabled',0)")
			if err != nil {
				return err
			}
			_, err = tx.Exec("INSERT OR IGNORE INTO profiles(member_id,display_name) SELECT id,'Deleted member' FROM members WHERE name='deleted_member'")
			if err != nil {
				return err
			}
			for _, q := range []string{"UPDATE topics SET member_id=(SELECT id FROM members WHERE name='deleted_member') WHERE member_id=?", "UPDATE replies SET member_id=(SELECT id FROM members WHERE name='deleted_member') WHERE member_id=?", "DELETE FROM sessions WHERE member_id=?", "DELETE FROM profiles WHERE member_id=?", "DELETE FROM members WHERE id=?"} {
				if _, err = tx.Exec(q, id); err != nil {
					return err
				}
			}
		default:
			return errors.New("Unknown action.")
		}
		if err != nil {
			return err
		}
		if err = tx.Commit(); err != nil {
			return err
		}
	}
	http.Redirect(w, r, "/community/admin", http.StatusSeeOther)
	return nil
}
