package providers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/cookiejar"

	ory "github.com/ory/kratos-client-go"
	"github.com/up9inc/mizu/shared/logger"
)

var client = getKratosClient("http://127.0.0.1:4433", "http://127.0.0.1:4434")

// returns session token if successful
func RegisterUser(username string, password string, ctx context.Context) (token *string, identityId string, err error, formErrorMessages map[string][]ory.UiText) {
	flow, _, err := client.V0alpha2Api.InitializeSelfServiceRegistrationFlowWithoutBrowser(ctx).Execute()
	if err != nil {
		return nil, "", err, nil
	}

	result, _, err := client.V0alpha2Api.SubmitSelfServiceRegistrationFlow(ctx).Flow(flow.Id).SubmitSelfServiceRegistrationFlowBody(
		ory.SubmitSelfServiceRegistrationFlowWithPasswordMethodBodyAsSubmitSelfServiceRegistrationFlowBody(&ory.SubmitSelfServiceRegistrationFlowWithPasswordMethodBody{
			Method:   "password",
			Password: password,
			Traits:   map[string]interface{}{"username": username},
		}),
	).Execute()

	if err != nil {
		parsedKratosError, parsingErr := parseKratosRegistrationFormError(err)
		if parsingErr != nil {
			logger.Log.Debugf("error parsing kratos error: %v", parsingErr)
			return nil, "", err, nil
		} else {
			return nil, "", err, parsedKratosError
		}
	}

	return result.SessionToken, result.Identity.Id, nil, nil
}

func PerformLogin(username string, password string, ctx context.Context) (*string, error) {
	flow, _, err := client.V0alpha2Api.InitializeSelfServiceLoginFlowWithoutBrowser(ctx).Execute()
	if err != nil {
		return nil, err
	}

	result, _, err := client.V0alpha2Api.SubmitSelfServiceLoginFlow(ctx).Flow(flow.Id).SubmitSelfServiceLoginFlowBody(
		ory.SubmitSelfServiceLoginFlowWithPasswordMethodBodyAsSubmitSelfServiceLoginFlowBody(&ory.SubmitSelfServiceLoginFlowWithPasswordMethodBody{
			Method:             "password",
			Password:           password,
			PasswordIdentifier: username,
		}),
	).Execute()

	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, errors.New("unknown error occured during login")
	}

	return result.SessionToken, nil
}

func VerifyToken(token string, ctx context.Context) (*ory.Session, error) {
	flow, _, err := client.V0alpha2Api.ToSession(ctx).XSessionToken(token).Execute()
	if err != nil {
		return nil, err
	}

	if flow == nil {
		return nil, nil
	}

	return flow, nil
}

func DeleteUser(identityId string, ctx context.Context) error {
	result, err := client.V0alpha2Api.AdminDeleteIdentity(ctx, identityId).Execute()
	if err != nil {
		return err
	}
	if result == nil {
		return fmt.Errorf("unknown error occured during user deletion %v", identityId)
	}

	if result.StatusCode < 200 || result.StatusCode > 299 {
		return fmt.Errorf("user deletion %v returned bad status %d", identityId, result.StatusCode)
	} else {
		return nil
	}
}

func AnyUserExists(ctx context.Context) (bool, error) {
	request := client.V0alpha2Api.AdminListIdentities(ctx)
	request.PerPage(1)

	if result, _, err := request.Execute(); err != nil {
		return false, err
	} else {
		return len(result) > 0, nil
	}
}

func Logout(token string, ctx context.Context) error {
	logoutRequest := client.V0alpha2Api.SubmitSelfServiceLogoutFlowWithoutBrowser(ctx)
	logoutRequest = logoutRequest.SubmitSelfServiceLogoutFlowWithoutBrowserBody(ory.SubmitSelfServiceLogoutFlowWithoutBrowserBody{
		SessionToken: token,
	})
	if response, err := logoutRequest.Execute(); err != nil {
		return err
	} else if response == nil || response.StatusCode < 200 || response.StatusCode > 299 {
		return errors.New("unknown error occured during logout")
	}

	return nil
}

func getKratosClient(url string, adminUrl string) *ory.APIClient {
	conf := ory.NewConfiguration()
	conf.Servers = ory.ServerConfigurations{{URL: url}}

	// this ensures kratos client uses the admin url for admin actions (any new admin action we use will have to be added here)
	conf.OperationServers = map[string]ory.ServerConfigurations{
		"V0alpha2ApiService.AdminDeleteIdentity": {{URL: adminUrl}},
		"V0alpha2ApiService.AdminListIdentities": {{URL: adminUrl}},
	}

	cj, _ := cookiejar.New(nil)
	conf.HTTPClient = &http.Client{Jar: cj}
	return ory.NewAPIClient(conf)
}

// returns map of form value key to error message
func parseKratosRegistrationFormError(err error) (map[string][]ory.UiText, error) {
	var openApiError *ory.GenericOpenAPIError
	if errors.As(err, &openApiError) {
		var registrationFlowModel *ory.SelfServiceRegistrationFlow
		if jsonErr := json.Unmarshal(openApiError.Body(), &registrationFlowModel); jsonErr != nil {
			return nil, jsonErr
		} else {
			formMessages := registrationFlowModel.Ui.Nodes
			parsedMessages := make(map[string][]ory.UiText)

			for _, message := range formMessages {
				if len(message.Messages) > 0 {
					if _, ok := parsedMessages[message.Group]; !ok {
						parsedMessages[message.Group] = make([]ory.UiText, 0)
					}
					parsedMessages[message.Group] = append(parsedMessages[message.Group], message.Messages...)
				}
			}
			return parsedMessages, nil
		}
	} else {
		return nil, errors.New("error is not a generic openapi error")
	}
}
