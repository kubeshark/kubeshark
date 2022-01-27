package controllers

import (
	"mizuserver/pkg/models"
	"mizuserver/pkg/providers"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/up9inc/mizu/shared/logger"

	ory "github.com/ory/kratos-client-go"
)

func Login(c *gin.Context) {
	if token, err := providers.PerformLogin(c.PostForm("username"), c.PostForm("password"), c.Request.Context()); err != nil {
		c.AbortWithStatusJSON(401, gin.H{"error": "bad login"})
	} else {
		c.JSON(200, gin.H{"token": token})
	}
}

func Logout(c *gin.Context) {
	token := c.GetHeader("x-session-token")
	if err := providers.Logout(token, c.Request.Context()); err != nil {
		logger.Log.Errorf("internal error while logging out %v", err)
		c.AbortWithStatusJSON(500, gin.H{"error": "error occured while logging out, the session might still be valid"})
	} else {
		c.JSON(200, "")
	}
}

func RecoverUserWithInviteToken(c *gin.Context) {
	token, err, formErrorMessages := providers.ResetPasswordWithInvite(c.PostForm("inviteToken"), c.PostForm("password"), c.Request.Context())
	handleRegistration(token, err, formErrorMessages, c)
}

func CreateUserAndInvite(c *gin.Context) {
	requestCreateUser := &models.InviteUserRequest{}

	if err := c.Bind(requestCreateUser); err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}

	if inviteToken, identityId, err := providers.CreateNewUserWithInvite(requestCreateUser.Username, requestCreateUser.Workspace, requestCreateUser.SystemRole, c.Request.Context()); err != nil {
		logger.Log.Errorf("internal error while creating user invite %v", err)
		c.JSON(http.StatusInternalServerError, err)
	} else {
		c.JSON(201, gin.H{"inviteToken": inviteToken, "userId": identityId})
	}
}

func CreateInviteForExistingUser(c *gin.Context) {
	if inviteToken, err := providers.CreateInvite(c.Param("userId"), c.Request.Context()); err != nil {
		logger.Log.Errorf("internal error while creating existing user invite %v", err)
		c.JSON(http.StatusInternalServerError, err)
	} else {
		c.JSON(201, gin.H{"inviteToken": inviteToken})
	}

}

func UpdateUser(c *gin.Context) {
	requestEditUser := &models.EditUserRequest{}

	if err := c.Bind(requestEditUser); err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}

	identityId := c.Param("userId")
	if err := providers.UpdateUserRoles(identityId, requestEditUser.Workspace, requestEditUser.SystemRole, c.Request.Context()); err != nil {
		logger.Log.Errorf("internal error while updating specific user %v", err)
		c.JSON(http.StatusInternalServerError, err)
	} else {
		c.JSON(200, "")
	}
}

func DeleteUser(c *gin.Context) {
	if err := providers.DeleteUser(c.Param("userId"), c.Request.Context()); err != nil {
		logger.Log.Errorf("internal error while deleting user %v", err)
		c.JSON(http.StatusInternalServerError, err)
	} else {
		c.JSON(200, "")
	}
}

func ListUsers(c *gin.Context) {
	if users, err := providers.ListUsers(c.Query("usernameFilter"), c.Request.Context()); err != nil {
		logger.Log.Errorf("internal error while listing users %v", err)
		c.JSON(http.StatusInternalServerError, err)
	} else {
		c.JSON(200, users)
	}
}

func GetUser(c *gin.Context) {
	if user, err := providers.GetUser(c.Param("userId"), c.Request.Context()); err != nil {
		logger.Log.Errorf("internal error while fetching specific user %v", err)
		c.JSON(http.StatusInternalServerError, err)
	} else {
		c.JSON(200, user)
	}
}

func handleRegistration(token *string, err error, formErrorMessages map[string][]ory.UiText, c *gin.Context) {
	if err != nil {
		if formErrorMessages != nil {
			logger.Log.Infof("user attempted to register but had form errors %v %v", formErrorMessages, err)
			c.AbortWithStatusJSON(400, formErrorMessages)
		} else {
			logger.Log.Errorf("unknown internal error registering user %v", err)
			c.AbortWithStatusJSON(500, gin.H{"error": "internal error occured while registering"})
		}
	} else {
		c.JSON(201, gin.H{"token": token})
	}
}
