package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-enjin/github-com-craftamap-atlas-gonnect"
	atlasjwt "github.com/go-enjin/github-com-craftamap-atlas-gonnect/atlas-jwt"
	"github.com/go-enjin/github-com-craftamap-atlas-gonnect/util"

	"github.com/golang-jwt/jwt"

	"github.com/go-enjin/be/pkg/log"
)

const JWT_PARAM = "jwt"
const AUTH_HEADER = "authorization"

type AuthenticationMiddleware struct {
	h       http.Handler
	addon   *gonnect.Addon
	skipQsh bool
}

func extractUnverifiedClaims(tokenStr string, validator jwt.Keyfunc) (jwt.MapClaims, bool) {
	token, _ := jwt.Parse(tokenStr, validator)
	if token == nil {
		return nil, false
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		return claims, true
	} else {
		log.ErrorF("Invalid JWT Token")
		return nil, false
	}
}

func ExtractJwt(r *http.Request) (string, bool) {
	var tokenInQuery = r.URL.Query().Get(JWT_PARAM)
	if tokenInQuery == "" && r.Body == http.NoBody {
		return "", false
	}

	tokenInBody := r.PostFormValue(JWT_PARAM)

	if tokenInQuery != "" && tokenInBody != "" {
		return "", false
	}

	var token string
	if tokenInBody != "" {
		token = tokenInBody
	} else {
		token = tokenInQuery
	}

	authHeader := r.Header.Get(AUTH_HEADER)
	if authHeader != "" && strings.HasPrefix(authHeader, "JWT ") {
		if token != "" {

		} else {
			token = strings.TrimPrefix(authHeader, "JWT ")
		}
	}

	// TODO: JS implements r.Query().Get(TOKEN_KEY_PARAM) and r.Query().Get(TOKEN_KEY_HEADER) as possible
	// Headers. However, it is marked as deprecated - we should follow the development of the js library
	// and see if it gets removed. For now, this should work

	if token == "" {
		return "", false
	}
	return token, true
}

func (h AuthenticationMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// TODO: Add better logging here
	// TODO: Add AC_OPTS no-auth
	// TODO: Refactor to be more compact
	// TODO: scoping

	token, ok := ExtractJwt(r)
	log.DebugF(r.URL.String())
	if !ok {
		util.SendError(w, r, h.addon, 401, "Could not find auth data on request")
		return
	}

	unverifiedClaims, ok := extractUnverifiedClaims(token, nil)

	if !ok {
		util.SendError(w, r, h.addon, 401, "JWT claim did not contain the issuer (iss) claim")
		return
	}

	if unverifiedClaims["iss"] == "" {
		util.SendError(w, r, h.addon, 401, "JWT claim did not contain the issuer (iss) claim")
		return
	}

	clientKey := unverifiedClaims["iss"].(string)

	// if unverifiedClaims["aud"] != nil && unverifiedClaims["aud"] != "" {
	// clientKey = unverifiedClaims["aud"].(string)
	// log.DebugF("using aud as clientKey: %v", unverifiedClaims["aud"])
	// w.WriteHeader(204)
	// return
	// }

	log.DebugF("using clientKey: %v", clientKey)

	queryStringHash := unverifiedClaims["qsh"]
	if queryStringHash == "" && !h.skipQsh {
		util.SendError(w, r, h.addon, 401, "JWT claim did not contain the query string hash (qsh) claim")
	}

	tenant, err := h.addon.Store.Get(clientKey)

	if err != nil {
		util.SendError(w, r, h.addon, 500, "Could not lookup stored client data for clientKey")
		return
	}

	secret := tenant.SharedSecret
	if secret == "" {
		util.SendError(w, r, h.addon, 401, "Could not find JQT sharedSecret in tenant clientKey")
		return
	}

	verifiedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {

		switch token.Header["alg"] {
		case "none":
			return nil, fmt.Errorf("alg is none, discard")
		case "HS256":
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("expected HS256 signing method, actual: %T", token.Method)
			}
		case "RS256":
			// when installing overtop another tenant situation, we're receiving
			// RS256 when atlas-gonnect is always expecting just HS256, this
			// problem is alleviated by changes to verify-installation ServeHTTP
			// where db lookups cause a different installation path
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("expected RS256 signing method, actual: %T", token.Method)
			}
		default:
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(secret), nil
	})

	if err != nil {
		log.ErrorF("JWT Token verification error: %v", err)
		util.SendError(w, r, h.addon, 500, "Could not verify JWT Token")
		return
	}

	err = verifiedToken.Claims.Valid()
	if err != nil {
		util.SendError(w, r, h.addon, 500, "Could not find verify JWT Claims; Auth request has expired")
		return
	}

	claims, ok := verifiedToken.Claims.(jwt.MapClaims)
	if !ok {
		util.SendError(w, r, h.addon, 500, "Could not cast Claims")
		return
	}

	ok = ValidateQshFromRequest(claims, r, h.addon, h.skipQsh)
	if !ok {
		util.SendError(w, r, h.addon, 401, "Auth failure: Query hash mismatch")
		return
	}

	log.DebugF("Auth successful")

	createSessionToken := func() (string, error) {
		verClaims := verifiedToken.Claims.(jwt.MapClaims)

		claims := &jwt.StandardClaims{
			Issuer: *h.addon.Key,
			// TODO: Check if subject can be asserted
			Audience: clientKey,
		}
		subject, ok := verClaims["subject"].(string)
		if ok {
			claims.Subject = subject
		}

		// TODO: We may have to add the context workaround, but lets ignore it for now

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		signedToken, err := token.SignedString([]byte(tenant.SharedSecret))
		if err != nil {
			return "", err
		}

		w.Header().Set("X-acpt", signedToken) // TODO: Do we really need to do this?

		return signedToken, nil
	}

	tokenString, err := createSessionToken()
	if err != nil {
		util.SendError(w, r, h.addon, 500, fmt.Sprintf("Could not create new access token %s", err))
		return
	}

	oldVerClaims := verifiedToken.Claims.(jwt.MapClaims)

	tenant, err = h.addon.Store.Get(clientKey)
	if err != nil {
		util.SendError(w, r, h.addon, 500, fmt.Sprintf("Could not create new access token %s", err))
		return
	}

	accountID := ""
	if oldVerClaims != nil && oldVerClaims["sub"] != nil {
		accountID = oldVerClaims["sub"].(string)
	}

	verifiedParams := map[string]string{
		"clientKey":   clientKey,
		"hostBaseUrl": tenant.BaseURL,
		"token":       tokenString,
		// TODO: We may have to add the context workaround instead of just using sub as userAccountId, but lets ignore it for now
		"userAccountId": accountID,
		"tenantContext": tenant.Context.String(),
	}

	requestHandler := NewRequestMiddleware(h.addon, verifiedParams)

	requestHandler(h.h).ServeHTTP(w, r)
}

func ValidateQshFromRequest(claims jwt.MapClaims, r *http.Request, addon *gonnect.Addon, skipQsh bool) bool {
	if !skipQsh && claims["qsh"] != "" {
		expectedHash := atlasjwt.CreateQueryStringHash(r, false, addon.Config.BaseUrl)
		if claims["qsh"] != expectedHash {

			expectedHash := atlasjwt.CreateQueryStringHash(r, true, addon.Config.BaseUrl)
			if claims["qsh"] != expectedHash {
				return false
			}
		}
	}
	return true
}

func NewAuthenticationMiddleware(addon *gonnect.Addon, skipQsh bool) func(h http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		return AuthenticationMiddleware{handler, addon, skipQsh}
	}
}