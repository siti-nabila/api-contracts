# gRPC Contracts

Shared Protocol Buffer contracts and generated Go clients for Siti Nabila's gRPC services.

## Packages

- `pb/user/v1`: authentication and user service contract
- `pb/auth/v1`: refresh, logout and password lifecycle contract
- `pb/session/v1`: self/admin session management contract
- `pb/profile/v1`: profile messages
- `pb/paginator/v1`: pagination messages

## Generate

```bash
make proto
go mod tidy
```

## Consume from Go

```bash
go get github.com/siti-nabila/api-contracts
```

```go
import userv1 "github.com/siti-nabila/api-contracts/pb/user/v1"
```

## Error Catalog

Error catalogs under `pkg/dictionary` own the error key, application code,
HTTP status, and localized messages. Locale keys are dynamic BCP-47 tags; new
languages do not require a new Go struct field.

```yaml
errors:
  already_exists:
    code: EX
    http_status: 409
    en: Already exists.
    id: Sudah ada.
    zh: "已存在。"
```

English is the required fallback. A regional request such as `zh-CN` falls
back to `zh` when that base-language message exists. gRPC status mappings are
declared separately in `pkg/grpcerror`; they are not derived from HTTP status
and are not stored in the YAML catalog.

Transaction consumers register `mapping.AuthCodeMappings()` and
`mapping.UserCodeMappings()` alongside `mapping.CommonCodeMappings()`.
`auth.ErrResourceBusy` and `user.ErrResourceBusy` have distinct qualified keys
and map to gRPC ABORTED / HTTP 409. Use `common.ErrDependencyUnavailable` for
an unavailable dependency (UNAVAILABLE / HTTP 503); the legacy
`common.ErrServiceUnavailable` mapping remains unchanged. Request cancellation
(including wrapped context errors and raw gRPC CANCELED) is localized as
`common.request_canceled`; deadline failures retain DEADLINE_EXCEEDED. Unknown
errors remain generic INTERNAL responses.

Publish the tested contract version before pinning it in grpc-auth. Parent
workspace verification uses this checkout and does not prove a published
version is available; consumers must also verify with `GOWORK=off` after
publishing.

## Session authorization foundation contract

The additive fields and security errors for grpc-auth are specified in
[docs/auth-session-contract.md](docs/auth-session-contract.md). The six existing
RPC names, old field numbers/types, numeric role codes, and cursor pagination
remain compatible. Publishing these contracts does not activate PENDING
registration, session authentication, or change legacy Login behavior by itself.
grpc-auth and rest-orc must complete their separate producer/caller/cutover gates.

## Session lifecycle and user listing contract

See [docs/session-lifecycle-contract.md](docs/session-lifecycle-contract.md) for
the nine additive lifecycle RPCs, credential modes, errors and ListUsers sorting,
filter/search/cursor semantics and additive created_at. These contracts require
separate backend implementation and Phase6 deployment readiness.

User management: [contract and compatibility](docs/user-management-contract.md).
