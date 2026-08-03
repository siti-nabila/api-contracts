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
