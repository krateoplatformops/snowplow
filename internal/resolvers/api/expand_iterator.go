package api

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	templates "github.com/krateoplatformops/snowplow/apis/templates/v1"
	xcontext "github.com/krateoplatformops/snowplow/plumbing/context"
	"github.com/krateoplatformops/snowplow/plumbing/ptr"
)

func expandIterator(ctx context.Context, in *templates.API, dict map[string]any) (all []*templates.API) {
	tpl := xcontext.JQTemplate(ctx)
	if tpl == nil {
		log := xcontext.Logger(ctx)
		log.Error("missing jq template engine")
		return []*templates.API{}
	}

	it := ptr.Deref(in.Iterator, "")

	tot := 1
	if len(it) > 0 {
		q := fmt.Sprintf("${ %s | length }", it)

		count, err := tpl.Execute(q, dict)
		if err != nil {
			log := xcontext.Logger(ctx)
			log.Error("unable to execute jq template", slog.String("query", q)) //, slog.Any("err", err))
			count = "1"
		}

		tot, err = strconv.Atoi(count)
		if err != nil {
			log := xcontext.Logger(ctx)
			log.Error("atoi failure", slog.Any("err", err))
			tot = 1
		}
	}

	render := func(i int, s string, ds map[string]any) string {
		exp := hackQueryFn(it, i, s)
		out, err := tpl.Execute(exp, ds)
		if err != nil {
			out = err.Error()
		}
		return out
	}

	all = make([]*templates.API, 0, tot)
	for i := 0; i < tot; i++ {
		el := &templates.API{
			DependOn: ptr.To(ptr.Deref(in.DependOn, "")),
			Name:     fmt.Sprintf("%s_%d", in.Name, i),
			Path:     render(i, in.Path, dict),
		}
		if in.Filter != nil {
			el.Filter = ptr.To(ptr.Deref(in.Filter, ""))
		}

		if in.Verb != nil {
			el.Verb = ptr.To(ptr.Deref(in.Verb, http.MethodGet))
		}

		if in.Payload != nil {
			el.Payload = ptr.To(ptr.Deref(in.Payload, ""))
		}

		if in.EndpointRef != nil {
			el.EndpointRef = in.EndpointRef.DeepCopy()
		}

		if in.Headers != nil {
			el.Headers = make([]string, len(in.Headers))
			copy(el.Headers, in.Headers)
		}

		all = append(all, el)
	}

	return all
}

func hackQueryFn(it string, i int, q string) string {
	if len(it) == 0 {
		return q
	}

	el := fmt.Sprintf("%s[%d]", it, i)
	q = strings.Replace(q, "${", fmt.Sprintf("${ %s | ", el), 1)
	return q
}

func appendToNestedMap(origin map[string]any, additions map[string]any, path string) {
	keys := strings.Split(path, ".")
	lastKey := keys[len(keys)-1]
	nested := origin

	// Naviga nella struttura fino al penultimo livello
	for _, key := range keys[:len(keys)-1] {
		if next, ok := nested[key]; ok {
			if nextMap, ok := next.(map[string]any); ok {
				nested = nextMap
			} else {
				// Gestione dell'errore: il path esiste ma non è una mappa
				fmt.Printf("Error: '%s' is not a map\n", key)
				return
			}
		} else {
			// Se il livello non esiste, crealo come mappa
			newMap := make(map[string]any)
			nested[key] = newMap
			nested = newMap
		}
	}

	// Ora siamo al livello giusto per fare append sulla chiave finale
	if existing, exists := nested[lastKey]; exists {
		if existingSlice, ok := existing.([]any); ok {
			if newSlice, ok := additions[lastKey].([]any); ok {
				nested[lastKey] = append(existingSlice, newSlice...)
			} else {
				nested[lastKey] = append(existingSlice, additions[lastKey])
			}
		} else {
			nested[lastKey] = []any{existing, additions[lastKey]}
		}
	} else {
		if slice, ok := additions[lastKey].([]any); ok {
			nested[lastKey] = slice
		} else {
			nested[lastKey] = []any{additions[lastKey]}
		}
	}
}
