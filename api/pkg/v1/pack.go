package v1

import "github.com/gogf/gf/v2/frame/g"

type PackUpdateImagesReq struct {
	g.Meta `path:"/pack/full-images" method:"post"  tag:"pack" summary:"path update full pkg"`
	Images []string `json:"images"`
}

type PackUpdateImagesRes struct {
}
