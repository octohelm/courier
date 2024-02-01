/*
Package example GENERATED BY gengo:client 
DON'T EDIT THIS FILE
*/
package example

import (
	context "context"
	io "io"

	github_com_octohelm_courier_pkg_courier "github.com/octohelm/courier/pkg/courier"
	github_com_octohelm_courier_pkg_courierhttp "github.com/octohelm/courier/pkg/courierhttp"
)

type UploadBlob struct {
	github_com_octohelm_courier_pkg_courierhttp.MethodPost `path:"/api/example/v0/blobs"`

	io.ReadCloser `in:"body" mime:"octet-stream"`
}

func (r *UploadBlob) Do(ctx context.Context, metas ...github_com_octohelm_courier_pkg_courier.Metadata) github_com_octohelm_courier_pkg_courier.Result {
	return github_com_octohelm_courier_pkg_courier.ClientFromContext(ctx, "example").Do(ctx, r, metas...)
}

func (r *UploadBlob) Invoke(ctx context.Context, metas ...github_com_octohelm_courier_pkg_courier.Metadata) (github_com_octohelm_courier_pkg_courier.Metadata, error) {
	return r.Do(ctx, metas...).Into(nil)
}

type GetStoreBlob struct {
	github_com_octohelm_courier_pkg_courierhttp.MethodGet `path:"/api/example/v0/store/:scope/blobs/:digest"`

	Scope string `name:"scope" in:"path"`

	Digest string `name:"digest" in:"path"`
}

func (r *GetStoreBlob) Do(ctx context.Context, metas ...github_com_octohelm_courier_pkg_courier.Metadata) github_com_octohelm_courier_pkg_courier.Result {
	return github_com_octohelm_courier_pkg_courier.ClientFromContext(ctx, "example").Do(ctx, r, metas...)
}

func (r *GetStoreBlob) Invoke(ctx context.Context, metas ...github_com_octohelm_courier_pkg_courier.Metadata) (*GetStoreBlobResponse, github_com_octohelm_courier_pkg_courier.Metadata, error) {
	var resp GetStoreBlobResponse
	meta, err := r.Do(ctx, metas...).Into(&resp)
	return &resp, meta, err
}

type GetFile struct {
	github_com_octohelm_courier_pkg_courierhttp.MethodPost `path:"/api/example/v0/blobs/:path"`

	Path string `name:"path" in:"path"`
}

func (r *GetFile) Do(ctx context.Context, metas ...github_com_octohelm_courier_pkg_courier.Metadata) github_com_octohelm_courier_pkg_courier.Result {
	return github_com_octohelm_courier_pkg_courier.ClientFromContext(ctx, "example").Do(ctx, r, metas...)
}

func (r *GetFile) Invoke(ctx context.Context, metas ...github_com_octohelm_courier_pkg_courier.Metadata) (*GetFileResponse, github_com_octohelm_courier_pkg_courier.Metadata, error) {
	var resp GetFileResponse
	meta, err := r.Do(ctx, metas...).Into(&resp)
	return &resp, meta, err
}

type ListOrg struct {
	github_com_octohelm_courier_pkg_courierhttp.MethodGet `path:"/api/example/v0/orgs"`
}

func (r *ListOrg) Do(ctx context.Context, metas ...github_com_octohelm_courier_pkg_courier.Metadata) github_com_octohelm_courier_pkg_courier.Result {
	return github_com_octohelm_courier_pkg_courier.ClientFromContext(ctx, "example").Do(ctx, r, metas...)
}

func (r *ListOrg) Invoke(ctx context.Context, metas ...github_com_octohelm_courier_pkg_courier.Metadata) (*ListOrgResponse, github_com_octohelm_courier_pkg_courier.Metadata, error) {
	var resp ListOrgResponse
	meta, err := r.Do(ctx, metas...).Into(&resp)
	return &resp, meta, err
}

type CreateOrg struct {
	github_com_octohelm_courier_pkg_courierhttp.MethodPost `path:"/api/example/v0/orgs"`

	OrgInfo `in:"body" mime:"json"`
}

func (r *CreateOrg) Do(ctx context.Context, metas ...github_com_octohelm_courier_pkg_courier.Metadata) github_com_octohelm_courier_pkg_courier.Result {
	return github_com_octohelm_courier_pkg_courier.ClientFromContext(ctx, "example").Do(ctx, r, metas...)
}

func (r *CreateOrg) Invoke(ctx context.Context, metas ...github_com_octohelm_courier_pkg_courier.Metadata) (github_com_octohelm_courier_pkg_courier.Metadata, error) {
	return r.Do(ctx, metas...).Into(nil)
}

type UploadStoreBlob struct {
	github_com_octohelm_courier_pkg_courierhttp.MethodPost `path:"/api/example/v0/store/:scope/blobs/uploads"`

	Scope string `name:"scope" in:"path"`

	io.ReadCloser `in:"body" mime:"octet-stream"`
}

func (r *UploadStoreBlob) Do(ctx context.Context, metas ...github_com_octohelm_courier_pkg_courier.Metadata) github_com_octohelm_courier_pkg_courier.Result {
	return github_com_octohelm_courier_pkg_courier.ClientFromContext(ctx, "example").Do(ctx, r, metas...)
}

func (r *UploadStoreBlob) Invoke(ctx context.Context, metas ...github_com_octohelm_courier_pkg_courier.Metadata) (github_com_octohelm_courier_pkg_courier.Metadata, error) {
	return r.Do(ctx, metas...).Into(nil)
}

type Cookie struct {
	github_com_octohelm_courier_pkg_courierhttp.MethodPost `path:"/api/example/v0/cookie-ping-pong"`

	Token string `name:"token,omitempty" in:"cookie"`
}

func (r *Cookie) Do(ctx context.Context, metas ...github_com_octohelm_courier_pkg_courier.Metadata) github_com_octohelm_courier_pkg_courier.Result {
	return github_com_octohelm_courier_pkg_courier.ClientFromContext(ctx, "example").Do(ctx, r, metas...)
}

func (r *Cookie) Invoke(ctx context.Context, metas ...github_com_octohelm_courier_pkg_courier.Metadata) (*CookieResponse, github_com_octohelm_courier_pkg_courier.Metadata, error) {
	var resp CookieResponse
	meta, err := r.Do(ctx, metas...).Into(&resp)
	return &resp, meta, err
}

type ListOrgOld struct {
	github_com_octohelm_courier_pkg_courierhttp.MethodGet `path:"/api/example/v0/org"`
}

func (r *ListOrgOld) Do(ctx context.Context, metas ...github_com_octohelm_courier_pkg_courier.Metadata) github_com_octohelm_courier_pkg_courier.Result {
	return github_com_octohelm_courier_pkg_courier.ClientFromContext(ctx, "example").Do(ctx, r, metas...)
}

func (r *ListOrgOld) Invoke(ctx context.Context, metas ...github_com_octohelm_courier_pkg_courier.Metadata) (*ListOrgOldResponse, github_com_octohelm_courier_pkg_courier.Metadata, error) {
	var resp ListOrgOldResponse
	meta, err := r.Do(ctx, metas...).Into(&resp)
	return &resp, meta, err
}

type DeleteOrg struct {
	github_com_octohelm_courier_pkg_courierhttp.MethodDelete `path:"/api/example/v0/orgs/:orgName"`

	OrgName string `name:"orgName" in:"path"`
}

func (r *DeleteOrg) Do(ctx context.Context, metas ...github_com_octohelm_courier_pkg_courier.Metadata) github_com_octohelm_courier_pkg_courier.Result {
	return github_com_octohelm_courier_pkg_courier.ClientFromContext(ctx, "example").Do(ctx, r, metas...)
}

func (r *DeleteOrg) Invoke(ctx context.Context, metas ...github_com_octohelm_courier_pkg_courier.Metadata) (github_com_octohelm_courier_pkg_courier.Metadata, error) {
	return r.Do(ctx, metas...).Into(nil)
}

type GetOrg struct {
	github_com_octohelm_courier_pkg_courierhttp.MethodGet `path:"/api/example/v0/orgs/:orgName"`

	OrgName string `name:"orgName" in:"path"`
}

func (r *GetOrg) Do(ctx context.Context, metas ...github_com_octohelm_courier_pkg_courier.Metadata) github_com_octohelm_courier_pkg_courier.Result {
	return github_com_octohelm_courier_pkg_courier.ClientFromContext(ctx, "example").Do(ctx, r, metas...)
}

func (r *GetOrg) Invoke(ctx context.Context, metas ...github_com_octohelm_courier_pkg_courier.Metadata) (*GetOrgResponse, github_com_octohelm_courier_pkg_courier.Metadata, error) {
	var resp GetOrgResponse
	meta, err := r.Do(ctx, metas...).Into(&resp)
	return &resp, meta, err
}

type OrgDataList struct {
	Data []OrgInfo `json:"data" name:"data" `

	Extra []any `json:"extra" name:"extra" `

	Total int `json:"total" name:"total" `
}

type OrgInfo struct {
	// 组织名称
	Name string `json:"name" name:"name"  validate:"@string[0,5]"`
	// 组织类型
	Type OrgType `json:"type,omitempty" name:"type,omitempty" `
}

type OrgType string

const (
	ORG_TYPE__Gov     OrgType = "GOV"
	ORG_TYPE__Company OrgType = "COMPANY"
)

type ListOrgResponse = OrgDataList

type Time string

type GetOrgResponse = OrgDetail

type GetStoreBlobResponse = string

type GetFileResponse = string

type CookieResponse = any

type ListOrgOldResponse = any

type OrgDetail struct {
	CreatedAt *Time `json:"createdAt,omitempty" name:"createdAt,omitempty" `
	// 组织名称
	Name string `json:"name" name:"name"  validate:"@string[0,5]"`
	// 组织类型
	Type OrgType `json:"type,omitempty" name:"type,omitempty" `
}
