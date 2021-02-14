package scanner

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/afero"
	"go.uber.org/zap"
	"golang.org/x/tools/godoc/util"
)

func ScanReader(
	z *zap.SugaredLogger,
	cfg BracketConfig,
	r io.Reader,
	f func(*RawElement) error, // will not include path. path is filled in [# .fill-path #].
) error {
	re, err := cfg.Regexp()
	if err != nil {
		return fmt.Errorf("invalid bracket: %w", err)
	}

	buf := make([]byte, 128)

	n, err := r.Read(buf)
	if err != nil && err != io.EOF {
		return fmt.Errorf("read: %w", err)
	}

	if !util.IsText(buf[:n]) {
		z.Debug("not a text file, ignoring")
		return nil
	}

	r = io.MultiReader(bytes.NewReader(buf[:n]), r)

	scanner := bufio.NewScanner(r)

	stopped := false

	for i := 0; scanner.Scan(); i++ {
		line := scanner.Text()

		ms := re.FindAllStringIndex(line, -1)

		for _, m := range ms {
			l, r := m[0], m[1]

			text := line[l:r]
			text = strings.TrimPrefix(text, cfg.Left)
			text = strings.TrimSuffix(text, cfg.Right)
			text = strings.TrimSpace(text)

			if strings.HasPrefix(text, "%") {
				switch text[1:] {
				case "stop":
					if stopped {
						return fmt.Errorf("already stopped")
					}

					stopped = true
				case "cont":
					if !stopped {
						return fmt.Errorf("not stopped")
					}

					stopped = false
				default:
					return fmt.Errorf("unknown pragma: %s", text)
				}

				continue
			}

			if stopped {
				continue
			}

			if err := f(&RawElement{
				Text: text,
				Loc: Loc{
					Line:        i + 1,
					StartColumn: l + 1,
					EndColumn:   r,
				},
			}); err != nil {
				return fmt.Errorf("%d.%d: %w", i+1, l+1, err)
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
}

func ScanFile(
	fs afero.Fs,
	z *zap.SugaredLogger,
	cfg BracketConfig,
	path string,
	f func(*RawElement) error,
) error {
	if fs == nil {
		fs = afero.NewOsFs()
	}

	var r io.Reader = os.Stdin

	if !(path == "" || path == "-" || path == "stdin") {
		fp, err := fs.Open(path)
		if err != nil {
			return err // do not wrap
		}

		defer fp.Close()

		r = fp
	}

	return ScanReader(
		z,
		cfg,
		r,
		func(e *RawElement) error {
			e.Loc.Path = path // [# .fill-path #]
			return f(e)
		},
	)
}
