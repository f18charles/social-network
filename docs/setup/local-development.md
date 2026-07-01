# Local Development

## Prerequisites

- Go 1.22 or newer
- Node.js 20.19 or newer, or 22.12 or newer
- npm
- A C compiler and SQLite development libraries for `go-sqlite3`

## Backend

From the repository root, create the local environment file:

```bash
cp backend/.env.example backend/.env
```

Replace the placeholder paths in `backend/.env` with working values:

```ini
PORT=8080
DATABASE_PATH=./db.sqlite
APP_ENV=development
ALLOWED_ORIGIN=http://localhost:5173
MIGRATIONS_DIR=./internal/db/migrations
```

Run the API from `backend/` so these relative paths resolve correctly:

```bash
cd backend
go run ./cmd/server
```

The server listens on `http://localhost:8080`. Database migrations are applied
automatically at startup, and the SQLite database is created at
`backend/db.sqlite`.

### E2E Fixture Data

Developer E2E data is not loaded by migrations or server startup. To explicitly
seed five fixture users, twenty posts, downloaded JPEG/PNG/GIF post and
comment media, downloaded user avatars, comments, nested replies, and post/comment
likes, run from `backend/`:

```bash
go run ./cmd/devdata seed
```

Every seeded user uses the password `Password123!`; emails use the
`e2e.<name>@example.test` pattern. The seed command downloads fixture media
from public image URLs into `uploads/images` and `uploads/avatars`, so it needs
network access. Remove the fixture rows and fixture media with:

```bash
go run ./cmd/devdata teardown
```

The command refuses to run when `APP_ENV=production`.

## Frontend

In a second terminal:

```bash
cd frontend
npm install
npm run dev
```

Vite serves the frontend at `http://localhost:5173` by default.

The Vite development server proxies `/api` requests to
`http://localhost:8080`. Uploaded media is served from `/uploads`; load returned
image URLs directly from the API origin or proxy `/uploads` to the backend in
frontend development. The shared request helper includes credentials for
authenticated routes so the session cookie is sent.

## Validation

Run backend checks:

```bash
cd backend
go test ./...
```

Run frontend checks:

```bash
cd frontend
npm test
npm run lint
npm run build
```

## Docker

Start the backend with persistent SQLite and upload volumes:

```bash
docker compose up --build
```

Uploaded files are mounted at `/app/uploads` in the container and survive a
backend container restart.
