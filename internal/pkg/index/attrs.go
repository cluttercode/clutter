package index

import (
	"fmt"
)

type Attrs map[string]string

func AttrToString(k, v string) string {
	if v == "" {
		return k
	}

	return fmt.Sprintf("%s=%s", k, v)
}
