package parser

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"text/scanner"
	"unicode"

	clutterScanner "github.com/cluttercode/clutter/internal/pkg/scanner"

	"github.com/cluttercode/clutter/pkg/clutter/clutterindex"
)

// [# %stop #]
// [# @attr1=v1 @attr2=v2 ./name attr3=v1 attr4 ... attrN #]
// [# %cont #]

const exact = "exact"

var (
	validNameRegexp     = regexp.MustCompile(`^[\w_][\w_\:\-\.\/]*$`)
	validAttrNameRegexp = regexp.MustCompile(`^[\w_][\w_\:\-]*$`)
)

func ParseElement(elem *clutterScanner.RawElement) (*clutterindex.Entry, error) {
	var s scanner.Scanner
	s.Init(strings.NewReader(elem.Text))
	s.Mode = scanner.ScanIdents | scanner.ScanStrings | scanner.ScanRawStrings

	var err error

	s.Error = func(_ *scanner.Scanner, msg string) { err = fmt.Errorf("%s", msg) }

	ent := clutterindex.Entry{Loc: elem.Loc, Attrs: map[string]string{}}

	search := ""

	var state func(string, bool) error

	addAttr := func(k, v string) error {
		if !validAttrNameRegexp.MatchString(k) {
			return fmt.Errorf("invalid attribute name: %q", k)
		}

		if vv, ok := ent.Attrs[k]; ok {
			return fmt.Errorf("attribute %q already set to %q", k, vv)
		}

		if k == "scope" {
			if v == "." {
				v = elem.Loc.Path
			} else if v == "./" {
				if v = filepath.Dir(elem.Loc.Path); v == "" || v == "." {
					// root
					return nil
				} else if v[len(v)-1] != '/' {
					v += "/"
				}
			}
		} else if k == "search" {
			// search types need to correspond to [# search-cli-exp-type-flags #].
			switch v {
			case "", exact:
				search = exact
			case "g", "gl":
				search = "glob"
			case "e", "re":
				search = "regexp"
			default:
				return fmt.Errorf("invalid search type: %q", v)
			}

			v = search
		}

		ent.Attrs[k] = v
		return nil
	}

	attr := func(k string, back func(string, bool) error) func(string, bool) error {
		eq := false

		return func(tok string, eol bool) error {
			if eol {
				return addAttr(k, "")
			}

			if tok[0] == '"' {
				var err error
				tok, err = strconv.Unquote(tok)
				if err != nil {
					return fmt.Errorf("invalid quotes: %w", err)
				}
			}

			if !eq {
				if eq = tok == "="; eq {
					return nil
				}

				if err = addAttr(k, ""); err != nil {
					return err
				}

				state = back
				return state(tok, false)
			}

			if err = addAttr(k, tok); err != nil {
				return err
			}

			state = back
			return nil
		}
	}

	post := func(tok string, eol bool) error {
		if eol {
			return nil
		}

		state = attr(tok, state)

		return nil
	}

	isPre := true

	pre := func(tok string, eol bool) error {
		if eol {
			return fmt.Errorf("empty element")
		}

		if tok[0] == '?' {
			return addAttr("search", tok[1:])
		}

		if tok[0] == '@' {
			state = attr(tok[1:], state)
			return nil
		}

		if strings.HasPrefix(tok, "./") {
			if err := addAttr("scope", "./"); err != nil {
				return err
			}

			tok = tok[2:]
		} else if tok[0] == '.' {
			if err := addAttr("scope", "."); err != nil {
				return err
			}

			tok = tok[1:]
		}

		if tok[0] == '"' {
			tok, err = strconv.Unquote(tok)
			if err != nil {
				return fmt.Errorf("invalid quotes: %w", err)
			}
		}

		// TODO: also validate patterns.
		if search == "" || search == exact {
			if !validNameRegexp.MatchString(tok) {
				return fmt.Errorf("invalid name: %q", tok)
			}
		}

		ent.Name = tok

		isPre = false

		state = post

		return nil
	}

	dot := false

	s.IsIdentRune = func(r rune, i int) bool {
		if i == 0 {
			dot = r == '.'

			return (isPre && (r == '@' || r == '.' || r == '?')) || unicode.IsLetter(r)
		}

		if i == 1 && dot && r == '/' {
			return true
		}

		dot = false

		return unicode.IsLetter(r) || unicode.IsDigit(r) || strings.ContainsRune("-_:/", r)
	}

	state = pre

	for tok := s.Scan(); tok != scanner.EOF; tok = s.Scan() {
		if err != nil {
			return nil, err
		}

		if err = state(s.TokenText(), false); err != nil {
			return nil, err
		}
	}

	if err = state("", true); err != nil {
		return nil, err
	}

	if len(ent.Attrs) == 0 {
		ent.Attrs = nil
	}

	return &ent, nil
}

func ParseElements(elems []*clutterScanner.RawElement) ([]*clutterindex.Entry, error) {
	ents := make([]*clutterindex.Entry, len(elems))
	for i, el := range elems {
		var err error
		ents[i], err = ParseElement(el)
		if err != nil {
			return nil, fmt.Errorf("parse %q@%v: %w", el.Text, el.Loc, err)
		}
	}

	return ents, nil
}
