package main

import (
	"github.com/up9inc/kubeshark/cli/cmd"
	"github.com/up9inc/kubeshark/cli/cmd/goUtils"
)

func main() {
	goUtils.HandleExcWrapper(cmd.Execute)
}
