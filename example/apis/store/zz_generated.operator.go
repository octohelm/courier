/*
Package store GENERATED BY gengo:operator
DON'T EDIT THIS FILE
*/
package store

import (
	courier "github.com/octohelm/courier/pkg/courier"
)

func init() {
	R.Register(courier.NewRouter(&GetStoreBlob{}))
}

func (GetStoreBlob) ResponseContent() any {
	return new(string)
}

func (GetStoreBlob) ResponseData() *string {
	return new(string)
}

func init() {
	R.Register(courier.NewRouter(&UploadStoreBlob{}))
}

func (UploadStoreBlob) ResponseContent() any {
	return nil
}

func (UploadStoreBlob) ResponseData() *courier.NoContent {
	return new(courier.NoContent)
}
