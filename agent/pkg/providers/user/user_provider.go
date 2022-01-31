package user

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"mizuserver/pkg/providers/database"
	"mizuserver/pkg/providers/workspace"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"time"

	"github.com/google/uuid"

	ory "github.com/ory/kratos-client-go"
	"github.com/up9inc/mizu/shared/logger"
)

const (
	inviteTTLSeconds = 60 * 60 * 24 * 14 // two weeks
	listUsersPerPage = 500
)

var client = getKratosClient("http://127.0.0.1:4433", "http://127.0.0.1:4434")

// returns session token if successful
func RegisterUser(username string, password string, inviteStatus InviteStatus, ctx context.Context) (token *string, identityId string, err error, formErrorMessages map[string][]ory.UiText) {
	flow, _, err := client.V0alpha2Api.InitializeSelfServiceRegistrationFlowWithoutBrowser(ctx).Execute()
	if err != nil {
		return nil, "", err, nil
	}

	result, _, err := client.V0alpha2Api.SubmitSelfServiceRegistrationFlow(ctx).Flow(flow.Id).SubmitSelfServiceRegistrationFlowBody(
		ory.SubmitSelfServiceRegistrationFlowWithPasswordMethodBodyAsSubmitSelfServiceRegistrationFlowBody(&ory.SubmitSelfServiceRegistrationFlowWithPasswordMethodBody{
			Method:   "password",
			Password: password,
			Traits:   map[string]interface{}{"username": username, "inviteStatus": inviteStatus},
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
	identity, _, err := client.V0alpha2Api.AdminGetIdentity(ctx, identityId).Execute()
	if err != nil {
		return err
	}

	traits := identity.Traits.(map[string]interface{})
	username := traits["username"].(string)

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
		return DeleteAllUserRoles(username)
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

func CreateNewUserWithInvite(username string, workspaceId string, systemRole string, ctx context.Context) (inviteToken string, identityId string, err error) {
	_, identityId, err, _ = RegisterUser(username, uuid.New().String(), PendingInviteStatus, ctx)
	if err != nil {
		return "", "", err
	}

	if err = SetUserSystemRole(username, systemRole); err != nil {
		DeleteUser(identityId, ctx)
		return "", "", err
	}

	if err = SetUserWorkspaceRole(username, workspaceId, UserRole); err != nil {
		DeleteUser(identityId, ctx)
		return "", "", err
	}

	if inviteToken, err := CreateInvite(identityId, ctx); err != nil {
		DeleteUser(identityId, ctx)
		return "", "", err
	} else {
		return inviteToken, identityId, nil
	}
}

func CreateInvite(identityId string, ctx context.Context) (inviteToken string, err error) {
	identity, _, err := client.V0alpha2Api.AdminGetIdentity(ctx, identityId).Execute()
	if err != nil {
		return "", err
	}

	traits := identity.Traits.(map[string]interface{})
	username := traits["username"].(string)

	inviteToken = uuid.New().String()

	if err := database.CreateInvite(inviteToken, username, identityId, time.Now().Unix()); err != nil {
		return "", err
	}

	return inviteToken, nil
}

func ResetPasswordWithInvite(inviteToken string, password string, ctx context.Context) (token *string, err error, formErrorMessages map[string][]ory.UiText) {
	invite, err := database.GetInviteByInviteToken(inviteToken)
	if err != nil {
		return nil, err, nil
	}

	if time.Now().Unix()-invite.CreatedAt > inviteTTLSeconds {
		return nil, errors.New("invite expired"), nil
	}

	sessionToken, err := GetUserSessionTokenUsingAdminAccess(invite.IdentityId, ctx)
	if err != nil {
		return nil, err, nil
	}

	err, formErrors := ChangePassword(sessionToken, password, ctx)
	if err != nil || formErrors != nil {
		return nil, err, formErrors
	}

	identity, _, err := client.V0alpha2Api.AdminGetIdentity(ctx, invite.IdentityId).Execute()
	if err != nil {
		return nil, err, nil
	}

	if err = database.DeleteInvite(invite.Token); err != nil {
		logger.Log.Warningf("error deleting invite: %v", err)
	}

	traits := identity.Traits.(map[string]interface{})
	traits["inviteStatus"] = AcceptedInviteStatus
	if err = UpdateUserTraits(identity.Id, traits, ctx); err != nil {
		logger.Log.Warningf("error updating user invite status: %v", err)
	}

	return &sessionToken, nil, nil
}

func GetUserSessionTokenUsingAdminAccess(identityId string, ctx context.Context) (string, error) {
	recoveryLink, _, err := client.V0alpha2Api.AdminCreateSelfServiceRecoveryLink(ctx).AdminCreateSelfServiceRecoveryLinkBody(ory.AdminCreateSelfServiceRecoveryLinkBody{
		IdentityId: identityId,
	}).Execute()

	if err != nil {
		return "", err
	}

	client := &http.Client{}

	resp, err := client.Get(recoveryLink.RecoveryLink)
	if err != nil {
		return "", err
	} else {
		defer resp.Body.Close()
		tokenResponse := &TokenResponse{}
		if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
			return "", err
		}
		return tokenResponse.Token, nil
	}
}

func ChangePassword(sessionToken string, newPassword string, ctx context.Context) (err error, formErrorMessages map[string][]ory.UiText) {
	flow, _, err := client.V0alpha2Api.InitializeSelfServiceSettingsFlowWithoutBrowser(ctx).XSessionToken(sessionToken).Execute()
	if err != nil {
		return err, nil
	}
	_, _, err = client.V0alpha2Api.SubmitSelfServiceSettingsFlow(ctx).Flow(flow.Id).XSessionToken(sessionToken).SubmitSelfServiceSettingsFlowBody(
		ory.SubmitSelfServiceSettingsFlowWithPasswordMethodBodyAsSubmitSelfServiceSettingsFlowBody(&ory.SubmitSelfServiceSettingsFlowWithPasswordMethodBody{
			Method:   "password",
			Password: newPassword,
		}),
	).Execute()

	if err != nil {
		parsedKratosError, parsingErr := parseKratosRegistrationFormError(err)
		if parsingErr != nil {
			logger.Log.Debugf("error parsing kratos error: %v", parsingErr)
			return err, nil
		} else {
			return err, parsedKratosError
		}
	}

	return nil, nil
}

func UpdateUserRoles(identityId string, workspaceId string, systemRole string, ctx context.Context) error {
	identity, _, err := client.V0alpha2Api.AdminGetIdentity(ctx, identityId).Execute()
	if err != nil {
		return err
	}

	traits := identity.Traits.(map[string]interface{})
	username := traits["username"].(string)

	if err = SetUserWorkspaceRole(username, workspaceId, UserRole); err != nil {
		return err
	}
	if err = SetUserSystemRole(username, systemRole); err != nil {
		return err
	}

	return nil
}

func UpdateUserTraits(identityId string, traits map[string]interface{}, ctx context.Context) error {
	_, _, err := client.V0alpha2Api.AdminUpdateIdentity(ctx, identityId).AdminUpdateIdentityBody(ory.AdminUpdateIdentityBody{Traits: traits}).Execute()
	return err
}

func GetUser(identityId string, ctx context.Context) (*User, error) {
	user := &User{
		UserId: identityId,
	}

	identity, _, err := client.V0alpha2Api.AdminGetIdentity(ctx, identityId).Execute()
	if err != nil {
		return nil, err
	}

	traits := identity.Traits.(map[string]interface{})
	username := traits["username"].(string)
	user.Username = username
	user.Status = InviteStatus(traits["inviteStatus"].(string))

	if systemRole, err := GetUserSystemRole(username); err != nil {
		return nil, err
	} else {
		user.SystemRole = systemRole
	}

	if workspaceId, err := GetUserWorkspaceId(username); err != nil {
		return nil, err
	} else {
		user.WorkspaceId = workspaceId
	}

	return user, nil
}

func ListUsers(usernameFilterQuery string, ctx context.Context) ([]UserListItem, error) {
	var users []UserListItem

	identities, err := getAllUsersRecursively(ctx, 0)
	if err != nil {
		return nil, err
	}

	for _, identity := range identities {
		traits := identity.Traits.(map[string]interface{})
		username := traits["username"].(string)

		if usernameFilterQuery == "" || strings.Contains(username, usernameFilterQuery) {
			inviteStatus := AcceptedInviteStatus
			if traits["inviteStatus"] != nil {
				inviteStatus = InviteStatus(traits["inviteStatus"].(string))
			}

			systemRole, err := GetUserSystemRole(username)
			if err != nil {
				return nil, err
			}

			users = append(users, UserListItem{
				Username:   username,
				UserId:     identity.Id,
				Status:     InviteStatus(inviteStatus),
				SystemRole: systemRole,
			})
		}
	}

	return users, nil
}

func WhoAmI(sessionToken string, ctx context.Context) (*WhoAmIResponse, error) {
	session, err := VerifyToken(sessionToken, ctx)
	if err != nil {
		return nil, err
	}

	user, err := GetUser(session.Identity.Id, ctx)
	if err != nil {
		return nil, err
	}

	var userWorkspace *workspace.WorkspaceResponse
	if user.WorkspaceId != "" {
		if userWorkspace, err = workspace.GetWorkspace(user.WorkspaceId); err != nil {
			return nil, err
		}
	}

	return &WhoAmIResponse{
		Username:   user.Username,
		SystemRole: user.SystemRole,
		Workspace:  userWorkspace,
	}, nil
}

func getAllUsersRecursively(ctx context.Context, page int) ([]ory.Identity, error) {
	var users []ory.Identity

	result, _, err := client.V0alpha2Api.AdminListIdentities(ctx).PerPage(listUsersPerPage).Page(int64(page)).Execute()
	if err != nil {
		return nil, err
	}

	users = append(users, result...)

	if len(result) == listUsersPerPage {
		nextPages, err := getAllUsersRecursively(ctx, page+1)
		if err != nil {
			return nil, err
		}
		return append(users, nextPages...), nil
	} else {
		return users, nil
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
