// Package runtime provides the means to execute on a given AST.
package runtime

import (
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/go-hclog"
	"github.com/saylorsolutions/nomlog/pkg/entries"
	"github.com/saylorsolutions/nomlog/pkg/iterator"
	"github.com/saylorsolutions/nomlog/plugin"
	"github.com/saylorsolutions/nomlog/runtime/dsl"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	ErrEmptyID        = errors.New("empty ID")
	ErrUndefined      = errors.New("undefined identifier")
	ErrConsumed       = errors.New("identifier has been consumed")
	ErrAlreadyDefined = errors.New("identifier is already in use")
	ErrInvalidState   = errors.New("invalid state")
	ErrUnknownSource  = errors.New("unknown source class")
	ErrUnknownSink    = errors.New("unknown sink class")
)

type runtimeState int

const (
	created runtimeState = iota
	started
	executing
	stopping
	done
)

var (
	stateStrings = map[runtimeState]string{
		created:   "Created",
		started:   "Started",
		executing: "Executing",
		stopping:  "Stopping",
		done:      "Done",
	}
)

type Runtime struct {
	log       hclog.Logger
	ctx       context.Context
	cancel    context.CancelFunc
	registry  *plugin.Registration
	plugins   []plugin.Plugin
	sources   []iterator.Iterator
	consumed  []bool
	sourceIDs map[string]int
	wg        sync.WaitGroup
	state     runtimeState
	dryRun    bool
}

func NewRuntime(log hclog.Logger, plugins ...plugin.Plugin) *Runtime {
	return &Runtime{
		log:       log.Named("runtime"),
		registry:  plugin.NewRegistration(),
		plugins:   plugins,
		sourceIDs: map[string]int{},
	}
}

func (r *Runtime) assertState(expected runtimeState, operation string) error {
	if r.state != expected {
		return fmt.Errorf("%w: current state is '%s', expected '%s' for operation '%s'", ErrInvalidState, stateStrings[r.state], stateStrings[expected], operation)
	}
	return nil
}

func (r *Runtime) Start(_ctx context.Context) error {
	start := time.Now()
	log := r.log
	log.Debug("Starting runtime")
	if err := r.assertState(created, "start"); err != nil {
		return err
	}
	log.Debug("Registering plugins")
	r.ctx, r.cancel = context.WithCancel(_ctx)
	for _, p := range r.plugins {
		start := time.Now()
		log := log.With("plugin-id", p.ID(), "started", start)
		log.Debug("Registering plugin")
		p.Register(r.registry)
		log.Debug("Done registering plugin", "duration", time.Now().Sub(start).String())
	}
	r.state = started
	completed := time.Now()
	dur := completed.Sub(start)
	log.Info("Runtime started", "start-duration", dur.String(), "started", completed)
	return nil
}

func (r *Runtime) Stop() (rerr error) {
	start := time.Now()
	log := r.log.With("stopping", start)
	log.Debug("Stopping runtime")
	if err := r.assertState(started, "stop"); err != nil {
		return err
	}
	r.state = stopping
	log.Debug("Cancelling runtime context")
	r.cancel()
	log.Debug("Waiting for operations to cease")
	r.wg.Wait()
	log.Debug("Shutting down plugins")
	for _, p := range r.plugins {
		log := log.With("plugin-id", p.ID())
		log.Debug("Stopping plugin")
		if err := p.Stopping(); err != nil {
			log.Error("Error stopping plugin", "error", err)
			if rerr == nil {
				rerr = err
			}
		}
		log.Debug("Plugin stopped")
	}
	r.state = done
	log.Info("Runtime stopped", "stop-duration", time.Now().Sub(start).String())
	return rerr
}

func (r *Runtime) ExecuteString(cmd string) error {
	ast, err := dsl.ParseString(cmd)
	if err != nil {
		return err
	}
	return r.Execute(ast...)
}

func (r *Runtime) Execute(asts ...dsl.AstNode) error {
	if len(asts) == 0 {
		return nil
	}
	if err := r.assertState(started, "execute"); err != nil {
		return err
	}

	start := time.Now()
	log := r.log.With("exec-start", start)
	log.Debug("Executing ASTs")
	defer func() {
		stop := time.Now()
		log.Debug("Completed AST executions", "exec-stop", stop, "exec-duration", stop.Sub(stop).String())
	}()

	for _, ast := range asts {
		astStart := time.Now()
		log := log.With("exec-ast-start", astStart, "type", ast.Type())
		switch ast := ast.(type) {
		case *dsl.Source:
			if err := r.validateNewSourceID(ast.ID); err != nil {
				log.Error("Invalid source ID", "error", err)
				return err
			}
			src, _, ok := r.registry.Source(ast.Class.Qualifier, ast.Class.SourceClass)
			if !ok {
				err := fmt.Errorf("%w: %s", ErrUnknownSource, ast.Class.Text())
				log.Error("Source class not found", "error", err)
				return err
			}
			if r.dryRun {
				log.Info("Dry run source", "class", ast.Class.Text(), "args", r.argString(ast.Args))
				r.addSource(ast.ID, nil)
				continue
			}
			log.Debug("Executing source AST")
			iter, err := src(r.ctx, ast.Args...)
			if err != nil {
				log.Error("Failed to create iterator", "error", err)
				return err
			}
			r.addSource(ast.ID, iter)
		case *dsl.Sink:
			if err := r.validateExistingSourceID(ast.Source); err != nil {
				log.Error("Invalid source ID", "error", err)
				return err
			}
			r.markConsumed(ast.Source)

			sink, _, ok := r.registry.Sink(ast.Class.Qualifier, ast.Class.SinkClass)
			if !ok {
				err := fmt.Errorf("%w: %s", ErrUnknownSink, ast.Class.Text())
				log.Error("Unknown sink", "error", err)
				return err
			}
			if r.dryRun {
				log.Info("Dry run sink", "class", ast.Class.Text(), "args", r.argString(ast.Args))
				continue
			}
			fn := func() error {
				if err := sink(r.ctx, r.getSource(ast.Source), ast.Args...); err != nil {
					log.Error("Failed to execute sink", "error", err)
					return err
				}
				return nil
			}
			if ast.Async {
				r.wg.Add(1)
				go func() {
					defer r.wg.Done()
					if err := fn(); err != nil {
						r.log.Error("Error running async sink operation", "sink", ast.Class.Text(), "args", r.argString(ast.Args), "error", err)
					}
				}()
				continue
			}
			if err := fn(); err != nil {
				return err
			}
		case *dsl.Merge:
			if err := r.validateExistingSourceID(ast.SourceA); err != nil {
				log.Error("Invalid source", "error", err)
				return err
			}
			if err := r.validateExistingSourceID(ast.SourceB); err != nil {
				log.Error("Invalid source", "error", err)
				return err
			}
			if err := r.validateNewSourceID(ast.ID); err != nil {
				log.Error("Invalid identifier", "error", err)
				return err
			}
			r.markConsumed(ast.SourceA, ast.SourceB)
			a := r.getSource(ast.SourceA)
			b := r.getSource(ast.SourceB)
			var merged iterator.Iterator
			if r.dryRun {
				log.Info("Dry run merge", "source-a", ast.SourceA, "source-b", ast.SourceB, "output", ast.ID)
			} else {
				merged = iterator.Merge(a, b)
			}
			r.addSource(ast.ID, merged)
		case *dsl.Dupe:
			if err := r.validateExistingSourceID(ast.Source); err != nil {
				log.Error("Invalid source", "error", err)
				return err
			}
			if err := r.validateNewSourceID(ast.TargetA); err != nil {
				log.Error("Invalid identifier", "error", err)
				return err
			}
			if err := r.validateNewSourceID(ast.TargetB); err != nil {
				log.Error("Invalid identifier", "error", err)
				return err
			}
			src := r.getSource(ast.Source)
			r.markConsumed(ast.Source)
			var (
				a, b iterator.Iterator
			)
			if r.dryRun {
				log.Info("Dry run dupe", "source", ast.Source, "output-a", ast.TargetA, "output-b", ast.TargetB)
			} else {
				a, b = iterator.Dupe(src)
			}
			r.addSource(ast.TargetA, a)
			r.addSource(ast.TargetB, b)
		case *dsl.Append:
			if err := r.validateExistingSourceID(ast.Source); err != nil {
				log.Error("Invalid source", "error", err)
				return err
			}
			if err := r.validateExistingSourceID(ast.Target); err != nil {
				log.Error("Invalid source", "error", err)
				return err
			}
			r.markConsumed(ast.Source)
			s, t := r.getSource(ast.Source), r.getSource(ast.Target)
			if r.dryRun {
				log.Info("Dry run append", "source", ast.Source, "target", ast.Target)
				continue
			}
			appended := iterator.Concat(t, s)
			r.replaceSource(ast.Target, appended)
		case *dsl.Cut:
			if err := r.validateExistingSourceID(ast.Source); err != nil {
				log.Error("Invalid source", "error", err)
				return err
			}
			src := r.getSource(ast.Source)
			if r.dryRun {
				log.Info("Dry run cut", "source", ast.Source, "mappings", ast.FieldSets, "delimiter", ast.Delimiter)
				continue
			}
			spec := entries.NewCutCollectSpec()
			for f, i := range ast.FieldSets {
				spec.Map(f, i)
			}

			cut := iterator.Cutter(src,
				entries.CutDelim(ast.Delimiter),
				entries.CutCollector(spec.Collector()),
			)
			r.replaceSource(ast.Source, cut)
		case *dsl.Fanout:
			if err := r.validateExistingSourceID(ast.Source); err != nil {
				log.Error("Invalid source", "error", err)
				return err
			}
			if err := r.validateNewSourceID(ast.TargetA); err != nil {
				log.Error("Invalid identifier", "error", err)
				return err
			}
			if err := r.validateNewSourceID(ast.TargetB); err != nil {
				log.Error("Invalid identifier", "error", err)
				return err
			}
			src := r.getSource(ast.Source)
			r.markConsumed(ast.Source)
			var (
				a, b iterator.Iterator
			)
			if r.dryRun {
				log.Info("Dry run fanout", "source", ast.Source, "output-a", ast.TargetA, "output-b", ast.TargetB)
			} else {
				a, b = iterator.Fanout(src)
			}
			r.addSource(ast.TargetA, a)
			r.addSource(ast.TargetB, b)
		case *dsl.Tag:
			if err := r.validateExistingSourceID(ast.Source); err != nil {
				log.Error("Invalid source", "error", err)
				return err
			}
			if r.dryRun {
				log.Info("Dry run tag", "source", ast.Source, "tag", ast.Tag)
			}
			src := r.getSource(ast.Source)
			src = iterator.Tag(src, ast.Tag)
			r.replaceSource(ast.Source, src)
		case *dsl.Join:
			if err := r.validateExistingSourceID(ast.Source); err != nil {
				log.Error("Invalid source", "error", err)
				return err
			}
			if r.dryRun {
				log.Info("Dry run join", "source", ast.Source, "patterns", ast.Patterns)
				continue
			}
			src := r.getSource(ast.Source)
			src = iterator.Joiner(src, ast.Patterns...)
			r.replaceSource(ast.Source, src)
		case *dsl.Eol:
		default:
			err := fmt.Errorf("likely bug, unhandled AST [%d] at line %d: %s", ast.Type(), ast.Line(), ast.Text())
			log.Error("Unrecognized AST", "error", err, "ast", ast)
			return err
		}
	}
	return nil
}

func (r *Runtime) argString(args []*dsl.Arg) string {
	var buf strings.Builder

	for i, arg := range args {
		if i > 0 {
			buf.WriteString(", ")
		}
		switch {
		case len(arg.String) > 0:
			buf.WriteString(arg.String)
		case len(arg.Identifier) > 0:
			buf.WriteString(arg.Identifier)
		case arg.Int != 0:
			buf.WriteString(strconv.FormatInt(arg.Int, 10))
		case arg.Number != 0:
			buf.WriteString(strconv.FormatFloat(arg.Number, 'f', 2, 64))
		default:
			buf.WriteString(arg.Text())
		}
	}
	return buf.String()
}

func (r *Runtime) validateNewSourceID(id string) error {
	if emptyID(id) {
		return ErrEmptyID
	}
	_, ok := r.sourceIDs[id]
	if ok {
		return fmt.Errorf("%w: %s", ErrAlreadyDefined, id)
	}
	return nil
}

func (r *Runtime) validateExistingSourceID(id string) error {
	if emptyID(id) {
		return ErrEmptyID
	}
	i, ok := r.sourceIDs[id]
	if !ok {
		return fmt.Errorf("%w: %s", ErrUndefined, id)
	}
	if r.consumed[i] {
		return fmt.Errorf("%w: %s", ErrConsumed, id)
	}
	return nil
}

func (r *Runtime) getSource(id string) iterator.Iterator {
	i := r.sourceIDs[id]
	return r.sources[i]
}

func (r *Runtime) addSource(id string, iter iterator.Iterator) {
	i := len(r.sources)
	r.sources = append(r.sources, iter)
	r.consumed = append(r.consumed, false)
	r.sourceIDs[id] = i
}

func (r *Runtime) replaceSource(id string, iter iterator.Iterator) {
	r.sources[r.sourceIDs[id]] = iter
}

func (r *Runtime) markConsumed(ids ...string) {
	for _, id := range ids {
		i := r.sourceIDs[id]
		r.consumed[i] = true
	}
}

func (r *Runtime) DryRun(ast ...dsl.AstNode) error {
	r.dryRun = true
	defer func() {
		r.dryRun = false
	}()
	return r.Execute(ast...)
}

func emptyID(id string) bool {
	return len(strings.TrimSpace(id)) == 0
}
