# API

[`openapi.json`](openapi.json) contains the OpenAPI 3.1 contract for the target
Social Network API.

The document is a design specification, not a guarantee that every path is
currently available. It can be opened with an OpenAPI compatible visualiser like
[swagger editor](https://editor.swagger.io/).

Consult `backend/internal/api/routers/router.go` when checking runtime endpoint
availability. The local API base URL is `http://localhost:8080/api`.

## Implemented Chat Runtime Endpoints

Authenticated requests use the `session_token` cookie.

- `GET /api/chats` returns active visible chat rooms with at least one saved message, sorted by latest message activity. The current response shape matches `ConversationResponse`: `thread_id` for DMs, `group_id` for groups, `target_id` for DM recipient user's ID, `type`, `target_name`, `target_avatar`, `last_message`, and `last_message_at`.
- `POST /api/chats/dm` opens or retrieves a DM thread. Body: `{ "recipient_id": "user-uuid" }`. At least one accepted follow relationship between the two users is required. Empty DM threads are not returned by `GET /api/chats` until a message is saved.
- `GET /api/chats/dm-candidates?limit=10` returns accepted followers/followees ordered by relationship recency, excluding users who already have a DM thread with at least one saved message.
- `GET /api/chats/{id}/messages?chat_type=dm|group&limit=100&offset=0` returns message history oldest-to-newest for the requested page. DM participants keep read access to existing history; group chat history requires current accepted membership.
- `DELETE /api/messages/{id}` soft deletes an individual message sent by the current user, updating its content to a tombstone `"This message is no longer available"` and broadcasting update notifications.
- `DELETE /api/messages?chat_id={chat_id}&chat_type={dm|group}` soft deletes all messages sent by the current user in the specified chat and broadcasts a `"messages.cleared"` WS event.
- `GET /api/ws` opens the authenticated WebSocket connection. Client send event: `{ "type": "message.send", "chat_id": "uuid", "chat_type": "dm|group", "content": "Hello", "client_message_id": "optional-client-id" }`. Server success event: `{ "type": "message.created", "data": MessageResponse, "payload": MessageResponse, "client_message_id": "optional-client-id" }`. Server error event: `{ "type": "error", "error": { "code": "CHAT_FORBIDDEN", "message": "..." }, "client_message_id": "optional-client-id" }`. Also broadcasts `"messages.cleared"` events.

Legacy runtime routes remain available: `GET /api/conversations` and `GET|POST /api/messages`.

## Implemented Group Runtime Fields and Admin Endpoints

Groups include optional `avatar` and accepted memberships include `role`, either `admin` or `member`. Existing creators are migrated to `admin`, and newly-created groups make the creator an admin automatically.

- `GET /api/groups/{id}` returns group details and the viewer's `status`/`role` when applicable.
- `PATCH /api/groups/{id}` lets group admins update `title`, `description`, and `avatar`.
- `GET /api/groups/{id}/members` returns accepted members with admins first, then members alphabetically; each member includes `role`.
- `GET /api/groups/{id}/requests` is admin-only and returns pending incoming join requests.
- `GET /api/groups/{id}/invitations` is admin-only and returns pending sent invitations.
- `POST /api/groups/{id}/members/{userID}/role/promote` promotes an accepted member to admin.
- `POST /api/groups/{id}/members/{userID}/role/demote` demotes an admin to member. The backend rejects demoting the last admin.
- `POST /api/groups/{id}/invite` is admin-only and creates a pending invitation; invited users still need to accept.
- `POST /api/groups/{id}/leave` lets an accepted member leave. If the user is the only accepted member, the group is deleted and group-scoped rows cascade. If the user is the last admin while other accepted members remain, the request is rejected until another admin is assigned.

Group feed access requires current accepted membership. Former members may view and delete their own group posts, but cannot edit them, read their comment/reply threads, create comments/replies, or navigate back into the group feed through comment context.
