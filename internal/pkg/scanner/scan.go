package scanner

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"golang.org/x/tools/godoc/util"

	"github.com/spf13/afero"
	"go.uber.org/zap"
)

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
