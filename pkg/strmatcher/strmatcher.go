package strmatcher

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/cluttercode/clutter/pkg/gitignore"
)

type Matcher func(string) bool
type Matchers []Matcher

type Compiler func(string) (Matcher, error)

// ---

var (
	_ Matcher = Matchers{}.All
	_ Matcher = Matchers{}.Any
)

func (ms Matchers) All(text string) bool {
	for _, fn := range ms {
		if !fn(text) {
			return false
		}
	}

	return true
}

func (ms Matchers) Any(text string) bool {
	for _, fn := range ms {
		if fn(text) {
			return true
		}
	}

	return false
}

// ---

func NewMatchers(compile Compiler, patterns []string) (Matchers, error) {
	fns := make([]Matcher, len(patterns))

	for i, p := range patterns {
		var err error

		if fns[i], err = compile(p); err != nil {
			return nil, fmt.Errorf("pattern #%d:%q compile error: %w", i, p, err)
		}
	}

	return fns, nil
}

// ---

func CompileExactMatcher(opt string) (Matcher, error) {
	return func(text string) bool { return opt == text }, nil
}

func CompileGlobMatcher(pattern string) (Matcher, error) {
	p := gitignore.ParsePattern(pattern, nil)

	return func(path string) bool {
		parts := strings.Split(path, string(filepath.Separator))
		isDir := len(path) > 0 && path[len(path)-1] == filepath.Separator

		// Don't care about inclusion/exclusion.
		return p.Match(parts, isDir) != gitignore.NoMatch
	}, nil
}

func CompileRegexpMatcher(pattern string) (Matcher, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}

	return re.MatchString, nil
}

// ---

func NewExactMatchers(options []string) Matchers {
	ms, _ := NewMatchers(CompileExactMatcher, options)
	return ms
}

func NewGlobMatchers(patterns []string) (Matchers, error) {
	return NewMatchers(CompileGlobMatcher, patterns)
}

func NewRegexpMatchers(patterns []string) (Matchers, error) {
	return NewMatchers(CompileRegexpMatcher, patterns)
}
