package user

type InviteStatus string

const (
	PendingInviteStatus  InviteStatus = "pending"
	AcceptedInviteStatus InviteStatus = "active"
)

type TokenResponse struct {
	Token string `json:"token"`
}

type User struct {
	Username   string       `json:"username"`
	UserId     string       `json:"userId"`
	Status     InviteStatus `json:"status"`
	Workspace  string       `json:"workspace"`
	SystemRole string       `json:"role"`
}

type UserListItem struct {
	Username   string       `json:"username"`
	UserId     string       `json:"userId"`
	Status     InviteStatus `json:"status"`
	SystemRole string       `json:"role"`
}

type InviteUserRequest struct {
	Username   string `json:"username"`
	Workspace  string `json:"workspace"`
	SystemRole string `json:"role"`
}

type EditUserRequest struct {
	Workspace  string `json:"workspace"`
	SystemRole string `json:"role"`
}

type WhoAmIResponse struct {
	Username   string `json:"username"`
	SystemRole string `json:"role"`
	Workspace  string `json:"workspace"`
}
