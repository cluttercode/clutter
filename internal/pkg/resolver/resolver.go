package resolver

import (
	"fmt"

	"go.uber.org/zap"

	"github.com/cluttercode/clutter/pkg/clutter/clutterindex"
)

func ResolveList(z *zap.SugaredLogger, what *clutterindex.Entry, index *clutterindex.Index) ([]*clutterindex.Entry, error) {
	return resolve(z, what, index, false, false)
}

func ResolveNext(z *zap.SugaredLogger, what *clutterindex.Entry, index *clutterindex.Index) ([]*clutterindex.Entry, error) {
	return resolve(z, what, index, true, false)
}

func ResolvePrev(z *zap.SugaredLogger, what *clutterindex.Entry, index *clutterindex.Index) ([]*clutterindex.Entry, error) {
	return resolve(z, what, index, false, true)
}

func resolve(z *zap.SugaredLogger, what *clutterindex.Entry, index *clutterindex.Index, next, prev bool) ([]*clutterindex.Entry, error) {
	if next && prev {
		z.Panic("prev and next are mutually exclusive")
	}

	matcher := func(ent *clutterindex.Entry) bool { return what.Name == ent.Name && ent.IsReferredBy(what) }

	if _, search := what.IsSearch(); search {
		if prev || next {
			prev, next = false, false
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

			if prev {
				if ent.Loc == what.Loc {
					if hold == nil {
						z.Debugw("found what, but nothing held")
						return clutterindex.ErrStop
					}

					z.Debugw("found what, emit held", "ent", hold)

					fmt.Println(hold.String())

					return clutterindex.ErrStop
				}

				hold = ent
				z.Debugw("holding", "ent", hold)

				return nil
			}

			if next {
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

			return nil
		},
	); err != nil {
		return nil, fmt.Errorf("filter: %w", err)
	}

	return ents, nil
}
