# Fantastic coffee (decaffeinated)

This repository contains the basic structure for [Web and Software Architecture](http://gamificationlab.uniroma1.it/en/wasa/) homework project.
It has been described in class.

"Fantastic coffee (decaffeinated)" is a simplified version for the WASA course, not suitable for a production environment.
The full version can be found in the "Fantastic Coffee" repository.

## Project structure

* `cmd/` contains all executables; Go programs here should only do "executable-stuff", like reading options from the CLI/env, etc.
	* `cmd/healthcheck` is an example of a daemon for checking the health of servers daemons; useful when the hypervisor is not providing HTTP readiness/liveness probes (e.g., Docker engine)
	* `cmd/webapi` contains an example of a web API server daemon
* `demo/` contains a demo config file
* `doc/` contains the documentation (usually, for APIs, this means an OpenAPI file)
* `service/` has all packages for implementing project-specific functionalities
	* `service/api` contains an example of an API server
	* `service/globaltime` contains a wrapper package for `time.Time` (useful in unit testing)
* `vendor/` is managed by Go, and contains a copy of all dependencies
* `webui/` is an example of a web frontend in Vue.js; it includes:
	* Bootstrap JavaScript framework
	* a customized version of "Bootstrap dashboard" template
	* feather icons as SVG
	* Go code for release embedding

Other project files include:
* `open-node.sh` starts a new (temporary) container using `node:20` image for safe and secure web frontend development (you don't want to use `node` in your system, do you?).

## Go vendoring

This project uses [Go Vendoring](https://go.dev/ref/mod#vendoring). You must use `go mod vendor` after changing some dependency (`go get` or `go mod tidy`) and add all files under `vendor/` directory in your commit.

For more information about vendoring:

* https://go.dev/ref/mod#vendoring
* https://www.ardanlabs.com/blog/2020/04/modules-06-vendoring.html

## Node/YARN vendoring

This repository uses `yarn` and a vendoring technique that exploits the ["Offline mirror"](https://yarnpkg.com/features/caching). As for the Go vendoring, the dependencies are inside the repository.

You should commit the files inside the `.yarn` directory.

## How to set up a new project from this template

You need to:

* Change the Go module path to your module path in `go.mod`, `go.sum`, and in `*.go` files around the project
* Rewrite the API documentation `doc/api.yaml`
* If no web frontend is expected, remove `webui` and `cmd/webapi/register-webui.go`
* Update top/package comment inside `cmd/webapi/main.go` to reflect the actual project usage, goal, and general info
* Update the code in `run()` function (`cmd/webapi/main.go`) to connect to databases or external resources
* Write API code inside `service/api`, and create any further package inside `service/` (or subdirectories)

## How to build

If you're not using the WebUI, or if you don't want to embed the WebUI into the final executable, then:

```shell
go build ./cmd/webapi/
```

If you're using the WebUI and you want to embed it into the final executable:

```shell
./open-node.sh
# (here you're inside the container)
yarn run build-embed
exit
# (outside the container)
go build -tags webui ./cmd/webapi/
```

## How to run (Development Mode)

### Local Run
1.  **Backend**:
    ```shell
    go run ./cmd/webapi/
    ```
2.  **Frontend**:
    ```shell
    cd webui
    npm install
    npm run dev
    ```
    The app will be available at `http://localhost:5173`. Use `npm` for consistent dependency management.

### Docker Compose (Recommended)
This starts both the backend and frontend with a single command and persists data.
```shell
docker compose up --build
```
-   **Frontend**: `http://localhost:5173`
-   **Backend**: `http://localhost:3000`
-   **Database**: Persisted in a Docker volume `wasa-data`.

## Final QA Checklist
Use this checklist to verify the system is fully functional:
- [ ] **Auth**: Login as two different users.
- [ ] **Messaging**: Alice sends DM to Bob; Bob receives it instantly.
- [ ] **Comments**: Alice adds/removes a comment on Bob's message.
- [ ] **Soft-delete**: Bob deletes his message; Alice sees it as "deleted".
- [ ] **Conflict (409)**: Alice tries to comment on Bob's deleted message (expect error).
- [ ] **Groups**: Create a group, rename it, add a member, and leave it.
- [ ] **Photos**: Upload a profile photo and group photo; verify they display correctly.
- [ ] **Persistence**: Restart containers (`docker compose restart`) and verify all data survives.

## How to reset the database

### Local
Remove the local database file:
```shell
rm data/wasa.db
```

### Docker
Remove the Docker volume:
```shell
docker compose down -v
```

## Regression Testing
A comprehensive automated test suite is available:
```shell
chmod +x scripts/regression_test.sh
./scripts/regression_test.sh
```
See [REGRESSION.md](REGRESSION.md) for detailed scenarios.

## License

See [LICENSE](LICENSE).

