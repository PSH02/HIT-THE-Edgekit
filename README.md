# EdgeKit

Go 기반 백엔드 API 스타터 킷 + CLI 프로젝트 생성기.

게임 백엔드, 웹 백엔드, 모바일 앱 백엔드 등 여러 종류의 서비스에 공통으로 사용하는 기본 뼈대입니다.

## Architecture

Hexagonal (Ports & Adapters) Architecture를 따릅니다.

```
internal/
├── core/           # 비즈니스 로직 (프레임워크/DB 의존 없음)
│   ├── auth/       # 인증/인가 도메인
│   ├── user/       # 사용자 도메인
│   └── session/    # 세션 도메인
├── adapters/       # 외부 시스템 어댑터
│   ├── http/       # Gin HTTP 핸들러/미들웨어/라우터
│   ├── grpc/       # gRPC 서비스/인터셉터
│   └── repository/ # 데이터 저장소 (postgres, redis, memory)
└── app/            # 애플리케이션 부트스트랩, DI, 설정
```

**Core (내부)**: 순수 비즈니스 로직. 외부 라이브러리 의존 없음.

**Adapters (외부)**: Core의 인터페이스를 구현하는 어댑터. HTTP/gRPC/DB 등 외부 시스템과 연결.

**App**: Config 로딩, DI 조립, 서버 시작/종료.

## Tech Stack

| Category | Technology |
|----------|-----------|
| Language | Go 1.23 |
| HTTP | Gin |
| gRPC | gRPC-Go |
| Database | PostgreSQL (pgx, raw SQL) |
| Cache | Redis (go-redis) |
| Auth | JWT (Ed25519) |
| Config | Viper |
| Logging | zap |
| CLI | Cobra + Huh |

## Quick Start

### Prerequisites

- Go 1.23+
- PostgreSQL 16+
- Redis 7+
- buf (for proto generation)

### Using Docker Compose

```bash
# 전체 인프라 + 서버 실행
make docker-up

# 종료
make docker-down
```

### Manual Setup

```bash
# 의존성 설치
go mod tidy

# JWT 키 생성
openssl genpkey -algorithm Ed25519 -out configs/jwt.key
openssl pkey -in configs/jwt.key -pubout -out configs/jwt.key.pub

# 데이터베이스 마이그레이션
export EDGEKIT_DATABASE_URL="postgres://edgekit:edgekit@localhost:5432/edgekit?sslmode=disable"
make migrate

# Proto 생성
make proto

# 빌드 & 실행
make run
```

### Environment Variables

환경 변수는 `EDGEKIT_` 접두사를 사용합니다. YAML 설정보다 환경 변수가 우선합니다.

| Variable | Default | Description |
|----------|---------|-------------|
| `EDGEKIT_APP_ENV` | local | 환경 (local/dev/staging/prod) |
| `EDGEKIT_HTTP_ADDR` | :8080 | HTTP 서버 주소 |
| `EDGEKIT_GRPC_ADDR` | :50051 | gRPC 서버 주소 |
| `EDGEKIT_DATABASE_URL` | - | PostgreSQL 연결 URL |
| `EDGEKIT_REDIS_URL` | - | Redis 연결 URL |
| `EDGEKIT_LOG_LEVEL` | info | 로그 레벨 (debug/info/warn/error) |

## API Endpoints

### Public

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/auth/register` | 회원가입 |
| POST | `/api/v1/auth/login` | 로그인 |
| GET | `/healthz` | Health check |
| GET | `/readyz` | Readiness check |

### Protected (Bearer Token)

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/users/me` | 내 프로필 조회 |
| PATCH | `/api/v1/users/me` | 프로필 수정 |
| POST | `/api/v1/sessions` | 세션 생성 |
| GET | `/api/v1/sessions` | 대기중 세션 목록 |
| GET | `/api/v1/sessions/:id` | 세션 상세 조회 |
| POST | `/api/v1/sessions/:id/join` | 세션 참가 |
| POST | `/api/v1/sessions/:id/leave` | 세션 퇴장 |

## gRPC Services

```protobuf
// user.v1.UserService
rpc Register(RegisterRequest) returns (RegisterResponse);
rpc Login(LoginRequest) returns (LoginResponse);
rpc GetProfile(GetProfileRequest) returns (GetProfileResponse);

// session.v1.SessionService
rpc CreateSession(CreateSessionRequest) returns (SessionResponse);
rpc GetSession(GetSessionRequest) returns (SessionResponse);
rpc JoinSession(JoinSessionRequest) returns (SessionResponse);
rpc LeaveSession(LeaveSessionRequest) returns (SessionResponse);
```

## CLI Generator

```bash
# 새 프로젝트 생성
go run ./cmd/edgekit new my-service

# 대화형 프롬프트로 기능 선택
# - HTTP API / gRPC API
# - PostgreSQL / Redis
# - JWT Auth / Rate Limiting
```

## Project Structure

```
.
├── cmd/
│   ├── server/         # 서버 엔트리포인트
│   └── edgekit/        # CLI 프로젝트 생성기
├── internal/
│   ├── core/           # 비즈니스 로직
│   ├── adapters/       # 외부 어댑터
│   └── app/            # 부트스트랩
├── pkg/                # 공용 패키지
│   ├── apperror/       # 에러 처리
│   ├── jwt/            # JWT 토큰 관리
│   ├── logger/         # 구조화 로깅
│   ├── ratelimit/      # 속도 제한
│   └── validator/      # 유효성 검사
├── proto/              # Protobuf 정의
├── configs/            # 설정 파일
├── deploy/             # 배포 설정 (Docker, k8s)
├── scripts/            # 유틸리티 스크립트
└── Makefile
```

## Middleware / Interceptors

### HTTP Middleware (Gin)

- **RequestID** — X-Request-ID 자동 생성/전파
- **Recovery** — 패닉 복구
- **CORS** — Cross-Origin 설정
- **Logging** — 요청/응답 구조화 로깅
- **RateLimit** — 속도 제한 (in-memory / Redis)
- **Auth** — JWT Bearer 토큰 검증

### gRPC Interceptors

- **Recovery** — 패닉 복구
- **RequestID** — 요청 ID 생성
- **Logging** — 메서드/시간/상태 로깅
- **RateLimit** — 속도 제한
- **Auth** — 메타데이터 기반 JWT 검증

## Development

```bash
make build      # 빌드
make run        # 실행
make test       # 테스트
make lint       # 린트
make fmt        # 포맷팅
make vet        # go vet
make tidy       # go mod tidy
make proto      # proto 재생성
make migrate    # DB 마이그레이션
```

## License

MIT
