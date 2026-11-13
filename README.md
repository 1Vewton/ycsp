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

## Repository layout

| Directory | Layer | Documentation |
| --- | --- | --- |
| `frontend/` | Frontend | [README](./frontend/README.md) |
| `backend/` | Data management | [README](./backend/README.md) |
| `compute/` | Data processing | [README](./compute/README.md) |

CI configuration is documented in [`.github/README-workflow.md`](./.github/README-workflow.md); the development plan is in [TASKS.md](./TASKS.md) (Chinese).
