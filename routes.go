package main

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/treeder/firetils"
	"github.com/treeder/gotils/v2"
	"github.com/treeder/quickstart/globals"
)

func setupRoutes(ctx context.Context, r chi.Router) {
  r.Get("/", gotils.ErrorHandler(hi))
  r.Route("/v1", func(r chi.Router) {
    r.Post("/msg", gotils.ErrorHandler(postMsg))
    r.Post("/assets/tokenize", gotils.ErrorHandler(tokenize))
    r.Post("/assets", gotils.ErrorHandler(addAssets))
    r.Put("/assets/tokenize", gotils.ErrorHandler(completeTokenization))
    r.Get("/assets/info/{id}", gotils.ErrorHandler(getAsset))
    r.Get("/assets/{uid}", gotils.ErrorHandler(getAssets))
    r.Get("/organizations/admin/{email}", gotils.ErrorHandler(getAdminOrgs))
    r.Get("/organizations/user/{uid}", gotils.ErrorHandler(getOrganizations))
    r.Get("/organizations/{orgid}/assets", gotils.ErrorHandler(getOrganizationAssets))
    r.Get("/organizations/{orgid}", gotils.ErrorHandler(getOrganization))
    r.Post("/organizations/{orgid}", gotils.ErrorHandler(inviteUser))
    r.Delete("/organizations/{orgid}", gotils.ErrorHandler(leaveUser))
    r.Post("/organizations", gotils.ErrorHandler(addOrganization))
    r.Get("/users/{orgid}", gotils.ErrorHandler(getOrgUsers))
    r.With(firetils.OptionalAuth).Get("/msgs", gotils.ErrorHandler(getMsgs))
    r.With(firetils.FireAuth).Post("/session", gotils.ErrorHandler(createSession))
  })
  r.Post("/data", gotils.ErrorHandler(getMsgs))
}

func createSession(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	expiresIn := time.Hour * 24 * 14
	// Create the session cookie. This will also verify the ID token in the process.
	// The session cookie will have the same claims as the ID token.
	// To only allow session cookie setting on recent sign-in, auth_time in ID token
	// can be checked to ensure user was recently signed in before creating a session cookie.
	idToken := r.Header.Get("Authorization")
	splitToken := strings.Split(idToken, " ")
	cookie, err := globals.App.Auth.SessionCookie(ctx, splitToken[1], expiresIn)
	if err != nil {
		gotils.NewHTTPError("Failed to create a cookie", http.StatusInternalServerError)
	}
	gotils.WriteObject(w, http.StatusOK, map[string]interface{}{"cookie": cookie, "expires": int(expiresIn.Seconds())})
	return nil
}

func hi(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	gotils.L(ctx).Info().Println("hi!")

	// TODO: store this in our own db as we build it up
	gotils.WriteObject(w, http.StatusOK, map[string]interface{}{"hello": "world"})

	return nil
}
