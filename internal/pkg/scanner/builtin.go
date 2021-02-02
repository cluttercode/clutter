package scanner

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"go.uber.org/zap"
	"golang.org/x/tools/godoc/util"
)

func makeBuiltin(cfg ToolConfig) (toolFunc, error) {
	re, err := cfg.Bracket.Regexp()
	if err != nil {
		return nil, fmt.Errorf("invalid bracket: %w", err)
	}

	buf := make([]byte, 128)

	return func(z *zap.SugaredLogger, path string, f func(*RawElement) error) error {
		fp, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("open %q: %w", path, err)
		}

		defer fp.Close()

		n, err := fp.Read(buf)
		if err != nil && err != io.EOF {
			return fmt.Errorf("read: %w", err)
		}

		if !util.IsText(buf[:n]) {
			z.Debug("not a text file, ignoring")
			return nil
		}

		if ofs, err := fp.Seek(0, 0); err != nil || ofs != 0 {
			return fmt.Errorf("seek: ofs=%d, err=%w", ofs, err)
		}

		scanner := bufio.NewScanner(fp)

		for i := 0; scanner.Scan(); i++ {
			line := scanner.Text()

			ms := re.FindAllStringIndex(line, -1)

			for _, m := range ms {
				l, r := m[0], m[1]

				text := line[l:r]
				text = strings.TrimPrefix(text, cfg.Bracket.Left)
				text = strings.TrimSuffix(text, cfg.Bracket.Right)
				text = strings.TrimSpace(text)

				if err := f(&RawElement{
					Text: text,
					Loc: Loc{
						Path:        path,
						Line:        i + 1,
						StartColumn: l + 1,
						EndColumn:   r,
					},
				}); err != nil {
					return fmt.Errorf("%s:%d.%d: %w", path, i+1, l+1, err)
				}
			}
		}

		if err := scanner.Err(); err == bufio.ErrTooLong {
			z.Warn("file has tokens that are too long")
			return nil
		} else if err != nil {
			return fmt.Errorf("scan: %w", err)
		}

		return nil
	}, nil
}
