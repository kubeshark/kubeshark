package providers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/cookiejar"

	ory "github.com/ory/kratos-client-go"
)

var client = getKratosClient("http://127.0.0.1:4433", "http://127.0.0.1:4434")

// returns bearer token if successful
func RegisterUser(username string, password string, ctx context.Context) (token *string, identityId string, err error) {
	flow, _, err := client.V0alpha2Api.InitializeSelfServiceRegistrationFlowWithoutBrowser(ctx).Execute()
	if err != nil {
		return nil, "", err
	}

	result, _, err := client.V0alpha2Api.SubmitSelfServiceRegistrationFlow(ctx).Flow(flow.Id).SubmitSelfServiceRegistrationFlowBody(
		ory.SubmitSelfServiceRegistrationFlowWithPasswordMethodBodyAsSubmitSelfServiceRegistrationFlowBody(&ory.SubmitSelfServiceRegistrationFlowWithPasswordMethodBody{
			Method:   "password",
			Password: password,
			Traits:   map[string]interface{}{"username": username},
		}),
	).Execute()

	if err != nil {
		return nil, "", err
	}

	return result.SessionToken, result.Identity.Id, nil
}

func PerformLogin(username string, password string, ctx context.Context) (*string, error) {
	// Initialize the flow
	flow, _, err := client.V0alpha2Api.InitializeSelfServiceLoginFlowWithoutBrowser(ctx).Execute()
	if err != nil {
		return nil, err
	}

	// Submit the form
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

func VerifyToken(token string, ctx context.Context) (bool, error) {
	flow, _, err := client.V0alpha2Api.ToSession(ctx).XSessionToken(token).Execute()
	if err != nil {
		return false, err
	}

	if flow == nil {
		return false, nil
	}

	return true, nil
}

func DeleteUser(identityId string, ctx context.Context) error {
	result, err := client.V0alpha2Api.AdminDeleteIdentity(ctx, identityId).Execute()
	if err != nil {
		return err
	}
	if result == nil {
		return errors.New("unknown error occured during user deletion")
	}

	if result.StatusCode < 200 || result.StatusCode > 299 {
		return errors.New(fmt.Sprintf("user deletion returned bad status %d", result.StatusCode))
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
