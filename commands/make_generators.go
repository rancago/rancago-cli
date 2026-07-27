package commands

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"
)

// ---- make:entity ----

func NewMakeEntityCommand() *SimpleCommand {
	fs := flag.NewFlagSet("make:entity", flag.ContinueOnError)
	return &SimpleCommand{
		use: "make:entity [name]", short: "Create a domain entity",
		argsMin: 1, argsMax: 1, flags: fs,
		runFn: func(_ *flag.FlagSet, args []string) error {
			name := args[0]
			pascal := toPascal(name)
			mod := ModuleName()
			gen := Generator{Name: name, Type: "Domain Entity", BasePath: "internal/domain/entities", Package: "entities"}
			return gen.writeFile(".go", `package entities

import (
	"time"

	"`+mod+`/internal/domain/valueobjects"
)

type `+pascal+` struct {
	ID        valueobjects.ID
	CreatedAt time.Time
	UpdatedAt time.Time
}

func New`+pascal+`() *`+pascal+` {
	now := time.Now()
	return &`+pascal+`{CreatedAt: now, UpdatedAt: now}
}
`)
		},
	}
}

// ---- make:value-object ----

func NewMakeValueObjectCommand() *SimpleCommand {
	fs := flag.NewFlagSet("make:value-object", flag.ContinueOnError)
	return &SimpleCommand{
		use: "make:value-object [name]", short: "Create a value object",
		argsMin: 1, argsMax: 1, flags: fs,
		runFn: func(_ *flag.FlagSet, args []string) error {
			name := args[0]
			pascal := toPascal(name)
			lower := strings.ToLower(pascal)
			gen := Generator{Name: name, Type: "Value Object", BasePath: "internal/domain/valueobjects", Package: "valueobjects"}
			return gen.writeFile(".go", `package valueobjects

import "fmt"

type `+pascal+` struct {
	value string
}

func New`+pascal+`(raw string) (`+pascal+`, error) {
	if raw == "" {
		return `+pascal+`{}, fmt.Errorf("`+lower+` cannot be empty")
	}
	return `+pascal+`{value: raw}, nil
}

func Must`+pascal+`(raw string) `+pascal+` {
	v, err := New`+pascal+`(raw)
	if err != nil { panic(err) }
	return v
}

func (v `+pascal+`) String() string              { return v.value }
func (v `+pascal+`) IsEmpty() bool               { return v.value == "" }
func (v `+pascal+`) Equals(o `+pascal+`) bool    { return v.value == o.value }
`)
		},
	}
}

// ---- make:port ----

func NewMakePortCommand() *SimpleCommand {
	fs := flag.NewFlagSet("make:port", flag.ContinueOnError)
	isDriving := fs.Bool("driving", false, "Generate as driving (inbound) port")
	return &SimpleCommand{
		use: "make:port [name]", short: "Create a port interface",
		argsMin: 1, argsMax: 1, flags: fs,
		runFn: func(f *flag.FlagSet, args []string) error {
			name := args[0]
			pascal := toPascal(name)
			portType, base := "driven", "internal/ports/driven"
			if *isDriving {
				portType, base = "driving", "internal/ports/driving"
			}
			mod := ModuleName()
			importBlock := `"context"`
			if *isDriving {
				importBlock = `"context"

	"` + mod + `/internal/domain/entities"
	"` + mod + `/internal/domain/valueobjects"`
			}
			gen := Generator{Name: name + "_port", Type: "Port (" + portType + ")", BasePath: base, Package: portType}
			return gen.writeFile(".go", `package `+portType+`

import (
	`+importBlock+`
)

type `+pascal+` interface {
	Example(ctx context.Context) error
}
`)
		},
	}
}

// ---- make:usecase ----

func NewMakeUsecaseCommand() *SimpleCommand {
	fs := flag.NewFlagSet("make:usecase", flag.ContinueOnError)
	return &SimpleCommand{
		use: "make:usecase [name]", short: "Create a use case interactor",
		argsMin: 1, argsMax: 1, flags: fs,
		runFn: func(_ *flag.FlagSet, args []string) error {
			name := args[0]
			pascal := toPascal(name)
			if !strings.HasSuffix(strings.ToLower(pascal), "usecase") &&
				!strings.HasSuffix(strings.ToLower(pascal), "interactor") {
				pascal += "Interactor"
			}
			mod := ModuleName()
			baseName := strings.TrimSuffix(strings.ToLower(pascal), "interactor") + "_usecase"
			gen := Generator{Name: baseName, Type: "Use Case", BasePath: "internal/application/usecases", Package: "usecases"}
			return gen.writeFile(".go", `package usecases

import (
	"context"

	"`+mod+`/internal/domain/entities"
	"`+mod+`/internal/domain/valueobjects"
	derrors "`+mod+`/internal/domain/errors"
	"`+mod+`/internal/ports/driven"
	"`+mod+`/internal/ports/driving"
)

type `+pascal+` struct{}

func New`+pascal+`() driving.XxxUseCase {
	return &`+pascal+`{}
}

func (uc *`+pascal+`) Example(ctx context.Context, id valueobjects.ID) (*entities.Entity, error) {
	_ = derrors.ErrNotFound
	_ = driven.Repository[entities.Entity](nil)
	return nil, nil
}
`)
		},
	}
}

// ---- make:adapter ----

func NewMakeAdapterCommand() *SimpleCommand {
	fs := flag.NewFlagSet("make:adapter", flag.ContinueOnError)
	direction := fs.String("direction", "driven", "driving|driven")
	return &SimpleCommand{
		use: "make:adapter [name]", short: "Create an infrastructure adapter",
		argsMin: 1, argsMax: 1, flags: fs,
		runFn: func(f *flag.FlagSet, args []string) error {
			name := args[0]
			pascal := toPascal(name)
			snake := toSnake(name)
			dir := *direction
			if dir != "driving" {
				dir = "driven"
			}
			mod := ModuleName()
			base := "internal/adapters/" + dir + "/" + snake
			gen := Generator{Name: snake + "_adapter", Type: "Adapter (" + dir + ")", BasePath: base, Package: snake}
			return gen.writeFile(".go", `package `+snake+`

import (
	"context"

	"`+mod+`/internal/ports/`+dir+`"
)

type `+pascal+`Adapter struct{}

func New`+pascal+`Adapter() `+dir+`.`+pascal+` {
	return &`+pascal+`Adapter{}
}

func (a *`+pascal+`Adapter) Example(ctx context.Context) error {
	return nil
}
`)
		},
	}
}

// ---- make:model ----

func NewMakeModelCommand() *SimpleCommand {
	fs := flag.NewFlagSet("make:model", flag.ContinueOnError)
	withMig := fs.Bool("migration", false, "Also generate a migration")
	fs.BoolVar(withMig, "m", false, "Also generate a migration (short)")
	return &SimpleCommand{
		use: "make:model [name]", short: "Create a GORM model",
		argsMin: 1, argsMax: 1, flags: fs,
		runFn: func(f *flag.FlagSet, args []string) error {
			name := args[0]
			pascal := toPascal(name)
			table := toSnake(pascal) + "s"
			gen := Generator{Name: name, Type: "Model", BasePath: "app/Models", Package: "Models"}
			content := `package Models

import (
	"time"
	"gorm.io/gorm"
)

type ` + pascal + ` struct {
	ID        uint           ` + "`gorm:\"primaryKey\" json:\"id\"`" + `
	CreatedAt time.Time      ` + "`json:\"created_at\"`" + `
	UpdatedAt time.Time      ` + "`json:\"updated_at\"`" + `
	DeletedAt gorm.DeletedAt ` + "`gorm:\"index\" json:\"deleted_at,omitempty\"`" + `
}

func (` + pascal + `) TableName() string { return "` + table + `" }
`
			if err := gen.writeFile(".go", content); err != nil {
				return err
			}
			if *withMig {
				return newMigration("create_" + table + "_table")
			}
			return nil
		},
	}
}

// ---- make:migration ----

func NewMakeMigrationCommand() *SimpleCommand {
	fs := flag.NewFlagSet("make:migration", flag.ContinueOnError)
	return &SimpleCommand{
		use: "make:migration [name]", short: "Create a migration file",
		argsMin: 1, argsMax: 1, flags: fs,
		runFn: func(_ *flag.FlagSet, args []string) error { return newMigration(args[0]) },
	}
}

func newMigration(name string) error {
	dir := "database/migrations"
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	ts := time.Now().Format("20060102150405")
	snake := toSnake(name)
	path := fmt.Sprintf("%s/%s_%s.go", dir, ts, snake)
	content := `package migrations

// Migration: ` + snake + `
// Generated: ` + time.Now().Format(time.RFC3339) + `

func Up() []string {
	return []string{
		"-- TODO: write UP SQL for ` + snake + `",
	}
}

func Down() []string {
	return []string{
		"-- TODO: write DOWN SQL for ` + snake + `",
	}
}
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return err
	}
	fmt.Printf("  ✅ Created Migration: %s\n", path)
	return nil
}

// ---- make:controller ----

func NewMakeControllerCommand() *SimpleCommand {
	fs := flag.NewFlagSet("make:controller", flag.ContinueOnError)
	resourceful := fs.Bool("resource", false, "Generate resource controller (Index/Show/Store/Update/Destroy)")
	fs.BoolVar(resourceful, "r", false, "Resource controller (short)")
	return &SimpleCommand{
		use: "make:controller [name]", short: "Create an HTTP controller",
		argsMin: 1, argsMax: 1, flags: fs,
		runFn: func(f *flag.FlagSet, args []string) error {
			name := args[0]
			pascal := toPascal(name)
			gen := Generator{Name: name, Type: "Controller", BasePath: "app/Http/Controllers", Package: "Controllers"}
			var content string
			if *resourceful {
				content = `package Controllers

import (
	"encoding/json"
	"net/http"
)

// ` + pascal + ` handles resourceful HTTP routes for ` + pascal + `.
type ` + pascal + ` struct{}

func New` + pascal + `() *` + pascal + ` { return &` + pascal + `{} }

// Index  GET /
func (c *` + pascal + `) Index(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, []interface{}{})
}

// Show   GET /{id}
func (c *` + pascal + `) Show(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{"id": r.PathValue("id")})
}

// Store  POST /
func (c *` + pascal + `) Store(w http.ResponseWriter, r *http.Request) {
	var body map[string]interface{}
	_ = json.NewDecoder(r.Body).Decode(&body)
	writeJSON(w, http.StatusCreated, body)
}

// Update PUT /{id}
func (c *` + pascal + `) Update(w http.ResponseWriter, r *http.Request) {
	var body map[string]interface{}
	_ = json.NewDecoder(r.Body).Decode(&body)
	body["id"] = r.PathValue("id")
	writeJSON(w, http.StatusOK, body)
}

// Destroy DELETE /{id}
func (c *` + pascal + `) Destroy(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{"deleted": r.PathValue("id")})
}

func writeJSON(w http.ResponseWriter, code int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}
`
			} else {
				content = `package Controllers

import (
	"encoding/json"
	"net/http"
)

// ` + pascal + ` handles HTTP requests for ` + pascal + `.
type ` + pascal + ` struct{}

func New` + pascal + `() *` + pascal + ` { return &` + pascal + `{} }

func (c *` + pascal + `) Handle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"message": "` + pascal + ` ok"})
}
`
			}
			return gen.writeFile(".go", content)
		},
	}
}

// ---- make:middleware ----

func NewMakeMiddlewareCommand() *SimpleCommand {
	fs := flag.NewFlagSet("make:middleware", flag.ContinueOnError)
	return &SimpleCommand{
		use: "make:middleware [name]", short: "Create an HTTP middleware",
		argsMin: 1, argsMax: 1, flags: fs,
		runFn: func(_ *flag.FlagSet, args []string) error {
			name := args[0]
			pascal := toPascal(name)
			gen := Generator{Name: name, Type: "Middleware", BasePath: "app/Http/Middleware", Package: "Middleware"}
			return gen.writeFile(".go", `package Middleware

import (
	"log"
	"net/http"
	"time"
)

// `+pascal+` is an HTTP middleware.
func `+pascal+`(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Printf("[`+pascal+`] %s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
		log.Printf("[`+pascal+`] done in %s", time.Since(start))
	})
}
`)
		},
	}
}

// ---- make:provider ----

func NewMakeProviderCommand() *SimpleCommand {
	fs := flag.NewFlagSet("make:provider", flag.ContinueOnError)
	return &SimpleCommand{
		use: "make:provider [name]", short: "Create a ServiceProvider stub",
		argsMin: 1, argsMax: 1, flags: fs,
		runFn: func(_ *flag.FlagSet, args []string) error {
			name := args[0]
			pascal := toPascal(name)
			gen := Generator{Name: name, Type: "ServiceProvider", BasePath: "app/Providers", Package: "Providers"}
			return gen.writeFile(".go", `package Providers

import (
	"github.com/rancago/framework/app/Contracts"
	"github.com/rancago/framework/framework/Container"
)

// `+pascal+` registers and boots the `+pascal+` module.
type `+pascal+` struct{}

func New`+pascal+`() Contracts.ServiceProvider {
	return &`+pascal+`{}
}

// Register binds services into the container. Keep this lightweight.
func (p *`+pascal+`) Register(c *Container.Container) error {
	// c.Singleton("service.xxx", func(c *Container.Container) (interface{}, error) {
	//     return Services.NewXxxService(), nil
	// })
	// c.Alias("service.xxx", "Contracts.XxxService")
	return nil
}

// Boot runs after all providers have registered. Wire drivers, seed data, etc.
func (p *`+pascal+`) Boot(c *Container.Container) error {
	return nil
}
`)
		},
	}
}
