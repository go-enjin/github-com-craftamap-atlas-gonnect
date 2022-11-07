package middleware

import (
	"context"
	"net/http"
	"net/url"

	"github.com/go-enjin/be/pkg/log"

	"github.com/go-enjin/github-com-craftamap-atlas-gonnect"
	"github.com/go-enjin/github-com-craftamap-atlas-gonnect/hostrequest"
)

type RequestMiddleware struct {
	h              http.Handler
	addon          *gonnect.Addon
	verifiedParams map[string]string
}

func (h RequestMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	getParam := func(key string) string {
		value := r.URL.Query().Get(key)
		if value != "" {
			err := r.ParseForm()
			if err != nil {
				return ""
			}
			v := r.Form
			value = v.Get(key)
		}
		return value
	}

	getHostBaseUrlFromQueryParams := func() string {
		hostUrl := getParam("xdm_e")
		if hostUrl != "" {
			return hostUrl + getParam("cp")
		} else {
			return ""
		}
	}

	getHostResourceUrl := func(isDev bool, baseUrl string, ext string) *url.URL {
		// again, bb handling is missing here
		var resource string
		if isDev {
			resource = "all-debug." + ext
		} else {
			resource = "all." + ext
		}

		uri, err := url.Parse(baseUrl + "/atlassian-connect/" + resource)
		if err != nil {
			return &url.URL{}
		} else {
			return uri
		}
	}

	log.DebugF("Setting Context Variables in Request Middleware")
	ctx := context.WithValue(r.Context(), "title", *h.addon.Name)
	ctx = context.WithValue(ctx, "addonKey", *h.addon.Key)
	ctx = context.WithValue(ctx, "localBaseUrl", h.addon.Config.BaseUrl)
	ctx = context.WithValue(ctx, "license", getParam("lic"))

	// if missing here: if isJira || isConfluence
	// Since this poc is for confluence only, this should be valid, for now
	hostBaseUrl := getHostBaseUrlFromQueryParams()
	ctx = context.WithValue(ctx, "hostBaseUrl", hostBaseUrl)

	if len(h.verifiedParams) > 0 {
		ctx = context.WithValue(ctx, "userAccountId", h.verifiedParams["userAccountId"])
		ctx = context.WithValue(ctx, "clientKey", h.verifiedParams["clientKey"])
		ctx = context.WithValue(ctx, "hostBaseUrl", h.verifiedParams["hostBaseUrl"])
		ctx = context.WithValue(ctx, "token", h.verifiedParams["token"])
		ctx = context.WithValue(ctx, "tenantContext", h.verifiedParams["tenantContext"])

		ctx = context.WithValue(ctx, "httpClient", &hostrequest.HostRequest{Addon: h.addon, ClientKey: h.verifiedParams["clientKey"]})
	} else {
		if tenant, err := h.addon.Store.GetByUrl(hostBaseUrl); err != nil {
			log.ErrorF("error getting tenant %v: %v", hostBaseUrl, err)
		} else {
			ctx = context.WithValue(ctx, "tenantContext", tenant.Context.String())
		}
	}

	ctx = context.WithValue(ctx, "hostUrl", hostBaseUrl)
	ctx = context.WithValue(ctx, "hostStylesheetUrl",
		// TODO: if dev...
		getHostResourceUrl(true, ctx.Value("hostBaseUrl").(string), "css"))
	ctx = context.WithValue(ctx, "hostScriptUrl", "https://connect-cdn.atl-paas.net/all.js")

	r = r.WithContext(ctx)

	h.h.ServeHTTP(w, r)
}

func NewRequestMiddleware(addon *gonnect.Addon, verifiedParameters map[string]string) func(h http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		return RequestMiddleware{handler, addon, verifiedParameters}
	}
}
