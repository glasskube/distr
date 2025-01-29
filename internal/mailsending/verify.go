package mailsending

import (
	"context"

	"github.com/glasskube/distr/internal/authjwt"
	internalctx "github.com/glasskube/distr/internal/context"
	"github.com/glasskube/distr/internal/mail"
	"github.com/glasskube/distr/internal/mailtemplates"
	"github.com/glasskube/distr/internal/types"
	"go.uber.org/zap"
)

func SendUserVerificationMail(ctx context.Context, userAccount types.UserAccount) error {
	mailer := internalctx.GetMailer(ctx)
	log := internalctx.GetLogger(ctx)

	// TODO: Should probably use a different mechanism for invite tokens but for now this should work OK
	if _, token, err := authjwt.GenerateVerificationTokenValidFor(userAccount); err != nil {
		log.Error("could not generate verification token for email verification", zap.Error(err))
		return err
	} else {
		mail := mail.New(
			mail.To(userAccount.Email),
			mail.Subject("Verify your Distr account"),
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
