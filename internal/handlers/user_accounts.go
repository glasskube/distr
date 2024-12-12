package handlers

import (
	"errors"
	"net/http"

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
	"go.uber.org/zap"
)

func UserAccountsRouter(r chi.Router) {
	r.With(requireUserRoleVendor).Group(func(r chi.Router) {
		r.Get("/", getUserAccountsHandler)
		r.Post("/", createUserAccountHandler)
	})
}

func getUserAccountsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if orgId, err := auth.CurrentOrgId(ctx); err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
	} else if userAccoutns, err := db.GetUserAccountsWithOrgID(ctx, orgId); err != nil {
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

	organization, err := db.GetCurrentOrg(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	userAccount := types.UserAccount{
		Email: body.Email,
		Name:  body.Name,
	}

	if err := db.CreateUserAccount(ctx, &userAccount); errors.Is(err, apierrors.ErrAlreadyExists) {
		// TODO: In the future this should not be an error, but we don't support multi-org users yet, so for now it is
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := db.CreateUserAccountOrganizationAssignment(ctx, userAccount.ID, organization.ID, body.UserRole); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// TODO: Should probably use a different mechanism for invite tokens but for now this should work OK
	_, token, err := auth.GenerateTokenValidFor(
		userAccount,
		types.OrganizationWithUserRole{Organization: *organization, UserRole: body.UserRole},
		env.InviteTokenValidDuration(),
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	if err := mailer.Send(ctx, mail.New(
		mail.To(userAccount.Email),
		mail.Subject("Welcome to Glasskube Cloud"),
		mail.HtmlBodyTemplate(mailtemplates.Invite(userAccount, *organization, token)),
	)); err != nil {
		log.Error("failed to send invite mail", zap.Error(err))
	}
}
