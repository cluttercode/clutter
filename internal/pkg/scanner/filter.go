package scanner

import (
	"os"
	"path/filepath"

	"github.com/go-git/go-git/plumbing/format/gitignore"
	"go.uber.org/zap"
)

func NewFilter(z *zap.SugaredLogger, cfg Config) (func(string, os.FileInfo) (bool, error), error) {
	ignores := make([]gitignore.Pattern, len(cfg.Ignore))

	for i, ig := range cfg.Ignore {
		ignores[i] = gitignore.ParsePattern(ig, nil)
	}

	exclude := gitignore.NewMatcher(ignores).Match

	return func(path string, fi os.FileInfo) (bool, error) {
		isDir := fi.IsDir()

		if exclude(filepath.SplitList(path), isDir) {
			if isDir {
				return false, filepath.SkipDir
			}

			return false, nil
		}

		return !isDir, nil
	}, nil
}
