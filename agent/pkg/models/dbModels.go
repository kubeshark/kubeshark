package models

import "gorm.io/gorm"

type Invite struct {
	gorm.Model
	Token      string
	IdentityId string
	Username   string
	CreatedAt  int64
}
