# rancago-cli

> Standalone CLI for the [RANCAGO Framework](https://github.com/rancago/rancago) - code generators, scaffolding, and development utilities.

[![Go Version](https://img.shields.io/badge/Go-1.23%2B-00ADD8?style=flat-square&logo=go)](https://go.dev)
[![Module](https://img.shields.io/badge/module-github.com%2Francago%2Francago--cli-blue?style=flat-square)](https://github.com/rancago/rancago-cli)

This binary is **independent from the framework** - zero dependency on `rancago/framework`. It generates files in whatever Go project you run it from, reading the module name from the local `go.mod`.

---

## About RANCAGO

**RANCAGO** = **R**esilient, **A**gnostic, & **N**ative **C**lean-**A**rchitecture **G**O Framework.

The name also draws from Sundanese roots:

- **Rancagé** — *skilled, precise, structured craftsmanship* → how Rancago enforces SOLID and hexagonal architecture step by step.
- **Ranca** — *a fertile, expansive wetland ecosystem* → the rich built-in feature set: pgvector, Redis, MinIO, Google Drive, Meet, Calendar, OAuth, RBAC.

---

## Install

```bash
go install github.com/rancago/rancago-cli@latest
```

Or build from source:

```bash
git clone https://github.com/rancago/rancago-cli.git
cd rancago-cli
go build -o rancago .
```

---

## Commands

```
rancago [command] [flags] [args]

SERVE
  serve                   Start HTTP (and optional gRPC) development server
    --port, -p  int       HTTP port (default: 8080)
    --grpc                Also start gRPC stub server

CODE GENERATORS
  make:entity       [name]          Domain entity → internal/domain/entities/
  make:value-object [name]          Value object  → internal/domain/valueobjects/
  make:port         [name]          Port interface → internal/ports/driven/
                    --driving       Inbound port   → internal/ports/driving/
  make:usecase      [name]          Use case interactor → internal/application/usecases/
  make:adapter      [name]          Adapter stub
                    --direction driving|driven
  make:model        [name]          GORM model → app/Models/
                    -m              Also generate migration
  make:migration    [name]          Migration file → database/migrations/
  make:controller   [name]          HTTP controller → app/Http/Controllers/
                    -r              Resourceful (Index/Show/Store/Update/Destroy)
  make:middleware   [name]          HTTP middleware → app/Http/Middleware/
  make:provider     [name]          ServiceProvider stub → app/Providers/

SCAFFOLD
  scaffold [name]                   Interactive bounded-context scaffolder
                                    Generates entity + port + use case + adapter in one go
    --no-interactive                Skip prompts; use --entity/--repo/--usecase/--http/--grpc flags

MIGRATIONS
  migrate                           Run pending migrations (stub - wire a GORM adapter)
  migrate --rollback                Rollback last batch

UTILITIES
  key:generate                      Generate a cryptographically secure APP_KEY
  storage:link                      Symlink public/storage → storage/app/public
  route:list                        Print all registered routes (HTTP + gRPC + WS)
  tinker                            Interactive REPL - explore container bindings and ports

  help                              Show this help
  version / -v                      Show version
```

---

## Usage examples

### Generate a bounded context in one shot

```bash
# Interactive
rancago scaffold Order

# Non-interactive
rancago scaffold Order --no-interactive --entity --repo --usecase --http
```

This creates:

```
internal/domain/entities/Order.go
internal/ports/driven/OrderRepository_port.go
internal/ports/driving/OrderUseCase_port.go
internal/application/usecases/Order_usecase.go
internal/adapters/driving/orderhandler/orderhandler_adapter.go
```

### Domain entities & value objects

```bash
rancago make:entity Product
rancago make:value-object Money
rancago make:value-object SKU
```

### Use cases & ports

```bash
rancago make:port ProductRepository          # driven (outbound)
rancago make:port ProductUseCase --driving   # driving (inbound)
rancago make:usecase Product
```

### Adapters

```bash
rancago make:adapter ProductHandler --direction driving   # HTTP handler
rancago make:adapter PostgresProduct --direction driven   # DB adapter
```

### GORM models & migrations

```bash
rancago make:model Product -m     # model + migration file
rancago make:migration add_slug_to_products
```

### HTTP controllers & middleware

```bash
rancago make:controller UserController -r   # resourceful: Index/Show/Store/Update/Destroy
rancago make:middleware AuthMiddleware
```

### ServiceProvider stub

```bash
rancago make:provider PaymentServiceProvider
```

### Secure key generation

```bash
rancago key:generate
# APP_KEY=base64:abc123...
```

---

## How it works

All `make:*` and `scaffold` commands read the module name from `go.mod` in the current working directory and generate idiomatic Go files following the Rancago hexagonal architecture layout. No network access, no side effects beyond writing files.

The `serve` command delegates to the project's own `go run .` - the CLI does not embed a server.

---

## Related

- [rancago/rancago](https://github.com/rancago/rancago) - the full framework

---

## License

Proprietary - Muhammad Ikhwan Fathulloh © 2026. rancago-cli 1.0.0.
