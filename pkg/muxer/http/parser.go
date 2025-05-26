package http

import (
	"errors"
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

type Options func(map[string]matcherBuilderFuncs)

func WithMatcher(syntax, matcherName string, builderFunc func(params ...string) (MatcherFunc, error)) Options {
	return func(syntaxFuncs map[string]matcherBuilderFuncs) {
		syntax = strings.ToLower(syntax)

		syntaxFuncs[syntax][matcherName] = func(tree *matchersTree, s ...string) error {
			matcher, err := builderFunc(s...)
			if err != nil {
				return fmt.Errorf("building matcher: %w", err)
			}

			tree.matcher = matcher
			return nil
		}
	}
}

func NewSyntaxParser(opts ...Options) (SyntaxParser, error) {
	syntaxFuncs := map[string]matcherBuilderFuncs{
		"v2": httpFuncsV2,
		"v3": httpFuncs,
	}

	for _, opt := range opts {
		opt(syntaxFuncs)
	}

	parsers := map[string]*parser{}
	for syntax, funcs := range syntaxFuncs {
		var err error
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

func newParser(funcs matcherBuilderFuncs) (*parser, error) {
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
	matcherFuncs matcherBuilderFuncs
}

func (p *parser) parse(rule string) (matchersTree, error) {
	parse, err := p.parser.Parse(rule)
	if err != nil {
		return matchersTree{}, fmt.Errorf("parsing rule %s: %w", rule, err)
	}
	buildTree, ok := parse.(rules.TreeBuilder)
	if !ok {
		return matchersTree{}, errors.New("obtaining build tree")
	}

	var matchers matchersTree
	err = matchers.addRule(buildTree(), p.matcherFuncs)
	if err != nil {
		return matchersTree{}, fmt.Errorf("adding rule %s: %w", rule, err)
	}

	return matchers, nil
}
