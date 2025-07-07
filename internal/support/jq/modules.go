package jq

import (
	"os"
	"strings"
	"sync"

	"github.com/itchyny/gojq"
	"github.com/krateoplatformops/plumbing/jqutil"
)

const (
	EnvModulesPath = "JQ_MODULES_PATH"
)

var (
	once         sync.Once
	cachedLoader gojq.ModuleLoader
)

func ModuleLoader() gojq.ModuleLoader {
	once.Do(func() {
		basePath, ok := os.LookupEnv(EnvModulesPath)
		if !ok {
			return
		}

		basePath = strings.TrimSpace(basePath)
		if basePath == "" {
			return
		}

		cachedLoader = jqutil.DirModuleLoader(basePath)
	})
	return cachedLoader
}
