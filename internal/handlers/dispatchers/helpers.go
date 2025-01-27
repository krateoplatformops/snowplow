package dispatchers

import (
	"net/http"

	"github.com/krateoplatformops/snowplow/internal/handlers/util"
	"github.com/krateoplatformops/snowplow/internal/objects"
	"github.com/krateoplatformops/snowplow/plumbing/http/response"
)

func fetchObject(req *http.Request) (got objects.Result) {
	gvr, err := util.ParseGVR(req)
	if err != nil {
		got.Err = response.New(http.StatusBadRequest, err)
		return
	}

	nsn, err := util.ParseNamespacedName(req)
	if err != nil {
		got.Err = response.New(http.StatusBadRequest, err)
		return
	}

	return objects.Get(req.Context(), objects.Reference{
		Name: nsn.Name, Namespace: nsn.Namespace,
		APIVersion: gvr.GroupVersion().String(),
		Resource:   gvr.Resource,
	})
}
