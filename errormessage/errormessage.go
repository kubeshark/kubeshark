package errormessage

import (
	"errors"
	"fmt"
	regexpsyntax "regexp/syntax"

	"github.com/kubeshark/kubeshark/config"
	"github.com/kubeshark/kubeshark/config/configStructs"
	"github.com/kubeshark/kubeshark/misc"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
)

// formatError wraps error with a detailed message that is meant for the user.
// While the errors are meant to be displayed, they are not meant to be exported as classes outsite of CLI.
func FormatError(err error) error {
	var errorNew error
	if k8serrors.IsForbidden(err) {
		errorNew = fmt.Errorf("insufficient permissions: %w. "+
			"supply the required permission or control %s's access to namespaces by setting %s "+
			"in the config file or setting the targeted namespace with --%s %s=<NAMESPACE>",
			err,
			misc.Software,
			configStructs.ReleaseNamespaceLabel,
			config.SetCommandName,
			configStructs.ReleaseNamespaceLabel)
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
