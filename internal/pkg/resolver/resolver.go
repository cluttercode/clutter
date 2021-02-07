package resolver

import (
	"fmt"

	"go.uber.org/zap"

	"github.com/cluttercode/clutter/pkg/clutter/clutterindex"
)

type params struct{ next, prev, cycle, first, last bool }

func ResolveList(z *zap.SugaredLogger, what *clutterindex.Entry, index *clutterindex.Index) ([]*clutterindex.Entry, error) {
	return resolve(z, what, index, params{})
}

func ResolveNext(z *zap.SugaredLogger, what *clutterindex.Entry, index *clutterindex.Index, cycle bool) ([]*clutterindex.Entry, error) {
	return resolve(z, what, index, params{next: true, cycle: cycle})
}

func ResolvePrev(z *zap.SugaredLogger, what *clutterindex.Entry, index *clutterindex.Index, cycle bool) ([]*clutterindex.Entry, error) {
	return resolve(z, what, index, params{prev: true, cycle: cycle})
}

func resolve(z *zap.SugaredLogger, what *clutterindex.Entry, index *clutterindex.Index, p params) ([]*clutterindex.Entry, error) {
	if p.next && p.prev {
		z.Panic("prev and next are mutually exclusive")
	}

	matcher := func(ent *clutterindex.Entry) bool { return what.Name == ent.Name && ent.IsReferredBy(what) }

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
		hold *clutterindex.Entry
		ents []*clutterindex.Entry
	)

	if err := clutterindex.ForEach(
		clutterindex.SliceSource(index),
		func(ent *clutterindex.Entry) error {
			match := matcher(ent)

			z.Debugw("considering", "ent", ent, "match", match)

			if !match {
				return nil
			}

			if p.prev {
				if ent.Loc == what.Loc {
					if hold == nil {
						z.Debugw("found what, but nothing held")
						return clutterindex.ErrStop
					}

					z.Debugw("found what, emit held", "ent", hold)

					ents = append(ents, hold)

					return clutterindex.ErrStop
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

				return clutterindex.ErrStop
			}

			ents = append(ents, ent)

			if p.first {
				return clutterindex.ErrStop
			}

			return nil
		},
	); err != nil {
		return nil, fmt.Errorf("filter: %w", err)
	}

	if p.cycle && len(ents) == 0 {
		if p.next {
			return resolve(z, what, index, params{first: true})
		} else if p.prev {
			return resolve(z, what, index, params{last: true})
		}
	}

	if p.last && len(ents) > 0 {
		return []*clutterindex.Entry{ents[len(ents)-1]}, nil
	}

	return ents, nil
}
