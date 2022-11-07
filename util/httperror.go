package util

import (
	"net/http"
	"path/filepath"
	"runtime"

	"github.com/go-enjin/github-com-craftamap-atlas-gonnect"

	"github.com/go-enjin/be/pkg/log"
)

func SendError(w http.ResponseWriter, addon *gonnect.Addon, errorCode int, message string) {
	w.WriteHeader(errorCode)
	_, _ = w.Write([]byte(message))
	pc, file, no, ok := runtime.Caller(1)
	details := runtime.FuncForPC(pc)

	if ok {
		log.ErrorF("%s:%d %s() %s", filepath.Base(file), no, details.Name(), message)
	} else {
		log.ErrorF(message)
	}
}
