package api

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"

	xcontext "github.com/krateoplatformops/plumbing/context"
	"github.com/krateoplatformops/plumbing/jqutil"
	"github.com/krateoplatformops/plumbing/ptr"
	jqsupport "github.com/krateoplatformops/snowplow/internal/support/jq"
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

		pig := map[string]any{
			opts.key: tmp,
		}
		if si, ok := opts.out["_slice_"]; ok {
			pig["_slice_"] = si
		}

		if opts.filter != nil {
			q := ptr.Deref(opts.filter, "")
			log.Debug("found local filter on api result", slog.String("filter", q))
			s, err := jqutil.Eval(context.TODO(), jqutil.EvalOptions{
				Query: q, Data: pig,
				ModuleLoader: jqsupport.ModuleLoader(),
			})
			if err != nil {
				log.Error("unable to evaluate JQ filter",
					slog.String("filter", q), slog.Any("error", err))
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
