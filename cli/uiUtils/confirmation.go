package uiUtils

import (
	"bufio"
	"fmt"
	"github.com/up9inc/mizu/cli/mizu"
	"log"
	"os"
	"strings"
)

func AskForConfirmation(s string) bool {
	reader := bufio.NewReader(os.Stdin)

	fmt.Printf(mizu.Magenta, s)

	response, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}
	response = strings.ToLower(strings.TrimSpace(response))
	if response == "" || response == "y" || response == "yes" {
		return true
	}
	return false
}
