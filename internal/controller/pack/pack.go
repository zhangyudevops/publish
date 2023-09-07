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

func (c *cPack) PackUpdateImagesPkg(ctx context.Context, req *apiv1.PackUpdateImagesReq) (res *apiv1.PackUpdateImagesRes, err error) {
	if len(req.Images) == 0 {
		return nil, fmt.Errorf("images is empty")
	}
	// create today's directory
	filePath, _ := g.Config().Get(ctx, "pkg.path")
	CurrentPackPath := filePath.String() + "/" + docker.TodayDate()

	// if the current pkg directory was existed, delete it
	if gfile.Exists(CurrentPackPath) {
		err = gfile.Remove(CurrentPackPath)
		if err != nil {
			return nil, err
		}
	}

	theNewestPath, err := service.File().GetNewestDir(ctx, filePath.String())
	if err != nil {
		return nil, err
	}
	g.Log().Debugf(ctx, "The newest path is: %s", theNewestPath)
	if err = service.Path().CopyFileAndDir(theNewestPath, CurrentPackPath); err != nil {
		_ = service.File().DeleteCurrentDir(ctx, CurrentPackPath)
		return
	}

	dstPath := CurrentPackPath + "/images"
	// uncompressed the images.tar.gz file
	if err = service.File().ExtraTarGzip(ctx, CurrentPackPath+"/images.tar.gz", dstPath); err != nil {
		_ = service.File().DeleteCurrentDir(ctx, CurrentPackPath)
		return
	} else {
		// if uncompressed success, delete the images.tar.gz file
		// delete the images.tar.gz file
		if err = gfile.Remove(CurrentPackPath + "/images.tar.gz"); err != nil {
			return
		}
	}

	// request images list pull from harbor and save it to local
	if err = service.Docker().PullImageAndSaveToLocal(ctx, dstPath, req.Images); err != nil {
		_ = service.File().DeleteCurrentDir(ctx, CurrentPackPath)
		return nil, err
	}

	// compress the today's directory
	if err = service.File().CompressTarGzip(ctx, dstPath, CurrentPackPath+"/"+"images.tar.gz"); err != nil {
		_ = service.File().DeleteCurrentDir(ctx, CurrentPackPath)
		g.Log().Errorf(ctx, "Compress the today's directory failed: %s", err.Error())
		return nil, err
	} else {
		// delete the today's directory
		_ = service.File().DeleteCurrentDir(ctx, dstPath)
		g.Log().Infof(ctx, "Compress the today's directory %s successfully", CurrentPackPath+"/"+"images.tar.gz")
	}

	return &apiv1.PackUpdateImagesRes{}, nil
}
