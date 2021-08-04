package errormessage

import (
	"errors"
	"fmt"
	regexpsyntax "regexp/syntax"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
)

// formatError wraps error with a detailed message that is meant for the user.
// While the errors are meant to be displayed, they are not meant to be exported as classes outsite of CLI.
func FormatError(err error) error {
	var errorNew error
	if k8serrors.IsForbidden(err) {
		errorNew = fmt.Errorf("Insufficient permissions: %w. Supply the required permission or control Mizu's access to namespaces by setting MizuNamespace in the config file or setting the tapped namespace with --set mizu-namespace=<NAMEPSACE>.", err)
	} else if syntaxError, isSyntaxError := asRegexSyntaxError(err); isSyntaxError {
		errorNew = fmt.Errorf("Regex %s is invalid: %w", syntaxError.Expr, err)
	} else {
		errorNew = err
	}

	return errorNew
}

func asRegexSyntaxError(err error) (*regexpsyntax.Error, bool) {
	var syntaxError *regexpsyntax.Error
	return syntaxError, errors.As(err, &syntaxError)
}
