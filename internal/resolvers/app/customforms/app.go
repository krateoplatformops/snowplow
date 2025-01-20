package customforms

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	templates "github.com/krateoplatformops/snowplow/apis/templates/v1"
	xcontext "github.com/krateoplatformops/snowplow/plumbing/context"
	"github.com/krateoplatformops/snowplow/plumbing/ptr"
	"github.com/krateoplatformops/snowplow/plumbing/tmpl"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func Resolve(ctx context.Context, app *templates.CustomFormAppTemplate, dict map[string]any) (uns *unstructured.Unstructured, err error) {
	uns = &unstructured.Unstructured{Object: map[string]any{}}

	if app == nil || len(dict) == 0 {
		return
	}

	log := xcontext.Logger(ctx)

	tpl := xcontext.JQTemplate(ctx)
	if tpl == nil {
		log.Error("unable to resolve customform app template",
			slog.String("err", "missing jq template engine"))
		return
	}

	query, _ := tpl.ParseQuery(app.Schema)
	schema, err := jqHandler(tpl, query, dict)
	if err != nil {
		return uns, err
	}
	log.Debug("successful handled 'app.template.scheme'", slog.Int("results", len(schema)))

	uns.Object = schema
	tmp, err := resolvePropertiesToHide(tpl, app.PropertiesToHide, schema)
	if err != nil {
		return uns, err
	}
	log.Debug("successful handled 'app.template.propertiesToHide'", slog.Int("results", len(tmp)))

	uns.Object = tmp

	tmp, err = resolvePropertiesToOverride(tpl, app.PropertiesToOverride, dict, uns.Object)
	if err != nil {
		return uns, err
	}
	uns.Object = tmp

	log.Debug("successful handled 'app.template.propertiesToOverride'", slog.Int("results", len(tmp)))

	return
}

func resolvePropertiesToHide(tpl tmpl.JQTemplate, in []string, dict map[string]any) (schema map[string]any, err error) {
	tot := len(in)
	if tot == 0 {
		return dict, nil
	}

	tmp := make([]string, tot)
	for i := range tot {
		el := in[i]
		if len(el) == 0 {
			continue
		}
		if el[0] != '.' {
			el = "." + el
		}
		tmp[i] = el
	}

	schema = dict
	for _, el := range tmp {
		query := fmt.Sprintf("del(%s)", el)
		schema, err = jqHandler(tpl, query, schema)
		if err != nil {
			return schema, fmt.Errorf("jq expression failure %q: %w", query, err)
		}
	}

	return
}

func resolvePropertiesToOverride(tpl tmpl.JQTemplate, in []templates.Data, dict, schema map[string]any) (map[string]any, error) {
	tot := len(in)
	if tot == 0 {
		return schema, nil
	}

	all := []templates.Data{}
	for _, el := range in {
		exp := strings.TrimSpace(el.Value)
		if len(exp) == 0 {
			continue
		}

		key := el.Name
		if len(key) == 0 {
			continue
		}
		if key[0] != '.' {
			key = "." + key
		}

		val := el.Value
		if !ptr.Deref(el.AsString, false) {
			var err error
			val, err = tpl.Execute(exp, dict)
			if err != nil {
				return schema, fmt.Errorf("jq expression failure %q: %w", exp, err)
			}
		}

		if len(val) > 0 {
			all = append(all, templates.Data{
				Name:     key,
				Value:    val,
				AsString: ptr.To(ptr.Deref(el.AsString, false)),
			})
		}
	}

	var err error
	for _, el := range all {
		query := fmt.Sprintf("%s |= %q", el.Name, el.Value)
		if !ptr.Deref(el.AsString, false) {
			query = fmt.Sprintf("%s |= %s", el.Name, el.Value)
		}

		schema, err = jqHandler(tpl, query, schema)
		if err != nil {
			return schema, fmt.Errorf("jq expression failure %q: %w", query, err)
		}
	}

	return schema, nil
}

func jqHandler(tpl tmpl.JQTemplate, q string, ds map[string]any) (map[string]any, error) {
	res, err := tpl.Q(q, ds)
	if err != nil {
		return nil, fmt.Errorf("jq expression failure %q: %w", q, err)
	}
	if len(res) == 0 {
		return ds, nil
	}

	val, ok := res[0].(map[string]any)
	if !ok {
		return ds, fmt.Errorf("jq del result is not (map[string]any)")
	}

	return val, nil
}
