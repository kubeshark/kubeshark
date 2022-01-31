package database

type Invite struct {
	Token      string `gorm:"primary_key"`
	IdentityId string
	Username   string
	CreatedAt  int64
}

type Namespace struct {
	Name        string `gorm:"primary_key"`
	WorkspaceID string `gorm:"primary_key"`
}

type Workspace struct {
	Id         string      `gorm:"primaryKey"`
	Name       string      `gorm:"unique"`
	Namespaces []Namespace `gorm:"foreignKey:WorkspaceID"`
}

type UniqueConstraintError struct {
	err error
}
