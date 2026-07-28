# gRPC Contracts

Shared Protocol Buffer contracts and generated Go clients for Siti Nabila's gRPC services.

## Packages

- `pb/user`: authentication and user service contract
- `pb/profile`: profile messages
- `pb/paginator`: pagination messages

## Generate

```bash
make proto
go mod tidy
```

## Consume from Go

```bash
go get github.com/siti-nabila/grpc-contracts
```

```go
import userpb "github.com/siti-nabila/grpc-contracts/pb/user"
```
