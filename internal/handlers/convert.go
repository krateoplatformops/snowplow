package handlers

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math"
	"net/http"
	"strings"

	xcontext "github.com/krateoplatformops/snowplow/plumbing/context"
	"github.com/krateoplatformops/snowplow/plumbing/http/response"
	"sigs.k8s.io/yaml"
)

func Converter() http.Handler {
	return &convertHandler{}
}

const MaxBodySize = 1 * 1024 * 1024 // 1MB

var _ http.Handler = (*convertHandler)(nil)

type convertHandler struct{}

// @Summary Convert YAML to JSON or JSON to YAML
// @Description This endpoint converts YAML to JSON or JSON to YAML based on the "Content-Type" header.
// @ID convert
// @Accept application/json, application/x-yaml, text/yaml
// @Produce application/json, application/x-yaml
// @Param data body string true "Input data in YAML or JSON format"
// @Success 200 {string} string "Converted output in the requested format"
// @Failure 400 {object} response.Status "Bad request, invalid input"
// @Failure 406 {object} response.Status "Unsupported 'Accept' header"
// @Failure 500 {object} response.Status "Internal server error"
// @Router /convert [post]
func (r *convertHandler) ServeHTTP(wri http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		response.MethodNotAllowed(wri, errors.New("only POST method is allowed"))
		return
	}

	body, err := io.ReadAll(io.LimitReader(req.Body, MaxBodySize))
	if err != nil {
		response.InternalError(wri, err)
		return
	}
	defer req.Body.Close()

	contentType := req.Header.Get("Content-Type")
	toJSON := (strings.Contains(contentType, "application/x-yaml") || strings.Contains(contentType, "text/yaml"))
	toYAML := strings.Contains(contentType, "application/json")

	log := xcontext.Logger(req.Context())

	if toJSON {
		limit := int(math.Min(128, float64(len(body))))
		log.Debug("converting data to JSON", slog.String("contentType", contentType),
			slog.String("data", fmt.Sprintf("%s", body[:limit])))

		dat, err := yaml.YAMLToJSON(body)
		if err != nil {
			response.BadRequest(wri, fmt.Errorf("failed to convert YAML to JSON: %w", err))
			return
		}

		wri.Header().Set("Content-Type", "application/json")
		wri.WriteHeader(http.StatusOK)
		wri.Write(dat)

		return
	}

	if toYAML {
		limit := int(math.Min(128, float64(len(body))))
		log.Debug("converting data to YAML", slog.String("contentType", contentType),
			slog.String("data", fmt.Sprintf("%s", body[:limit])))

		dat, err := yaml.JSONToYAML(body)
		if err != nil {
			response.BadRequest(wri, fmt.Errorf("failed to convert JSON to YAML: %w", err))
			return
		}

		wri.Header().Set("Content-Type", "application/x-yaml")
		wri.WriteHeader(http.StatusOK)
		wri.Write(dat)
		return
	}

	response.NotAcceptable(wri,
		fmt.Errorf("unsupported content type '%s' use 'application/json' or 'application/x-yaml'", contentType))

	return
}
