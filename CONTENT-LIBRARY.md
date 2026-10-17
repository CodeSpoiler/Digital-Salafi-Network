# Articles and learning

Local URLs: `/en/articles`, `/en/learn`, and their `/ar/` equivalents.

Sign in as `network_admin` at `http://localhost:8095/community/admin` for Windows development. Open **Articles & Learning → Add content**. Choose Article, Book, Scholar or Learning path, enter a unique URL name, title, summary and sections, then publish. Arabic fields are optional and fall back to English. Learning paths can include multiple steps and links such as `/articles/a-steady-study-routine`.

Expand an existing item to edit or delete it. Publication and deletion take effect immediately; no restart is needed. Only locally connected administrators can manage content. All mutations require a valid session and CSRF token. Source links must use HTTPS. Content is plain text, safely escaped on display.

With `COMMUNITY_DB` enabled, the library is stored in the SQLite database alongside the community. Include this database in backups. `content/catalog.json` seeds the library once; deleted entries do not reappear on restart. A server without a database uses the read-only embedded catalog.

The Members section deletes a regular member's account, profile and sessions, while preserving their posts under `deleted_member`. Administrator accounts cannot be deleted. Deleting a topic removes all its comments; deleting a comment preserves its child replies. Each deletion requires its confirmation checkbox. Content deletion clears links to that entry in learning steps, preserving the steps themselves.

Starter content: four original study-habit guides, two practical reading paths,
two book references, and two official scholar website references. These are not
the sample articles or usage statistics in the design. Book graphics are
decorative covers, not representations of a particular printed edition.

Directory sources:
- https://sunnah.com/riyadussalihin
- https://sunnah.com/bulugh
- https://binbaz.org.sa/
- https://binothaimeen.net/

Search,
filters, and sorting run on the server. Learning checklists are stored only in
the current browser, not synchronized with accounts.

Run `go test ./...` to check localized routes, resource links, filters, escaped
search input, unknown routes, and existing forum functionality.

## Illustration

Both sections use the existing simple illustration in `static/forum-skyline.svg`.
