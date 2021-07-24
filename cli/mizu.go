package main

import (
	"github.com/up9inc/mizu/cli/cmd"
	"github.com/up9inc/mizu/cli/mizu"
)

func main() {
	mizu.InitLogger()
	cmd.Execute()
}
