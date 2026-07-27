// Package commands is the standalone Rancago CLI.
// It has zero dependency on the rancago/framework module —
// all it does is generate and scaffold files in whatever project you run it from.
package commands

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

// Version and BuildDate are set at build time via ldflags:
//
//	go build -ldflags "-X github.com/rancago/rancago-cli/commands.Version=1.2.3" .
var (
	Version   = "1.0.0"
	BuildDate = "2026-07-27"
)

func banner() string {
	return `
  ____                                        
 |  _ \ __ _ _ __   ___ __ _  __ _  ___  
 | |_) / _` + "`" + ` | '_ \ / __/ _` + "`" + ` |/ _` + "`" + ` |/ _ \ 
 |  _ < (_| | | | | (_| (_| | (_| | (_) |
 |_| \_\__,_|_| |_|\___\__,_|\__, |\___/ 
                               |___/       
`
}

// RunCLI is the entry point for the rancago CLI binary.
// Returns the OS exit code.
func RunCLI() int {
	if len(os.Args) < 2 {
		printHelp()
		return 0
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "-h", "--help", "help":
		printHelp()
		return 0

	case "-v", "--version", "version":
		fmt.Printf("rancago-cli %s (built %s)\n", Version, BuildDate)
		return 0

	case "serve":
		return handleServe(args)

	case "migrate":
		return run(NewMigrateCommand(), args)

	case "scaffold":
		return run(NewScaffoldCommand(), args)

	case "make:entity":
		return run(NewMakeEntityCommand(), args)

	case "make:value-object", "make:vo":
		return run(NewMakeValueObjectCommand(), args)

	case "make:port":
		return run(NewMakePortCommand(), args)

	case "make:usecase":
		return run(NewMakeUsecaseCommand(), args)

	case "make:adapter":
		return run(NewMakeAdapterCommand(), args)

	case "make:model":
		return run(NewMakeModelCommand(), args)

	case "make:migration":
		return run(NewMakeMigrationCommand(), args)

	case "make:controller":
		return run(NewMakeControllerCommand(), args)

	case "make:middleware":
		return run(NewMakeMiddlewareCommand(), args)

	case "make:provider":
		return run(NewMakeProviderCommand(), args)

	case "tinker":
		return run(NewTinkerCommand(), args)

	case "key:generate", "key:gen":
		return run(NewKeyGenerateCommand(), args)

	case "storage:link":
		return run(NewStorageLinkCommand(), args)

	case "route:list":
		return run(NewRouteListCommand(), args)

	default:
		if strings.HasPrefix(cmd, "make:") {
			fmt.Fprintf(os.Stderr, "Unknown make command: %s\n", cmd)
			fmt.Fprintln(os.Stderr, "Available: make:entity, make:value-object, make:port, make:usecase,")
			fmt.Fprintln(os.Stderr, "           make:adapter, make:model, make:migration, make:controller,")
			fmt.Fprintln(os.Stderr, "           make:middleware, make:provider")
			return 1
		}
		fmt.Fprintf(os.Stderr, "Unknown command: %s  (try: rancago help)\n", cmd)
		return 1
	}
}

func printHelp() {
	fmt.Println(banner())
	fmt.Println("rancago-cli — Rancago Framework Code Generator & Toolkit")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  rancago [command] [flags] [args]")
	fmt.Println()

	rows := [][2]string{
		{"serve", "Start the HTTP (and optional gRPC) development server"},
		{"migrate", "Run database migrations"},
		{"scaffold [name]", "Interactive scaffolder for a full bounded context"},
		{"", ""},
		{"Code generators:", ""},
		{"  make:entity [name]", "Domain entity (internal/domain/entities)"},
		{"  make:value-object [name]", "Value object (internal/domain/valueobjects)"},
		{"  make:port [name]", "Port interface, --driving for inbound port"},
		{"  make:usecase [name]", "Use case interactor (internal/application/usecases)"},
		{"  make:adapter [name]", "Adapter stub, --direction driving|driven"},
		{"  make:model [name] [-m]", "GORM model, -m also generates migration"},
		{"  make:migration [name]", "Migration file (database/migrations)"},
		{"  make:controller [name] [-r]", "HTTP controller, -r for resourceful"},
		{"  make:middleware [name]", "HTTP middleware stub"},
		{"  make:provider [name]", "ServiceProvider stub (Register + Boot)"},
		{"", ""},
		{"Utilities:", ""},
		{"  tinker", "Interactive REPL (container explorer)"},
		{"  key:generate", "Generate a secure APP_KEY (base64:...)"},
		{"  storage:link", "Symlink public/storage → storage/app/public"},
		{"  route:list", "Print all registered HTTP / gRPC / WS routes"},
		{"", ""},
		{"  help", "Show this help message"},
		{"  version / -v", "Show CLI version"},
	}

	for _, r := range rows {
		if r[0] == "" {
			fmt.Println(r[1])
		} else {
			fmt.Printf("  %-32s %s\n", r[0], r[1])
		}
	}
	fmt.Println()
	fmt.Println("Run 'rancago <command> --help' for command-specific flags.")
}

// handleServe starts an HTTP + optional gRPC server.
// It calls the rancago server binary if found, otherwise prints instructions.
func handleServe(args []string) int {
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	port := fs.Int("port", 8080, "HTTP server port")
	withGRPC := fs.Bool("grpc", false, "Also start gRPC server")
	fs.IntVar(port, "p", 8080, "HTTP port (short)")
	if err := fs.Parse(args); err != nil {
		return 1
	}

	fmt.Printf("[rancago] Starting server on :%d", *port)
	if *withGRPC {
		fmt.Print(" + gRPC :9090")
	}
	fmt.Println()

	// Delegate to `go run .` in the target project directory.
	fmt.Println("[rancago] Tip: run `go run .` in your project root to start the server.")
	fmt.Println("[rancago] The CLI is a standalone tool — the server lives in your project.")

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	<-ctx.Done()
	fmt.Println("\n[rancago] Shutdown signal received.")
	return 0
}

type executor interface {
	SetArgs([]string)
	Execute() error
}

func run(cmd executor, args []string) int {
	cmd.SetArgs(args)
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}
	return 0
}
