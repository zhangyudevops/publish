package pack

import (
	"context"
	"fmt"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gfile"
	apiv1 "publish/api/pkg/v1"
	"publish/internal/service"
	"publish/utility/docker"
)

type cPack struct{}

func Pack() *cPack {
	return &cPack{}
}

// copyOldDirectoryToNew copy the newest directory to package directory
func (c *cPack) copyOldDirectoryToNew(ctx context.Context, src, dst string) (err error) {
	// if the current pkg directory was existed, delete it
	if gfile.Exists(dst) {
		if err = gfile.Remove(dst); err != nil {
			return
		}
	}

	theNewestPath, err := service.File().GetNewestDir(ctx, src)
	if err != nil {
		return
	}
	g.Log().Debugf(ctx, "The newest path is: %s", theNewestPath)
	if err = service.Path().CopyFileAndDir(theNewestPath, dst); err != nil {
		_ = service.File().DeleteCurrentDir(ctx, dst)
		g.Log().Errorf(ctx, "复制目录失败: %s", err.Error())

		return
	}

	return
}

func (c *cPack) PackUpdateImagesPkg(ctx context.Context, req *apiv1.PackUpdateImagesReq) (res *apiv1.PackUpdateImagesRes, err error) {
	if len(req.Images) == 0 {
		return nil, fmt.Errorf("images is empty")
	}
	// create today's directory
	filePath, _ := g.Config().Get(ctx, "pkg.path")
	CurrentPackPath := filePath.String() + "/" + docker.TodayDate()

	if err = c.copyOldDirectoryToNew(ctx, filePath.String(), CurrentPackPath); err != nil {
		return nil, err
	}

	dstPath := CurrentPackPath + "/images"
	gzipImageFilePath := CurrentPackPath + "/images.tar.gz"
	// delete old images.tar.gz file
	if gfile.Exists(gzipImageFilePath) {
		if err = gfile.Remove(gzipImageFilePath); err != nil {
			return nil, err
		}
	}

	// request images list pull from harbor and save it to local
	if err = service.Docker().PullImageAndSaveToLocal(ctx, dstPath, req.Images); err != nil {
		_ = service.File().DeleteCurrentDir(ctx, CurrentPackPath)
		return nil, err
	}

	// compress the today's directory
	if err = service.File().CompressTarGzip(ctx, dstPath, gzipImageFilePath); err != nil {
		_ = service.File().DeleteCurrentDir(ctx, CurrentPackPath)
		g.Log().Errorf(ctx, "Compress the images folder failed: %s", err.Error())
		return nil, err
	} else {
		// delete the today's directory
		_ = service.File().DeleteCurrentDir(ctx, dstPath)
		g.Log().Info(ctx, "Compress the images folder successfully")
	}

	return &apiv1.PackUpdateImagesRes{}, nil
}
