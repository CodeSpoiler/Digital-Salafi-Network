# Salafi Network

A Go website and community forum, using server-rendered HTML, CSS and SQLite. This branch contains the simpler responsive design for local development.

## Run locally

With Go installed, run `bash scripts/dev.sh` from this folder. On Windows with Ubuntu WSL and Go installed inside it, run `powershell -ExecutionPolicy Bypass -File .\dev.ps1`.

Open http://localhost:8095/community. Create a test account; the development database is stored in ignored data/development.db. Stop with Ctrl+C, and restart after source changes. This launcher runs Go directly, bound to localhost, without containers or a public tunnel.

## Tests

Run `go test ./...` in WSL/Linux or a terminal with Go installed.

## Git workflow

Work on `design/simple-responsive-forum`. Commit your selected changes and run `git push` to update this branch. When ready, open a pull request into main. Production configuration, private data and earlier deployment history are not included in this repository.

## Features

Username/password accounts, categories, topics, threaded comments, search and basic administration. Password recovery, DMs, uploads, pagination and expanded moderation are still in development. Local test accounts are separate from any hosted installation.
