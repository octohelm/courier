package mem

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"sync"

	storeservice "github.com/octohelm/courier/internal/example/domain/store/service"
	storev1 "github.com/octohelm/courier/internal/example/pkg/apis/store/v1"
)

const (
	blobMediaType     storev1.MediaType = "application/octet-stream"
	manifestMediaType storev1.MediaType = "application/vnd.example.manifest.v1+json"
)

var _ storeservice.Service = (*Service)(nil)

// Service 提供制品仓库域内存版实现。
type Service struct {
	mu        sync.RWMutex
	blobs     map[storev1.Namespace]map[storev1.Digest][]byte
	manifests map[storev1.Namespace]map[storev1.Digest]*storev1.Manifest
}

// New 创建制品仓库域内存服务。
func New() *Service {
	return &Service{
		blobs:     map[storev1.Namespace]map[storev1.Digest][]byte{},
		manifests: map[storev1.Namespace]map[storev1.Digest]*storev1.Manifest{},
	}
}

func (s *Service) UploadBlob(_ context.Context, namespace storev1.Namespace, body io.ReadCloser) (*storev1.Descriptor, error) {
	if body == nil {
		return nil, fmt.Errorf("blob body is nil")
	}
	defer body.Close()

	data, err := io.ReadAll(body)
	if err != nil {
		return nil, err
	}

	digest := digestFor(data)

	s.mu.Lock()
	defer s.mu.Unlock()

	ns := ensureBlobNamespace(s.blobs, namespace)
	ns[digest] = bytes.Clone(data)

	return &storev1.Descriptor{
		MediaType: blobMediaType,
		Digest:    digest,
		Size:      int64(len(data)),
	}, nil
}

func (s *Service) GetBlob(_ context.Context, namespace storev1.Namespace, digest storev1.Digest) (io.ReadCloser, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, err := s.getBlobLocked(namespace, digest)
	if err != nil {
		return nil, err
	}

	return io.NopCloser(bytes.NewReader(bytes.Clone(data))), nil
}

func (s *Service) DeleteBlob(_ context.Context, namespace storev1.Namespace, digest storev1.Digest) (*storev1.Descriptor, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	ns, ok := s.blobs[namespace]
	if !ok {
		return nil, &storev1.ErrBlobNotFound{Namespace: namespace, Digest: digest}
	}

	data, ok := ns[digest]
	if !ok {
		return nil, &storev1.ErrBlobNotFound{Namespace: namespace, Digest: digest}
	}

	delete(ns, digest)

	return &storev1.Descriptor{
		MediaType: blobMediaType,
		Digest:    digest,
		Size:      int64(len(data)),
	}, nil
}

func (s *Service) PutManifest(_ context.Context, namespace storev1.Namespace, digest storev1.Digest, manifest *storev1.Manifest) (*storev1.Descriptor, error) {
	if manifest == nil {
		return nil, &storev1.ErrManifestInvalid{Reason: "manifest 不能为空"}
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.validateManifestLocked(namespace, manifest); err != nil {
		return nil, err
	}

	payload, err := json.Marshal(manifest)
	if err != nil {
		return nil, err
	}

	ns := ensureManifestNamespace(s.manifests, namespace)
	ns[digest] = cloneManifest(manifest)

	return &storev1.Descriptor{
		MediaType: manifestMediaType,
		Digest:    digest,
		Size:      int64(len(payload)),
	}, nil
}

func (s *Service) GetManifest(_ context.Context, namespace storev1.Namespace, digest storev1.Digest) (*storev1.Manifest, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ns, ok := s.manifests[namespace]
	if !ok {
		return nil, &storev1.ErrManifestNotFound{Namespace: namespace, Digest: digest}
	}

	manifest, ok := ns[digest]
	if !ok {
		return nil, &storev1.ErrManifestNotFound{Namespace: namespace, Digest: digest}
	}

	return cloneManifest(manifest), nil
}

func (s *Service) DeleteManifest(_ context.Context, namespace storev1.Namespace, digest storev1.Digest) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	ns, ok := s.manifests[namespace]
	if !ok {
		return &storev1.ErrManifestNotFound{Namespace: namespace, Digest: digest}
	}

	if _, ok := ns[digest]; !ok {
		return &storev1.ErrManifestNotFound{Namespace: namespace, Digest: digest}
	}

	delete(ns, digest)
	return nil
}

func (s *Service) getBlobLocked(namespace storev1.Namespace, digest storev1.Digest) ([]byte, error) {
	ns, ok := s.blobs[namespace]
	if !ok {
		return nil, &storev1.ErrBlobNotFound{Namespace: namespace, Digest: digest}
	}

	data, ok := ns[digest]
	if !ok {
		return nil, &storev1.ErrBlobNotFound{Namespace: namespace, Digest: digest}
	}

	return data, nil
}

func (s *Service) validateManifestLocked(namespace storev1.Namespace, manifest *storev1.Manifest) error {
	if manifest.Config.Digest == "" {
		return &storev1.ErrManifestInvalid{Reason: "config.digest 不能为空"}
	}

	if _, err := s.getBlobLocked(namespace, manifest.Config.Digest); err != nil {
		return err
	}

	for i, asset := range manifest.Assets {
		if asset.Digest == "" {
			return &storev1.ErrManifestInvalid{Reason: fmt.Sprintf("assets[%d].digest 不能为空", i)}
		}
		if _, err := s.getBlobLocked(namespace, asset.Digest); err != nil {
			return err
		}
	}

	return nil
}

func ensureBlobNamespace(m map[storev1.Namespace]map[storev1.Digest][]byte, namespace storev1.Namespace) map[storev1.Digest][]byte {
	if _, ok := m[namespace]; !ok {
		m[namespace] = map[storev1.Digest][]byte{}
	}
	return m[namespace]
}

func ensureManifestNamespace(m map[storev1.Namespace]map[storev1.Digest]*storev1.Manifest, namespace storev1.Namespace) map[storev1.Digest]*storev1.Manifest {
	if _, ok := m[namespace]; !ok {
		m[namespace] = map[storev1.Digest]*storev1.Manifest{}
	}
	return m[namespace]
}

func cloneManifest(v *storev1.Manifest) *storev1.Manifest {
	if v == nil {
		return nil
	}

	c := &storev1.Manifest{
		Config: v.Config,
		Assets: make([]storev1.Descriptor, len(v.Assets)),
	}
	copy(c.Assets, v.Assets)
	return c
}

func digestFor(data []byte) storev1.Digest {
	sum := sha256.Sum256(data)
	return storev1.Digest("sha256:" + hex.EncodeToString(sum[:]))
}
