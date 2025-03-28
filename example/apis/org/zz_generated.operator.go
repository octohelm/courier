/*
Package org GENERATED BY gengo:operator
DON'T EDIT THIS FILE
*/
package org

import (
	courier "github.com/octohelm/courier/pkg/courier"
	statuserror "github.com/octohelm/courier/pkg/statuserror"
)

func init() {
	R.Register(courier.NewRouter(&Cookie{}))
}

func (Cookie) ResponseContent() any {
	return new(any)
}

func (Cookie) ResponseData() *any {
	return new(any)
}

func init() {
	R.Register(courier.NewRouter(&CreateOrg{}))
}

func (CreateOrg) ResponseContent() any {
	return nil
}

func (CreateOrg) ResponseData() *courier.NoContent {
	return new(courier.NoContent)
}

func init() {
	R.Register(courier.NewRouter(&DeleteOrg{}))
}

func (DeleteOrg) ResponseContent() any {
	return nil
}

func (DeleteOrg) ResponseData() *courier.NoContent {
	return new(courier.NoContent)
}

func init() {
	R.Register(courier.NewRouter(&GetOrg{}))
}

func (GetOrg) ResponseContent() any {
	return new(Detail)
}

func (GetOrg) ResponseData() *Detail {
	return new(Detail)
}

func (GetOrg) ResponseErrors() []error {
	return []error{
		&(statuserror.Descriptor{
			Code:    "org.ErrNotFound",
			Message: "{OrgName}: 组织不存在",
			Status:  404,
		}),
	}
}

func init() {
	R.Register(courier.NewRouter(&ListOrg{}))
}

func (ListOrg) ResponseStatusCode() int {
	return 200
}

func (ListOrg) ResponseContent() any {
	return new(DataList[Info])
}

func (ListOrg) ResponseData() *DataList[Info] {
	return new(DataList[Info])
}

func init() {
	R.Register(courier.NewRouter(&ListOrgOld{}))
}

func (ListOrgOld) ResponseStatusCode() int {
	return 302
}

func (ListOrgOld) ResponseContent() any {
	return new(any)
}

func (ListOrgOld) ResponseData() *any {
	return new(any)
}
