package workspace

import "mizuserver/pkg/models"

func CreateWorkspace(name string, namespaces []string) error {
	workspace := models.Workspace{
		Name:       name,
		Namespaces: namespaces,
	}

	if err = db.Create(workspace).Error; err != nil {
		return err
	}
}
