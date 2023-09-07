package model

type ImageTag struct {
	Stream string `json:"stream"` // 获取整个tag，需要重新修饰
}
