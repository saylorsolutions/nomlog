package dsl

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var (
	ErrUnexpectedToken     = errors.New("unexpected token")
	ErrUndefinedIdentifier = errors.New("undefined identifier")
	ErrAlreadyDefined      = errors.New("identifier is already defined")
	ErrAlreadyConsumed     = errors.New("iterator is no longer consumable")
	errNotAMatch           = errors.New("not a match")
)

func errUndefined(id string) error {
	return fmt.Errorf("%w '%s'", ErrUndefinedIdentifier, id)
}

func errAlreadyDefined(id string) error {
	return fmt.Errorf("'%s' %w", id, ErrAlreadyDefined)
}

func errAlreadyConsumed(id string) error {
	return fmt.Errorf("'%s' %w", id, ErrAlreadyConsumed)
}

type AstType int

const (
	EOL AstType = iota
	ARG
	SOURCE_CLASS
	SOURCE
	SINK_CLASS
	SINK
	ASYNC_SINK
	MERGE
	DUPE
	APPEND
	CUT
	FANOUT
	TAG
)

func ParseString(s string) ([]AstNode, error) {
	p := newParser(lexString(s))
	dsl, err := p.parse()
	if err != nil {
		consumeTokens(p.l.tokens)
	}
	return dsl, err
}

func ParseFile(file string) ([]AstNode, error) {
	l, err := lexFile(file)
	if err != nil {
		return nil, err
	}
	p := newParser(l)
	dsl, err := p.parse()
	if err != nil {
		consumeTokens(p.l.tokens)
	}
	return dsl, err
}

func consumeTokens(ch <-chan token) {
	for range ch {
	}
}

func (p *parser) parse() ([]AstNode, error) {
	var (
		str   = p.l.stream()
		nodes []AstNode
	)
	go func() {
		p.l.lex()
	}()

	for {
		t := str.peek()
		switch t.Type {
		case tEof:
			return nodes, nil
		case tErr:
			return nodes, errors.New(t.Text)
		case tEol:
			eol, err := p.parseEol(str)
			if err != nil {
				return nil, err
			}
			nodes = append(nodes, eol)
			continue
		case tSource:
			source, err := p.parseSource(str)
			if err != nil {
				return nil, err
			}
			nodes = append(nodes, source)
		case tSink:
			sink, err := p.parseSink(str)
			if err != nil {
				return nil, err
			}
			nodes = append(nodes, sink)
		case tMerge:
			merge, err := p.parseMerge(str)
			if err != nil {
				return nil, err
			}
			nodes = append(nodes, merge)
		case tDupe:
			dupe, err := p.parseDupe(str)
			if err != nil {
				return nil, err
			}
			nodes = append(nodes, dupe)
		case tAppend:
			_append, err := p.parseAppend(str)
			if err != nil {
				return nil, err
			}
			nodes = append(nodes, _append)
		case tCut:
			cut, err := p.parseCut(str)
			if err != nil {
				return nil, err
			}
			nodes = append(nodes, cut)
		case tFanout:
			fanout, err := p.parseFanout(str)
			if err != nil {
				return nil, err
			}
			nodes = append(nodes, fanout)
		case tTag:
			tag, err := p.parseTag(str)
			if err != nil {
				return nil, err
			}
			nodes = append(nodes, tag)
		default:
			return nil, unexpected(str.next(), "EOL", "EOF", "source", "sink", "merge", "dupe", "append", "cut", "fanout")
		}
	}
}

func unexpected(t token, expected ...string) error {
	expect := "one of " + strings.Join(expected, ", ")
	return fmt.Errorf("%w: expected %s at line %d position %d", ErrUnexpectedToken, expect, t.Line, t.Pos)
}

func semantic(t token, err error) error {
	return fmt.Errorf("%w at line %d position %d", err, t.Line, t.Pos)
}

func notAMatch(err error) bool {
	return errors.Is(err, errNotAMatch)
}

// AstNode represents a node of an AST graph.
type AstNode interface {
	Line() int
	Pos() int
	Text() string
	Type() AstType
}

type ast struct {
	AstLine int     `json:"line"`
	AstPos  int     `json:"pos"`
	AstText string  `json:"text"`
	AstType AstType `json:"type"`
}

func (a *ast) Line() int {
	return a.AstLine
}
func (a *ast) Pos() int {
	return a.AstPos
}
func (a *ast) Text() string {
	return a.AstText
}
func (a *ast) Type() AstType {
	return a.AstType
}

func (a *ast) setVals(t token, typ AstType) {
	a.AstLine = t.Line
	a.AstPos = t.Pos
	a.AstText = t.Text
	a.AstType = typ
}
func (a *ast) append(t token) {
	a.AstText += t.Text
}
func (a *ast) appendSpace(t token) {
	a.AstText += " " + t.Text
}
func (a *ast) appendText(s string) {
	a.AstText += s
}
func (a *ast) appendTextSpace(s string) {
	a.AstText += " " + s
}

type parser struct {
	l        *lexer
	sources  map[string]bool
	consumed map[string]bool
	sinks    map[string]bool
}

func newParser(l *lexer) *parser {
	return &parser{
		l:        l,
		sources:  map[string]bool{},
		consumed: map[string]bool{},
		sinks:    map[string]bool{},
	}
}

type Eol struct {
	ast
}

func (p *parser) parseRequiredEol(str *tokenStream) (*Eol, error) {
	eol, err := p.parseEol(str)
	if notAMatch(err) {
		return nil, unexpected(str.peek(), "end of file", "end of line")
	}
	return eol, err
}

func (p *parser) parseEol(str *tokenStream) (*Eol, error) {
	t := str.next()
	if t.Type == tEof || t.Type == tEol {
		eol := new(Eol)
		eol.setVals(t, EOL)
		return eol, nil
	}
	str.pushBack(t)
	return nil, errNotAMatch
}

type Arg struct {
	ast
	String     string  `json:"string"`
	Number     float64 `json:"number"`
	Int        int64   `json:"int"`
	Identifier string  `json:"identifier"`
}

func escapeString(s string) string {
	s = strings.TrimPrefix(strings.TrimSuffix(s, "\""), "\"")
	s = strings.ReplaceAll(s, `\r`, "\r")
	s = strings.ReplaceAll(s, `\n`, "\n")
	s = strings.ReplaceAll(s, `\t`, "\t")
	s = strings.ReplaceAll(s, `\"`, "\"")
	s = strings.ReplaceAll(s, `\\`, "\\")
	return s
}

func (p *parser) parseArg(str *tokenStream) (*Arg, error) {
	t := str.next()
	switch t.Type {
	case tString:
		a := &Arg{String: escapeString(t.Text)}
		a.setVals(t, ARG)
		return a, nil
	case tNumber:
		n, err := strconv.ParseFloat(t.Text, 64)
		if err != nil {
			return nil, errors.New("invalid float")
		}
		a := &Arg{Number: n}
		a.setVals(t, ARG)
		return a, nil
	case tInt:
		i, err := strconv.ParseInt(t.Text, 10, 64)
		if err != nil {
			return nil, errors.New("invalid int")
		}
		a := &Arg{Int: i}
		a.setVals(t, ARG)
		return a, nil
	case tIdentifier:
		id := t.Text
		if !p.sources[id] && !p.sinks[id] {
			return nil, semantic(t, errUndefined(id))
		}
		a := &Arg{Identifier: id}
		a.setVals(t, ARG)
		return a, nil
	default:
		str.pushBack(t)
		return nil, errNotAMatch
	}
}

func (p *parser) parseArgs(str *tokenStream) ([]*Arg, error) {
	var (
		args []*Arg
	)

	for {
		if len(args) > 0 {
			t := str.next()
			if t.Type != tComma {
				str.pushBack(t)
				return args, nil
			}
		}
		a, err := p.parseArg(str)
		if err != nil {
			if notAMatch(err) {
				if len(args) == 0 {
					return nil, unexpected(str.peek(), "argument")
				}
				return args, nil
			}
			return nil, err
		}
		args = append(args, a)
	}
}

type SourceClass struct {
	ast
	Qualifier   string `json:"qualifier"`
	SourceClass string `json:"class"`
}

func (p *parser) parseSourceClass(str *tokenStream) (*SourceClass, error) {
	sc := new(SourceClass)
	qual := str.next()
	if qual.Type != tIdentifier {
		str.pushBack(qual)
		return nil, unexpected(qual, "source class qualifier")
	}
	sc.Qualifier = qual.Text
	sc.setVals(qual, SOURCE_CLASS)

	dot := str.next()
	if dot.Type != tDot {
		str.pushBack(dot, qual)
		return nil, unexpected(dot, "dot separator")
	}
	sc.append(dot)

	id := str.next()
	if id.Type != tIdentifier {
		str.pushBack(id, dot, qual)
		return nil, unexpected(id, "source class identifier")
	}
	sc.SourceClass = id.Text
	sc.append(id)
	return sc, nil
}

type Source struct {
	ast
	ID    string       `json:"id"`
	Class *SourceClass `json:"sourceClass"`
	Args  []*Arg       `json:"args"`
}

func (p *parser) parseSource(str *tokenStream) (*Source, error) {
	src := new(Source)
	s := str.next()
	if s.Type != tSource {
		str.pushBack(s)
		return nil, errNotAMatch
	}
	src.setVals(s, SOURCE)

	as := str.next()
	if as.Type != tAs {
		return nil, unexpected(as, "as")
	}
	src.appendSpace(as)
	id := str.next()
	if id.Type != tIdentifier {
		return nil, unexpected(id, "source identifier")
	}
	if p.sources[id.Text] {
		return nil, semantic(id, errAlreadyDefined(id.Text))
	}
	p.sources[id.Text] = true
	src.ID = id.Text
	src.appendSpace(id)

	sc, err := p.parseSourceClass(str)
	if err != nil {
		return nil, err
	}
	src.Class = sc
	src.appendTextSpace(sc.AstText)

	args, err := p.parseArgs(str)
	if err != nil {
		return nil, err
	}
	src.Args = args

	for i, a := range args {
		if i > 0 {
			src.appendText(",")
		}
		src.appendTextSpace(a.AstText)
	}

	_, err = p.parseRequiredEol(str)
	if err != nil {
		return nil, err
	}
	return src, nil
}

type SinkClass struct {
	ast
	Qualifier string `json:"qualifier"`
	SinkClass string `json:"class"`
}

func (p *parser) parseSinkClass(str *tokenStream) (*SinkClass, error) {
	sc := new(SinkClass)
	qual := str.next()
	if qual.Type != tIdentifier {
		str.pushBack(qual)
		return nil, unexpected(qual, "sink class qualifier")
	}
	sc.Qualifier = qual.Text
	sc.setVals(qual, SINK_CLASS)

	dot := str.next()
	if dot.Type != tDot {
		str.pushBack(dot, qual)
		return nil, unexpected(dot, "dot separator")
	}
	sc.append(dot)

	id := str.next()
	if id.Type != tIdentifier {
		str.pushBack(id, dot, qual)
		return nil, unexpected(id, "sink class identifier")
	}
	sc.SinkClass = id.Text
	sc.append(id)
	return sc, nil
}

type Sink struct {
	ast
	Source string     `json:"source"`
	Async  bool       `json:"async"`
	ID     string     `json:"id,omitempty"`
	Class  *SinkClass `json:"sinkClass"`
	Args   []*Arg     `json:"args"`
}

func (p *parser) parseSink(str *tokenStream) (*Sink, error) {
	sink := new(Sink)
	s := str.next()
	if s.Type != tSink {
		str.pushBack(s)
		return nil, errNotAMatch
	}
	sink.setVals(s, SINK)

	iterID := str.next()
	if iterID.Type != tIdentifier {
		return nil, unexpected(iterID, "iterator identifier")
	}
	if !p.sources[iterID.Text] {
		return nil, semantic(iterID, errUndefined(iterID.Text))
	}
	if p.consumed[iterID.Text] {
		return nil, semantic(iterID, errAlreadyConsumed(iterID.Text))
	}
	sink.Source = iterID.Text
	p.consumed[iterID.Text] = true
	sink.appendSpace(iterID)

	asyncTo := str.next()
	if asyncTo.Type == tAsync {
		sink.AstType = ASYNC_SINK
		sink.appendSpace(asyncTo)
		as := str.next()
		if as.Type != tAs {
			return nil, unexpected(as, "as")
		}
		sink.appendSpace(as)
		id := str.next()
		if id.Type != tIdentifier {
			return nil, unexpected(id, "sink identifier")
		}
		if p.sources[id.Text] {
			return nil, semantic(id, ErrAlreadyDefined)
		}
		if p.sinks[id.Text] {
			return nil, semantic(id, errAlreadyDefined(id.Text))
		}
		sink.ID = id.Text
		p.sinks[id.Text] = true
		sink.appendSpace(id)

		to := str.next()
		if to.Type != tTo {
			return nil, unexpected(to, "to")
		}
		sink.appendSpace(to)
	} else if asyncTo.Type != tTo {
		return nil, unexpected(asyncTo, "to", "async")
	} else {
		sink.appendSpace(asyncTo)
	}

	sc, err := p.parseSinkClass(str)
	if err != nil {
		return nil, err
	}
	sink.Class = sc
	sink.appendTextSpace(sc.AstText)

	args, err := p.parseArgs(str)
	if err != nil {
		return nil, err
	}
	sink.Args = args

	for i, a := range args {
		if i > 0 {
			sink.appendText(",")
		}
		sink.appendTextSpace(a.AstText)
	}

	_, err = p.parseRequiredEol(str)
	if err != nil {
		return nil, err
	}
	return sink, nil
}

type Merge struct {
	ast
	SourceA string `json:"sourceA"`
	SourceB string `json:"sourceB"`
	ID      string `json:"id"`
}

func (p *parser) parseMerge(str *tokenStream) (*Merge, error) {
	merge := new(Merge)
	m := str.next()
	if m.Type != tMerge {
		str.pushBack(m)
		return nil, errNotAMatch
	}
	merge.setVals(m, MERGE)

	a := str.next()
	if a.Type != tIdentifier {
		return nil, unexpected(a, "source identifier")
	}
	if !p.sources[a.Text] {
		return nil, semantic(a, errUndefined(a.Text))
	}
	if p.consumed[a.Text] {
		return nil, semantic(a, errAlreadyConsumed(a.Text))
	}
	p.consumed[a.Text] = true
	merge.SourceA = a.Text
	merge.appendSpace(a)

	and := str.next()
	if and.Type != tAnd {
		return nil, unexpected(and, "and")
	}
	merge.appendSpace(and)

	b := str.next()
	if b.Type != tIdentifier {
		return nil, unexpected(b, "source identifier")
	}
	if !p.sources[b.Text] {
		return nil, semantic(b, errUndefined(b.Text))
	}
	if p.consumed[b.Text] {
		return nil, semantic(b, errAlreadyConsumed(b.Text))
	}
	p.consumed[b.Text] = true
	merge.SourceB = b.Text
	merge.appendSpace(b)

	as := str.next()
	if as.Type != tAs {
		return nil, unexpected(as, "as")
	}
	merge.appendSpace(as)

	id := str.next()
	if id.Type != tIdentifier {
		return nil, unexpected(id, "merged identifier")
	}
	if p.sources[id.Text] {
		return nil, semantic(id, errAlreadyDefined(id.Text))
	}
	merge.ID = id.Text
	p.sources[id.Text] = true
	merge.appendSpace(id)

	_, err := p.parseRequiredEol(str)
	if err != nil {
		return nil, err
	}
	return merge, nil
}

type Dupe struct {
	ast
	Source  string `json:"source"`
	TargetA string `json:"targetA"`
	TargetB string `json:"targetB"`
}

func (p *parser) parseDupe(str *tokenStream) (*Dupe, error) {
	dupe := new(Dupe)
	d := str.next()
	if d.Type != tDupe {
		str.pushBack(d)
		return nil, errNotAMatch
	}
	dupe.setVals(d, DUPE)

	src := str.next()
	if src.Type != tIdentifier {
		return nil, unexpected(src, "source identifier")
	}
	if !p.sources[src.Text] {
		return nil, semantic(src, errUndefined(src.Text))
	}
	if p.consumed[src.Text] {
		return nil, semantic(src, errAlreadyConsumed(src.Text))
	}
	p.consumed[src.Text] = true
	dupe.Source = src.Text
	dupe.appendSpace(src)

	as := str.next()
	if as.Type != tAs {
		return nil, unexpected(as, "as")
	}
	dupe.appendSpace(as)

	a := str.next()
	if a.Type != tIdentifier {
		return nil, unexpected(a, "target identifier")
	}
	if p.sources[a.Text] {
		return nil, semantic(a, errAlreadyDefined(a.Text))
	}
	dupe.TargetA = a.Text
	p.sources[a.Text] = true
	dupe.appendSpace(a)

	and := str.next()
	if and.Type != tAnd {
		return nil, unexpected(and, "and")
	}
	dupe.appendSpace(and)

	b := str.next()
	if b.Type != tIdentifier {
		return nil, unexpected(b, "target identifier")
	}
	if p.sources[b.Text] {
		return nil, semantic(b, errAlreadyDefined(b.Text))
	}
	dupe.TargetB = b.Text
	p.sources[b.Text] = true
	dupe.appendSpace(b)

	_, err := p.parseRequiredEol(str)
	if err != nil {
		return nil, err
	}
	return dupe, nil
}

type Append struct {
	ast
	Source string `json:"source"`
	Target string `json:"target"`
}

func (p *parser) parseAppend(str *tokenStream) (*Append, error) {
	apnd := new(Append)
	a := str.next()
	if a.Type != tAppend {
		str.pushBack(a)
		return nil, errNotAMatch
	}
	apnd.setVals(a, APPEND)

	src := str.next()
	if src.Type != tIdentifier {
		return nil, unexpected(src, "source identifier")
	}
	if !p.sources[src.Text] {
		return nil, semantic(src, errUndefined(src.Text))
	}
	if p.consumed[src.Text] {
		return nil, semantic(src, errAlreadyConsumed(src.Text))
	}
	apnd.Source = src.Text
	p.consumed[src.Text] = true
	apnd.appendSpace(src)

	to := str.next()
	if to.Type != tTo {
		return nil, unexpected(to, "to")
	}
	apnd.appendSpace(to)

	trg := str.next()
	if trg.Type != tIdentifier {
		return nil, unexpected(trg, "target identifier")
	}
	if !p.sources[trg.Text] {
		return nil, semantic(trg, errUndefined(trg.Text))
	}
	apnd.Target = trg.Text
	apnd.appendSpace(trg)

	_, err := p.parseRequiredEol(str)
	if err != nil {
		return nil, err
	}
	return apnd, nil
}

type Cut struct {
	ast
	Delimiter string         `json:"delimiter"`
	Source    string         `json:"source"`
	FieldSets map[string]int `json:"fieldSets"`
}

func (p *parser) parseCut(str *tokenStream) (*Cut, error) {
	cut := &Cut{FieldSets: map[string]int{}}
	cut.Delimiter = " "

	c := str.next()
	if c.Type != tCut {
		str.pushBack(c)
		return nil, errNotAMatch
	}
	cut.setVals(c, CUT)

	withId := str.next()
	if withId.Type == tWith {
		cut.appendSpace(withId)
		s := str.next()
		if s.Type != tString {
			return nil, unexpected(s, "string delimiter")
		}
		cut.Delimiter = escapeString(s.Text)
		cut.appendSpace(s)
		withId = str.next()
	}
	if withId.Type != tIdentifier {
		return nil, unexpected(withId, "source identifier")
	}
	if p.consumed[withId.Text] {
		return nil, semantic(withId, errAlreadyConsumed(withId.Text))
	}
	cut.Source = withId.Text
	cut.appendSpace(withId)

	set := str.next()
	if set.Type != tSet {
		return nil, unexpected(set, "set")
	}
	cut.appendSpace(set)

	lp := str.next()
	if lp.Type != tLpar {
		return nil, unexpected(lp, "(")
	}
	cut.appendSpace(lp)

loop:
	for first := true; ; first = false {
		if !first {
			commaParen := str.next()
			switch commaParen.Type {
			case tComma:
				cut.append(commaParen)
			case tRpar:
				cut.append(commaParen)
				break loop
			default:
				return nil, unexpected(commaParen, ",", ")")
			}
		}

		id := str.next()
		if id.Type != tIdentifier {
			if id.Type == tRpar {
				str.pushBack(id)
				break
			}
			return nil, unexpected(id, "field set identifier")
		}

		eq := str.next()
		if eq.Type != tEq {
			return nil, unexpected(eq, "=")
		}

		num := str.next()
		if num.Type != tInt {
			return nil, unexpected(num, "int field number")
		}
		i, err := strconv.Atoi(num.Text)
		if err != nil {
			return nil, err
		}
		cut.FieldSets[id.Text] = i
		if first {
			cut.append(id)
		} else {
			cut.appendSpace(id)
		}
		cut.appendSpace(eq)
		cut.appendSpace(num)
	}

	_, err := p.parseRequiredEol(str)
	if err != nil {
		return nil, err
	}
	return cut, nil
}

type Fanout struct {
	ast
	Source  string `json:"source"`
	TargetA string `json:"targetA"`
	TargetB string `json:"targetB"`
}

func (p *parser) parseFanout(str *tokenStream) (*Fanout, error) {
	fanout := new(Fanout)

	f := str.next()
	if f.Type != tFanout {
		return nil, errNotAMatch
	}
	fanout.setVals(f, FANOUT)

	src := str.next()
	if src.Type != tIdentifier {
		return nil, unexpected(src, "source identifier")
	}
	if !p.sources[src.Text] {
		return nil, semantic(src, ErrUndefinedIdentifier)
	}
	if p.consumed[src.Text] {
		return nil, semantic(src, ErrAlreadyConsumed)
	}
	fanout.Source = src.Text
	p.consumed[src.Text] = true
	fanout.appendSpace(src)

	as := str.next()
	if as.Type != tAs {
		return nil, unexpected(as, "as")
	}
	fanout.appendSpace(as)

	a := str.next()
	if a.Type != tIdentifier {
		return nil, unexpected(a, "target identifier")
	}
	if p.sources[a.Text] {
		return nil, semantic(a, errAlreadyDefined(a.Text))
	}
	fanout.TargetA = a.Text
	p.sources[a.Text] = true
	fanout.appendSpace(a)

	and := str.next()
	if and.Type != tAnd {
		return nil, unexpected(and, "and")
	}
	fanout.appendSpace(and)

	b := str.next()
	if b.Type != tIdentifier {
		return nil, unexpected(b, "target identifier")
	}
	if p.sources[b.Text] {
		return nil, semantic(b, errAlreadyDefined(b.Text))
	}
	fanout.TargetB = b.Text
	p.sources[b.Text] = true
	fanout.appendSpace(b)

	_, err := p.parseRequiredEol(str)
	if err != nil {
		return nil, err
	}
	return fanout, nil
}

type Tag struct {
	ast
	Source string
	Tag    string
}

func (p *parser) parseTag(str *tokenStream) (*Tag, error) {
	t := new(Tag)

	tagKw := str.next()
	if tagKw.Type != tTag {
		return nil, errNotAMatch
	}
	t.setVals(tagKw, TAG)

	src := str.next()
	if src.Type != tIdentifier {
		return nil, unexpected(src, "source identifier")
	}
	t.Source = src.Text
	t.appendSpace(src)

	with := str.next()
	if with.Type != tWith {
		return nil, unexpected(with, "with")
	}
	t.appendSpace(with)

	tag := str.next()
	if tag.Type != tString {
		return nil, unexpected(tag, "tag string")
	}
	t.Tag = escapeString(tag.Text)
	t.appendSpace(tag)

	_, err := p.parseRequiredEol(str)
	if err != nil {
		return nil, err
	}
	return t, nil
}
