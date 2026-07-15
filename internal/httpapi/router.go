// Package httpapi is the REST surface described in docs/plan.md: orgs and
// the instances (tenants) each org owns, backed by internal/provisioning.
package httpapi

import (
	"net/http"

	_ "listnun/internal/apidocs" // swagger docs, registered via side-effect init
	"listnun/internal/authn"
	"listnun/internal/provisioning"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger"
)

type API struct {
	svc      *provisioning.Service
	verifier *authn.Verifier
}

func New(svc *provisioning.Service, verifier *authn.Verifier) http.Handler {
	a := &API{svc: svc, verifier: verifier}

	r := chi.NewRouter()
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)

	r.Get("/api/health", a.health)
	r.Get("/swagger/*", httpSwagger.WrapHandler)

	r.Route("/api/v1", func(r chi.Router) {
		r.Use(a.authMiddleware)

		r.Get("/me", a.getMe)

		r.Route("/orgs", func(r chi.Router) {
			r.Get("/", a.listOrgs)
			r.Post("/", a.createOrg)

			r.Route("/{orgID}", func(r chi.Router) {
				r.Use(a.requireOrgMembership)

				r.Route("/instances", func(r chi.Router) {
					r.Get("/", a.listInstances)
					r.Post("/", a.createInstance)

					r.Route("/{instanceID}", func(r chi.Router) {
						r.Get("/", a.getInstance)
						r.Get("/events", a.listEvents)
						r.Post("/setup-link", a.resendSetupLink)
						r.Get("/sender-identity", a.getSenderIdentity)
						r.Post("/sender-identity", a.addSenderIdentity)
						r.Delete("/sender-identity", a.deleteSenderIdentity)
						r.Get("/postmark-server", a.getPostmarkServer)
						r.Delete("/postmark-server", a.deletePostmarkServer)
						r.Post("/postmark-server/resync", a.resyncPostmarkServer)
						r.Get("/custom-domain", a.getCustomDomain)
						r.Post("/custom-domain", a.addCustomDomain)
						r.Delete("/custom-domain", a.deleteCustomDomain)
					})
				})

				r.Route("/members", func(r chi.Router) {
					r.Get("/", a.listMembers)
					r.Post("/", a.inviteMember)
				})
			})
		})

		r.Route("/admin", func(r chi.Router) {
			r.Use(a.requireSuperAdmin)

			r.Get("/orgs", a.adminListOrgs)
			r.Get("/instances", a.adminListInstances)
			r.Route("/instances/{instanceID}", func(r chi.Router) {
				r.Get("/", a.adminGetInstance)
				r.Delete("/", a.adminDeleteInstance)
				r.Put("/status", a.adminSetTenantStatus)
				r.Post("/setup-link", a.adminResendSetupLink)
				r.Delete("/postmark-server", a.adminDeletePostmarkServer)
			})
		})
	})

	return r
}

// health godoc
//
//	@Summary	Health check
//	@Tags		health
//	@Produce	json
//	@Success	200	{object}	healthResponse
//	@Router		/health [get]
func (a *API) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
