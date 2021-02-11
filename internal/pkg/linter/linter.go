package linter

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/Knetic/govaluate"
	"github.com/gobwas/glob"
	"go.uber.org/zap"

	"github.com/cluttercode/clutter/internal/pkg/index"
)

type internalRule struct {
	checkPath func(string) bool
	eval      func(context.Context, *index.Entry) (bool, error)
}

type Linter struct {
	z *zap.SugaredLogger

	config Config

	rules []internalRule
}

func (ir *internalRule) init(l *Linter, r Rule) error {
	l.z.Debugw("init rule", "r", r)

	if r.PathGlob != "" && r.PathRegexp != "" {
		return fmt.Errorf("path-pattern and path-re are mutuallye exclusive")
	}

	if len(r.Shell) != 0 && r.Expr != "" {
		return fmt.Errorf("shell and expr are mutually exclusive")
	}

	if len(r.Shell) == 0 && r.Expr == "" {
		return fmt.Errorf("either shell or expr are required")
	}

	ir.checkPath = func(string) bool { return true }

	if g := r.PathGlob; g != "" {
		m, err := glob.Compile(g)
		if err != nil {
			return fmt.Errorf("path-glob: %w", err)
		}

		ir.checkPath = func(path string) bool {
			ok := m.Match(path)

			l.z.Debugw("glob match", "path", path, "ok", ok, "glob", g)

			return ok
		}
	}

	if re := r.PathRegexp; re != "" {
		m, err := regexp.Compile(re)
		if err != nil {
			return fmt.Errorf("path-re: %w", err)
		}

		ir.checkPath = func(path string) bool {
			ok := m.MatchString(path)

			l.z.Debugw("re match", "path", path, "ok", ok, "re", re)

			return ok
		}
	}

	if expr := r.Expr; expr != "" {
		funcs := map[string]govaluate.ExpressionFunction{
			"re_match": func(vs ...interface{}) (interface{}, error) {
				if len(vs) != 2 {
					return nil, fmt.Errorf("expecting two arguments")
				}

				p, ok := vs[0].(string)
				if !ok {
					return nil, fmt.Errorf("pattern argument must be a string")
				}

				s, ok := vs[1].(string)
				if !ok {
					return nil, fmt.Errorf("text argument must be a string")
				}

				return regexp.MatchString(p, s)
			},
			"glob_match": func(vs ...interface{}) (interface{}, error) {
				if len(vs) != 2 {
					return nil, fmt.Errorf("expecting two arguments")
				}

				p, ok := vs[0].(string)
				if !ok {
					return nil, fmt.Errorf("pattern argument must be a string")
				}

				s, ok := vs[1].(string)
				if !ok {
					return nil, fmt.Errorf("text argument must be a string")
				}

				g, err := glob.Compile(p)
				if err != nil {
					return nil, err
				}

				return g.Match(s), nil
			},
		}

		evalexpr, err := govaluate.NewEvaluableExpressionWithFunctions(expr, funcs)
		if err != nil {
			return fmt.Errorf("expr: %w", err)
		}

		ir.eval = func(_ context.Context, ent *index.Entry) (bool, error) {
			return l.eval(evalexpr, ent)
		}
	}

	if cmd := r.Shell; len(cmd) > 0 {
		ir.eval = func(ctx context.Context, ent *index.Entry) (bool, error) {
			return l.shell(ctx, cmd, ent)
		}
	}

	return nil
}

func NewLinter(z *zap.SugaredLogger, cfg Config) (*Linter, error) {
	l := &Linter{
		z:      z,
		config: cfg,
		rules:  make([]internalRule, len(cfg.Rules)),
	}

	for i, r := range cfg.Rules {
		if err := l.rules[i].init(l, r); err != nil {
			return nil, fmt.Errorf("rule %d: %w", i, err)
		}
	}

	return l, nil
}

func (l *Linter) Rule(i int) *Rule {
	if i >= len(l.config.Rules) {
		return nil
	}

	return &l.config.Rules[i]
}

func (l *Linter) LintRule(ctx context.Context, i int, ent *index.Entry) (bool, error) {
	rule := l.rules[i]

	if !rule.checkPath(ent.Loc.Path) {
		l.z.Debugw("not matching path", "path", ent.Loc.Path)
		return true, nil
	}

	pass, err := rule.eval(ctx, ent)
	if err != nil {
		return false, err
	}

	return pass, nil
}

func (l *Linter) Lint(ctx context.Context, ent *index.Entry) ([]int, error) {
	var fails []int

	for i, rule := range l.config.Rules {
		name := fmt.Sprintf("%q", rule.Name)
		if name == "" {
			name = fmt.Sprintf("#%d", i)
		}

		title := fmt.Sprintf("rule %s: ", name)

		l.z.Debugw("checking rule", "name", name, "i", i)

		ok, err := l.LintRule(ctx, i, ent)
		if err != nil {
			return nil, fmt.Errorf("%s%w", title, err)
		}

		if !ok {
			fails = append(fails, i)
		}
	}

	return fails, nil
}

func entVars(ent *index.Entry) map[string]interface{} {
	m := map[string]interface{}{
		"NAME": ent.Name,
		"PATH": ent.Loc.Path,
	}

	for k, v := range ent.Attrs {
		m[strings.ToUpper(fmt.Sprintf("ATTR_%s", k))] = v
	}

	return m
}

func (l *Linter) shell(ctx context.Context, cmdParts []string, ent *index.Entry) (bool, error) {
	vars := entVars(ent)

	varsList := make([]string, 0, len(vars))
	for k, v := range vars {
		varsList = append(varsList, fmt.Sprintf("ENT_%s=%s", k, v))
	}

	cmd := exec.CommandContext(ctx, cmdParts[0], cmdParts[1:]...)

	cmd.Env = append(os.Environ(), varsList...)

	l.z.Infow("execution shell lint rule", "cmd", cmd)

	if err := cmd.Run(); err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			l.z.Info("returned non-zero")
			return false, nil
		}

		l.z.Errorw("shell error", "err", err)

		return false, fmt.Errorf("shell: %w", err)
	}

	l.z.Info("returned zero")

	return true, nil
}

func (l *Linter) eval(eval *govaluate.EvaluableExpression, ent *index.Entry) (bool, error) {
	l.z.Infow("checking expression")

	res, err := eval.Evaluate(map[string]interface{}{"name": ent.Name, "path": ent.Loc.Path, "attrs": ent.Attrs.ToStruct()}) // [# govaluate-params #]
	if err != nil {
		l.z.Errorw("eval error", "err", err)
		return false, err
	}

	l.z.Infow("returned", "res", res)

	pass, ok := res.(bool)

	if !ok {
		l.z.Errorw("non-boolean result", "res", res)
		return false, fmt.Errorf("expression result is not a boolean: %v", res)
	}

	return pass, nil
}
