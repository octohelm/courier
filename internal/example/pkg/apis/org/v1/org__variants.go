package v1

// OrgForCreateRequest 表示创建组织请求体。
type OrgForCreateRequest struct {
	// 待创建的组织规格
	Spec OrgSpec `json:"spec"`
}

// OrgForUpdateRequest 表示更新组织请求体。
type OrgForUpdateRequest struct {
	// 待更新的组织规格
	Spec OrgSpec `json:"spec"`
}

// OrgForListRequest 表示组织列表查询条件。
type OrgForListRequest struct {
	// 按组织 ID 过滤
	OrgID *OrgID `json:"orgID,omitempty"`
	// 按组织名称过滤
	OrgName *OrgName `json:"orgName,omitempty"`
	// 按组织类型过滤
	OrgType *OrgType `json:"orgType,omitempty"`
}
