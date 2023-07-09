package auth

import (
	"errors"
	"mini-bank/models"
	"mini-bank/repository"
	"mini-bank/utils"

	"github.com/gin-gonic/gin"
)

func Login(ctx *gin.Context) {
	var login LoginDTO
	var user models.User

	ctx.BindJSON(&login)
	err := repository.GetUserByIdentifier(&user, login.Identifier)
	if err != nil {
		utils.ResponseNotFound(ctx, err)
		return
	}

	err = utils.CheckPasswordHash(login.Password, user.Password)

	if err != nil {
		utils.ResponseUnauthorized(ctx, err)
		return
	}

	token, err := utils.GenerateToken(utils.TokenPayload{Id: user.Id, Email: user.Email})
	if err != nil {
		utils.ResponseUnauthorized(ctx, errors.New("failed to generate token"))
		return
	}

	utils.ResponseSuccess(ctx, gin.H{"token": token})
}

func Register(ctx *gin.Context) {
	var user models.User

	err := ctx.BindJSON(&user)
	if err != nil {
		utils.ResponseBadRequest(ctx, err)
		return
	}

	user.Password = utils.HashPassword(user.Password)

	err = repository.Create(&user)
	if err != nil {
		utils.ResponseBadRequest(ctx, err)
		return
	}

	//Auto login after register
	token, err := utils.GenerateToken(utils.TokenPayload{Id: user.Id, Email: user.Email})
	if err != nil {
		utils.ResponseUnauthorized(ctx, errors.New("failed to generate token"))
		return
	}

	// Strip password from response
	user.Password = ""

	utils.ResponseCreated(ctx, map[string]interface{}{"token": token, "user": user})
}

func GetProfile(ctx *gin.Context) {
	var user models.User

	id := ctx.MustGet("Id").(string)
	err := repository.GetUsersPermission(&user, id)

	if err != nil {
		utils.ResponseNotFound(ctx, err)
		return
	}

	user.Password = ""

	var permissions []string
	uniquePermissions := make(map[string]bool)

	for i, v := range user.Roles {
		for _, permission := range v.Permissions {
			if !uniquePermissions[permission.Name] {
				permissions = append(permissions, permission.Name)
				uniquePermissions[permission.Name] = true
			}
		}
		user.Roles[i].Permissions = nil
	}

	utils.ResponseSuccess(ctx, gin.H{"user": user, "permissions": permissions})
}
