# Salafi Network

A Go website and community forum using server-rendered HTML, CSS and SQLite.

## Run locally

Install Go (1.22 or newer) and Git, then clone this repository. No Docker, WSL, Cloudflare account or hosting configuration is required.

### Windows

Double-click `start-dev.cmd`, or run:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File .\dev.ps1
```

### Linux and macOS

```sh
bash scripts/dev.sh
```

Open http://localhost:8095/community and create a local test account. Each developer has their own database in `data/development.db`, which Git ignores. Stop with Ctrl+C and restart after changing Go, HTML or CSS. On Windows, `stop-dev.cmd` can also stop a background development server. See [Windows development notes](WINDOWS-DEVELOPMENT.md).

The first run downloads Go dependencies; later runs reuse the local cache. These launchers bind to localhost and do not publish the site or deploy changes.

## Get other contributors' changes

Stop the local server, save your own work in a commit or stash, then run:

```sh
git switch main
git pull --ff-only origin main
```

Start the local server again. You do not need to edit the launcher when someone changes the design or application code. Your ignored local database stays on your computer.

## Contribute

Create a branch for your changes:

```sh
git switch -c design/my-change
git add <files-you-changed>
git commit -m "Describe your change"
git push -u origin design/my-change
```

Open a pull request into `main`. Contributors without repository write access can fork the repository and open a pull request from their fork. Once changes are merged, other developers can pull `main` using the commands above.

## Tests

```sh
go test ./...
```

## Development and deployment

This repository contains shared application code and local development launchers. Production deployment is handled separately by the project owner. Pushing or pulling code does not start Docker, create a tunnel or update production. Do not commit databases, credentials, tokens, private configuration or backup archives.
