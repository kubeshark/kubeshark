package main

import (
	"github.com/kubeshark/kubeshark/cmd"
	"github.com/kubeshark/kubeshark/cmd/goUtils"
)

func main() {
	goUtils.HandleExcWrapper(cmd.Execute)
}
