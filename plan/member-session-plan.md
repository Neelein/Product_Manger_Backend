# Member & Session Feature Plan

## Overview

Implement membership (會員功能) with two database tables and cookie-based session authentication.

- `members` — member profile data (email, password, name)
- `sessions` — session keys bound to members with 1-day expiry

Session key is stored in `HttpOnly` cookie. Expired sessions return 401 — re-login required.

---

## Table Definitions

### `members`

| Column | Type | Constraints |
|---|---|---|
| id | UUID | PK, DEFAULT gen_random_uuid() |
| email | VARCHAR(255) | NOT NULL, UNIQUE |
| password | VARCHAR(255) | NOT NULL (bcrypt hash) |
| name | VARCHAR(255) | NOT NULL |
| created_at | TIMESTAMPTZ | NOT NULL DEFAULT now() |
| updated_at | TIMESTAMPTZ | NOT NULL DEFAULT now() |

### `sessions`

| Column | Type | Constraints |
|---|---|---|
| id | UUID | PK, DEFAULT gen_random_uuid() |
| member_id | UUID | FK → members(id) ON DELETE CASCADE, NOT NULL |
| session_key | VARCHAR(255) | NOT NULL UNIQUE, DEFAULT gen_random_uuid()::text |
| created_at | TIMESTAMPTZ | NOT NULL DEFAULT now() |
| expires_at | TIMESTAMPTZ | NOT NULL DEFAULT now() + interval '1 day' |

Indexes: `(member_id)`, `(session_key)`

---

## Domain Model (`src/domain/member.go`)

```go
type Member struct {
    ID        string    `json:"id"`
    Email     string    `json:"email"`
    Password  string    `json:"-"`
    Name      string    `json:"name"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

type Session struct {
    ID         string    `json:"id"`
    MemberID   string    `json:"member_id"`
    SessionKey string    `json:"session_key"`
    CreatedAt  time.Time `json:"created_at"`
    ExpiresAt  time.Time `json:"expires_at"`
}
```

### Request / Response Types

```go
type RegisterRequest struct {
    Email    string `json:"email"`
    Password string `json:"password"`
    Name     string `json:"name"`
}

type LoginRequest struct {
    Email    string `json:"email"`
    Password string `json:"password"`
}

type LoginResponse struct {
    Member MemberResponse `json:"member"`
}

type MemberResponse struct {
    ID    string `json:"id"`
    Email string `json:"email"`
    Name  string `json:"name"`
}
```

---

## Domain Errors (`src/domain/errors.go`)

| Sentinel Error | Condition |
|---|---|
| `ErrEmailAlreadyExists` | Register with duplicate email |
| `ErrInvalidCredentials` | Login with wrong email/password |
| `ErrSessionNotFound` | Cookie key not in DB |
| `ErrSessionExpired` | Session key found but expires_at < now() |

---

## Repository Interfaces (`src/domain/repository.go`)

```go
type MemberRepository interface {
    Create(ctx context.Context, member *Member) error
    GetByEmail(ctx context.Context, email string) (*Member, error)
    GetByID(ctx context.Context, id string) (*Member, error)
}

type SessionRepository interface {
    Create(ctx context.Context, session *Session) error
    GetByKey(ctx context.Context, sessionKey string) (*Session, error)
    Delete(ctx context.Context, sessionKey string) error
    DeleteByMemberID(ctx context.Context, memberID string) error
}
```

---

## Repository Implementations

### `src/database/member_repo.go` — `MemberRepositoryPGX`

| Method | SQL |
|---|---|
| Create | `INSERT INTO members (email, password, name) VALUES ($1, $2, $3) RETURNING id, created_at, updated_at` |
| GetByEmail | `SELECT * FROM members WHERE email = $1` |
| GetByID | `SELECT * FROM members WHERE id = $1` |

Duplicate email check: `pgx` unique violation error code `23505` → map to `ErrEmailAlreadyExists`.

### `src/database/session_repo.go` — `SessionRepositoryPGX`

| Method | SQL |
|---|---|
| Create | `INSERT INTO sessions (member_id) VALUES ($1) RETURNING id, session_key, created_at, expires_at` |
| GetByKey | `SELECT * FROM sessions WHERE session_key = $1` |
| Delete | `DELETE FROM sessions WHERE session_key = $1` |
| DeleteByMemberID | `DELETE FROM sessions WHERE member_id = $1` |

---

## API Handlers (`src/api/member_handler.go`)

| Method | Path | Handler | Description |
|---|---|---|---|
| POST | `/api/members/register` | `RegisterMember` | 建立會員，不回傳 session |
| POST | `/api/members/login` | `LoginMember` | 驗證密碼 → 建立 session → Set-Cookie |
| POST | `/api/members/logout` | `LogoutMember` | 刪除 session → 清除 cookie |
| GET | `/api/members/me` | `GetCurrentMember` | 讀 cookie → 驗證 session → 回傳會員資料 |

### Session Validation Flow (for `GetCurrentMember`)

```
1. r.Cookie("session_key") → 無 cookie → 401
2. SessionRepository.GetByKey(ctx, key) → 無此 key → 401
3. session.ExpiresAt.Before(time.Now()) → 過期 → 401
4. MemberRepository.GetByID(ctx, session.MemberID) → 回傳 member
```

### Cookie Settings

```go
http.Cookie{
    Name:     "session_key",
    Value:    session.SessionKey,
    Path:     "/",
    HttpOnly: true,
    Secure:   true,
    SameSite: http.SameSiteLaxMode,
    Expires:  session.ExpiresAt,
}
```

Logout clears by setting `MaxAge: -1` (or `Expires: time.Unix(0, 0)`).

---

## Routes (`src/api/router.go`)

```go
func RegisterMemberRoutes(r *mux.Router, memberRepo domain.MemberRepository, sessionRepo domain.SessionRepository)
```

Registered in `main.go` alongside existing `RegisterProductRoutes`.

---

## New Dependency

```bash
go get golang.org/x/crypto
```

Used for `bcrypt.GenerateFromPassword` (register) and `bcrypt.CompareHashAndPassword` (login).

---

## Implementation Order

1. SQL migration file (`db/migrations/002_create_members.sql`)
2. Add domain structs to `src/domain/member.go`
3. Add sentinel errors to `src/domain/errors.go`
4. Add repository interfaces to `src/domain/repository.go`
5. Implement `src/database/member_repo.go`
6. Implement `src/database/session_repo.go`
7. Implement `src/api/member_handler.go`
8. Update `src/api/router.go` with member routes
9. Update `main.go` to wire member + session repos
10. Add `golang.org/x/crypto` dependency
11. Tests: `member_repo_test.go`, `session_repo_test.go`, `member_handler_test.go`

---

## Test Plan

### Unit Tests (no build tag)

- `TestMember_RegisterRequest` — validate JSON tags and field validation

### Integration Tests (`//go:build integration`)

- `src/test/database/member_repo_test.go`
  - `TestMemberRepositoryPGX_Create` — create, get by email, get by ID
  - `TestMemberRepositoryPGX_Create_DuplicateEmail` — expect ErrEmailAlreadyExists
  - `TestMemberRepositoryPGX_GetByEmail_NotFound` — expect nil
- `src/test/database/session_repo_test.go`
  - `TestSessionRepositoryPGX_Create` — create session, validate expires_at ≈ now + 24h
  - `TestSessionRepositoryPGX_GetByKey` — retrieve by key
  - `TestSessionRepositoryPGX_GetByKey_Expired` — past expires_at
  - `TestSessionRepositoryPGX_Delete` — delete after create
- `src/test/api/member_handler_test.go`
  - `TestHandler_Register` — POST /api/members/register
  - `TestHandler_Register_DuplicateEmail` — 409 Conflict
  - `TestHandler_Login` — POST /api/members/login → Set-Cookie header
  - `TestHandler_Login_WrongPassword` — 401
  - `TestHandler_Me` — GET /api/members/me with valid cookie
  - `TestHandler_Me_NoCookie` — 401
  - `TestHandler_Me_ExpiredSession` — 401
  - `TestHandler_Logout` — POST /api/members/logout → cleared cookie
