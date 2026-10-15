# Windows development

This checkout is for native Windows development. Double-click start-dev.cmd or run `powershell -ExecutionPolicy Bypass -File .\dev.ps1` in this folder. Visit http://localhost:8095/community. Stop with Ctrl+C; restart to pick up Go, HTML and CSS changes.

Go for Windows is required. Initial dependencies download once and are cached. No Docker, WSL, Cloudflare or public deployment is started by this launcher. The server listens only on 127.0.0.1.

Development accounts and posts live in the ignored data/development.db. Create a test account in the local site. Live credentials and data are not used.

The organization repository is the `origin` remote. Nothing is automatically committed, pushed, or deployed. Review changes before pushing, and keep deployment configuration and private data outside the public repository.

Use `go test ./...` to run tests. If Go is not yet on your terminal PATH, reopen the terminal or use `& 'C:\Program Files\Go\bin\go.exe' test ./...`.

The initially launched server runs in the background. Double-click stop-dev.cmd to stop it, then start-dev.cmd to run it in a visible terminal. Both helpers operate only on the local Windows development server.
