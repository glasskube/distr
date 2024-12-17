package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/getsentry/sentry-go"
	"github.com/glasskube/cloud/api"
	"github.com/glasskube/cloud/internal/apierrors"
	"github.com/glasskube/cloud/internal/auth"
	internalctx "github.com/glasskube/cloud/internal/context"
	"github.com/glasskube/cloud/internal/db"
	"github.com/glasskube/cloud/internal/env"
	"github.com/glasskube/cloud/internal/mail"
	"github.com/glasskube/cloud/internal/mailtemplates"
	"github.com/glasskube/cloud/internal/types"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

func UserAccountsRouter(r chi.Router) {
	r.With(requireUserRoleVendor).Group(func(r chi.Router) {
		r.Get("/", getUserAccountsHandler)
		r.Post("/", createUserAccountHandler)
		r.Route("/{userId}", func(r chi.Router) {
			r.Use(userAccountMiddleware)
			r.Delete("/", deleteUserAccountHandler)
		})
	})
}

func getUserAccountsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if orgId, err := auth.CurrentOrgId(ctx); err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
	} else if userAccoutns, err := db.GetUserAccountsWithOrgID(ctx, orgId); err != nil {
		sentry.GetHubFromContext(ctx).CaptureException(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		RespondJSON(w, userAccoutns)
	}
}

func createUserAccountHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	mailer := internalctx.GetMailer(ctx)
	log := internalctx.GetLogger(ctx)

	body, err := JsonBody[api.CreateUserAccountRequest](w, r)
	if err != nil {
		return
	}

	var organization types.Organization
	userAccount := types.UserAccount{
		Email: body.Email,
		Name:  body.Name,
	}

	if err := db.RunTx(ctx, pgx.TxOptions{}, func(ctx context.Context) error {
		if result, err := db.GetCurrentOrg(ctx); err != nil {
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
		return nil
	}); err != nil {
		log.Warn("could not create user", zap.Error(err))
		return
	}

	// TODO: Should probably use a different mechanism for invite tokens but for now this should work OK
	_, token, err := auth.GenerateVerificationTokenValidFor(
		userAccount,
		types.OrganizationWithUserRole{Organization: organization, UserRole: body.UserRole},
		env.InviteTokenValidDuration(),
	)
	if err != nil {
		sentry.GetHubFromContext(ctx).CaptureException(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	var email mail.Mail
	if body.UserRole == types.UserRoleCustomer {
		if currentUser, err := db.GetCurrentUserWithRole(ctx); err != nil {
			log.Error("failed to generate invite mail", zap.Error(err))
			return
		} else {
			email = mail.New(
				mail.To(userAccount.Email),
				mail.Bcc(currentUser.Email),
				mail.ReplyTo(currentUser.Email),
				mail.Subject("Welcome to Glasskube Cloud"),
				mail.HtmlBodyTemplate(mailtemplates.InviteCustomer(userAccount, organization, token, body.ApplicationName)),
			)
		}
	} else {
		email = mail.New(
			mail.To(userAccount.Email),
			mail.Subject("Welcome to Glasskube Cloud"),
			mail.HtmlBodyTemplate(mailtemplates.InviteUser(userAccount, organization, token)),
		)
	}

	if err := mailer.Send(ctx, email); err != nil {
		sentry.GetHubFromContext(ctx).CaptureException(err)
		log.Error("failed to send invite mail", zap.Error(err))
	}
}

func deleteUserAccountHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := internalctx.GetLogger(ctx)
	userAccount := internalctx.GetUserAccount(ctx)
	if currentUserID, err := auth.CurrentUserId(ctx); err != nil {
		log.Warn("error getting current user", zap.Error(err))
		sentry.GetHubFromContext(ctx).CaptureException(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else if userAccount.ID == currentUserID {
		http.Error(w, "UserAccount deleting themselves is not allowed", http.StatusForbidden)
	} else if err := db.DeleteUserAccountWithID(ctx, userAccount.ID); err != nil {
		log.Warn("error deleting user", zap.Error(err))
		sentry.GetHubFromContext(ctx).CaptureException(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusNoContent)
	}
}

func userAccountMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		log := internalctx.GetLogger(ctx)
		if userId := r.PathValue("userId"); userId == "" {
			http.Error(w, "missing userId", http.StatusBadRequest)
		} else if userAccount, err := db.GetUserAccountWithID(ctx, userId); err != nil {
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
