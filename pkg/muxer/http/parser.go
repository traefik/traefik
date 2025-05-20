package http

import (
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/traefik/traefik/v3/pkg/rules"
	"github.com/vulcand/predicate"
)

type SyntaxParser struct {
	parsers map[string]*parser
}

type Options func(map[string]matcherFuncs)

func WithMatcher(syntax, matcherName string, matcherFunc MatcherFunc) Options {
	return func(syntaxFuncs map[string]matcherFuncs) {
		syntax = strings.ToLower(syntax)
		syntaxFuncs[syntax][matcherName] = matcherFunc
	}
}

func NewSyntaxParser(opts ...Options) (SyntaxParser, error) {
	syntaxFuncs := map[string]matcherFuncs{
		"v2": httpFuncsV2,
		"v3": httpFuncs,
	}

	for _, opt := range opts {
		opt(syntaxFuncs)
	}

	var err error
	parsers := map[string]*parser{}
	for syntax, funcs := range syntaxFuncs {
		parsers[syntax], err = newParser(funcs)
		if err != nil {
			return SyntaxParser{}, err
		}
	}

	return SyntaxParser{
		parsers: parsers,
	}, nil
}

func (s SyntaxParser) parse(syntax string, rule string) (matchersTree, error) {
	parser, ok := s.parsers[syntax]
	if !ok {
		parser = s.parsers["v3"]
	}

	return parser.parse(rule)
}

func newParser(funcs matcherFuncs) (*parser, error) {
	p, err := rules.NewParser(slices.Collect(maps.Keys(funcs)))
	if err != nil {
		return nil, err
	}

	return &parser{
		parser:       p,
		matcherFuncs: funcs,
	}, nil
}

type parser struct {
	parser       predicate.Parser
	matcherFuncs matcherFuncs
}

func (p *parser) parse(rule string) (matchersTree, error) {
	parse, err := p.parser.Parse(rule)
	if err != nil {
		return matchersTree{}, fmt.Errorf("error while parsing rule %s: %w", rule, err)
	}
	buildTree, ok := parse.(rules.TreeBuilder)
	if !ok {
		return matchersTree{}, fmt.Errorf("error while parsing rule %s", rule)
	}

	var matchers matchersTree
	err = matchers.addRule(buildTree(), p.matcherFuncs)
	if err != nil {
		return matchersTree{}, fmt.Errorf("error while adding rule %s: %w", rule, err)
	}

	return matchers, nil
}
