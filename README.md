# Salafi Network

A Go website and community forum, using server-rendered HTML, CSS and SQLite. This branch contains the simpler responsive design for local development.

## Run locally

With Go installed, run `bash scripts/dev.sh` from this folder. On Windows with Ubuntu WSL and Go installed inside it, run `powershell -ExecutionPolicy Bypass -File .\dev.ps1`.

Open http://localhost:8095/community. Create a test account; the development database is stored in ignored data/development.db. Stop with Ctrl+C, and restart after source changes. This launcher runs Go directly, bound to localhost, without containers or a public tunnel.

## Tests

Run `go test ./...` in WSL/Linux or a terminal with Go installed.

