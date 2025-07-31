package widgets

import (
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/krateoplatformops/plumbing/maps"
	templatesv1 "github.com/krateoplatformops/snowplow/apis/templates/v1"
)

const (
	widgetDataKey            = "widgetData"
	widgetDataTemplateKey    = "widgetDataTemplate"
	apiRefKey                = "apiRef"
	resourcesRefsKey         = "resourcesRefs"
	resourcesRefsTemplateKey = "resourcesRefsTemplate"
)

func GetAPIVersion(obj map[string]any) string {
	val, err := maps.NestedString(obj, "apiVersion")
	if err != nil {
		return ""
	}
	return val
}

func GetKind(obj map[string]any) string {
	val, err := maps.NestedString(obj, "kind")
	if err != nil {
		return ""
	}
	return val
}

func GetNamespace(obj map[string]any) string {
	val, err := maps.NestedString(obj, "metadata", "namespace")
	if err != nil {
		return ""
	}
	return val
}

func GetName(obj map[string]any) string {
	val, err := maps.NestedString(obj, "metadata", "name")
	if err != nil {
		return ""
	}
	return val
}

func GetUID(obj map[string]any) string {
	val, err := maps.NestedString(obj, "metadata", "uid")
	if err != nil {
		return ""
	}
	return val
}

func GetWidgetData(obj map[string]any) map[string]any {
	data, ok, err := maps.NestedMap(obj, "spec", widgetDataKey)
	if !ok || err != nil {
		return map[string]any{}
	}
	return data
}

func GetWidgetDataTemplate(obj map[string]any) ([]templatesv1.WidgetDataTemplate, error) {
	data, ok, err := maps.NestedSliceNoCopy(obj, "spec", widgetDataTemplateKey)
	if !ok || err != nil {
		return nil, err
	}

	items, err := maps.ToMapSlice(data)
	if err != nil {
		return nil, err
	}

	return maps.MapSliceToStructSlice[templatesv1.WidgetDataTemplate](items)
}

func GetApiRef(obj map[string]any) (templatesv1.ObjectReference, error) {
	src, ok, err := maps.NestedMapNoCopy(obj, "spec", apiRefKey)
	if !ok || err != nil {
		return templatesv1.ObjectReference{}, err
	}

	dat, err := json.Marshal(src)
	if err != nil {
		return templatesv1.ObjectReference{}, err
	}

	ref := templatesv1.ObjectReference{
		Resource:   "restactions",
		APIVersion: fmt.Sprintf("%s/%s", templatesv1.Group, templatesv1.Version),
	}
	err = json.Unmarshal(dat, &ref)

	return ref, err
}

func GetResourcesRefs(obj map[string]any) ([]templatesv1.ResourceRef, error) {
	arr, ok, err := maps.NestedSlice(obj, "spec", resourcesRefsKey, "items")
	if !ok || err != nil {
		return []templatesv1.ResourceRef{}, err
	}

	mapSlice, err := maps.ToMapSlice(arr)
	if err != nil {
		return []templatesv1.ResourceRef{}, err
	}

	return maps.MapSliceToStructSlice[templatesv1.ResourceRef](mapSlice)
}

func GetResourcesRefsTemplate(obj map[string]any) ([]templatesv1.ResourceRefTemplate, error) {
	arr, ok, err := maps.NestedSlice(obj, "spec", resourcesRefsTemplateKey)
	if !ok || err != nil {
		return []templatesv1.ResourceRefTemplate{}, err
	}

	mapSlice, err := maps.ToMapSlice(arr)
	if err != nil {
		return []templatesv1.ResourceRefTemplate{}, err
	}

	return maps.MapSliceToStructSlice[templatesv1.ResourceRefTemplate](mapSlice)
}

func loggerAttr(obj map[string]any) slog.Attr {
	return slog.Group("widget",
		slog.String("name", GetName(obj)),
		slog.String("namespace", GetNamespace(obj)),
		slog.String("apiVersion", GetAPIVersion(obj)),
		slog.String("kind", GetKind(obj)),
	)
}
