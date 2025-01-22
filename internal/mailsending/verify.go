package mailsending

import (
	"context"

	"github.com/glasskube/cloud/internal/auth"
	internalctx "github.com/glasskube/cloud/internal/context"
	"github.com/glasskube/cloud/internal/mail"
	"github.com/glasskube/cloud/internal/mailtemplates"
	"github.com/glasskube/cloud/internal/types"
	"go.uber.org/zap"
)

func SendUserVerificationMail(ctx context.Context, userAccount types.UserAccount) error {
	mailer := internalctx.GetMailer(ctx)
	log := internalctx.GetLogger(ctx)

	// TODO: Should probably use a different mechanism for invite tokens but for now this should work OK
	if _, token, err := auth.GenerateVerificationTokenValidFor(userAccount); err != nil {
		log.Error("could not generate verification token for email verification", zap.Error(err))
		return err
	} else {
		mail := mail.New(
			mail.To(userAccount.Email),
			mail.Subject("Verify your Glasskube Cloud Email"),
			mail.HtmlBodyTemplate(mailtemplates.VerifyEmail(userAccount, token)),
		)
		if err := mailer.Send(ctx, mail); err != nil {
			log.Error(
				"could not send verification mail",
				zap.Error(err),
				zap.String("user", userAccount.Email),
			)
			return err
		} else {
			log.Info("verification mail has been sent", zap.String("user", userAccount.Email))
			return nil
		}
	}
}
