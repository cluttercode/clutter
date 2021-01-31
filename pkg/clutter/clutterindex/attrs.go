package clutterindex

import (
	"fmt"
)

type Attrs map[string]string

// used for [# govaluate-params #].
type AttrsStruct struct{ as Attrs }

func (a Attrs) ToStruct() *AttrsStruct {
	return &AttrsStruct{as: a}
}

func AttrToString(k, v string) string {
	if v == "" {
		return k
	}

	return fmt.Sprintf("%s=%s", k, v)
}

func (a AttrsStruct) Has(k string) bool {
	_, ok := a.as[k]
	return ok
}
