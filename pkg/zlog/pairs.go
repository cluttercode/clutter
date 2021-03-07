package zlog

import (
	"fmt"
	"strings"
)

type Pair struct {
	K string
	V interface{}
}

func (p Pair) String() string { return fmt.Sprintf("%q: %q", p.K, PairValueToString(p.V)) }

type Pairs []Pair

func (ps Pairs) Strings() []string {
	strs := make([]string, len(ps))
	for i, p := range ps {
		strs[i] = p.String()
	}
	return strs
}

func (ps Pairs) String() string { return strings.Join(ps.Strings(), ", ") }

// TODO: might want to use JSON? Allow custom formatting?
func PairValueToString(v interface{}) string { return fmt.Sprintf("%+v", v) }

func toPairs(kvs []interface{}) (Pairs, error) {
	pairs := make([]Pair, 0, len(kvs)/2)

	for len(kvs) >= 2 {
		k, v := kvs[0], kvs[1]

		ks, ok := k.(string)
		if !ok {
			return nil, fmt.Errorf("key must be a string")
		}

		pairs = append(pairs, Pair{K: ks, V: v})

		kvs = kvs[2:]
	}

	if len(kvs) > 0 {
		return nil, fmt.Errorf("odd number of args")
	}

	return Pairs(pairs), nil
}
