package routes

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/go-enjin/github-com-craftamap-atlas-gonnect"
	"github.com/go-enjin/github-com-craftamap-atlas-gonnect/middleware"
	"github.com/go-enjin/github-com-craftamap-atlas-gonnect/store"
	"github.com/go-enjin/github-com-craftamap-atlas-gonnect/util"
	"github.com/go-enjin/be/pkg/log"
)

type AtlassianConnectHandler struct {
	Addon *gonnect.Addon
}

func (h AtlassianConnectHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(h.Addon.AddonDescriptor)
}

func NewAtlassianConnectHandler(addon *gonnect.Addon) http.Handler {
	return AtlassianConnectHandler{addon}
}

type InstalledHandler struct {
	Addon *gonnect.Addon
}

func (h InstalledHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	tenant, err := store.NewTenantFromReader(r.Body)
	if err != nil {
		util.SendError(w, h.Addon, 500, err.Error())
		return
	}
	_, err = h.Addon.Store.Set(tenant)
	if err != nil {
		util.SendError(w, h.Addon, 500, err.Error())
		return
	}
	log.InfoF("installed new tenant %s\n", tenant.BaseURL)
	_, _ = w.Write([]byte("OK"))
}

func NewInstalledHandler(addon *gonnect.Addon) http.Handler {
	return InstalledHandler{addon}
}

type UninstalledHandler struct {
	Addon *gonnect.Addon
}

func (h UninstalledHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	tenant, err := store.NewTenantFromReader(r.Body)
	if err != nil {
		util.SendError(w, h.Addon, 500, err.Error())
		return
	}
	_, err = h.Addon.Store.Set(tenant)
	if err != nil {
		util.SendError(w, h.Addon, 500, err.Error())
		return
	}
	log.InfoF("uninstalled tenant %s\n", tenant.BaseURL)
	_, _ = w.Write([]byte("OK"))
}

func NewUninstalledHandler(addon *gonnect.Addon) http.Handler {
	return UninstalledHandler{addon}
}

var RegisteredRoutes []string

func RegisterRoutes(base string, addon *gonnect.Addon, mux chi.Router) {
	base = strings.Trim(base, " \t/")
	if base == "" {
		base = "/"
	} else {
		base = "/" + base + "/"
	}
	RegisteredRoutes = append(RegisteredRoutes, base+"atlassian-connect.json", base+"installed", base+"uninstalled")
	mux.Handle(base+"atlassian-connect.json", NewAtlassianConnectHandler(addon))
	mux.Handle(base+"installed", middleware.NewVerifyInstallationMiddleware(addon)(NewInstalledHandler(addon)))
	mux.Handle(base+"uninstalled", middleware.NewAuthenticationMiddleware(addon, false)(NewUninstalledHandler(addon)))
}
