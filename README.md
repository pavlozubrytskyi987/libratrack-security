# LibraTrack Security

A Go/Gin REST API for library management with JWT authentication and role-based access control.

## Features

- User registration with bcrypt password hashing
- JWT token authentication (HS256, 24h expiry)
- RBAC middleware with 3 roles: reader, librarian, admin
- Book and loan management endpoints
- User enumeration prevention on login

## How to run

```bash
export JWT_SECRET="your-secret-key"
go run ./cmd/server
```

Server starts on `:8080`.

## API endpoints

| Method | Path | Auth | Roles |
|--------|------|------|-------|
| POST | /api/v1/auth/register | - | public |
| POST | /api/v1/auth/login | - | public |
| GET | /api/v1/books | - | public |
| GET | /api/v1/me/loans | JWT | any |
| POST | /api/v1/books | JWT | librarian, admin |
| PUT | /api/v1/books/:id | JWT | librarian, admin |
| POST | /api/v1/loans | JWT | librarian, admin |
| POST | /api/v1/loans/:id/return | JWT | librarian, admin |
| GET | /api/v1/users | JWT | admin |
| PUT | /api/v1/users/:id/role | JWT | admin |

## Quick test

```bash
# register
curl -X POST localhost:8080/api/v1/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"email":"test@example.com","password":"password123"}'

# login
curl -X POST localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"test@example.com","password":"password123"}'

# use the token
TOKEN="<paste token here>"
curl localhost:8080/api/v1/me/loans -H "Authorization: Bearer $TOKEN"
```

## Tests

```bash
go test -v ./...
```

## Project structure

```
cmd/server/         entry point with route groups
internal/auth/      bcrypt + JWT
internal/handler/   request handlers
internal/middleware/ auth + RBAC middleware
internal/model/     data models + DTOs
internal/store/     in-memory storage
migrations/         SQL migration files
```
