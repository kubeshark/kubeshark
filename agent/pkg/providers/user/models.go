package user

type InviteStatus string

const (
	PendingInviteStatus  InviteStatus = "pending"
	AcceptedInviteStatus InviteStatus = "accepted"
)

type TokenResponse struct {
	Token string `json:"token"`
}

type User struct {
	Username     string       `json:"username"`
	UserId       string       `json:"userId"`
	InviteStatus InviteStatus `json:"inviteStatus"`
	Workspace    string       `json:"workspace"`
	SystemRole   string       `json:"systemRole"`
}

type UserListItem struct {
	Username     string       `json:"username"`
	UserId       string       `json:"userId"`
	InviteStatus InviteStatus `json:"inviteStatus"`
}

type InviteUserRequest struct {
	Username   string `json:"username"`
	Workspace  string `json:"workspace"`
	SystemRole string `json:"systemRole"`
}

type EditUserRequest struct {
	Workspace  string `json:"workspace"`
	SystemRole string `json:"systemRole"`
}

type WhoAmIResponse struct {
	Username   string `json:"username"`
	SystemRole string `json:"systemRole"`
	Workspace  string `json:"workspace"`
}
