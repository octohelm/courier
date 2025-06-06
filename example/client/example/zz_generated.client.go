/*
Package example GENERATED BY gengo:client
DON'T EDIT THIS FILE
*/
package example

import (
	io "io"
	time "time"

	org "github.com/octohelm/courier/example/apis/org"
	domainorg "github.com/octohelm/courier/example/pkg/domain/org"
	filter "github.com/octohelm/courier/example/pkg/filter"
	courier "github.com/octohelm/courier/pkg/courier"
	courierhttp "github.com/octohelm/courier/pkg/courierhttp"
)

type Cookie struct {
	courierhttp.MethodPost `path:"/api/example/v0/cookie-ping-pong"`

	CookieParameters
}

type CookieParameters struct {
	Token string `name:"token,omitzero" in:"cookie"`
}

func (Cookie) ResponseData() *courier.NoContent {
	return new(courier.NoContent)
}

type CreateOrg struct {
	courierhttp.MethodPost `path:"/api/example/v0/orgs"`

	CreateOrgParameters
}

type CreateOrgParameters struct {
	RequestBody OrgInfo `in:"body" mime:"application/json"`
}

func (CreateOrg) ResponseData() *courier.NoContent {
	return new(courier.NoContent)
}

type ListOrg struct {
	courierhttp.MethodGet `path:"/api/example/v0/orgs"`

	ListOrgParameters
}

type ListOrgParameters struct {
	OrgID *OrgIDAsFilter `name:"org~id,omitzero" in:"query"`
}

func (ListOrg) ResponseData() *ListOrgResponse {
	return new(ListOrgResponse)
}

type DeleteOrg struct {
	courierhttp.MethodDelete `path:"/api/example/v0/orgs/{orgName}"`

	DeleteOrgParameters
}

type DeleteOrgParameters struct {
	OrgName string `name:"orgName" in:"path"`
}

func (DeleteOrg) ResponseData() *courier.NoContent {
	return new(courier.NoContent)
}

type GetOrg struct {
	courierhttp.MethodGet `path:"/api/example/v0/orgs/{orgName}"`

	GetOrgParameters
}

type GetOrgParameters struct {
	OrgName OrgName `name:"orgName" in:"path"`
}

func (GetOrg) ResponseData() *GetOrgResponse {
	return new(GetOrgResponse)
}

type ListOrgOld struct {
	courierhttp.MethodGet `path:"/api/example/v0/org"`

	ListOrgOldParameters
}

type ListOrgOldParameters struct{}

func (ListOrgOld) ResponseData() *courier.NoContent {
	return new(courier.NoContent)
}

type GetStoreBlob struct {
	courierhttp.MethodGet `path:"/api/example/v0/store/{scope}/blobs/{digest}"`

	GetStoreBlobParameters
}

type GetStoreBlobParameters struct {
	Scope string `name:"scope" in:"path"`

	Digest string `name:"digest" in:"path"`
}

func (GetStoreBlob) ResponseData() *GetStoreBlobResponse {
	return new(GetStoreBlobResponse)
}

type UploadStoreBlob struct {
	courierhttp.MethodPost `path:"/api/example/v0/store/{scope}/blobs/uploads"`

	UploadStoreBlobParameters
}

type UploadStoreBlobParameters struct {
	Scope string `name:"scope" in:"path"`

	RequestBody IoReadCloser `in:"body" mime:"application/octet-stream"`
}

func (UploadStoreBlob) ResponseData() *courier.NoContent {
	return new(courier.NoContent)
}

type GetFile struct {
	courierhttp.MethodPost `path:"/api/example/v0/blobs/{path}"`

	GetFileParameters
}

type GetFileParameters struct {
	Path string `name:"path" in:"path"`
}

func (GetFile) ResponseData() *GetFileResponse {
	return new(GetFileResponse)
}

type UploadBlob struct {
	courierhttp.MethodPost `path:"/api/example/v0/blobs"`

	UploadBlobParameters
}

type UploadBlobParameters struct {
	RequestBody IoReadCloser `in:"body" mime:"application/octet-stream"`
}

func (UploadBlob) ResponseData() *courier.NoContent {
	return new(courier.NoContent)
}

type GetFileResponse = string

type (
	GetOrgResponse       = org.Detail
	GetStoreBlobResponse = string
)

type (
	IoReadCloser      = io.ReadCloser
	ListOrgResponse   = org.DataList[org.Info]
	OrgDetail         = org.Detail
	OrgIDAsFilter     = filter.Filter[org.ID]
	OrgInfo           = org.Info
	OrgInfoAsDataList = org.DataList[org.Info]
	OrgName           = org.Name
	OrgType           = domainorg.Type
	Time              = time.Time
)
