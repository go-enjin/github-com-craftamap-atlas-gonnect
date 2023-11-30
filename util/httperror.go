package util

import (
	"net/http"

	"github.com/go-enjin/github-com-craftamap-atlas-gonnect"

	"github.com/go-enjin/be/pkg/log"
)

func SendError(w http.ResponseWriter, r *http.Request, addon *gonnect.Addon, errorCode int, message string) {
	w.WriteHeader(errorCode)
	_, _ = w.Write([]byte(message))
	log.ErrorRDF(r, 1, "%s", message)
}