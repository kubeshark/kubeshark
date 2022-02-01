package providers

/*
This provider abstracts keto role management down to what we need for mizu

Keto, in the configuration we use it, is basically a tuple database. Each tuple consists of 4 strings (namespace, object, relation, subjectID) - for example ("workspaces", "sock-shop-workspace", "viewer", "ramiberman")

namespace - used to organize tuples into groups - we currently use "system" for defining admins and "workspaces" for defining workspace permissions
objects - represents something one can have permissions to (files, mizu workspaces etc)
relation - represents the permission (viewer, editor, owner etc) - we currently use only viewer and admin
subject - represents the user or group that has the permission - we currently use usernames

more on keto here: https://www.ory.sh/keto/docs/
*/

import (
	"fmt"

	"github.com/up9inc/mizu/agent/pkg/utils"

	ketoClient "github.com/ory/keto-client-go/client"
	ketoRead "github.com/ory/keto-client-go/client/read"
	ketoWrite "github.com/ory/keto-client-go/client/write"
	ketoModels "github.com/ory/keto-client-go/models"
)

const (
	ketoHost      = "localhost"
	ketoReadPort  = 4466
	ketoWritePort = 4467
)

var (
	readClient              = getKetoClient(fmt.Sprintf("%s:%d", ketoHost, ketoReadPort))
	writeClient             = getKetoClient(fmt.Sprintf("%s:%d", ketoHost, ketoWritePort))
	systemRoleNamespace     = "system"
	workspacesRoleNamespace = "workspaces"

	systemObject = "system"

	AdminRole  = "admin"
	ViewerRole = "viewer"
)

func GetUserSystemRoles(username string) ([]string, error) {
	return getObjectRelationsForSubjectID(systemRoleNamespace, systemObject, username)
}

func CheckIfUserHasSystemRole(username string, role string) (bool, error) {
	systemRoles, err := GetUserSystemRoles(username)
	if err != nil {
		return false, err
	}

	for _, systemRole := range systemRoles {
		if systemRole == role {
			return true, nil
		}
	}

	return false, nil
}

func GetUserWorkspaceRole(username string, workspace string) ([]string, error) {
	return getObjectRelationsForSubjectID(workspacesRoleNamespace, workspace, username)
}

func SetUserWorkspaceRole(username string, workspace string, role string) error {
	return createObjectRelationForSubjectID(workspacesRoleNamespace, workspace, username, role)
}

func SetUserSystemRole(username string, role string) error {
	return createObjectRelationForSubjectID(systemRoleNamespace, systemObject, username, role)
}

func DeleteAllUserWorkspaceRoles(username string) error {
	return deleteAllNamespacedRelationsForSubjectID(workspacesRoleNamespace, username)
}

func createObjectRelationForSubjectID(namespace string, object string, subjectID string, relation string) error {
	tuple := ketoModels.RelationQuery{
		Namespace: &namespace,
		Object:    object,
		Relation:  relation,
		SubjectID: subjectID,
	}

	_, err := writeClient.Write.CreateRelationTuple(ketoWrite.
		NewCreateRelationTupleParams().
		WithPayload(&tuple))

	if err != nil {
		return err
	}

	return nil
}

func getObjectRelationsForSubjectID(namespace string, object string, subjectID string) ([]string, error) {
	relationTuples, err := queryRelationTuples(&namespace, &object, &subjectID, nil)
	if err != nil {
		return nil, err
	}

	relations := make([]string, 0)

	for _, clientRelation := range relationTuples {
		relations = append(relations, *clientRelation.Relation)
	}

	return utils.UniqueStringSlice(relations), nil
}

func deleteAllNamespacedRelationsForSubjectID(namespace string, subjectID string) error {
	relationTuples, err := queryRelationTuples(&namespace, nil, &subjectID, nil)
	if err != nil {
		return err
	}

	for _, clientRelation := range relationTuples {
		_, err := writeClient.Write.DeleteRelationTuple(ketoWrite.
			NewDeleteRelationTupleParams().
			WithNamespace(*clientRelation.Namespace).
			WithObject(*clientRelation.Object).
			WithRelation(*clientRelation.Relation).
			WithSubjectID(&clientRelation.SubjectID))

		if err != nil {
			return err
		}
	}

	return nil
}

func queryRelationTuples(namespace *string, object *string, subjectID *string, role *string) ([]*ketoModels.InternalRelationTuple, error) {
	relationTuplesQuery := ketoRead.NewGetRelationTuplesParams()
	if namespace != nil {
		relationTuplesQuery = relationTuplesQuery.WithNamespace(*namespace)
	}
	if object != nil {
		relationTuplesQuery = relationTuplesQuery.WithObject(object)
	}
	if subjectID != nil {
		relationTuplesQuery = relationTuplesQuery.WithSubjectID(subjectID)
	}
	if role != nil {
		relationTuplesQuery = relationTuplesQuery.WithRelation(role)
	}

	return recursiveKetoPagingTraverse(relationTuplesQuery, make([]*ketoModels.InternalRelationTuple, 0), "")
}

func recursiveKetoPagingTraverse(queryParams *ketoRead.GetRelationTuplesParams, tuples []*ketoModels.InternalRelationTuple, pagingToken string) ([]*ketoModels.InternalRelationTuple, error) {
	params := queryParams
	if pagingToken != "" {
		params = queryParams.WithPageToken(&pagingToken)
	}

	clientRelationsResponse, err := readClient.Read.GetRelationTuples(params)

	if err != nil {
		return nil, err
	}

	tuples = append(tuples, clientRelationsResponse.Payload.RelationTuples...)

	if clientRelationsResponse.Payload.NextPageToken != "" {
		return recursiveKetoPagingTraverse(queryParams, tuples, clientRelationsResponse.Payload.NextPageToken)
	}

	return tuples, nil
}

func getKetoClient(url string) *ketoClient.OryKeto {
	return ketoClient.NewHTTPClientWithConfig(nil,
		ketoClient.
			DefaultTransportConfig().
			WithSchemes([]string{"http"}).
			WithHost(url))
}
