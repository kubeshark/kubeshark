package errormessage

import (
	"errors"
	"fmt"
	"github.com/up9inc/mizu/cli/config"
	regexpsyntax "regexp/syntax"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
)

// formatError wraps error with a detailed message that is meant for the user.
// While the errors are meant to be displayed, they are not meant to be exported as classes outsite of CLI.
func FormatError(err error) error {
	var errorNew error
	if k8serrors.IsForbidden(err) {
		errorNew = fmt.Errorf("insufficient permissions: %w. "+
			"supply the required permission or control Mizu's access to namespaces by setting %s "+
			"in the config file or setting the tapped namespace with --%s %s=<NAMEPSACE>",
			err,
			config.MizuResourcesNamespaceConfigName,
			config.SetCommandName,
			config.MizuResourcesNamespaceConfigName)
	} else if syntaxError, isSyntaxError := asRegexSyntaxError(err); isSyntaxError {
		errorNew = fmt.Errorf("regex %s is invalid: %w", syntaxError.Expr, err)
	} else {
		errorNew = err
	}

	return errorNew
}

func asRegexSyntaxError(err error) (*regexpsyntax.Error, bool) {
	var syntaxError *regexpsyntax.Error
	return syntaxError, errors.As(err, &syntaxError)
}
