package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"

	health "astuart.co/go-healthcheck"
	"github.com/davecgh/go-spew/spew"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"golang.org/x/oauth2"
)

var ctx, cancel = context.WithCancel(context.Background())

func init() {
	go func() {
		ch := make(chan os.Signal, 2)
		signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT)
		i := 0
		for range ch {
			i++
			if i > 1 {
				os.Exit(1)
			}
			cancel()
		}
	}()
}

func main() {
	cfg := setupConfig()
	lg.Info("started")
	clientID := cfg.GetString("client.id")
	lg.WithField("clientid", clientID).Info("got clientid")

	oauthConf := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: cfg.GetString("client.secret"),
		Endpoint: oauth2.Endpoint{
			TokenURL: "https://api.id.me/oauth/token",
			AuthURL:  "https://groups.id.me/",
		},
		Scopes:      []string{"alumni", "employee", "responder", "government", "student", "nurse"},
		RedirectURL: "https://idme.astuart.co:8444/api/v1/openid/callback",
	}

	reg := health.NewRegistry()
	r := mux.NewRouter()
	r.Path("/health").Handler(reg)

	api := r.PathPrefix("/api/v1").Subrouter()

	api.PathPrefix("/openid/callback").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t, err := oauthConf.Exchange(r.Context(), r.URL.Query().Get("code"))
		if err != nil {
			log.Println(err)
			http.Error(w, "couldn't exchange code for token", 401)
			return
		}

		spew.Dump(t)

		cli := oauth2.NewClient(ctx, oauth2.StaticTokenSource(t))

		res, err := cli.Get("https://api.id.me/api/public/v3/attributes.json")
		if err != nil {
			lg.WithError(err).Error("could not get groups with bearer token")
			return
		}
		defer res.Body.Close()

		var ifc interface{}
		json.NewDecoder(res.Body).Decode(&ifc)

		json.NewEncoder(w).Encode(map[string]interface{}{
			"token": t,
			"info":  ifc,
		})
	})

	r.Path("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uStr := oauthConf.AuthCodeURL("foo")
		u, _ := url.Parse(uStr)
		q := u.Query()
		delete(q, "scope")
		q["scopes"] = []string{strings.Join(oauthConf.Scopes, ",")}
		u.RawQuery = q.Encode()

		http.Redirect(w, r, u.String(), http.StatusTemporaryRedirect)
	})

	go lg.Fatal(http.ListenAndServeTLS(
		":8444",
		"./idme.astuart.co.crt",
		"./idme.astuart.co.key",
		handlers.CORS(
			handlers.AllowedHeaders([]string{"Authorization"}),
			handlers.AllowedOrigins([]string{"*"}),
		)(r),
	))
	<-ctx.Done()
}
