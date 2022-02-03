package database

import (
	"errors"
	"fmt"

	"github.com/up9inc/mizu/agent/pkg/config"

	"github.com/google/uuid"
	"github.com/mattn/go-sqlite3"
	"github.com/up9inc/mizu/shared/logger"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const (
	databaseFileName               = "kratos.sqlite"
	gormRecordNotFoundErrorMessage = "record not found"
)

var db *gorm.DB

func InitializeApplicationDatabase() {
	databasePath := config.Config.AgentDatabasePath + databaseFileName

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
		return handleDatabaseError(err)
	}
	return nil
}

func GetInviteByInviteToken(inviteToken string) (*Invite, error) {
	var invite Invite
	if err := db.Where("token = ?", inviteToken).First(&invite).Error; err != nil {
		return nil, handleDatabaseError(err)
	}
	return &invite, nil
}

func DeleteInvite(inviteToken string) error {
	if err := db.Where("token = ?", inviteToken).Delete(&Invite{}).Error; err != nil {
		return handleDatabaseError(err)
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
		return nil, handleDatabaseError(err)
	}

	return workspace, nil
}

func ListWorkspaces() ([]*Workspace, error) {
	var workspaces []*Workspace
	if err := db.Find(&workspaces).Error; err != nil {
		return nil, handleDatabaseError(err)
	}
	return workspaces, nil
}

func ListWorkspacesWithRelations() ([]*Workspace, error) {
	var workspaces []*Workspace
	if err := db.Preload("Namespaces").Find(&workspaces).Error; err != nil {
		return nil, handleDatabaseError(err)
	}
	return workspaces, nil
}

func GetWorkspaceWithRelations(workspaceId string) (*Workspace, error) {
	var workspace Workspace
	if err := db.Preload("Namespaces").First(&workspace, "id = ?", workspaceId).Error; err != nil {
		return nil, handleDatabaseError(err)
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
		// ensure workspace exists, tx.Save will create it if it doesn't exist so we must check
		if err := tx.First(&Workspace{}, "id = ?", workspaceId).Error; err != nil {
			return err
		}
		// delete old namespaces
		if err := tx.Delete(&Namespace{}, "workspace_id = ?", workspaceId).Error; err != nil {
			return err
		}
		// this also creates the new namespace rows
		if err := tx.Save(workspace).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, handleDatabaseError(err)
	}

	return workspace, nil
}

func DeleteWorkspace(workspaceId string) error {
	if err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&Namespace{}, "workspace_id = ?", workspaceId).Error; err != nil {
			return err
		}
		if err := tx.Delete(&Workspace{}, "id = ?", workspaceId).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		return handleDatabaseError(err)
	}
	return nil
}

func GetAllUniqueNamespaces() ([]string, error) {
	var namespaces []string
	if err := db.Model(&Namespace{}).Distinct("Name").Pluck("Name", &namespaces).Error; err != nil {
		return nil, handleDatabaseError(err)
	}
	return namespaces, nil
}

func handleDatabaseError(err error) error {
	var sqliteError sqlite3.Error
	if errors.As(err, &sqliteError) {
		if sqliteError.ExtendedCode == sqlite3.ErrConstraintUnique {
			return &ErrorUniqueConstraintViolation{}
		}
	} else if err.Error() == gormRecordNotFoundErrorMessage {
		return &ErrorNotFound{}
	}
	return err
}
