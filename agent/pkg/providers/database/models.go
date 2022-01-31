package database

import "gorm.io/gorm"

type Invite struct {
	gorm.Model
	Token      string
	IdentityId string
	Username   string
	CreatedAt  int64
}

type Workspace struct {
	gorm.Model
	Name       string
	Namespaces []string
}
