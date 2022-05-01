package uiUtils

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/up9inc/mizu/logger"
)

func AskForConfirmation(s string) bool {
	reader := bufio.NewReader(os.Stdin)

	fmt.Printf(Magenta, s)

	response, err := reader.ReadString('\n')
	if err != nil {
		logger.Log.Fatalf("Error while reading confirmation string, err: %v", err)
	}
	response = strings.ToLower(strings.TrimSpace(response))
	if response == "" || response == "y" || response == "yes" {
		return true
	}
	return false
}
