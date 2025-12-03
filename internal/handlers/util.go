package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"

	"github.com/getsentry/sentry-go"
	"github.com/glasskube/distr/internal/contenttype"
	internalctx "github.com/glasskube/distr/internal/context"
	"go.uber.org/zap"
)

func JsonBody[T any](w http.ResponseWriter, r *http.Request) (T, error) {
	var t T
	err := json.NewDecoder(r.Body).Decode(&t)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	return t, err
}

func RespondJSON(w http.ResponseWriter, data any) {
	w.Header().Add("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func readMultipartFile(w http.ResponseWriter, r *http.Request, formKey string) ([]byte, bool) {
	log := internalctx.GetLogger(r.Context())
	if file, head, err := r.FormFile(formKey); err != nil {
		if !errors.Is(err, http.ErrMissingFile) {
			log.Error("failed to get file from upload", zap.Error(err))
			sentry.GetHubFromContext(r.Context()).CaptureException(err)
			w.WriteHeader(http.StatusInternalServerError)
			return nil, false
		} else {
			return nil, true
		}
	} else {
		log.Sugar().Debugf("got file %v with type %v and size %v", head.Filename, head.Header, head.Size)
		// max file size is 100KiB
		if head.Size > 102400 {
			log.Debug("large body was rejected")
			w.WriteHeader(http.StatusRequestEntityTooLarge)
			fmt.Fprintln(w, "file too large (max 100 KiB)")
			return nil, false
		} else if err := contenttype.IsYaml(head.Header); err != nil {
			w.WriteHeader(http.StatusUnsupportedMediaType)
			fmt.Fprint(w, html.EscapeString(err.Error()))
			return nil, false
		} else if data, err := io.ReadAll(file); err != nil {
			log.Error("failed to read file from upload", zap.Error(err))
			sentry.GetHubFromContext(r.Context()).CaptureException(err)
			w.WriteHeader(http.StatusInternalServerError)
			return nil, false
		} else {
			return data, true
		}
	}
}
