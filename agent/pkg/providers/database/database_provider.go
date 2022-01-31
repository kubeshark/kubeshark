package database

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/up9inc/mizu/shared/logger"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const databasePath = "/app/data/kratos.sqlite"

var db *gorm.DB

func init() {
	var err error
	db, err = gorm.Open(sqlite.Open(databasePath), &gorm.Config{})
	if err != nil {
		panic(fmt.Sprintf("failed to connect local sqlite database at %s", databasePath))
	}
	if db.Error != nil {
		logger.Log.Errorf("db error %v", db.Error)
	}

	err = db.AutoMigrate(&Invite{}, &Namespace{}, &Workspace{})
	if err != nil {
		panic(fmt.Sprintf("failed to migrate schema to local sqlite database at %s", databasePath))
	}
}

func CreateInvite(inviteToken string, username string, identityId string, creationTime int64) error {
	invite := &Invite{
		Token:      inviteToken,
		Username:   username,
		IdentityId: identityId,
		CreatedAt:  creationTime,
	}

	if err := db.Create(invite).Error; err != nil {
		return err
	}
	return nil
}

func GetInviteByInviteToken(inviteToken string) (*Invite, error) {
	var invite Invite
	if err := db.Where("token = ?", inviteToken).First(&invite).Error; err != nil {
		return nil, err
	}
	return &invite, nil
}

func DeleteInvite(inviteToken string) error {
	if err := db.Where("token = ?", inviteToken).Delete(&Invite{}).Error; err != nil {
		return err
	}
	return nil
}

func CreateWorkspace(name string, namespaces []string) (*Workspace, error) {
	namespaceRows := make([]Namespace, len(namespaces))

	for i, namespace := range namespaces {
		namespaceRows[i] = Namespace{
			Name: namespace,
		}
	}

	workspace := &Workspace{
		Id:         uuid.New().String(),
		Name:       name,
		Namespaces: namespaceRows,
	}

	if err := db.Create(workspace).Error; err != nil {
		return nil, err
	}

	return workspace, nil
}

func ListWorkspaces() ([]*Workspace, error) {
	var workspaces []*Workspace
	if err := db.Find(&workspaces).Error; err != nil {
		return nil, err
	}
	return workspaces, nil
}

func GetWorkspaceWithRelations(workspaceId string) (*Workspace, error) {
	var workspace Workspace
	if err := db.Preload("Namespaces").First(&workspace, "id = ?", workspaceId).Error; err != nil {
		return nil, err
	}
	return &workspace, nil
}

func UpdateWorkspace(workspaceId string, name string, namespaces []string) (*Workspace, error) {
	namespaceRows := make([]Namespace, len(namespaces))

	for i, namespace := range namespaces {
		namespaceRows[i] = Namespace{
			Name: namespace,
		}
	}

	workspace := &Workspace{
		Id:         workspaceId,
		Name:       name,
		Namespaces: namespaceRows,
	}

	if err := db.Transaction(func(tx *gorm.DB) error {
		// delete old namespaces
		if err := tx.Delete(&Namespace{}, "workspace_id = ?", workspaceId).Error; err != nil {
			return err
		}
		if err := tx.Save(workspace).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return workspace, nil
}
