# WASAText Messenger

WASAText Messenger is a lightweight, web-based social messaging platform designed for the **Web and Software Architecture (WASA)** course. It features a robust Go backend, a reactive Vue 3/Vite frontend, and reliable SQLite persistence using a pure-Go driver.

## Features
- **Direct & Group Messaging**: Seamless one-on-one and multi-user conversations.
- **Message Lifecycle**: Send, receive, and **soft-delete** messages.
- **Interactions**: Add or remove comments on any message.
- **Forwarding**: Easily forward messages across different conversations.
- **User Discovery**: Search for users to start new DMs.
- **Managed Profiles**: Upload and display profile and group photos (PNG/JPEG).
- **SQLite Persistence**: Data survives restarts via `modernc.org/sqlite`.
- **Docker Compose Orchestration**: Start the entire stack with a single command.

## Technologies Used

### Backend
- **Go 1.24**: Fast, statically-typed compiled language.
- **httprouter**: High-performance request multiplexer.
- **Logrus**: Structured logger for Go.
- **Gorilla Handlers**: Standard middleware for logging and recovery.
- **UUID**: Unique identifier generation for messages and conversations.
- **modernc.org/sqlite**: Pure-Go SQLite implementation (no CGO required).
- **ardanlabs/conf**: Support for configuration via environment variables and flags.

### Frontend
- **Vue.js 3**: Progressive JavaScript framework.
- **Vite**: Next-generation frontend tooling.
- **Vue Router**: Official router for Vue.js.
- **Axios**: Promised-based HTTP client for the browser.
- **Bootstrap**: Modern styling and responsive components.

## Setup & Running the Application on your local machine

### Backend
**Prerequisites**: [Go 1.24+](https://go.dev/dl/)

**Build and Run**:
```bash
go run ./cmd/webapi/
```
The backend listens on port **3000** by default. Data is stored in `./data/wasa.db` (the folder is created automatically).

### Frontend
**Prerequisites**: [Node.js 20+](https://nodejs.org/)

**Build and Run**:
```bash
cd webui
npm install
npm run dev
```
The frontend dev server starts on **http://localhost:5173**. Requests to `/api` are automatically proxied to the backend.

### Docker
The project is best run using **Docker Compose**, which handles all networking and persistence setup.

**Start the Stack**:
```bash
docker compose up --build
```
- **Web UI**: [http://localhost:5173](http://localhost:5173)
- **API Server**: [http://localhost:3000](http://localhost:3000)
- **Persistence**: Data is mapped to a named volume `wasa-data`.

**Stop the Stack**:
```bash
docker compose down
```

## Configuration
WASAText can be configured via environment variables (inherited by Docker) or command-line flags.

**Key Environment Variables**:
- `CFG_DB_FILENAME`: Path to the SQLite database (default: `./data/wasa.db`).
- `CFG_WEB_APIHOST`: Host and port for the API server (default: `0.0.0.0:3000`).
- `CFG_DEBUG`: Enable verbose logging (default: `false`).

**Database Reset**:
- **Local**: `rm data/wasa.db`
- **Docker**: `docker compose down -v`

## Testing
An automated regression suite is provided to verify the core messaging and group flows.

**Run All Tests**:
```bash
chmod +x scripts/regression_test.sh
./scripts/regression_test.sh
```
The script validates authentication, message synchronization, soft-deletion conflict logic, group management, and photo uploads.

## Troubleshooting
- **Port Conflict**: If port 3000 or 5173 is in use, the application will fail to start. Ensure these ports are available.
- **Script Permissions**: If you cannot run scripts, use `chmod +x scripts/*.sh` to grant execution rights.
- **Docker Health**: If the backend is "unhealthy" in Docker, ensure it can reach its own `/liveness` endpoint.
- **Proxy Issues**: In local dev, the frontend relies on the Vite proxy defined in `vite.config.js` to reach the backend.
- **Stale Database**: If you encounter schema errors after an update, perform a **Database Reset** as described above.

---
*Developed for the WASA course assignment.*
