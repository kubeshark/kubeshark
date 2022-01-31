package workspace

import (
	"errors"
	"mizuserver/pkg/providers/database"

	"github.com/mattn/go-sqlite3"
)

func CreateWorkspace(name string, namespaces []string) (*WorkspaceResponse, error) {
	if workspace, err := database.CreateWorkspace(name, namespaces); err != nil {
		return nil, handleDatabaseError(err)
	} else {
		return &WorkspaceResponse{
			Id:         workspace.Id,
			Name:       workspace.Name,
			Namespaces: namespaceSliceToStringSlice(workspace.Namespaces),
		}, nil
	}
}

func ListWorkspaces() ([]*WorkspaceListItemResponse, error) {
	if workspaces, err := database.ListWorkspaces(); err != nil {
		return nil, handleDatabaseError(err)
	} else {
		workspaceResponseListItems := make([]*WorkspaceListItemResponse, len(workspaces))
		for i, workspace := range workspaces {
			workspaceResponseListItems[i] = &WorkspaceListItemResponse{
				Id:   workspace.Id,
				Name: workspace.Name,
			}
		}
		return workspaceResponseListItems, nil
	}
}

func GetWorkspace(workspaceId string) (*WorkspaceResponse, error) {
	if workspace, err := database.GetWorkspaceWithRelations(workspaceId); err != nil {
		return nil, handleDatabaseError(err)
	} else {
		return &WorkspaceResponse{
			Id:         workspace.Id,
			Name:       workspace.Name,
			Namespaces: namespaceSliceToStringSlice(workspace.Namespaces),
		}, nil
	}
}

func UpdateWorkspace(workspaceId string, name string, namespaces []string) (*WorkspaceResponse, error) {
	if workspace, err := database.UpdateWorkspace(workspaceId, name, namespaces); err != nil {
		return nil, handleDatabaseError(err)
	} else {
		return &WorkspaceResponse{
			Id:         workspace.Id,
			Name:       workspace.Name,
			Namespaces: namespaceSliceToStringSlice(workspace.Namespaces),
		}, nil
	}
}

func namespaceSliceToStringSlice(namespacesObjects []database.Namespace) []string {
	namespaces := make([]string, len(namespacesObjects))

	for i, namespaceObject := range namespacesObjects {
		namespaces[i] = namespaceObject.Name
	}

	return namespaces
}

func handleDatabaseError(err error) error {
	var sqliteError sqlite3.Error
	if errors.As(err, &sqliteError) {
		if sqliteError.ExtendedCode == sqlite3.ErrConstraintUnique {
			return &ErrorWorkspaceNameAlreadyExists{}
		}
	}
	return err
}
