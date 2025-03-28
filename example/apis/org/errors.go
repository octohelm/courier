package org

import (
	"fmt"

	"github.com/octohelm/courier/example/pkg/statuserr"
)

type ErrNotFound struct {
	statuserr.NotFound

	OrgName Name
}

func (e ErrNotFound) Error() string {
	return fmt.Sprintf("%s: 组织不存在", e.OrgName)
}
