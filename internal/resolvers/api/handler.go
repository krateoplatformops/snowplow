package api

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"

	xcontext "github.com/krateoplatformops/snowplow/plumbing/context"
	"github.com/krateoplatformops/snowplow/plumbing/jqutil"
	"github.com/krateoplatformops/snowplow/plumbing/ptr"
)

type jsonHandlerOptions struct {
	key    string
	out    map[string]any
	filter *string
}

func jsonHandler(ctx context.Context, opts jsonHandlerOptions) func(io.ReadCloser) error {
	return func(in io.ReadCloser) error {
		log := xcontext.Logger(ctx)

		dat, err := io.ReadAll(in)
		if err != nil {
			return err
		}

		var tmp any
		if err := json.Unmarshal(dat, &tmp); err != nil {
			return err
		}

		if opts.filter != nil {
			q := ptr.Deref(opts.filter, "")
			log.Debug("found local filter on api results", slog.String("filter", q))
			s, err := jqutil.Eval(context.TODO(), jqutil.EvalOptions{
				Query: q, Data: tmp,
			})
			if err != nil {
				log.Error("evaluating JQ", slog.Any("error", err))
			} else {
				if err := json.Unmarshal([]byte(s), &tmp); err != nil {
					return err
				}
			}
		}

		got, ok := opts.out[opts.key]
		if !ok {
			opts.out[opts.key] = tmp
			return nil
		}

		switch existingSlice := got.(type) {
		case []any:
			if v := wrapAsSlice(tmp); len(v) > 0 {
				opts.out[opts.key] = append(existingSlice, v...)
			}
		default:
			opts.out[opts.key] = []any{got, tmp}

			switch v := tmp.(type) {
			case []any:
				all := []any{got}
				all = append(all, v...)
				opts.out[opts.key] = all
			default:
				opts.out[opts.key] = []any{got, v}
			}
		}

		return nil
	}
}

func wrapAsSlice(value any) []any {
	switch v := value.(type) {
	case []any:
		return v
	default:
		return []any{v}
	}
}
