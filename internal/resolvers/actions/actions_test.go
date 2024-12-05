//go:build integration
// +build integration

package actions

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/krateoplatformops/snowplow/apis/templates/v1alpha1"
	xcontext "github.com/krateoplatformops/snowplow/plumbing/context"
)

func TestResolveActions(t *testing.T) {
	log := slog.New(slog.NewJSONHandler(os.Stdout,
		&slog.HandlerOptions{Level: slog.LevelDebug}))

	ctx := xcontext.BuildContext(context.TODO(),
		xcontext.WithTraceId("test"),
		xcontext.WithLogger(log),
	)

	res, err := Resolve(ctx, []*v1alpha1.Action{
		{
			Template: &v1alpha1.ActionTemplate{
				ID:         "test-id",
				Name:       "nginx",
				Namespace:  "demo-system",
				Resource:   "deployments",
				APIVersion: "apps/v1",
				Verb:       "put",
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	spew.Dump(res)
}
