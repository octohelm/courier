package mem

import (
	"context"
	"slices"
	"sync"
	"time"

	orgservice "github.com/octohelm/courier/internal/example/domain/org/service"
	metav1 "github.com/octohelm/courier/internal/example/pkg/apis/meta/v1"
	orgv1 "github.com/octohelm/courier/internal/example/pkg/apis/org/v1"
)

var _ orgservice.Service = (*Service)(nil)

// Service 提供组织域内存版实现。
type Service struct {
	mu       sync.RWMutex
	nextID   orgv1.OrgID
	orgs     map[orgv1.OrgID]*orgv1.Org
	idByName map[orgv1.OrgName]orgv1.OrgID
}

// New 创建组织域内存服务。
func New() *Service {
	return &Service{
		nextID:   1,
		orgs:     map[orgv1.OrgID]*orgv1.Org{},
		idByName: map[orgv1.OrgName]orgv1.OrgID{},
	}
}

func (s *Service) Create(_ context.Context, req *orgv1.OrgForCreateRequest) (*orgv1.Org, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if req == nil {
		req = &orgv1.OrgForCreateRequest{}
	}

	if _, ok := s.idByName[req.Spec.Name]; ok {
		return nil, &orgv1.ErrOrgNameConflict{OrgName: req.Spec.Name}
	}

	now := time.Now()
	org := &orgv1.Org{
		ID:        s.nextID,
		CreatedAt: &now,
		UpdatedAt: &now,
		Spec:      req.Spec,
	}

	s.orgs[org.ID] = cloneOrg(org)
	s.idByName[org.Spec.Name] = org.ID
	s.nextID++

	return cloneOrg(org), nil
}

func (s *Service) Update(_ context.Context, orgID orgv1.OrgID, req *orgv1.OrgForUpdateRequest) (*orgv1.Org, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	org, ok := s.orgs[orgID]
	if !ok {
		return nil, &orgv1.ErrOrgNotFound{OrgID: &orgID}
	}

	if req == nil {
		req = &orgv1.OrgForUpdateRequest{}
	}

	if currentID, exists := s.idByName[req.Spec.Name]; exists && currentID != orgID {
		return nil, &orgv1.ErrOrgNameConflict{OrgName: req.Spec.Name}
	}

	delete(s.idByName, org.Spec.Name)
	org.Spec = req.Spec
	now := time.Now()
	org.UpdatedAt = &now
	s.idByName[org.Spec.Name] = org.ID

	return cloneOrg(org), nil
}

func (s *Service) Delete(_ context.Context, orgID orgv1.OrgID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	org, ok := s.orgs[orgID]
	if !ok {
		return &orgv1.ErrOrgNotFound{OrgID: &orgID}
	}

	delete(s.idByName, org.Spec.Name)
	delete(s.orgs, orgID)
	return nil
}

func (s *Service) Get(_ context.Context, orgID orgv1.OrgID) (*orgv1.Org, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	org, ok := s.orgs[orgID]
	if !ok {
		return nil, &orgv1.ErrOrgNotFound{OrgID: &orgID}
	}

	return cloneOrg(org), nil
}

func (s *Service) List(_ context.Context, req *orgv1.OrgForListRequest, pager *metav1.Pager) (*orgv1.OrgList, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := make([]*orgv1.Org, 0, len(s.orgs))

	for _, org := range s.orgs {
		if !matchOrg(org, req) {
			continue
		}
		items = append(items, cloneOrg(org))
	}

	slices.SortFunc(items, func(a, b *orgv1.Org) int {
		switch {
		case a.ID < b.ID:
			return -1
		case a.ID > b.ID:
			return 1
		default:
			return 0
		}
	})

	total := int64(len(items))
	offset, limit := normalizePager(pager, total)
	items = items[offset:limit]

	return &orgv1.OrgList{
		Items: items,
		Total: total,
	}, nil
}

func matchOrg(org *orgv1.Org, req *orgv1.OrgForListRequest) bool {
	if req == nil {
		return true
	}
	if req.OrgID != nil && org.ID != *req.OrgID {
		return false
	}
	if req.OrgName != nil && org.Spec.Name != *req.OrgName {
		return false
	}
	if req.OrgType != nil && org.Spec.Type != *req.OrgType {
		return false
	}
	return true
}

func normalizePager(pager *metav1.Pager, total int64) (int64, int64) {
	if pager == nil {
		return 0, total
	}

	offset := pager.Offset
	if offset < 0 {
		offset = 0
	}
	if offset > total {
		offset = total
	}

	limit := pager.Limit
	if limit <= 0 {
		limit = total - offset
	}

	end := offset + limit
	if end > total {
		end = total
	}
	return offset, end
}

func cloneOrg(v *orgv1.Org) *orgv1.Org {
	if v == nil {
		return nil
	}

	c := *v
	if v.CreatedAt != nil {
		t := *v.CreatedAt
		c.CreatedAt = &t
	}
	if v.UpdatedAt != nil {
		t := *v.UpdatedAt
		c.UpdatedAt = &t
	}
	return &c
}
