package main

import (
	"github.com/up9inc/mizu/cli/cmd"
	"github.com/up9inc/mizu/shared/goUtils"
)

func main() {
	goUtils.HandleExcWrapper(cmd.Execute)
}
