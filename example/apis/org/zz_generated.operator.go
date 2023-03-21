/*
Package org GENERATED BY gengo:operator 
DON'T EDIT THIS FILE
*/
package org

import (
	github_com_octohelm_courier_pkg_courier "github.com/octohelm/courier/pkg/courier"
	github_com_octohelm_courier_pkg_statuserror "github.com/octohelm/courier/pkg/statuserror"
)

func init() {
	R.Register(github_com_octohelm_courier_pkg_courier.NewRouter(&Cookie{}))
}

func init() {
	R.Register(github_com_octohelm_courier_pkg_courier.NewRouter(&CreateOrg{}))
}

func (*CreateOrg) ResponseContent() any {
	return nil
}

func init() {
	R.Register(github_com_octohelm_courier_pkg_courier.NewRouter(&DeleteOrg{}))
}

func (*DeleteOrg) ResponseContent() any {
	return nil
}

func init() {
	R.Register(github_com_octohelm_courier_pkg_courier.NewRouter(&GetOrg{}))
}

func (*GetOrg) ResponseContent() any {
	return nil
}

func (*GetOrg) ResponseErrors() []error {
	return []error{
		&(github_com_octohelm_courier_pkg_statuserror.StatusErr{
			Code: 404,
			Key:  "NotFound",
			Msg:  "NotFound",
		}),
	}
}

func init() {
	R.Register(github_com_octohelm_courier_pkg_courier.NewRouter(&ListOrg{}))
}

func (*ListOrg) ResponseContent() any {
	return new(DataList)
}

func init() {
	R.Register(github_com_octohelm_courier_pkg_courier.NewRouter(&ListOrgOld{}))
}
