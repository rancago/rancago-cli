package commands

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ---- SimpleCommand ----

// SimpleCommand is a minimal cobra-like command backed by flag.FlagSet.
type SimpleCommand struct {
	use     string
	short   string
	argsMin int
	argsMax int // -1 = unlimited
	flags   *flag.FlagSet
	runFn   func(fs *flag.FlagSet, args []string) error
}

func (c *SimpleCommand) SetArgs(a []string) {
	_ = c.flags.Parse(a)
}

func (c *SimpleCommand) Execute() error {
	args := c.flags.Args()
	if c.argsMin > 0 && len(args) < c.argsMin {
		return fmt.Errorf("usage: rancago %s — %s (missing required args)", c.use, c.short)
	}
	if c.argsMax >= 0 && len(args) > c.argsMax {
		return fmt.Errorf("usage: rancago %s — too many arguments", c.use)
	}
	return c.runFn(c.flags, args)
}

// ---- Name helpers ----

func toPascal(s string) string {
	parts := splitName(s)
	for i, p := range parts {
		if p == "" {
			continue
		}
		parts[i] = strings.ToUpper(p[:1]) + strings.ToLower(p[1:])
	}
	return strings.Join(parts, "")
}

func toSnake(s string) string {
	return strings.ToLower(strings.Join(splitName(s), "_"))
}

func splitName(s string) []string {
	s = strings.ReplaceAll(s, "/", "_")
	s = strings.ReplaceAll(s, "-", "_")
	return strings.Split(s, "_")
}

func strRepeat(s string, n int) string {
	out := ""
	for i := 0; i < n; i++ {
		out += s
	}
	return out
}

// ---- Generator ----

// Generator writes a generated file to BasePath/PascalName+ext.
type Generator struct {
	Name     string
	Type     string
	BasePath string
	Package  string
}

func (g Generator) writeFile(ext, content string) error {
	if err := os.MkdirAll(g.BasePath, 0755); err != nil {
		return err
	}
	filename := filepath.Join(g.BasePath, toPascal(g.Name)+ext)
	if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
		return err
	}
	fmt.Printf("  ✅ Created %s: %s\n", g.Type, filename)
	return nil
}

// ModuleName reads the Go module name from go.mod in the current directory.
func ModuleName() string {
	b, err := os.ReadFile("go.mod")
	if err != nil {
		return "github.com/your-org/your-project"
	}
	for _, line := range strings.Split(string(b), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module "))
		}
	}
	return "github.com/your-org/your-project"
}
