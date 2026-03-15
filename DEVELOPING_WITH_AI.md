# AI와 함께 EdgeKit 개발하기

이 문서는 AI 코딩 도구(Copilot, Cursor, Claude 등)와 함께 EdgeKit을 효과적으로 개발하기 위한 가이드입니다.

## 아키텍처 개요

EdgeKit은 Hexagonal Architecture (Ports & Adapters)를 따릅니다. 코드를 수정할 때 이 원칙을 지켜주세요:

### Core Layer (`internal/core/`)

**규칙: 외부 프레임워크/DB 패키지 import 금지**

- `auth/` — 인증/인가. `Subject`, `AuthContext`, `TokenService`
- `user/` — 사용자 도메인. `User` 엔티티, `Service`, 포트 인터페이스 (`UserRepository`, `UserCache`, `PasswordHasher`)
- `session/` — 세션 도메인. `Session` 엔티티, `Service`, 포트 인터페이스 (`SessionRepository`)

Core에 새 기능을 추가할 때:
1. `entity.go`에 엔티티와 입/출력 타입 정의
2. `ports.go`에 리포지토리/외부 의존 인터페이스 정의
3. `service.go`에 비즈니스 로직 구현 — 인터페이스만 사용

### Adapter Layer (`internal/adapters/`)

**규칙: Core의 인터페이스를 구현. Core를 import하되, Core가 Adapter를 import하면 안 됨**

- `http/` — Gin 기반 HTTP 어댑터
  - `handler/` — HTTP 핸들러. Core Service를 호출하고 HTTP 응답 반환
  - `middleware/` — 미들웨어 (auth, cors, logging, ratelimit, recovery, requestid)
  - `response/` — 통일된 JSON 응답 형식 (`Response`, `OK`, `Fail`, `Created`)
  - `router.go` — 라우트 정의. 패키지명 `httpadapter`

- `grpc/` — gRPC 어댑터
  - `server.go` — gRPC 서버 생성 + 인터셉터 체이닝. 패키지명 `grpcadapter`
  - `interceptor/` — gRPC 인터셉터 (auth, logging, ratelimit, recovery, requestid)
  - `service/` — gRPC 서비스 구현 (proto → core 매핑)

- `repository/` — 데이터 저장소 어댑터
  - `postgres/` — PostgreSQL 구현 (pgx, raw SQL)
  - `redis/` — Redis 캐시 + 속도 제한
  - `memory/` — 인메모리 구현 (테스트/개발용)

### pkg Layer (`pkg/`)

**규칙: 독립적인 유틸리티 패키지. internal 패키지를 import하면 안 됨**

- `apperror/` — 애플리케이션 에러 타입 + HTTP/gRPC 상태 매핑
- `jwt/` — Ed25519 기반 JWT 토큰 관리
- `logger/` — 로거 인터페이스 + zap 구현
- `ratelimit/` — Limiter 인터페이스 + Config
- `validator/` — 유효성 검사 유틸리티

## 새 도메인 추가하기

예: `notification` 도메인 추가

```
1. internal/core/notification/entity.go    — 엔티티 정의
2. internal/core/notification/ports.go     — 리포지토리 인터페이스
3. internal/core/notification/service.go   — 비즈니스 로직
4. internal/adapters/repository/postgres/notification.go  — DB 구현
5. internal/adapters/http/handler/notification.go         — HTTP 핸들러
6. internal/adapters/http/router.go                       — 라우트 추가
7. internal/app/app.go                                    — DI 조립
```

## 에러 처리 패턴

```go
// Core에서 에러 생성
return nil, apperror.New(apperror.CodeNotFound, "user not found")

// Adapter(HTTP)에서 에러 응답
response.Fail(c, err)  // 자동으로 AppError → HTTP 상태 매핑

// Adapter(gRPC)에서 에러 응답
return nil, apperror.ToGRPCError(apperror.As(err))  // 자동으로 AppError → gRPC 상태 매핑
```

에러 코드: `CodeBadRequest`, `CodeUnauthorized`, `CodeForbidden`, `CodeNotFound`, `CodeConflict`, `CodeRateLimited`, `CodeInternal`, `CodeUnavailable`

## DI (Dependency Injection)

수동 생성자 주입 방식입니다 (wire/fx 미사용).

```go
// internal/app/app.go에서 조립
userRepo := postgres.NewUserRepository(pool)
userCache := redisrepo.NewUserCache(redisClient)
userSvc := user.NewService(userRepo, userCache, hasher, tokenSvc)
```

새 서비스 추가 시 `internal/app/app.go`의 `Run()` 메서드에서 생성자를 호출하고 의존성을 주입하세요.

## 핵심 인터페이스

AI가 코드를 생성할 때 이 인터페이스를 반드시 참고해야 합니다:

```go
// internal/core/user/ports.go
type UserRepository interface {
    Create(ctx context.Context, user *User) error
    FindByID(ctx context.Context, id string) (*User, error)
    FindByEmail(ctx context.Context, email string) (*User, error)
    Update(ctx context.Context, user *User) error
    Delete(ctx context.Context, id string) error
    List(ctx context.Context, offset, limit int) ([]*User, int64, error)
}

type UserCache interface {
    Get(ctx context.Context, id string) (*User, error)
    Set(ctx context.Context, user *User) error
    Invalidate(ctx context.Context, id string) error
}

type PasswordHasher interface {
    Hash(password string) (string, error)
    Verify(password, hash string) error
}
```

```go
// internal/core/session/ports.go
type SessionRepository interface {
    Create(ctx context.Context, session *Session) error
    FindByID(ctx context.Context, id string) (*Session, error)
    Update(ctx context.Context, session *Session) error
    ListByStatus(ctx context.Context, status Status, offset, limit int) ([]*Session, error)
    Delete(ctx context.Context, id string) error
}
```

```go
// pkg/ratelimit/limiter.go
type Limiter interface {
    Allow(ctx context.Context, key string) (Result, error)
}
```

## 코딩 컨벤션

1. **코멘트 최소화** — 코드 자체가 설명이 되도록 작성
2. **에러 래핑** — `fmt.Errorf("context: %w", err)` 패턴 사용
3. **Context 전파** — 모든 I/O 함수에 `context.Context` 첫 번째 매개변수
4. **인터페이스 준수** — Core 포트 인터페이스를 정확히 구현
5. **ORM 미사용** — pgx로 raw SQL 작성
6. **테스트** — stdlib `testing` 패키지만 사용, 수동 mock
