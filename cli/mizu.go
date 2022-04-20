package main

import (
	"fmt"

	carbon "github.com/golang-module/carbon/v2"
	"github.com/up9inc/mizu/cli/cmd"
	"github.com/up9inc/mizu/cli/cmd/goUtils"
)

func main() {
	fmt.Sprintf("%s", carbon.Now())
	goUtils.HandleExcWrapper(cmd.Execute)
}
