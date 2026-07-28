package commands

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"
)

// featureEntry describes a generated file for the docs/features/<name>.md table.
type featureEntry struct{ layer, path, role string }

// NewMakeFeatureCommand scaffolds a complete hexagonal feature and generates
// docs/features/<name>.md as a compact context file for future edits.
func NewMakeFeatureCommand() *SimpleCommand {
	fs := flag.NewFlagSet("make:feature", flag.ContinueOnError)
	noInteractive := fs.Bool("no-interactive", false, "Skip prompts; use explicit flags")
	entity := fs.Bool("entity", true, "Generate domain entity")
	repo := fs.Bool("repo", true, "Generate driven repository port")
	usecase := fs.Bool("usecase", true, "Generate use case port + interactor")
	httpFlag := fs.Bool("http", true, "Generate HTTP driving adapter")
	grpcFlag := fs.Bool("grpc", false, "Generate gRPC driving adapter")
	desc := fs.String("desc", "", "Short description (for docs/features/<name>.md)")

	return &SimpleCommand{
		use:     "make:feature [name]",
		short:   "Scaffold a feature + generate docs/features/<name>.md",
		argsMin: 0,
		argsMax: 1,
		flags:   fs,
		runFn: func(f *flag.FlagSet, args []string) error {
			spec := scaffoldSpec{}
			if len(args) > 0 {
				spec.Name = args[0]
			}

			if !*noInteractive {
				r := bufio.NewReader(os.Stdin)
				if spec.Name == "" {
					spec.Name = prompt(r, "  🏗️  Feature name (e.g. Order, Invoice, Product): ")
				}
				if *desc == "" {
					*desc = prompt(r, "  📝 One-line description (for docs): ")
				}
				spec.HasEntity = promptBool(r, "  → Create domain entity? [Y/n] ", true)
				spec.HasRepo = promptBool(r, "  → Create repository port? [Y/n] ", true)
				spec.HasUC = promptBool(r, "  → Create use case port + interactor? [Y/n] ", true)
				spec.HasHTTP = promptBool(r, "  → Create HTTP driving adapter? [Y/n] ", true)
				spec.HasGRPC = promptBool(r, "  → Create gRPC driving adapter? [y/N] ", false)
			} else {
				spec.HasEntity = *entity
				spec.HasRepo = *repo
				spec.HasUC = *usecase
				spec.HasHTTP = *httpFlag
				spec.HasGRPC = *grpcFlag
			}

			if spec.Name == "" {
				return fmt.Errorf("feature name is required")
			}

			fmt.Printf("\n  🚀 Scaffolding feature %q (hexagonal)\n", spec.Name)
			fmt.Println("  " + strRepeat("═", 60))

			runSub := func(cmd *SimpleCommand, argStr string) error {
				cmd.SetArgs(strings.Fields(argStr))
				return cmd.Execute()
			}

			var generated []featureEntry

			if spec.HasEntity {
				if err := runSub(NewMakeEntityCommand(), spec.Name); err != nil {
					return err
				}
				generated = append(generated, featureEntry{
					"Domain Entity",
					"internal/domain/entities/" + toPascal(spec.Name) + ".go",
					"Core business object with state and behavior",
				})
			}
			if spec.HasRepo {
				if err := runSub(NewMakePortCommand(), spec.Name+"Repository"); err != nil {
					return err
				}
				generated = append(generated, featureEntry{
					"Driven Port (Repository)",
					"internal/ports/driven/" + toPascal(spec.Name) + "Repository.go",
					"Outbound persistence contract",
				})
			}
			if spec.HasUC {
				if err := runSub(NewMakePortCommand(), spec.Name+"UseCase --driving"); err != nil {
					return err
				}
				generated = append(generated, featureEntry{
					"Driving Port (Use Case)",
					"internal/ports/driving/" + toPascal(spec.Name) + "UseCase.go",
					"Inbound contract — what HTTP/gRPC/CLI can call",
				})
				if err := runSub(NewMakeUsecaseCommand(), spec.Name); err != nil {
					return err
				}
				generated = append(generated, featureEntry{
					"Application Use Case",
					"internal/application/usecases/" + toSnake(spec.Name) + "_usecase.go",
					"Business logic — orchestrates domain + driven ports",
				})
			}
			if spec.HasHTTP {
				if err := runSub(NewMakeAdapterCommand(), spec.Name+"Handler --direction driving"); err != nil {
					return err
				}
				generated = append(generated, featureEntry{
					"Driving Adapter (HTTP)",
					"internal/adapters/driving/" + toSnake(spec.Name) + "handler/" + toSnake(spec.Name) + "handler_adapter.go",
					"HTTP entry point",
				})
			}
			if spec.HasGRPC {
				if err := runSub(NewMakeAdapterCommand(), spec.Name+"Grpc --direction driving"); err != nil {
					return err
				}
				generated = append(generated, featureEntry{
					"Driving Adapter (gRPC)",
					"internal/adapters/driving/" + toSnake(spec.Name) + "grpc/" + toSnake(spec.Name) + "grpc_adapter.go",
					"gRPC entry point stub",
				})
			}

			mdPath, err := writeFeatureDoc(spec.Name, *desc, generated)
			if err != nil {
				fmt.Printf("  ⚠️  Could not write feature doc: %v\n", err)
			} else {
				fmt.Printf("  📄 Context doc: %s\n", mdPath)
			}

			fmt.Printf("\n  ✅ Feature %q scaffolded!\n", spec.Name)
			fmt.Println("  Next: wire adapters in bootstrap/app.go → RegisterCore().")
			fmt.Println()
			return nil
		},
	}
}

func writeFeatureDoc(name, desc string, files []featureEntry) (string, error) {
	pascal := toPascal(name)
	snake := toSnake(name)
	mod := ModuleName()
	dir := "docs/features"
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	path := dir + "/" + snake + ".md"
	if desc == "" {
		desc = pascal + " bounded context"
	}

	var sb strings.Builder

	sb.WriteString("<!-- CONTEXT_START\n")
	sb.WriteString("module: " + mod + "\n")
	sb.WriteString("feature: " + pascal + "\n")
	sb.WriteString("generated: " + time.Now().Format("2006-01-02") + "\n")
	sb.WriteString("arch: hexagonal (ports-and-adapters)\n")
	sb.WriteString("CONTEXT_END -->\n\n")

	sb.WriteString("# Feature: " + pascal + "\n\n")
	sb.WriteString("> " + desc + "\n\n")

	sb.WriteString("## 📋 Instructions\n\n")
	sb.WriteString("<!-- INSTRUCTION\n")
	sb.WriteString("Read this file FIRST before any edit. Rules:\n")
	sb.WriteString("1. Domain entities must not import ports or adapters.\n")
	sb.WriteString("2. Ports are Go interfaces only — no implementations.\n")
	sb.WriteString("3. Use cases depend on driven ports via constructor injection.\n")
	sb.WriteString("4. Adapters depend on driving ports — never on use case structs.\n")
	sb.WriteString("5. Wire new bindings in bootstrap/app.go (Container.Singleton).\n")
	sb.WriteString("6. Use derrors.New(op, sentinel, msg) for domain errors.\n")
	sb.WriteString("7. IDs use valueobjects.ID — NewIDStr() or NewIDUint().\n")
	sb.WriteString("INSTRUCTION -->\n\n")

	sb.WriteString("## 📁 Generated Files\n\n")
	sb.WriteString("| Layer | File | Role |\n")
	sb.WriteString("|-------|------|------|\n")
	for _, f := range files {
		sb.WriteString("| " + f.layer + " | `" + f.path + "` | " + f.role + " |\n")
	}
	sb.WriteString("\n")

	sb.WriteString("## 🏗️ Layer Flow\n\n")
	sb.WriteString("```\n")
	sb.WriteString("HTTP/gRPC/CLI\n")
	sb.WriteString("  └─ Driving Adapter  (internal/adapters/driving/)\n")
	sb.WriteString("       └─ Driving Port   (internal/ports/driving/" + pascal + "UseCase)\n")
	sb.WriteString("            └─ Use Case      (internal/application/usecases/" + snake + "_usecase.go)\n")
	sb.WriteString("                 └─ Driven Port   (internal/ports/driven/" + pascal + "Repository)\n")
	sb.WriteString("                      └─ Adapter      (internal/adapters/driven/persistence/)\n")
	sb.WriteString("```\n\n")

	sb.WriteString("## 🔌 Bootstrap Wiring\n\n")
	sb.WriteString("Add to `bootstrap/app.go` → `RegisterCore()`:\n\n")
	sb.WriteString("```go\n")
	sb.WriteString("a.Container.Singleton(\"repo." + snake + "\", func(c *kernel.Container) (interface{}, error) {\n")
	sb.WriteString("    return inmemory.NewInMemory" + pascal + "Repo(), nil\n")
	sb.WriteString("})\n")
	sb.WriteString("a.Container.Singleton(\"uc." + snake + "\", func(c *kernel.Container) (interface{}, error) {\n")
	sb.WriteString("    r, _ := c.Resolve(\"repo." + snake + "\")\n")
	sb.WriteString("    return usecases.New" + pascal + "Interactor(r.(driven." + pascal + "Repository)), nil\n")
	sb.WriteString("})\n")
	sb.WriteString("a.Container.Alias(\"uc." + snake + "\", \"driving." + pascal + "UseCase\")\n")
	sb.WriteString("```\n\n")

	sb.WriteString("## ⚡ Quick Tasks\n\n")
	sb.WriteString("<!-- OUTPUT_HINTS\n")
	sb.WriteString("Add method: edit port → implement in use case → add HTTP route in adapter\n")
	sb.WriteString("Add entity field: edit entity → update constructor → rancago make:migration add_field_to_" + snake + "s\n")
	sb.WriteString("OUTPUT_HINTS -->\n\n")
	sb.WriteString("| Task | Where |\n")
	sb.WriteString("|------|-------|\n")
	sb.WriteString("| Add entity field | `internal/domain/entities/" + pascal + ".go` |\n")
	sb.WriteString("| Add use case method | Port + interactor |\n")
	sb.WriteString("| Add HTTP route | Driving adapter `RegisterRoutes()` |\n")
	sb.WriteString("| Add migration | `rancago make:migration add_..._to_" + snake + "s` |\n\n")

	sb.WriteString("## 🚨 Domain Errors\n\n")
	sb.WriteString("```go\n")
	sb.WriteString("derrors.New(\"" + snake + ".create\", derrors.ErrValidation, \"name is required\")\n")
	sb.WriteString("// ErrNotFound · ErrUnauthorized · ErrForbidden · ErrValidation · ErrConflict · ErrAlreadyExists\n")
	sb.WriteString("```\n")

	if err := os.WriteFile(path, []byte(sb.String()), 0644); err != nil {
		return "", err
	}
	fmt.Printf("  📄 Created Feature Doc: %s\n", path)
	return path, nil
}
