package providers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"mizuserver/pkg/models"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	ory "github.com/ory/kratos-client-go"
	"github.com/up9inc/mizu/shared/logger"
)

const (
	databasePath     = "/app/data/kratos.sqlite"
	inviteTTLSeconds = 60 * 60 * 24 * 14 // two weeks
	listUsersPerPage = 500
)

var (
	client = getKratosClient("http://127.0.0.1:4433", "http://127.0.0.1:4434")
	db     *gorm.DB
)

func init() {
	var err error
	db, err = gorm.Open(sqlite.Open(databasePath), &gorm.Config{})
	if err != nil {
		panic(fmt.Sprintf("failed to connect local sqlite database at %s", databasePath))
	}
	if db.Error != nil {
		logger.Log.Errorf("db error %v", db.Error)
	}

	err = db.AutoMigrate(&models.Invite{})
	if err != nil {
		panic(fmt.Sprintf("failed to migrate schema to local sqlite database at %s", databasePath))
	}
}

// returns session token if successful
func RegisterUser(username string, password string, inviteStatus models.InviteStatus, ctx context.Context) (token *string, identityId string, err error, formErrorMessages map[string][]ory.UiText) {
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

func CreateUserInvite(username string, workspace string, systemRole string, ctx context.Context) (inviteToken string, identityId string, err error) {
	inviteToken = uuid.New().String()

	// use inviteToken as a temporary password
	_, identityId, err, _ = RegisterUser(username, inviteToken, models.PendingInviteStatus, ctx)
	if err != nil {
		return "", "", err
	}
	invite := &models.Invite{
		Token:      inviteToken,
		Username:   username,
		IdentityId: identityId,
		CreatedAt:  time.Now().Unix(),
	}

	if err = db.Create(invite).Error; err != nil {
		//Delete the user to prevent a locked user scenario
		DeleteUser(identityId, ctx)

		return "", "", err
	}

	return inviteToken, identityId, nil
}

func RegisterWithInvite(inviteToken string, password string, ctx context.Context) (token *string, err error, formErrorMessages map[string][]ory.UiText) {
	invite := models.Invite{}
	if err = db.Where("token = ?", inviteToken).First(&invite).Error; err != nil {
		return nil, err, nil
	}

	if time.Now().Unix()-invite.CreatedAt > inviteTTLSeconds {
		return nil, errors.New("invite expired"), nil
	}

	// TODO: kratos are planning to add an admin endpoint that allows changing user passwords without having to log in
	sessionToken, err := PerformLogin(invite.Username, inviteToken, ctx)
	if err != nil {
		return nil, err, nil
	}

	flow, _, err := client.V0alpha2Api.InitializeSelfServiceSettingsFlowWithoutBrowser(ctx).XSessionToken(*sessionToken).Execute()
	if err != nil {
		return nil, err, nil
	}
	_, _, err = client.V0alpha2Api.SubmitSelfServiceSettingsFlow(ctx).Flow(flow.Id).XSessionToken(*sessionToken).SubmitSelfServiceSettingsFlowBody(
		ory.SubmitSelfServiceSettingsFlowWithPasswordMethodBodyAsSubmitSelfServiceSettingsFlowBody(&ory.SubmitSelfServiceSettingsFlowWithPasswordMethodBody{
			Method:   "password",
			Password: password,
		}),
	).Execute()

	if err != nil {
		parsedKratosError, parsingErr := parseKratosRegistrationFormError(err)
		if parsingErr != nil {
			logger.Log.Debugf("error parsing kratos error: %v", parsingErr)
			return nil, err, nil
		} else {
			return nil, err, parsedKratosError
		}
	}

	if err = db.Delete(&invite).Error; err != nil {
		logger.Log.Warningf("error deleting invite: %v", err)
	}

	traits := flow.Identity.Traits.(map[string]interface{})
	traits["inviteStatus"] = models.AcceptedInviteStatus
	if err = UpdateUserTraits(flow.Identity.Id, traits, ctx); err != nil {
		logger.Log.Warningf("error updating user invite status: %v", err)
	}

	return sessionToken, nil, nil
}

func UpdateUserRoles(identityId string, workspace string, systemRole string, ctx context.Context) error {
	identity, _, err := client.V0alpha2Api.AdminGetIdentity(ctx, identityId).Execute()
	if err != nil {
		return err
	}

	traits := identity.Traits.(map[string]interface{})
	username := traits["username"].(string)

	if err = SetUserWorkspaceRole(username, workspace, UserRole); err != nil {
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

func GetUser(identityId string, ctx context.Context) (*models.User, error) {
	user := &models.User{
		UserId: identityId,
	}

	identity, _, err := client.V0alpha2Api.AdminGetIdentity(ctx, identityId).Execute()
	if err != nil {
		return nil, err
	}

	traits := identity.Traits.(map[string]interface{})
	username := traits["username"].(string)
	user.Username = username
	user.InviteStatus = models.InviteStatus(traits["inviteStatus"].(string))

	if systemRole, err := GetUserSystemRole(username); err != nil {
		return nil, err
	} else {
		user.SystemRole = systemRole
	}

	if workspace, err := GetUserWorkspace(username); err != nil {
		return nil, err
	} else {
		user.Workspace = workspace
	}

	return user, nil
}

func ListUsers(usernameFilterQuery string, ctx context.Context) ([]models.UserListItem, error) {
	var users []models.UserListItem

	identities, err := getAllUsersRecursively(ctx, 0)
	if err != nil {
		return nil, err
	}

	for _, identity := range identities {
		traits := identity.Traits.(map[string]interface{})
		username := traits["username"].(string)

		if usernameFilterQuery == "" || strings.Contains(username, usernameFilterQuery) {
			inviteStatus := models.AcceptedInviteStatus
			if traits["inviteStatus"] != nil {
				inviteStatus = models.InviteStatus(traits["inviteStatus"].(string))
			}

			users = append(users, models.UserListItem{
				Username:     username,
				UserId:       identity.Id,
				InviteStatus: models.InviteStatus(inviteStatus),
			})
		}
	}

	return users, nil
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
