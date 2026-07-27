package commands

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/base64"
	"flag"
	"fmt"
	"os"
	"strings"
)

// ---- migrate ----

func NewMigrateCommand() *SimpleCommand {
	fs := flag.NewFlagSet("migrate", flag.ContinueOnError)
	rollback := fs.Bool("rollback", false, "Rollback the last batch of migrations")
	return &SimpleCommand{
		use: "migrate", short: "Run database migrations",
		argsMin: 0, argsMax: 0, flags: fs,
		runFn: func(f *flag.FlagSet, _ []string) error {
			action := "applying"
			if *rollback {
				action = "rolling back"
			}
			fmt.Printf("\n  🗄️  Migrate: %s migrations\n", action)
			fmt.Println("  " + strRepeat("─", 60))
			fmt.Println("  No DB adapter wired — plug in a GORM-backed driven adapter to run real migrations.")
			fmt.Println("  Migration status: stub (no-op)")
			fmt.Println()
			return nil
		},
	}
}

// ---- scaffold ----

type scaffoldSpec struct {
	Name      string
	HasEntity bool
	HasRepo   bool
	HasUC     bool
	HasHTTP   bool
	HasGRPC   bool
}

func NewScaffoldCommand() *SimpleCommand {
	fs := flag.NewFlagSet("scaffold", flag.ContinueOnError)
	name := fs.String("name", "", "Component name")
	fs.StringVar(name, "n", "", "Component name (short)")
	noInteractive := fs.Bool("no-interactive", false, "Skip prompts, use flags")
	entity := fs.Bool("entity", true, "Scaffold entity")
	repo := fs.Bool("repo", true, "Scaffold repository port")
	usecase := fs.Bool("usecase", true, "Scaffold use case + port")
	httpFlag := fs.Bool("http", true, "Scaffold HTTP adapter")
	grpcFlag := fs.Bool("grpc", false, "Scaffold gRPC adapter")
	return &SimpleCommand{
		use: "scaffold [name]", short: "Interactive full bounded-context scaffolder",
		argsMin: 0, argsMax: 1, flags: fs,
		runFn: func(f *flag.FlagSet, args []string) error {
			spec := scaffoldSpec{Name: *name}
			if len(args) > 0 && spec.Name == "" {
				spec.Name = args[0]
			}

			if !*noInteractive {
				r := bufio.NewReader(os.Stdin)
				if spec.Name == "" {
					spec.Name = prompt(r, "  🏗️  Component name (e.g. Order, Product, Invoice): ")
				}
				spec.HasEntity = promptBool(r, "  → Create domain entity? [Y/n] ", true)
				spec.HasRepo = promptBool(r, "  → Create repository port? [Y/n] ", true)
				spec.HasUC = promptBool(r, "  → Create use case + port? [Y/n] ", true)
				spec.HasHTTP = promptBool(r, "  → Create HTTP adapter? [Y/n] ", true)
				spec.HasGRPC = promptBool(r, "  → Create gRPC adapter? [y/N] ", false)
			} else {
				spec.HasEntity = *entity
				spec.HasRepo = *repo
				spec.HasUC = *usecase
				spec.HasHTTP = *httpFlag
				spec.HasGRPC = *grpcFlag
			}

			if spec.Name == "" {
				return fmt.Errorf("component name is required")
			}

			fmt.Printf("\n  🚀 Scaffolding bounded context %q\n", spec.Name)
			fmt.Println("  " + strRepeat("═", 60))

			runSub := func(cmd *SimpleCommand, argStr string) error {
				cmd.SetArgs(strings.Fields(argStr))
				return cmd.Execute()
			}

			if spec.HasEntity {
				if err := runSub(NewMakeEntityCommand(), spec.Name); err != nil {
					return err
				}
			}
			if spec.HasRepo {
				if err := runSub(NewMakePortCommand(), spec.Name+"Repository"); err != nil {
					return err
				}
			}
			if spec.HasUC {
				if err := runSub(NewMakePortCommand(), spec.Name+"UseCase --driving"); err != nil {
					return err
				}
				if err := runSub(NewMakeUsecaseCommand(), spec.Name); err != nil {
					return err
				}
			}
			if spec.HasHTTP {
				if err := runSub(NewMakeAdapterCommand(), spec.Name+"Handler --direction driving"); err != nil {
					return err
				}
			}
			if spec.HasGRPC {
				if err := runSub(NewMakeAdapterCommand(), spec.Name+"Grpc --direction driving"); err != nil {
					return err
				}
			}

			fmt.Printf("\n  ✅ Scaffold %q complete!\n", spec.Name)
			fmt.Println("  Next: wire the new adapters in bootstrap/app.go and RegisterCore().")
			fmt.Println()
			return nil
		},
	}
}

// ---- tinker ----

func NewTinkerCommand() *SimpleCommand {
	fs := flag.NewFlagSet("tinker", flag.ContinueOnError)
	return &SimpleCommand{
		use: "tinker", short: "Interactive REPL (architecture explorer)",
		argsMin: 0, argsMax: 0, flags: fs,
		runFn: func(_ *flag.FlagSet, _ []string) error {
			fmt.Println(`
  🔮 Rancago Tinker — minimal REPL
  Commands: help  ports  ls  module  quit
`)
			reader := bufio.NewReader(os.Stdin)
			for {
				fmt.Print("rancago> ")
				line, err := reader.ReadString('\n')
				if err != nil {
					break
				}
				switch strings.TrimSpace(line) {
				case "":
					continue
				case "quit", "exit":
					fmt.Println("  Goodbye!")
					return nil
				case "help":
					fmt.Println("  Commands: help  ports  ls  module  quit")
				case "module":
					fmt.Println("  Module:", ModuleName())
				case "ports":
					fmt.Println(`
  Driving (inbound) ports:
    NotificationUseCase · UserUseCase · DocumentUseCase

  Driven (outbound) ports:
    Repository[T] · VectorRepository[T]
    NotificationRepository · UserRepository · RoleRepository
    PermissionRepository · DocumentRepository
    CachePort · StorageDriver · StorageManagerPort
    WebSocketPort · AuthProviderPort · SocialitePort
    DatabasePort · TransactionPort
`)
				case "ls":
					fmt.Println(`
  Default container bindings (bootstrap/app.go):
    config            *config.Config
    redis             Cache.RedisManager
    ws.hub            WebSocket.Hub
    storage           Storage.Manager
    auth.socialite    Auth.SocialiteManager
    auth.rbac         Auth.RBACService
    google.calendar   Contracts.CalendarService
    google.meet       Contracts.MeetService
    google.scheduler  Contracts.MeetingScheduler
    service.notification  Contracts.NotificationService
    repo.*            driven.XxxRepository (in-memory)
    uc.*              driving.XxxUseCase
`)
				default:
					fmt.Printf("  unknown: %q  (type 'help')\n", strings.TrimSpace(line))
				}
			}
			return nil
		},
	}
}

// ---- key:generate ----

func NewKeyGenerateCommand() *SimpleCommand {
	fs := flag.NewFlagSet("key:generate", flag.ContinueOnError)
	return &SimpleCommand{
		use: "key:generate", short: "Generate a secure APP_KEY",
		argsMin: 0, argsMax: 0, flags: fs,
		runFn: func(_ *flag.FlagSet, _ []string) error {
			b := make([]byte, 32)
			if _, err := rand.Read(b); err != nil {
				return fmt.Errorf("key:generate: %w", err)
			}
			key := "base64:" + base64.StdEncoding.EncodeToString(b)
			fmt.Printf("\n  🔑  APP_KEY=%s\n\n", key)
			fmt.Println("  Set this in your environment or config/config.go.")
			return nil
		},
	}
}

// ---- storage:link ----

func NewStorageLinkCommand() *SimpleCommand {
	fs := flag.NewFlagSet("storage:link", flag.ContinueOnError)
	return &SimpleCommand{
		use: "storage:link", short: "Symlink public/storage → storage/app/public",
		argsMin: 0, argsMax: 0, flags: fs,
		runFn: func(_ *flag.FlagSet, _ []string) error {
			if err := os.MkdirAll("storage/app/public", 0755); err != nil {
				return err
			}
			link := "public/storage"
			if _, err := os.Lstat(link); err == nil {
				fmt.Println("  ℹ️  Link already exists:", link)
				return nil
			}
			if err := os.Symlink("../storage/app/public", link); err != nil {
				return fmt.Errorf("could not create symlink (run as admin on Windows): %w", err)
			}
			fmt.Println("  ✅ Created:", link, "→ storage/app/public")
			return nil
		},
	}
}

// ---- route:list ----

func NewRouteListCommand() *SimpleCommand {
	fs := flag.NewFlagSet("route:list", flag.ContinueOnError)
	return &SimpleCommand{
		use: "route:list", short: "List registered routes",
		argsMin: 0, argsMax: 0, flags: fs,
		runFn: func(_ *flag.FlagSet, _ []string) error {
			fmt.Println("\n  Rancago Default Routes")
			fmt.Println("  " + strRepeat("─", 80))
			fmt.Printf("  %-8s  %-48s  %s\n", "METHOD", "PATH", "HANDLER / NOTES")
			fmt.Println("  " + strRepeat("─", 80))
			rows := [][3]string{
				{"GET", "/", "HealthHandler.Welcome"},
				{"GET", "/api/v1/health", "HealthHandler.Health"},
				{"POST", "/api/v1/notifications/send", "NotificationHandler.Send"},
				{"POST", "/api/v1/notifications/broadcast", "NotificationHandler.Broadcast"},
				{"GET", "/api/v1/notifications/list", "NotificationHandler.List"},
				{"GET", "/api/v1/notifications/count", "NotificationHandler.Count"},
				{"POST", "/api/v1/notifications/read", "NotificationHandler.MarkRead"},
				{"GET", "/ws", "WebSocket Hub (stub — extend with gorilla/websocket)"},
				{"gRPC", ":9090 /rancago.NotificationService/*", "GRPCNotificationAdapter (stub)"},
			}
			for _, r := range rows {
				fmt.Printf("  %-8s  %-48s  %s\n", r[0], r[1], r[2])
			}
			fmt.Println()
			return nil
		},
	}
}

// ---- helpers ----

func prompt(r *bufio.Reader, q string) string {
	fmt.Print(q)
	ans, _ := r.ReadString('\n')
	return strings.TrimSpace(ans)
}

func promptBool(r *bufio.Reader, q string, def bool) bool {
	switch strings.ToLower(prompt(r, q)) {
	case "y", "yes":
		return true
	case "n", "no":
		return false
	}
	return def
}

// unused suppresses "imported and not used" for context in serve command.
var _ = context.Background
