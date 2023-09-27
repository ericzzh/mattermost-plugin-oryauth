package main

import (
	"fmt"
	"net/http"
	"net/url"
	"sync"

	"context"
	"github.com/mattermost/mattermost-server/v6/plugin"
	client "github.com/ory/hydra-client-go"
	"strings"
)

// Plugin implements the interface expected by the Mattermost server to communicate between the server and plugin processes.
type Plugin struct {
	plugin.MattermostPlugin

	// configurationLock synchronizes access to the configuration.
	configurationLock sync.RWMutex

	// configuration is the active plugin configuration. Consult getConfiguration and
	// setConfiguration for usage.
	configuration *configuration
}

// ServeHTTP demonstrates a plugin that handles HTTP requests by greeting the world.
func (p *Plugin) ServeHTTP(c *plugin.Context, w http.ResponseWriter, r *http.Request) {

	userID := r.Header.Get("Mattermost-User-ID")

	if userID == "" {
		p.redirectToLogin(r, w)
		return
	}

	// fmt.Fprintf(w, "path:%v, parameters:%v", r.URL.Path, r.URL.RawQuery)

	switch r.URL.Path {
	case "/login":
		p.handleLoginRequest(c, w, r)
	case "/consent":
		p.handleConsentRequest(c, w, r)
	default:
		http.NotFound(w, r)
	}

}

func (p *Plugin) redirectToLogin(r *http.Request, w http.ResponseWriter) {
	rurl := strings.Replace(r.RequestURI, "/plugins/", "/plug/", 1)
	http.Redirect(w, r, *p.API.GetConfig().ServiceSettings.SiteURL+"/login?redirect_to="+url.QueryEscape(rurl), http.StatusFound)
}

func (p *Plugin) handleLoginRequest(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
	apiClient := p.getHydraAPIClient()
	lc := r.URL.Query().Get("login_challenge")
	hr := apiClient.OAuth2Api.GetOAuth2LoginRequest(context.Background()).LoginChallenge(lc)
	lreq, lres, err := apiClient.OAuth2Api.GetOAuth2LoginRequestExecute(hr)
	if err != nil {
		fmt.Fprintf(w, "get login challenge data failed: %v, response:%v", err, lres)
		return
	}

	ar := apiClient.OAuth2Api.AcceptOAuth2LoginRequest(context.Background()).LoginChallenge(lc)

	if lreq.Skip {
		ar = ar.AcceptOAuth2LoginRequest(*client.NewAcceptOAuth2LoginRequest(lreq.Subject))
	} else {
		ses, perr := p.API.GetSession(c.SessionId)
		if perr != nil {
			fmt.Fprintf(w, "get session failed: %v", err)
			return
		}
		ar = ar.AcceptOAuth2LoginRequest(*client.NewAcceptOAuth2LoginRequest(ses.UserId))
	}

	to, lres, err := apiClient.OAuth2Api.AcceptOAuth2LoginRequestExecute(ar)
	if err != nil {
		fmt.Fprintf(w, "send accept login request failed: %v, response: %v", err, lres)
		return
	}
	http.Redirect(w, r, to.RedirectTo, http.StatusFound)
	return

}

func (p *Plugin) handleConsentRequest(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
	apiClient := p.getHydraAPIClient()
	cc := r.URL.Query().Get("consent_challenge")
	hr := apiClient.OAuth2Api.GetOAuth2ConsentRequest(context.Background()).ConsentChallenge(cc)
	lreq, lres, err := apiClient.OAuth2Api.GetOAuth2ConsentRequestExecute(hr)
	if err != nil {
		fmt.Fprintf(w, "get consent challenge data failed: %v, respose: %v", err, lres)
		return
	}

	ar := apiClient.OAuth2Api.AcceptOAuth2ConsentRequest(context.Background()).ConsentChallenge(cc)
	newcr := client.NewAcceptOAuth2ConsentRequest()
	newcr.SetGrantScope(lreq.GetRequestedScope())

	if !(*lreq.Skip) {
		ses, perr := p.API.GetSession(c.SessionId)
		if perr != nil {
			fmt.Fprintf(w, "get session failed: %v", err)
			return
		}

		ccs := client.NewAcceptOAuth2ConsentRequestSession()
		ccs.SetAccessToken(map[string]string{
			"access_token": ses.Token,
		})

		u, apperr := p.API.GetUser(ses.UserId)
		if apperr != nil {
			fmt.Fprintf(w, "get user failed: %v", apperr)
			return
		}

		var name string
		if u.Nickname != "" {
			name = u.Nickname
		} else {
			name = u.LastName + u.FirstName
		}
		ccs.SetIdToken(map[string]string{
			"email":              u.Email,
			"preferred_username": u.Username,
			"name":               name,
		})
		newcr.SetSession(*ccs)
	}

	ar = ar.AcceptOAuth2ConsentRequest(*newcr)

	to, lres, err := apiClient.OAuth2Api.AcceptOAuth2ConsentRequestExecute(ar)
	if err != nil {
		fmt.Fprintf(w, "send accept consent request failed: %v, response: %v", err, lres)
		return
	}
	http.Redirect(w, r, to.RedirectTo, http.StatusFound)
	return
}

func (p *Plugin) getHydraAPIClient() *client.APIClient {
	configuration := client.NewConfiguration()
	configuration.Servers = []client.ServerConfiguration{
		{
			URL: p.configuration.OryHydraAdminEndPoint,
		},
	}

	return client.NewAPIClient(configuration)
}

// See https://developers.mattermost.com/extend/plugins/server/reference/
