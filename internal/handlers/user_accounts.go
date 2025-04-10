package handlers

import (
	"context"
	"errors"
	"fmt"
	"github.com/glasskube/distr/internal/util"
	"io"
	"net/http"
	"net/url"

	"github.com/getsentry/sentry-go"
	"github.com/glasskube/distr/api"
	"github.com/glasskube/distr/internal/apierrors"
	"github.com/glasskube/distr/internal/auth"
	"github.com/glasskube/distr/internal/authjwt"
	internalctx "github.com/glasskube/distr/internal/context"
	"github.com/glasskube/distr/internal/db"
	"github.com/glasskube/distr/internal/env"
	"github.com/glasskube/distr/internal/mailsending"
	"github.com/glasskube/distr/internal/middleware"
	"github.com/glasskube/distr/internal/types"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func UserAccountsRouter(r chi.Router) {
	r.With(requireUserRoleVendor, middleware.RequireOrgAndRole).Group(func(r chi.Router) {
		r.Get("/", getUserAccountsHandler)
		r.Post("/", createUserAccountHandler)
		r.Route("/{userId}", func(r chi.Router) {
			r.Use(userAccountMiddleware)
			r.Delete("/", deleteUserAccountHandler)
			r.Patch("/image", patchImageUserAccountHandler)
		})
	})
	r.Get("/status", getUserAccountStatusHandler)
}

func getUserAccountsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	auth := auth.Authentication.Require(ctx)
	if userAccoutns, err := db.GetUserAccountsByOrgID(ctx, *auth.CurrentOrgID()); err != nil {
		sentry.GetHubFromContext(ctx).CaptureException(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		RespondJSON(w, userAccoutns)
	}
}

func getUserAccountStatusHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	auth := auth.Authentication.Require(ctx)
	userAccount := auth.CurrentUser()
	RespondJSON(w, map[string]any{
		"active": userAccount.PasswordHash != nil,
	})
}

func createUserAccountHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := internalctx.GetLogger(ctx)
	auth := auth.Authentication.Require(ctx)

	body, err := JsonBody[api.CreateUserAccountRequest](w, r)
	if err != nil {
		return
	}

	var organization types.OrganizationWithBranding
	userAccount := types.UserAccount{
		Email: body.Email,
		Name:  body.Name,
	}
	var inviteURL string

	if err := db.RunTx(ctx, func(ctx context.Context) error {
		if result, err := db.GetOrganizationWithBranding(ctx, *auth.CurrentOrgID()); err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return err
		} else {
			organization = *result
		}

		if err := db.CreateUserAccount(ctx, &userAccount); errors.Is(err, apierrors.ErrAlreadyExists) {
			// TODO: In the future this should not be an error, but we don't support multi-org users yet, so for now it is
			http.Error(w, err.Error(), http.StatusBadRequest)
			return err
		} else if err != nil {
			sentry.GetHubFromContext(ctx).CaptureException(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return err
		}

		if err := db.CreateUserAccountOrganizationAssignment(
			ctx,
			userAccount.ID,
			organization.ID,
			body.UserRole,
		); err != nil {
			sentry.GetHubFromContext(ctx).CaptureException(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return err
		}

		// TODO: Should probably use a different mechanism for invite tokens but for now this should work OK
		if _, token, err := authjwt.GenerateVerificationTokenValidFor(userAccount); err != nil {
			sentry.GetHubFromContext(ctx).CaptureException(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return err
		} else {
			inviteURL = fmt.Sprintf("%v/join?jwt=%v", env.Host(), url.QueryEscape(token))
			if err := mailsending.SendUserInviteMail(
				ctx,
				userAccount,
				organization,
				body.UserRole,
				body.ApplicationName,
				inviteURL,
			); err != nil {
				sentry.GetHubFromContext(ctx).CaptureException(err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return err
			}
		}

		return nil
	}); err != nil {
		log.Warn("could not create user", zap.Error(err))
		return
	}

	RespondJSON(w, api.CreateUserAccountResponse{ID: userAccount.ID, InviteURL: inviteURL})
}

func deleteUserAccountHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := internalctx.GetLogger(ctx)
	userAccount := internalctx.GetUserAccount(ctx)
	auth := auth.Authentication.Require(ctx)
	if userAccount.ID == auth.CurrentUserID() {
		http.Error(w, "UserAccount deleting themselves is not allowed", http.StatusForbidden)
	} else if err := db.DeleteUserAccountWithID(ctx, userAccount.ID); err != nil {
		log.Warn("error deleting user", zap.Error(err))
		if errors.Is(err, apierrors.ErrNotFound) {
			w.WriteHeader(http.StatusNoContent)
		} else if errors.Is(err, apierrors.ErrConflict) {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else {
			sentry.GetHubFromContext(ctx).CaptureException(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	} else {
		w.WriteHeader(http.StatusNoContent)
	}
}

func patchImageUserAccountHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := internalctx.GetLogger(ctx)
	userAccount := internalctx.GetUserAccount(ctx)

	if image, err := getImageFromMultipartForm(r); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	} else if err := db.UpdateUserAccountImage(ctx, userAccount.ID, *image); err != nil {
		log.Warn("error uploading user image", zap.Error(err))
		if errors.Is(err, apierrors.ErrNotFound) {
			w.WriteHeader(http.StatusNoContent)
		} else if errors.Is(err, apierrors.ErrConflict) {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else {
			sentry.GetHubFromContext(ctx).CaptureException(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func getImageFromMultipartForm(r *http.Request) (*types.Image, error) {
	if err := r.ParseMultipartForm(102400); err != nil {
		return nil, fmt.Errorf("failed to parse form: %w", err)
	}

	image := types.Image{}

	if file, head, err := r.FormFile("image"); err != nil {
		return nil, errors.New("image not found")
	} else if head.Size > 102400 {
		return nil, errors.New("file too large (max 100 KiB)")
	} else if data, err := io.ReadAll(file); err != nil {
		return nil, err
	} else {
		image.Image = data
		image.ImageFileName = &head.Filename
		image.ImageContentType = util.PtrTo(head.Header.Get("Content-Type"))
	}

	return &image, nil
}

func userAccountMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		log := internalctx.GetLogger(ctx)
		if userId, err := uuid.Parse(r.PathValue("userId")); err != nil {
			http.NotFound(w, r)
		} else if userAccount, err := db.GetUserAccountByID(ctx, userId); err != nil {
			if errors.Is(err, apierrors.ErrNotFound) {
				http.NotFound(w, r)
			} else {
				log.Warn("error getting user", zap.Error(err))
				sentry.GetHubFromContext(ctx).CaptureException(err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		} else {
			h.ServeHTTP(w, r.WithContext(internalctx.WithUserAccount(ctx, userAccount)))
		}
	})
}
