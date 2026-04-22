package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	xcontext "github.com/krateoplatformops/plumbing/context"
	"github.com/krateoplatformops/plumbing/http/response"
	"github.com/krateoplatformops/plumbing/jqutil"
	jqsupport "github.com/krateoplatformops/snowplow/internal/support/jq"
	"github.com/krateoplatformops/snowplow/internal/telemetry"
)

const (
	MaxBodySize = 1 * 1024 * 1024 // 1MB
)

// @Summary     Evaluate a JQ query against JSON input
// @Description This endpoint accepts a JSON body containing a JQ `query` and some `data`.
// @Description It evaluates the query against the data and returns the result as formatted JSON.
// @Tags        jq
// @Accept      json
// @Produce     json
// @Param       body  body   jqin  true  "Input payload containing JQ query and JSON data"
// @Success     200   {object}  any    "Successfully evaluated JQ query"
// @Failure 400 {object} response.Status
// @Failure 401 {object} response.Status
// @Failure 404 {object} response.Status
// @Failure 500 {object} response.Status
// @Router      /jq [post]
func JQ(metrics *telemetry.Metrics) http.HandlerFunc {
	return func(wri http.ResponseWriter, req *http.Request) {
		start := time.Now()

		if req.Method != http.MethodPost {
			metrics.IncJQError(req.Context(), "method_not_allowed", http.StatusMethodNotAllowed)
			metrics.RecordJQRequest(req.Context(), http.StatusMethodNotAllowed, time.Since(start))
			response.MethodNotAllowed(wri, errors.New("only POST method is allowed"))
			return
		}

		log := xcontext.Logger(req.Context())

		_, err := xcontext.UserConfig(req.Context())
		if err != nil {
			log.Error("unable to get user endpoint", slog.Any("err", err))
			metrics.IncJQError(req.Context(), "user_config", http.StatusUnauthorized)
			metrics.RecordJQRequest(req.Context(), http.StatusUnauthorized, time.Since(start))
			response.Unauthorized(wri, err)
			return
		}

		contentType := req.Header.Get("Content-Type")
		if !strings.Contains(contentType, "application/json") {
			err := fmt.Errorf("unsupported content type '%s' use 'application/json'", contentType)
			log.Error(err.Error())
			metrics.IncJQError(req.Context(), "content_type", http.StatusNotAcceptable)
			metrics.RecordJQRequest(req.Context(), http.StatusNotAcceptable, time.Since(start))
			response.NotAcceptable(wri, err)
			return
		}

		in := &jqin{}
		decodeStart := time.Now()
		dec := json.NewDecoder(io.LimitReader(req.Body, MaxBodySize))
		err = dec.Decode(in)
		metrics.RecordJQDecodeDuration(req.Context(), time.Since(decodeStart))
		if err != nil {
			log.Error("unable to decode JSON body", slog.Any("err", err))
			metrics.IncJQError(req.Context(), "decode", http.StatusInternalServerError)
			metrics.RecordJQRequest(req.Context(), http.StatusInternalServerError, time.Since(start))
			response.InternalError(wri, err)
			return
		}

		evalStart := time.Now()
		res, err := jqutil.Eval(req.Context(), jqutil.EvalOptions{
			Query:        in.Query,
			Data:         in.Data,
			ModuleLoader: jqsupport.ModuleLoader(),
		})
		metrics.RecordJQEvalDuration(req.Context(), time.Since(evalStart))
		if err != nil {
			log.Error("unable to evaluate JQ query", slog.Any("err", err))
			metrics.IncJQError(req.Context(), "eval", http.StatusInternalServerError)
			metrics.RecordJQRequest(req.Context(), http.StatusInternalServerError, time.Since(start))
			response.InternalError(wri, err)
			return
		}

		val := jqutil.InferType(res)

		wri.Header().Set("Content-Type", "application/json")
		wri.WriteHeader(http.StatusOK)
		enc := json.NewEncoder(wri)
		enc.SetIndent("", "  ")
		if err := enc.Encode(&val); err != nil {
			metrics.IncJQError(req.Context(), "encode_response", http.StatusInternalServerError)
			return
		}

		metrics.RecordJQRequest(req.Context(), http.StatusOK, time.Since(start))
	}
}

type jqin struct {
	Query string `json:"query"`
	Data  any    `json:"data"`
}
