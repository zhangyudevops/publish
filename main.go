package main

import (
	_ "publish/internal/packed"

	"github.com/gogf/gf/v2/os/gctx"

	"publish/internal/cmd"
)

func main() {
	cmd.Main.Run(gctx.GetInitCtx())
}
