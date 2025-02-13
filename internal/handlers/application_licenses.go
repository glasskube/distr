package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/getsentry/sentry-go"
	"github.com/glasskube/distr/internal/apierrors"
	"github.com/glasskube/distr/internal/auth"
	"github.com/glasskube/distr/internal/authn/authinfo"
	internalctx "github.com/glasskube/distr/internal/context"
	"github.com/glasskube/distr/internal/db"
	"github.com/glasskube/distr/internal/middleware"
	"github.com/glasskube/distr/internal/types"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

func ApplicationLicensesRouter(r chi.Router) {
	r.Use(middleware.RequireOrgID, middleware.RequireUserRole)
	r.Get("/", getApplicationLicenses)
	r.With(requireUserRoleVendor).Post("/", createApplicationLicense)
	r.Route("/{applicationLicenseId}", func(r chi.Router) {
		r.With(applicationLicenseMiddleware).Group(func(r chi.Router) {
			r.Get("/", getApplicationLicense)
			r.With(requireUserRoleVendor).Delete("/", deleteApplicationLicense)
			r.With(requireUserRoleVendor).Put("/", updateApplicationLicense)
		})
	})
}

func createApplicationLicense(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := internalctx.GetLogger(ctx)
	auth := auth.Authentication.Require(ctx)
	license, err := JsonBody[types.ApplicationLicenseWithVersions](w, r)
	if err != nil {
		return
	}
	license.OrganizationID = *auth.CurrentOrgID()

	// TODO registry validatin probably

	err = db.RunTx(ctx, pgx.TxOptions{}, func(ctx context.Context) error {
		if err := db.CreateApplicationLicense(ctx, &license.ApplicationLicense); errors.Is(err, apierrors.ErrConflict) {
			http.Error(w, "A license with this name already exists", http.StatusBadRequest)
			return err
		} else if err != nil {
			log.Warn("could not create license", zap.Error(err))
			sentry.GetHubFromContext(ctx).CaptureException(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return err
		}
		for _, version := range license.Versions {
			if err := db.AddVersionToApplicationLicense(ctx, &license.ApplicationLicense, version.ID); err != nil {
				log.Warn("could not add version to license", zap.Error(err))
				sentry.GetHubFromContext(ctx).CaptureException(err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return err
			}
		}
		return nil
	})
	if err == nil {
		RespondJSON(w, license)
	}
}

func updateApplicationLicense(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := internalctx.GetLogger(ctx)
	auth := auth.Authentication.Require(ctx)
	license, err := JsonBody[types.ApplicationLicenseWithVersions](w, r)
	if err != nil {
		return
	}
	license.OrganizationID = *auth.CurrentOrgID()

	existing := internalctx.GetApplicationLicense(ctx)
	if IsEmptyUUID(license.ID) {
		license.ID = existing.ID
	} else if license.ID != existing.ID {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if existing.OwnerUserAccountID != nil &&
		(license.OwnerUserAccountID == nil || *existing.OwnerUserAccountID != *license.OwnerUserAccountID) {
		http.Error(w, "Changing the license owner is not allowed", http.StatusBadRequest)
		return
	}

	// TODO registry validatin probably

	txErr := db.RunTx(ctx, pgx.TxOptions{}, func(ctx context.Context) error {
		if err := db.UpdateApplicationLicense(ctx, &license.ApplicationLicense); errors.Is(err, apierrors.ErrConflict) {
			http.Error(w, "A license with this name already exists", http.StatusBadRequest)
			return err
		} else if err != nil {
			log.Warn("could not update license", zap.Error(err))
			sentry.GetHubFromContext(ctx).CaptureException(err)
			w.WriteHeader(http.StatusInternalServerError)
			return err
		}

		for _, version := range license.Versions {
			alreadyExists := false
			for _, existingVersion := range existing.Versions {
				if version.ID == existingVersion.ID {
					alreadyExists = true
					break
				}
			}
			if !alreadyExists {
				if len(existing.Versions) == 0 {
					// we don't allow narrowing down the scope yet. If the existing license allows all versions,
					// setting some specific ones is not possible anymore
					err = errors.New("narrowing down license scope is not allowed yet")
					http.Error(w, err.Error(), http.StatusBadRequest)
					return err
				} else {
					if err := db.AddVersionToApplicationLicense(ctx, &license.ApplicationLicense, version.ID); err != nil {
						log.Warn("could not add version to license", zap.Error(err))
						sentry.GetHubFromContext(ctx).CaptureException(err)
						http.Error(w, err.Error(), http.StatusInternalServerError)
						return err
					}
				}
			}
		}

		for _, existingVersion := range existing.Versions {
			stillExists := false
			for _, version := range license.Versions {
				if version.ID == existingVersion.ID {
					stillExists = true
					break
				}
			}
			if !stillExists {
				if len(license.Versions) > 0 {
					// for now, removing specific versions from the license is not possible
					// for removal we also would have to check whether this version is used in some deployment target
					err = errors.New("narrowing down license scope is not allowed yet")
					http.Error(w, err.Error(), http.StatusBadRequest)
					return err
				} else {
					// however removing the relations is possible iff the user chose "all versions" by versions = []
					if err := db.RemoveVersionFromApplicationLicense(
						ctx, &license.ApplicationLicense, existingVersion.ID); err != nil {
						log.Warn("could not remove version from license", zap.Error(err))
						sentry.GetHubFromContext(ctx).CaptureException(err)
						http.Error(w, err.Error(), http.StatusInternalServerError)
						return err
					}
				}
			}
		}

		return nil
	})
	if txErr == nil {
		// TODO versions?
		RespondJSON(w, license)
	}
}

func getApplicationLicenses(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	auth := auth.Authentication.Require(ctx)
	if *auth.CurrentUserRole() == types.UserRoleVendor {
		if licenses, err := db.GetApplicationLicensesWithOrganizationID(ctx, *auth.CurrentOrgID()); err != nil {
			internalctx.GetLogger(ctx).Error("failed to get licenses", zap.Error(err))
			sentry.GetHubFromContext(ctx).CaptureException(err)
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			RespondJSON(w, licenses)
		}
	} else {
		if licenses, err :=
			db.GetApplicationLicensesWithOwnerID(ctx, auth.CurrentUserID(), *auth.CurrentOrgID()); err != nil {
			internalctx.GetLogger(ctx).Error("failed to get licenses", zap.Error(err))
			sentry.GetHubFromContext(ctx).CaptureException(err)
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			RespondJSON(w, licenses)
		}
	}
}

func getApplicationLicense(w http.ResponseWriter, r *http.Request) {
	license := internalctx.GetApplicationLicense(r.Context())
	RespondJSON(w, license)
}

func deleteApplicationLicense(w http.ResponseWriter, r *http.Request) {
	// TODO
	/*ctx := r.Context()
	log := internalctx.GetLogger(ctx)
	application := internalctx.GetApplication(ctx)
	auth := auth.Authentication.Require(ctx)
	if application.OrganizationID != *auth.CurrentOrgID() {
		http.NotFound(w, r)
	} else if err := db.DeleteApplicationWithID(ctx, application.ID); err != nil {
		log.Warn("error deleting application", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusNoContent)
	}*/
}

func applicationLicenseMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		auth := auth.Authentication.Require(ctx)
		if licenseId, err := uuid.Parse(r.PathValue("applicationLicenseId")); err != nil {
			http.Error(w, "applicationLicenseId is not a valid UUID", http.StatusBadRequest)
		} else if license, err := db.GetApplicationLicenseByID(ctx, licenseId); errors.Is(err, apierrors.ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
		} else if err != nil {
			internalctx.GetLogger(r.Context()).Error("failed to get license", zap.Error(err))
			sentry.GetHubFromContext(ctx).CaptureException(err)
			w.WriteHeader(http.StatusInternalServerError)
		} else if !canSeeLicense(auth, license) {
			w.WriteHeader(http.StatusForbidden)
		} else {
			ctx = internalctx.WithApplicationLicense(ctx, license)
			next.ServeHTTP(w, r.WithContext(ctx))
		}
	})
}

func canSeeLicense(auth authinfo.AuthInfo, license *types.ApplicationLicenseWithVersions) bool {
	if license.OrganizationID != *auth.CurrentOrgID() {
		return false
	}
	if *auth.CurrentUserRole() == types.UserRoleCustomer {
		if license.OwnerUserAccountID == nil || *license.OwnerUserAccountID != auth.CurrentUserID() {
			return false
		}
	}
	return true
}
