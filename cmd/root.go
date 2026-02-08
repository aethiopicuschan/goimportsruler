package cmd

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"runtime/debug"
	"sort"
	"strings"

	"github.com/aethiopicuschan/goimportsruler/pkg/config"
	"github.com/aethiopicuschan/goimportsruler/pkg/ctxrw"
	"github.com/aethiopicuschan/goimportsruler/pkg/errs"
	"github.com/aethiopicuschan/goimportsruler/pkg/values"
	"github.com/aethiopicuschan/goimportsruler/pkg/walker"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:           "goimportsruler <path>",
	Short:         "A tool to manage import rules for Go projects",
	Args:          cobra.MinimumNArgs(0),
	RunE:          run,
	SilenceUsage:  true,
	SilenceErrors: true,
	Example:       "  goimportsruler ./...",
}

type groupKey struct {
	file string
	pkg  string
	rule string
	desc string
}

type entry struct {
	line int
	col  int
	path string
}

func run(cmd *cobra.Command, args []string) (err error) {
	start := "."
	if len(args) > 0 {
		start = args[0]
	}

	// --- Load configuration ---
	cfg, fp, err := config.LoadConfig(start)
	if err != nil {
		// If the configuration file is not found, inform the user to create one.
		if errors.Is(err, errs.ErrConfigNotFound) {
			fmt.Println("Configuration file not found. Please create a configuration file to use goimportsruler.")
			candidates := values.GetAllConfigFileNames()
			fmt.Printf("Following file names are supported:\n\n")
			for _, candidate := range candidates {
				fmt.Printf("  - %s\n", candidate)
			}
			fmt.Println()
		}

		// If the configuration file is invalid, show an example configuration.
		if errors.Is(err, errs.ErrInvalidFile) {
			fmt.Println("Invalid configuration file. Please check the syntax and content of your configuration file.")
			example := config.ExampleConfig()
			var buf []byte
			buffer := bytes.NewBuffer(buf)
			var err2 error
			switch filepath.Ext(fp) {
			case ".json":
				err2 = example.ToJSON(buffer)
			case ".yaml", ".yml":
				err2 = example.ToYAML(buffer)
			default:
				// Default to JSON if the extension is unrecognized.
				err2 = example.ToJSON(buffer)
			}
			if err2 == nil {
				fmt.Printf("Here is an example configuration:\n\n")
				fmt.Printf("```\n%s\n```\n\n", buffer.String())
			}
		}
		return
	}

	ctx := ctxrw.NewContextWriter().SetConfig(context.Background(), *cfg)

	w, err := walker.NewWalker(ctx, start)
	if err != nil {
		return
	}

	ve, err := w.Walk(ctx, args...)
	if err != nil {
		return
	}

	if ve.Len() == 0 {
		return nil
	}

	// Group violations by (file, package, rule, description) and format a stable, readable output.
	groups := map[groupKey][]entry{}

	for v := range ve.Violations() {
		k := groupKey{
			file: v.File(),
			pkg:  v.SourcePkg(),
			rule: v.RuleName(),
			desc: v.RuleDescription(),
		}
		groups[k] = append(groups[k], entry{
			line: v.Line(),
			col:  v.Column(),
			path: v.ImportPath(),
		})
	}

	keys := make([]groupKey, 0, len(groups))
	for k := range groups {
		keys = append(keys, k)
	}

	sort.Slice(keys, func(i, j int) bool {
		if keys[i].file != keys[j].file {
			return keys[i].file < keys[j].file
		}
		if keys[i].pkg != keys[j].pkg {
			return keys[i].pkg < keys[j].pkg
		}
		if keys[i].rule != keys[j].rule {
			return keys[i].rule < keys[j].rule
		}
		return keys[i].desc < keys[j].desc
	})

	const (
		indentBlock = "  "
		indentItems = "    "
	)

	for index, k := range keys {
		ents := groups[k]

		// Sort entries by position, then path.
		sort.Slice(ents, func(i, j int) bool {
			if ents[i].line != ents[j].line {
				return ents[i].line < ents[j].line
			}
			if ents[i].col != ents[j].col {
				return ents[i].col < ents[j].col
			}
			return ents[i].path < ents[j].path
		})

		// Deduplicate consecutive identical (line,col,path) entries.
		dedup := make([]entry, 0, len(ents))
		var last *entry
		for i := range ents {
			if last != nil && last.line == ents[i].line && last.col == ents[i].col && last.path == ents[i].path {
				continue
			}
			e := ents[i]
			dedup = append(dedup, e)
			last = &e
		}
		ents = dedup

		// Header line (single line to keep it compact).
		fmt.Printf("%s (package %s)\n", k.file, k.pkg)

		// Rule block.
		ruleLine := fmt.Sprintf("%s- rule: %q", indentBlock, k.rule)
		if k.desc != "" && k.desc != "<no description>" {
			// Put description in parentheses to keep it on the same line.
			ruleLine += fmt.Sprintf(" (%s)", k.desc)
		}
		fmt.Println(ruleLine)

		// Items.
		for _, e := range ents {
			// Prefer a compact location format: :line:col
			loc := fmt.Sprintf(":%d:%d", e.line, e.col)
			fmt.Printf("%s• %s%s  ←  %s\n", indentItems, k.file, loc, e.path)
		}

		if index < len(keys)-1 {
			fmt.Println()
		}
	}

	// Return a sentinel error for CI usage.
	err = fmt.Errorf("%w: %d violation(s) found", errs.ErrViolationsFound, ve.Len())

	// Keep cobra.Command referenced to avoid unused import issues in some build tags.
	_ = strings.TrimSpace(cmd.Use)

	return
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	bi, ok := debug.ReadBuildInfo()
	if ok && bi.Main.Version != "" {
		rootCmd.Version = bi.Main.Version
	} else {
		rootCmd.Version = "unknown"
	}
}
