package mailsending

import (
	"context"
	"fmt"

	"github.com/glasskube/distr/internal/auth"
	internalctx "github.com/glasskube/distr/internal/context"
	"github.com/glasskube/distr/internal/db"
	imail "github.com/glasskube/distr/internal/mail"
	"github.com/glasskube/distr/internal/mailtemplates"
	"github.com/glasskube/distr/internal/types"
	"go.uber.org/zap"
)

func SendUserInviteMail(
	ctx context.Context,
	userAccount types.UserAccount,
	organization types.OrganizationWithBranding,
	userRole types.UserRole,
	applicationName string,
	inviteURL string,
) error {
	mailer := internalctx.GetMailer(ctx)
	log := internalctx.GetLogger(ctx)
	auth := auth.Authentication.Require(ctx)

	from, err := organization.EmailFromAddressParsedOrDefault()
	if err != nil {
		return err
	}
	from.Name = organization.Name
	var email imail.Mail
	switch userRole {
	case types.UserRoleCustomer:
		if currentUser, err := db.GetUserAccountByID(ctx, auth.CurrentUserID()); err != nil {
			log.Error("could not get current user for invite mail", zap.Error(err))
			return err
		} else {
			email = imail.New(
				imail.To(userAccount.Email),
				imail.From(*from),
				imail.Bcc(currentUser.Email),
				imail.ReplyTo(currentUser.Email),
				imail.Subject("Welcome to Distr"),
				imail.HtmlBodyTemplate(mailtemplates.InviteCustomer(userAccount, organization, inviteURL, applicationName)),
			)
		}
	case types.UserRoleVendor:
		email = imail.New(
			imail.To(userAccount.Email),
			imail.From(*from),
			imail.Subject("Welcome to Distr"),
			imail.HtmlBodyTemplate(mailtemplates.InviteUser(userAccount, organization, inviteURL)),
		)
	default:
		return fmt.Errorf("unknown UserRole: %v", userRole)
	}

	if err := mailer.Send(ctx, email); err != nil {
		log.Error(
			"could not send invite mail",
			zap.Error(err),
			zap.String("user", userAccount.Email),
		)
		return err
	} else {
		log.Info("invite mail has been sent", zap.String("user", userAccount.Email))
		return nil
	}
}
