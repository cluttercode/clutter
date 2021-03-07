package scanner

import (
	"fmt"
	"io"

	"github.com/cluttercode/clutter/pkg/zlog"
)

func ScanSitterReader(
	z *zlog.Logger,
	cfg BracketConfig,
	r io.Reader,
	f func(*RawElement) error, // will not include path. path is filled in [# ./fill-path #].
) error {
	_, err := cfg.Regexp()
	if err != nil {
		return fmt.Errorf("invalid bracket: %w", err)
	}

	return nil
}
