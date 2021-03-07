package resolver

import (
	"fmt"

	"github.com/cluttercode/clutter/pkg/zlog"

	"github.com/cluttercode/clutter/internal/pkg/index"
)

type params struct{ next, prev, cycle, first, last bool }

func ResolveList(z *zlog.Logger, what *index.Entry, idx *index.Index) ([]*index.Entry, error) {
	return resolve(z, what, idx, params{})
}

func ResolveNext(z *zlog.Logger, what *index.Entry, idx *index.Index, cycle bool) ([]*index.Entry, error) {
	return resolve(z, what, idx, params{next: true, cycle: cycle})
}

func ResolvePrev(z *zlog.Logger, what *index.Entry, idx *index.Index, cycle bool) ([]*index.Entry, error) {
	return resolve(z, what, idx, params{prev: true, cycle: cycle})
}

func resolve(z *zlog.Logger, what *index.Entry, idx *index.Index, p params) ([]*index.Entry, error) {
	if p.next && p.prev {
		z.Panic("prev and next are mutually exclusive")
	}

	matcher := func(ent *index.Entry) bool { return what.Name == ent.Name && ent.IsReferredBy(what) }

	if _, search := what.IsSearch(); search {
		if p.prev || p.next {
			p.prev, p.next = false, false
			z.Warn("--next and --prev are ignored when resolving a search tag")
		}

		var err error
		matcher, err = what.Matcher()
		if err != nil {
			return nil, fmt.Errorf("invalid search tag")
		}
	}

	var (
		hold *index.Entry
		ents []*index.Entry
	)

	if err := index.ForEach(
		idx,
		func(ent *index.Entry) error {
			match := matcher(ent)

			z.Debugw("considering", "ent", ent, "match", match)

			if !match {
				return nil
			}

			if p.prev {
				if ent.Loc == what.Loc {
					if hold == nil {
						z.Debugw("found what, but nothing held")
						return index.ErrStop
					}

					z.Debugw("found what, emit held", "ent", hold)

					ents = append(ents, hold)

					return index.ErrStop
				}

				hold = ent
				z.Debugw("holding", "ent", hold)

				return nil
			}

			if p.next {
				if hold == nil {
					if ent.Loc == what.Loc {
						hold = ent
						z.Debugw("found what", "ent", hold)
					}

					return nil
				}

				z.Debugw("emit current", "ent", hold)

				ents = append(ents, ent)

				return index.ErrStop
			}

			ents = append(ents, ent)

			if p.first {
				return index.ErrStop
			}

			return nil
		},
	); err != nil {
		return nil, fmt.Errorf("filter: %w", err)
	}

	if p.cycle && len(ents) == 0 {
		if p.next {
			return resolve(z, what, idx, params{first: true})
		} else if p.prev {
			return resolve(z, what, idx, params{last: true})
		}
	}

	if p.last && len(ents) > 0 {
		return []*index.Entry{ents[len(ents)-1]}, nil
	}

	return ents, nil
}
