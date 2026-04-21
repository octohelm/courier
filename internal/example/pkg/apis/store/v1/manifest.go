package v1

// Manifest 表示镜像清单。
type Manifest struct {
	// 配置对象描述
	Config Descriptor `json:"config"`
	// 资源对象描述列表
	Assets []Descriptor `json:"assets"`
}

// Descriptor 表示对象描述信息。
type Descriptor struct {
	// 媒体类型
	MediaType MediaType `json:"mediaType"`
	// 内容摘要
	Digest Digest `json:"digest"`
	// 内容大小
	Size int64 `json:"size"`
}

// MediaType 表示对象媒体类型。
type MediaType string

// Digest 表示内容摘要。
type Digest string
