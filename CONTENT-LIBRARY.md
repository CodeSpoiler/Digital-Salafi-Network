# Articles and learning

Local URLs: `/en/articles`, `/en/learn`, and their `/ar/` equivalents.

Edit `content/catalog.json` to add or revise content, then restart Go. The catalog
is embedded in the application. Each item has a unique slug, a kind (`article`,
`path`, `book`, or `scholar`), and English/Arabic text. Learning steps reference
existing pages with paths excluding the language prefix.

Starter content: four original study-habit guides, two practical reading paths,
two book references, and two official scholar website references. These are not
the sample articles or usage statistics in the design. Book graphics are
decorative covers, not representations of a particular printed edition.

Directory sources:
- https://sunnah.com/riyadussalihin
- https://sunnah.com/bulugh
- https://binbaz.org.sa/
- https://binothaimeen.net/

The admin dashboard currently manages the forum, not this catalog. Search,
filters, and sorting run on the server. Learning checklists are stored only in
the current browser, not synchronized with accounts. No existing member or
discussion data is modified.

Run `go test ./...` to check localized routes, resource links, filters, escaped
search input, unknown routes, and existing forum functionality.

## Illustration

Both sections use the existing simple illustration in `static/forum-skyline.svg`.
