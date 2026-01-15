package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/distr-sh/distr/internal/auth"
	"github.com/distr-sh/distr/internal/billing"
	internalctx "github.com/distr-sh/distr/internal/context"
	"github.com/distr-sh/distr/internal/handlerutil"
	"go.uber.org/zap"
)

func CreateBillingPortalSessionHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := internalctx.GetLogger(ctx)
	auth := auth.Authentication.Require(ctx)
	org := auth.CurrentOrg()

	// Check if organization has a Stripe customer ID
	if org.StripeCustomerID == nil || *org.StripeCustomerID == "" {
		http.Error(w, "no stripe customer found for organization", http.StatusConflict)
		return
	}

	var body struct {
		ReturnURL string `json:"returnUrl"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		log.Debug("bad json payload", zap.Error(err))
		http.Error(w, "bad json payload", http.StatusBadRequest)
		return
	}

	session, err := billing.CreateBillingPortalSession(ctx, billing.BillingPortalSessionParams{
		CustomerID: *org.StripeCustomerID,
		ReturnURL:  fmt.Sprintf("%v/subscription", handlerutil.GetRequestSchemeAndHost(r)),
	})
	if err != nil {
		log.Error("failed to create billing portal session", zap.Error(err))
		http.Error(w, "failed to create billing portal session", http.StatusInternalServerError)
		return
	}

	RespondJSON(w, map[string]string{
		"url": session.URL,
	})
}
