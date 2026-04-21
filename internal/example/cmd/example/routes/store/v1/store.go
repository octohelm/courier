package v1

import (
	"context"

	storeservice "github.com/octohelm/courier/internal/example/domain/store/service"
	endpointstorev1 "github.com/octohelm/courier/internal/example/pkg/endpoints/store/v1"
)

// +gengo:injectable
type UploadBlob struct {
	endpointstorev1.UploadBlob

	svc storeservice.Service `inject:""`
}

func (r *UploadBlob) Output(ctx context.Context) (any, error) {
	return r.svc.UploadBlob(ctx, r.Namespace, r.Body)
}

// +gengo:injectable
type GetBlob struct {
	endpointstorev1.GetBlob

	svc storeservice.Service `inject:""`
}

func (r *GetBlob) Output(ctx context.Context) (any, error) {
	body, err := r.svc.GetBlob(ctx, r.Namespace, r.Digest)
	if err != nil {
		return nil, err
	}
	return body, nil
}

// +gengo:injectable
type DeleteBlob struct {
	endpointstorev1.DeleteBlob

	svc storeservice.Service `inject:""`
}

func (r *DeleteBlob) Output(ctx context.Context) (any, error) {
	return r.svc.DeleteBlob(ctx, r.Namespace, r.Digest)
}

// +gengo:injectable
type PutManifest struct {
	endpointstorev1.PutManifest

	svc storeservice.Service `inject:""`
}

func (r *PutManifest) Output(ctx context.Context) (any, error) {
	return r.svc.PutManifest(ctx, r.Namespace, r.Digest, r.Body)
}

// +gengo:injectable
type GetManifest struct {
	endpointstorev1.GetManifest

	svc storeservice.Service `inject:""`
}

func (r *GetManifest) Output(ctx context.Context) (any, error) {
	return r.svc.GetManifest(ctx, r.Namespace, r.Digest)
}

// +gengo:injectable
type DeleteManifest struct {
	endpointstorev1.DeleteManifest

	svc storeservice.Service `inject:""`
}

func (r *DeleteManifest) Output(ctx context.Context) (any, error) {
	return nil, r.svc.DeleteManifest(ctx, r.Namespace, r.Digest)
}
