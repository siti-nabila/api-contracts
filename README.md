# gRPC Contracts

Shared Protocol Buffer contracts and generated Go clients for Siti Nabila's gRPC services.

## Packages

- `pb/user/v1`: authentication and user service contract
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
