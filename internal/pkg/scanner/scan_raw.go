package scanner

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/cluttercode/clutter/pkg/zlog"
)

func ScanRawReader(
	z *zlog.Logger,
	cfg BracketConfig,
	r io.Reader,
	f func(*RawElement) error, // will not include path. path is filled in [# ./fill-path #].
) error {
	re, err := cfg.Regexp()
	if err != nil {
		return fmt.Errorf("invalid bracket: %w", err)
	}

	scanner := bufio.NewScanner(r)

	stopped := false

S:
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
				case "stop!":
					// hard stop will stop scanning the rest of the file.
					break S

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
