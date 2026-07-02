# Social Network

A Facebook-style social networking application built with a Go backend and a
React frontend.

The project is under active development. It currently includes the application
shell and the first backend features; the broader social-network feature set is
documented and being implemented incrementally.

## Current Status

Present in the current codebase:

- React 19 and Vite UI scaffold with routes and page components for the home
  feed, profile, friends, groups, events, messages, and notifications
- Go REST API with account registration, login, logout, and current-user
  endpoints
- Public and private profile support in the user model
- Follow, unfollow, accept, reject, followers, and following API operations
- Posts, comments, privacy controls, selected audiences, and reactions
- Groups with invitations, membership requests, admin roles, member leaving, group feeds, and events
- Direct and group chat with WebSocket delivery, DM entry points, and DM starter candidates
- In-app notifications for follow, group, and event workflows
- Cookie-backed sessions, SQLite persistence, automatic SQL migrations, and explicit developer fixture data

## Technology

- **Frontend:** React 19, React Router 7, Vite 8, and plain JavaScript/JSX
- **Backend:** Go 1.22, `net/http`, and SQLite
- **Authentication:** Cookie-backed sessions with bcrypt password hashing
- **Data:** SQLite with application-managed migrations

## Documentation

- [Documentation index](docs/README.md)
- [Local development setup](docs/setup/local-development.md)
- [API specification](docs/api/README.md)
- [Database documentation](docs/database/README.md)
- [Original project requirements](docs/reference/project-requirements.md)
- [Contribution guide](CONTRIBUTING.md)

## Project Layout

```text
.
├── backend/       Go API, business logic, persistence, and migrations
├── frontend/      React single-page application
├── docs/          Setup, API, database, and reference documentation
└── CONTRIBUTING.md
```

## Project Design

https://miro.com/app/board/uXjVHKn_cVY=/?share_link_id=101358376740
