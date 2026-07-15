# Chirpy API Endpoint Documentation

This document lists every HTTP endpoint exposed by the Chirpy server, grouped by category.

Base URL: `http://localhost:8080` (default)

---

## Table of Contents

- [Public status](#public-status)
- [Users](#users)
- [Authentication](#authentication)
- [Chirps](#chirps)
- [Webhooks](#webhooks)
- [Admin](#admin)
- [Frontend](#frontend)

---

## Public status

### GET /api/healthz

Returns a simple health check response.

- **Auth:** none
- **Response:** `text/plain`

**200 OK**
```
OK
```

Implemented in: `admin.go:15`

---

## Users

### POST /api/users

Create a new user account.

- **Auth:** none
- **Body:** JSON

**Request**
```json
{
  "email": "user@example.com",
  "password": "securepassword"
}
```

**201 Created**
```json
{
  "id": "uuid",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z",
  "email": "user@example.com",
  "is_chirpy_red": false
}
```

**Errors:** `400 Bad Request` for invalid input or a database error.

Implemented in: `users.go:26`

### PUT /api/users

Update the authenticated user's email and password.

- **Auth:** Bearer token required (`Authorization: Bearer <access-token>`)
- **Body:** JSON

**Request**
```json
{
  "email": "newemail@example.com",
  "password": "newpassword"
}
```

**200 OK**
```json
{
  "id": "uuid",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z",
  "email": "newemail@example.com",
  "is_chirpy_red": false
}
```

**Errors:** `400 Bad Request`, `401 Unauthorized`, `500 Internal Server Error`.

Implemented in: `users.go:221`

---

## Authentication

### POST /api/login

Log in with email and password. Returns a short-lived access token and a long-lived refresh token.

- **Auth:** none
- **Body:** JSON

**Request**
```json
{
  "email": "user@example.com",
  "password": "securepassword"
}
```

**200 OK**
```json
{
  "id": "uuid",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z",
  "email": "user@example.com",
  "token": "<access-token-jwt>",
  "refresh_token": "<refresh-token>",
  "is_chirpy_red": false
}
```

**Errors:** `401 Unauthorized` for incorrect email or password.

Implemented in: `users.go:82`

### POST /api/refresh

Exchange a valid refresh token for a new access token.

- **Auth:** Bearer token required (`Authorization: Bearer <refresh-token>`)
- **Body:** none

**200 OK**
```json
{
  "token": "<new-access-token-jwt>"
}
```

**Errors:** `401 Unauthorized` for missing, expired, or revoked refresh tokens.

Implemented in: `users.go:154`

### POST /api/revoke

Revoke a refresh token so it can no longer be used to mint new access tokens.

- **Auth:** Bearer token required (`Authorization: Bearer <refresh-token>`)
- **Body:** none

**204 No Content**

**Errors:** `401 Unauthorized` for missing or invalid refresh tokens.

Implemented in: `users.go:190`

---

## Chirps

### POST /api/chirps

Create a new chirp. Chirps are limited to 140 characters. The server automatically replaces the words `kerfuffle`, `sharbert`, and `fornax` with `****`.

- **Auth:** Bearer token required (`Authorization: Bearer <access-token>`)
- **Body:** JSON

**Request**
```json
{
  "body": "Hello, Chirpy!"
}
```

**201 Created**
```json
{
  "id": "uuid",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z",
  "body": "Hello, Chirpy!",
  "user_id": "uuid"
}
```

**Errors:** `400 Bad Request` for invalid input, chirps over 140 characters, or a database error.

Implemented in: `chirps.go:25`

### GET /api/chirps/

List all chirps. Optionally filter by author and sort order.

- **Auth:** none
- **Query parameters:**
  - `author_id` (optional): UUID of a user to filter chirps by
  - `sort` (required): `asc` or `desc`

**Request**
```
GET /api/chirps/?sort=desc
GET /api/chirps/?author_id=<uuid>&sort=asc
```

**200 OK**
```json
[
  {
    "id": "uuid",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z",
    "body": "Hello, Chirpy!",
    "user_id": "uuid"
  }
]
```

**Errors:** `400 Bad Request` for invalid sort value; `404 Not Found` if the author filter matches no chirps.

Implemented in: `chirps.go:77`

### GET /api/chirps/{chirpID}

Get a single chirp by its UUID.

- **Auth:** none

**200 OK**
```json
{
  "id": "uuid",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z",
  "body": "Hello, Chirpy!",
  "user_id": "uuid"
}
```

**Errors:** `400 Bad Request` for invalid UUID; `404 Not Found` if chirp does not exist.

Implemented in: `chirps.go:150`

### DELETE /api/chirps/{chirpID}

Delete a chirp. Only the chirp's owner can delete it.

- **Auth:** Bearer token required (`Authorization: Bearer <access-token>`)

**204 No Content**

**Errors:** `400 Bad Request` for invalid UUID; `401/403 Unauthorized` for missing token or non-owner; `404 Not Found` if chirp does not exist.

Implemented in: `chirps.go:196`

---

## Webhooks

### POST /api/polka/webhooks

External webhook for upgrading a user to Chirpy Red. The request must include the configured API key in the `Authorization` header.

- **Auth:** API key required (`Authorization: ApiKey <api-key>`)
- **Body:** JSON

**Request**
```json
{
  "event": "user.upgraded",
  "data": {
    "user_id": "uuid"
  }
}
```

**204 No Content**

The handler returns `204 No Content` for any non-`user.upgraded` event without performing an upgrade.

**Errors:** `401 Unauthorized` for missing or invalid API key; `404 Not Found` if the user does not exist.

Implemented in: `webhooks.go:12`

---

## Admin

### GET /admin/metrics

Returns a simple HTML page showing how many times the `/app/` file server has been hit.

- **Auth:** none
- **Response:** `text/html`

**200 OK**
```html
<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited 42 times!</p>
  </body>
</html>
```

Implemented in: `admin.go:21`

### POST /admin/reset

Resets all users and the hit counter. Only available when `PLATFORM=dev`.

- **Auth:** none
- **Platform guard:** `PLATFORM=dev`

**200 OK**
```
Reset successfully! Hits: 0
Users reset successfully!
```

**Errors:** `403 Forbidden` in non-dev environments.

Implemented in: `admin.go:35`

---

## Frontend

The server also serves a static frontend from the project root.

- `GET /app/` — serves static files via the project root directory
- `GET /assets/logo.png` — serves the Chirpy logo

Both routes are wrapped in a metrics middleware that increments the visit counter.

Implemented in: `main.go:51-52`

---

## Common Errors

All JSON endpoints return errors in this shape:

```json
{
  "error": "human readable message"
}
```

The `Content-Type` is set to `application/json` unless otherwise noted.

### Error status codes

| Status | Meaning |
|--------|---------|
| `400` | Bad request / invalid input |
| `401` | Unauthorized / missing or invalid token |
| `403` | Forbidden / insufficient permissions |
| `404` | Resource not found |
| `500` | Internal server error |

---

## Authentication Helpers

Authorization helpers are located in `internal/auth/auth.go`:

- `HashPassword(password string) (string, error)` — Argon2id password hashing
- `CheckPasswordHash(password, hash string) (bool, error)` — Argon2id verification
- `MakeJWT(userID, secret, expiresIn)` — create a signed JWT
- `ValidateJWT(tokenString, secret)` — validate a JWT and return the user ID
- `GetBearerToken(headers)` — extract the token from `Authorization: Bearer ...`
- `GetAPIKey(headers)` — extract the key from `Authorization: ApiKey ...`
- `MakeRefreshToken()` — generate a random 256-bit refresh token

The JWT is signed with `HS256` and includes the user ID as the subject, with a default access-token lifetime of 1 hour. Refresh tokens expire after 60 days and are stored in the `refresh_tokens` table.
