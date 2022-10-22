package main

import (
	"github.com/kubeshark/kubeshark/cli/cmd"
	"github.com/kubeshark/kubeshark/cli/cmd/goUtils"
)

func main() {
	goUtils.HandleExcWrapper(cmd.Execute)
}
