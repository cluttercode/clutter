package scanner

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/plumbing/format/gitignore"
	"go.uber.org/zap"
)

var defaultIgnores = []string{
	".git",
}

func NewFilter(z *zap.SugaredLogger, cfg Config) (func(string, os.FileInfo) (bool, error), error) {
	if len(cfg.Ignore) == 0 {
		cfg.Ignore = defaultIgnores
	}

	ignores := make([]gitignore.Pattern, len(cfg.Ignore))

	for i, ig := range cfg.Ignore {
		ignores[i] = gitignore.ParsePattern(ig, nil)
	}

	exclude := gitignore.NewMatcher(ignores).Match

	return func(path string, fi os.FileInfo) (bool, error) {
		isDir := fi.IsDir()

		split := strings.Split(path, string(filepath.Separator))

		z := z.With("path", split, "is_dir", isDir)

		if exclude(split, isDir) {
			z.Debug("exclude")

			if isDir {
				return false, filepath.SkipDir
			}

			return false, nil
		}

		z.Debug("include")

		return !isDir, nil
	}, nil
}
