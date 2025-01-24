package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"

	xcontext "github.com/krateoplatformops/snowplow/plumbing/context"
	"github.com/krateoplatformops/snowplow/plumbing/ptr"
)

func jsonResponseHandlerSmart(ctx context.Context, key string, out map[string]any, filter *string) func(io.ReadCloser) error {
	log := xcontext.Logger(ctx)

	return func(in io.ReadCloser) error {
		dat, err := io.ReadAll(in)
		if err != nil {
			return err
		}

		var tmp any
		if err := json.Unmarshal(dat, &tmp); err != nil {
			return err
		}
		log.Debug("api response", slog.Any("json", tmp))

		if filter != nil {
			tpl := xcontext.JQ(ctx)
			if tpl != nil {
				q := fmt.Sprintf("${ %s }", ptr.Deref(filter, ""))
				s, err := tpl.Execute(q, tmp)
				if err != nil {
					return err
				}

				if err := json.Unmarshal([]byte(s), &tmp); err != nil {
					return err
				}
			}
		}

		got, ok := out[key]
		if !ok {
			out[key] = tmp // wrapAsSlice(tmp)
			return nil
		}

		switch existingSlice := got.(type) {
		case []any:
			if v := wrapAsSlice(tmp); len(v) > 0 {
				out[key] = append(existingSlice, v...)
			}
		default:
			out[key] = []any{got, tmp}

			switch v := tmp.(type) {
			case []any:
				all := []any{got}
				all = append(all, v...)
				out[key] = all
			default:
				out[key] = []any{got, v}
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

func flattenNestedSlices(nested [][]any) []any {
	var flat []any
	for _, inner := range nested {
		flat = append(flat, inner...)
	}
	return flat
}

func jsonResponseHandler(ctx context.Context, key string, out map[string]any, filter *string) func(io.ReadCloser) error {
	return func(in io.ReadCloser) error {
		dat, err := io.ReadAll(in)
		if err != nil {
			return err
		}

		var tmp any
		if err := json.Unmarshal(dat, &tmp); err != nil {
			return err
		}

		if filter != nil {
			tpl := xcontext.JQ(ctx)
			if tpl != nil {
				q := fmt.Sprintf("${ %s }", ptr.Deref(filter, ""))
				s, err := tpl.Execute(q, tmp)
				if err != nil {
					return err
				}
				fmt.Println("==> F", s)
				if err := json.Unmarshal([]byte(s), &tmp); err != nil {
					return err
				}
			}
		}

		out[key] = tmp

		//t := reflect.TypeOf(got)
		//fmt.Println(" ==>> type: ", t)
		//spew.Dump(got)
		//fmt.Printf("\n\n\n")
		return nil
	}
}

func jsonResponseHandler2(ctx context.Context, key string, out map[string]any, filter *string) func(io.ReadCloser) error {
	return func(in io.ReadCloser) error {
		dat, err := io.ReadAll(in)
		if err != nil {
			return err
		}

		x := bytes.TrimSpace(dat)
		isArray := len(x) > 0 && x[0] == '['
		if !isArray {
			var tmp any
			if err := json.Unmarshal(dat, &tmp); err != nil {
				return err
			}

			if filter != nil {
				tpl := xcontext.JQ(ctx)
				if tpl != nil {
					q := fmt.Sprintf("${ %s }", ptr.Deref(filter, ""))
					s, err := tpl.Execute(q, tmp)
					if err != nil {
						return err
					}
					fmt.Println("==> F", s)
					if err := json.Unmarshal([]byte(s), &tmp); err != nil {
						return err
					}
				}
			}

			out[key] = tmp
			return nil
		}

		v := []any{}
		if err = json.Unmarshal(dat, &v); err != nil {
			return err
		}

		if filter != nil {
			tpl := xcontext.JQ(ctx)
			if tpl != nil {
				q := fmt.Sprintf("${ %s }", ptr.Deref(filter, ""))
				s, err := tpl.Execute(q, v)
				if err != nil {
					return err
				}
				fmt.Println("==> F2", s)
				if err := json.Unmarshal([]byte(s), &v); err != nil {
					return err
				}
			}
		}

		got, ok := out[key]
		if !ok {
			out[key] = map[string]any{
				"items": v,
			}
			return nil
		}

		src, ok := got.(map[string][]any)
		if ok {
			items, exists := src["items"]
			if !exists {
				src["items"] = v
				return nil
			}

			src["items"] = append(items, v...)
			return nil
		}

		mmm, ok := got.(map[string]any)
		if ok {
			items, exists := mmm["items"]
			if !exists {
				mmm["items"] = v
				return nil
			}

			if aaa, ok := items.([]any); ok {
				mmm["items"] = append(aaa, v)
				return nil
			}
		}

		//t := reflect.TypeOf(got)
		//fmt.Println(" ==>> type: ", t)
		//spew.Dump(got)
		//fmt.Printf("\n\n\n")
		return nil
	}
}

func filterMapByKey(input map[string]any, keepKey string) map[string]any {
	tmp := make(map[string]any)

	if value, exists := input[keepKey]; exists {
		tmp[keepKey] = value
	}

	return tmp
}
