# Database

The backend uses SQLite. Its executable schema is built from migrations in
`backend/internal/db/migrations/`, which are applied automatically when the
server starts.

[`schema.dbml`](schema.dbml) describes the target data model, including planned
features that do not have migrations yet. It can be opened with a DBML-compatible
tool to visualize relationships like [dbdiagram.io/d](https://dbdiagram.io/d).

## Current Migrations

The implemented migration set currently creates:

- `users`
- `sessions`
- `followers`
- `groups`
- `group_members`
- `posts`
- `post_audiences`
- `comments`
- `post_votes`
- `comment_votes`

The group tables are the minimal cross-team contract required by group-scoped
posts and membership checks. Full group workflows, invitations, events, and
related APIs are implemented separately.

When changing the database, add paired `.up.sql` and `.down.sql` migration files
and update the DBML document when the conceptual model changes. See
[`CONTRIBUTING.md`](../../CONTRIBUTING.md) for naming and validation guidance.

## Developer Fixture Data

E2E fixture data is managed outside migrations so it is never inserted by
production startup. From `backend/`, use `go run ./cmd/devdata seed` to insert
fixture users, posts, downloaded post/comment media, user avatars, comments,
replies, and post/comment votes into the configured SQLite database. Use
`go run ./cmd/devdata teardown` to remove only fixture-owned rows and fixture
media files.
