> English | [简体中文](./README.zh-CN.md)

# ycsp

A crystal structure prediction (CSP) teaching platform for teenagers: enter a chemical composition and see the crystal structure it is most likely to form. Students need no prior experience with computation or programming.

## What this project is

The platform is for teenagers, and assumes no background in computation or programming. The only acceptance criterion is whether a middle-school student can understand and operate it independently.

## Architecture

| Layer | Responsibility | Stack |
| --- | --- | --- |
| Frontend | Student interface and crystal structure visualization | TypeScript + Vue |
| Data management | Single public API, data storage, forwarding compute requests | Go + Gin + GORM |
| Data processing | Crystal structure prediction and numerical work | Python + FastAPI (NumPy only when numerical work needs it) |

Request flow: `Frontend (Vue) → Data management (Go) → Data processing (Python)`. The frontend talks only to the data management layer, which in turn calls the data processing layer.

For local development and tests the data management layer uses SQLite (zero install). Production uses PostgreSQL; see the [backend README](./backend/README.md) for the production configuration and migration step.

## Languages

The product must switch interface languages so that users of different languages can work independently:

- Chinese by default, with English also available (switchable).
- Copy is separate from code: the frontend uses language packs (see `frontend/src/locales/`), and no text is hard-coded in components.
- The backend returns only error codes and parameters; the frontend translates them into the current language, so no language-specific message is baked into the backend.

## Install and run

Requires Go 1.25+, Node.js 22+, [uv](https://docs.astral.sh/uv/) (the data processing layer uses it to manage Python dependencies and the virtual environment), and [Redis](https://redis.io/) (stores login sessions; without a local Redis you can temporarily use in-memory sessions, see below).

Ports and addresses have defaults, so everything works as-is. To change them, copy the corresponding `.env.example` to `.env` and edit it:

```powershell
Copy-Item compute\.env.example compute\.env
Copy-Item backend\.env.example backend\.env
Copy-Item frontend\.env.example frontend\.env
```

Then start the three layers in three terminals:

```powershell
# Data processing layer (default 8001)
cd compute
uv run python -m app.main

# Data management layer (8000)
cd backend
go mod tidy; go run .

# Frontend (5173)
cd frontend
npm install; npm run dev
```

Open <http://localhost:5173/> and click "Check background services" on the home page. If it shows "The background services are running, so you can start predicting.", all three layers are connected.

> Without a local Redis, set `YCSP_SESSION_DRIVER=memory` in `backend/.env` to run registration and login
> (sessions live only in process memory and are lost after a restart).

## Docker Compose (development)

Requires Docker Desktop, or Docker Engine with Compose v2. This is a development stack: the source is bind-mounted into the containers, so Vite, uvicorn and Air hot-reload; PostgreSQL runs in a container, and the host does not need Go / Node / uv / Redis.

```powershell
docker compose up --build
```

The first build downloads Go / Python / npm dependencies; later starts are faster. Open <http://localhost:5173/> and click "Check background services".

| Service | Address | Notes |
| --- | --- | --- |
| `frontend` | <http://localhost:5173> | Vite dev server |
| `backend` | <http://localhost:8000> | Go API |
| `compute` | <http://localhost:8001> | FastAPI; API docs at `/docs` |
| `postgres` | `127.0.0.1:5432` | Development user/password `ycsp` / `ycsp`, database `ycsp` |
| `redis` | `127.0.0.1:6380` | Login sessions (still 6379 inside the container; the host port avoids a local Redis) |

Common commands:

```powershell
docker compose logs -f backend
docker compose logs -f compute
docker compose down        # keep the database and dependency volumes
docker compose down -v     # also delete PostgreSQL, Redis and dependency volumes
```

Notes:

- This Compose stack uses PostgreSQL, not SQLite, so development exercises the production database path.
- The credentials and ports are for local development only; do not use them in production.
- After changing `go.mod` / `uv.lock` / `package.json`, the dependency volumes may be stale: run `docker compose down -v`, then `docker compose up --build`.
- If a host port is already in use, change the left-hand side of the corresponding `ports` entry in `docker-compose.yml`.

## Repository layout

| Directory | Layer | Documentation |
| --- | --- | --- |
| `frontend/` | Frontend | [README](./frontend/README.md) |
| `backend/` | Data management | [README](./backend/README.md) |
| `compute/` | Data processing | [README](./compute/README.md) |
| `slides/` | Presentation | [README](./slides/README.md) |
| `docker-compose.yml` | Development stack | See "Docker Compose (development)" above |

CI configuration is documented in [`.github/README-workflow.md`](./.github/README-workflow.md); the development plan is in [TASKS.md](./TASKS.md) (Chinese).
