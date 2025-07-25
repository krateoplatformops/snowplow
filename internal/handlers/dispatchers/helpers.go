package dispatchers

import (
	"log/slog"
	"net/http"
	"strconv"

	xcontext "github.com/krateoplatformops/plumbing/context"
	"github.com/krateoplatformops/plumbing/http/response"
	templatesv1 "github.com/krateoplatformops/snowplow/apis/templates/v1"
	"github.com/krateoplatformops/snowplow/internal/handlers/util"
	"github.com/krateoplatformops/snowplow/internal/objects"
)

func fetchObject(req *http.Request) (got objects.Result) {
	log := xcontext.Logger(req.Context())

	gvr, err := util.ParseGVR(req)
	if err != nil {
		got.Err = response.New(http.StatusBadRequest, err)
		return
	}
	log.Debug("GVR from request query parameters", slog.Any("gvr", gvr))

	nsn, err := util.ParseNamespacedName(req)
	if err != nil {
		got.Err = response.New(http.StatusBadRequest, err)
		return
	}
	log.Debug("Name and Namespace from request query parameters", slog.Any("nsn", nsn))

	return objects.Get(req.Context(), templatesv1.ObjectReference{
		Reference: templatesv1.Reference{
			Name: nsn.Name, Namespace: nsn.Namespace,
		},
		APIVersion: gvr.GroupVersion().String(),
		Resource:   gvr.Resource,
	})
}

func paginationInfo(log *slog.Logger, req *http.Request) (perPage, page int) {
	perPage, page = -1, -1

	if val := req.URL.Query().Get("per_page"); val != "" {
		var err error
		perPage, err = strconv.Atoi(val)
		if err != nil {
			log.Error("unable convert per_page parameter to int",
				slog.Any("err", err))
		}
	}

	if val := req.URL.Query().Get("page"); val != "" {
		var err error
		page, err = strconv.Atoi(val)
		if err != nil {
			log.Error("unable convert page parameter to int",
				slog.Any("err", err))
		}
	}

	if perPage > 0 && page <= 0 {
		page = 1
	}

	return
}
