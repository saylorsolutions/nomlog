package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/go-hclog"
	"github.com/saylorsolutions/nomlog/plugin"
	"github.com/saylorsolutions/nomlog/plugin/file"
	"github.com/saylorsolutions/nomlog/plugin/store"
	"github.com/saylorsolutions/nomlog/runtime"
	"github.com/saylorsolutions/nomlog/runtime/dsl"
	"os"
	"strings"
	"time"
)

func main() {
	log := hclog.Default()
	if len(os.Args) <= 1 {
		usage()
		return
	}
	args := os.Args[1:]
	if len(args) >= 1 {
		switch args[0] {
		case "exec":
			start := time.Now()
			if err := doExec(log, args[1:]...); err != nil {
				exitError("Failed to execute script: %v", err)
			}
			dur := time.Now().Sub(start)
			var durStr string
			if dur < time.Millisecond {
				durStr = dur.Round(time.Microsecond).String()
			} else if dur < time.Second {
				durStr = dur.Round(time.Millisecond).String()
			} else {
				durStr = dur.Round(time.Second).String()
			}
			fmt.Printf("Script executed successfully in %s\n", durStr)
			return
		case "vet":
			if err := doVet(log, args[1:]...); err != nil {
				exitError("Dry run failed: %v", err)
			}
			fmt.Println("Dry run ran successfully")
		case "plugins":
			doPrintPlugins()
		case "help":
			usage()
		default:
			exitError("Unrecognized command: '%s'", args[0])
		}
	} else {
		exitError("No command specified")
	}
}

func exitError(format string, args ...any) {
	if !strings.HasSuffix(format, "\n") {
		format += "\n"
	}
	fmt.Printf("Error: "+format, args...)
	usage()
	os.Exit(-1)
}

func usage() {
	text := `
nomlog is a log management tool that is able to execute scripts.

  nomlog help
  nomlog plugins
  nomlog exec FILE
  nomlog vet FILE

The 'help' subcommand will print this usage information.
The 'plugins' subcommand will print information about plugins, and the documentation for all plugins loaded into the runtime for this program.
The 'exec' subcommand will execute FILE as a nomlog script. Any errors that occur during execution will be reported.
The 'vet' subcommand will dry run FILE as a nomlog script. Errors will still be reported as if the script were really executed, but no action will be taken.
`
	fmt.Print(text)
}

func plugins() []plugin.Plugin {
	return []plugin.Plugin{
		file.Plugin(),
		store.Plugin(),
	}
}

func doPrintPlugins() {
	reg := plugin.NewRegistration()
	for _, p := range plugins() {
		p.Register(reg)
	}
	fmt.Println("Plugins are used to extend the functionality of a nomlog based applications with technology specific functionality")
	fmt.Println()
	fmt.Print(reg.AllDocs())
}

func doExec(log hclog.Logger, args ...string) (rerr error) {
	if len(args) >= 1 {
		r := runtime.NewRuntime(log, plugins()...)
		if err := r.Start(context.Background()); err != nil {
			return err
		}
		defer func() {
			err := r.Stop()
			if err != nil {
				log.Error("Error while stopping runtime", "error", err)
				rerr = err
			}
		}()
		ast, err := dsl.ParseFile(args[0])
		if err != nil {
			return err
		}
		err = r.Execute(ast...)
		if err != nil {
			return err
		}
		return nil
	}
	return errors.New("not enough arguments for exec")
}

func doVet(log hclog.Logger, args ...string) (rerr error) {
	if len(args) >= 1 {
		r := runtime.NewRuntime(log, plugins()...)
		if err := r.Start(context.Background()); err != nil {
			return err
		}
		defer func() {
			err := r.Stop()
			if err != nil {
				log.Error("Error while stopping runtime", "error", err)
				rerr = err
			}
		}()
		ast, err := dsl.ParseFile(args[0])
		if err != nil {
			return err
		}
		err = r.DryRun(ast...)
		if err != nil {
			return err
		}
		return nil
	}
	return errors.New("not enough arguments for exec")
}
