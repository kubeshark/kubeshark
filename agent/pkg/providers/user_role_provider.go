package providers

import (
	"errors"
	"fmt"

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
	readClient               = getKetoClient(fmt.Sprintf("%s:%d", ketoHost, ketoReadPort))
	writeClient              = getKetoClient(fmt.Sprintf("%s:%d", ketoHost, ketoWritePort))
	ketoUsersNamespace       = "users"
	ketoUserRoleRelationName = "userRole"
)

func GetUserRoles(username string) ([]string, error) {
	clientRelationsResponse, err := readClient.Read.GetRelationTuples(ketoRead.
		NewGetRelationTuplesParams().
		WithNamespace(ketoUsersNamespace).
		WithRelation(&ketoUserRoleRelationName).
		WithSubjectID(&username))

	if err != nil {
		return nil, err
	}
	if clientRelationsResponse.Error() != "" {
		return nil, errors.New(clientRelationsResponse.Error())
	}

	relationObjects := make([]string, 0)

	for _, clientRelation := range clientRelationsResponse.Payload.RelationTuples {
		relationObjects = append(relationObjects, *clientRelation.Object)
	}

	return relationObjects, nil
}

func SetUserRole(username string, roleName string) error {
	tuple := ketoModels.RelationQuery{
		Namespace: &ketoUsersNamespace,
		Object:    roleName,
		Relation:  ketoUserRoleRelationName,
		SubjectID: username,
	}

	tupleCreateResponse, err := writeClient.Write.CreateRelationTuple(ketoWrite.
		NewCreateRelationTupleParams().
		WithPayload(&tuple))

	if err != nil {
		return err
	}
	if tupleCreateResponse.Error() != "" {
		return errors.New(tupleCreateResponse.Error())
	}

	return nil
}

func getKetoClient(url string) *ketoClient.OryKeto {
	return ketoClient.NewHTTPClientWithConfig(nil,
		ketoClient.
			DefaultTransportConfig().
			WithSchemes([]string{"http"}).
			WithHost(url))
}
