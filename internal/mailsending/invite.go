package mailsending

import (
	"context"
	"fmt"

	"github.com/glasskube/cloud/internal/auth"
	internalctx "github.com/glasskube/cloud/internal/context"
	"github.com/glasskube/cloud/internal/db"
	"github.com/glasskube/cloud/internal/env"
	"github.com/glasskube/cloud/internal/mail"
	"github.com/glasskube/cloud/internal/mailtemplates"
	"github.com/glasskube/cloud/internal/types"
	"go.uber.org/zap"
)

func SendUserInviteMail(
	ctx context.Context,
	userAccount types.UserAccount,
	organization types.OrganizationWithBranding,
	userRole types.UserRole,
	applicationName string,
) error {
	mailer := internalctx.GetMailer(ctx)
	log := internalctx.GetLogger(ctx)

	// TODO: Should probably use a different mechanism for invite tokens but for now this should work OK
	if _, token, err := auth.GenerateVerificationTokenValidFor(userAccount); err != nil {
		log.Error("could not get current user for invite mail", zap.Error(err))
		return err
	} else {
		from := env.GetMailerConfig().FromAddress
		from.Name = organization.Name
		var email mail.Mail
		switch userRole {
		case types.UserRoleCustomer:
			if currentUser, err := db.GetCurrentUser(ctx); err != nil {
				log.Error("could not get current user for invite mail", zap.Error(err))
				return err
			} else {
				email = mail.New(
					mail.To(userAccount.Email),
					mail.From(from),
					mail.Bcc(currentUser.Email),
					mail.ReplyTo(currentUser.Email),
					mail.Subject("Welcome to Glasskube Cloud"),
					mail.HtmlBodyTemplate(mailtemplates.InviteCustomer(userAccount, organization, token, applicationName)),
				)
			}
		case types.UserRoleVendor:
			email = mail.New(
				mail.To(userAccount.Email),
				mail.From(from),
				mail.Subject("Welcome to Glasskube Cloud"),
				mail.HtmlBodyTemplate(mailtemplates.InviteUser(userAccount, organization, token)),
			)
		default:
			return fmt.Errorf("unknown UserRole: %v", userRole)
		}

		if err := mailer.Send(ctx, email); err != nil {
			log.Error(
				"could not send invite mail",
				zap.Error(err),
				zap.String("user", userAccount.Email),
				zap.String("token", token),
			)
			return err
		} else {
			log.Info("invite mail has been sent", zap.String("user", userAccount.Email))
			return nil
		}
	}
}
