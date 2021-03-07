package scanner

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"unicode/utf8"

	"github.com/cluttercode/clutter/pkg/zlog"
)

func isText(s []byte) bool {
	const max = 1024 // at least utf8.UTFMax
	if len(s) > max {
		s = s[0:max]
	}
	for i, c := range string(s) {
		if i+utf8.UTFMax > len(s) {
			// last char may be incomplete - ignore
			break
		}
		if c == 0xFFFD || c < ' ' && c != '\n' && c != '\t' && c != '\f' {
			// decoding error or control character - not a text file
			return false
		}
	}
	return true
}

func ScanFile(
	z *zlog.Logger,
	cfg BracketConfig,
	path string,
	f func(*RawElement) error,
) error {
	var r io.Reader = os.Stdin

	if !(path == "" || path == "-" || path == "stdin") {
		fp, err := os.Open(path)
		if err != nil {
			return err // do not wrap
		}

		defer fp.Close()

		r = fp
	}

	buf := make([]byte, 128)

	n, err := r.Read(buf)
	if err != nil && err != io.EOF {
		return fmt.Errorf("read: %w", err)
	}

	if !isText(buf[:n]) {
		z.Debug("not a text file, ignoring")
		return nil
	}

	r = io.MultiReader(bytes.NewReader(buf[:n]), r)

	return ScanRawReader(
		z,
		cfg,
		r,
		func(e *RawElement) error {
			e.Loc.Path = path // [# .fill-path #]
			return f(e)
		},
	)
}
