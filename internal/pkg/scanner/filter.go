package scanner

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/cluttercode/clutter/pkg/gitignore"
	"github.com/cluttercode/clutter/pkg/zlog"
)

var defaultIgnores = []string{
	".git",
}

func NewFilter(z *zlog.Logger, cfg Config) (func(string, os.FileInfo) (bool, error), error) {
	if len(cfg.Ignore) == 0 {
		cfg.Ignore = defaultIgnores
	}

	ignores := make([]gitignore.Pattern, len(cfg.Ignore))

	for i, ig := range cfg.Ignore {
		ignores[i] = gitignore.ParsePattern(ig, nil)
	}

	exclude := gitignore.NewMatcher(ignores).Match

	return func(path string, fi os.FileInfo) (bool, error) {
		var (
			isDir, isLink bool
			mode          os.FileMode
		)

		if fi != nil {
			mode = fi.Mode()
			isDir = fi.IsDir()
			isLink = mode&os.ModeSymlink != 0
		}

		split := strings.Split(path, string(filepath.Separator))

		z := z.With("path", split, "is_link", isLink, "is_dir", isDir, "mode", mode)

		if isLink {
			z.Debug("links are excluded")

			return false, nil
		}

		if exclude(split, isDir) {
			z.Debug("exclude dir")

			if isDir {
				return false, filepath.SkipDir
			}

			return false, nil
		}

		z.Debug("include")

		return !isDir, nil
	}, nil
}
