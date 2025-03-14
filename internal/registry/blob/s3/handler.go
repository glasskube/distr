package s3

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/glasskube/distr/internal/registry/blob"
	"github.com/glasskube/distr/internal/util"
	v1 "github.com/google/go-containerregistry/pkg/v1"
)

const (
	accessKey       = "distr"
	accessKeySecret = "distr123"
	bucket          = "distr"
)

type blobHandler struct {
	s3Client        *s3.Client
	s3PresignClient *s3.PresignClient
	allowRedirect   bool
}

var _ blob.BlobHandler = &blobHandler{}
var _ blob.BlobStatHandler = &blobHandler{}
var _ blob.BlobPutHandler = &blobHandler{}
var _ blob.BlobDeleteHandler = &blobHandler{}

func NewBlobHandler(allowRedirect bool) blob.BlobHandler {
	s3Client := s3.New(s3.Options{
		Region:       "eu-central-1",
		UsePathStyle: true,
		BaseEndpoint: util.PtrTo("http://localhost:9000"),
		Credentials: aws.NewCredentialsCache(
			credentials.NewStaticCredentialsProvider(accessKey, accessKeySecret, ""),
		),
	})
	return &blobHandler{
		s3Client:        s3Client,
		s3PresignClient: s3.NewPresignClient(s3Client),
		allowRedirect:   allowRedirect,
	}
}

// Get implements blob.BlobHandler.
func (handler *blobHandler) Get(
	ctx context.Context,
	repo string,
	h v1.Hash,
	allowRedirect bool,
) (io.ReadCloser, error) {
	key := h.String()
	if handler.allowRedirect && allowRedirect {
		resp, err := handler.s3PresignClient.PresignGetObject(ctx,
			&s3.GetObjectInput{Bucket: util.PtrTo(bucket), Key: &key})
		if err != nil {
			return nil, convertErrNotFound(err)
		} else {
			return nil, blob.RedirectError{
				Code:     http.StatusTemporaryRedirect,
				Location: resp.URL,
			}
		}
	} else {
		obj, err := handler.s3Client.GetObject(ctx, &s3.GetObjectInput{Bucket: util.PtrTo(bucket), Key: &key})
		if err != nil {
			return nil, convertErrNotFound(err)
		}
		return obj.Body, nil
	}
}

// Stat implements blob.BlobStatHandler.
func (handler *blobHandler) Stat(ctx context.Context, repo string, h v1.Hash) (int64, error) {
	key := h.String()
	obj, err := handler.s3Client.HeadObject(ctx, &s3.HeadObjectInput{Bucket: util.PtrTo(bucket), Key: &key})
	if err != nil {
		return 0, convertErrNotFound(err)
	}
	return *obj.ContentLength, nil
}

// Put implements blob.BlobPutHandler.
func (handler *blobHandler) Put(ctx context.Context, repo string, h v1.Hash, contentType string, r io.Reader) error {
	key := h.String()
	if rc, ok := r.(io.Closer); ok {
		defer rc.Close()
	}

	// The AWS S3 SDK requires a io.ReadSeeker event though the interface only specifies io.Reader
	if _, ok := r.(io.Seeker); !ok {
		if data, err := io.ReadAll(r); err != nil {
			return err
		} else {
			r = bytes.NewReader(data)
		}
	}

	input := s3.PutObjectInput{
		Bucket: util.PtrTo(bucket),
		Key:    &key,
		Body:   r,
	}

	if contentType != "" {
		input.ContentType = &contentType
	}

	_, err := handler.s3Client.PutObject(ctx, &input)
	if err != nil {
		return convertErrNotFound(err)
	}
	return nil
}

// Delete implements blob.BlobDeleteHandler.
func (handler *blobHandler) Delete(ctx context.Context, repo string, h v1.Hash) error {
	key := h.String()
	_, err := handler.s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{Bucket: util.PtrTo(bucket), Key: &key})
	if err != nil {
		return convertErrNotFound(err)
	}
	return nil
}

func convertErrNotFound(err error) error {
	var nf *types.NotFound
	var nsk *types.NoSuchKey
	if errors.As(err, &nf) || errors.As(err, &nsk) {
		err = fmt.Errorf("%w: %w", blob.ErrNotFound, err)
	}
	return err
}
