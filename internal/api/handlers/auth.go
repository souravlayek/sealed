package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/souravlayek/sealed-server/internal/database"
	"github.com/souravlayek/sealed-server/internal/middleware"
	"github.com/souravlayek/sealed-server/pkg/auth"
	"github.com/souravlayek/sealed-server/pkg/logger"
	"github.com/souravlayek/sealed-server/pkg/sqlc"
	"github.com/souravlayek/sealed-server/utils"
)

// AuthResponse is the response for authentication endpoints.
type AuthResponse struct {
	Message      string `json:"message"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type UserProfileResponse struct {
	ID    int64  `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

// RegisterUser handles new user registration.
//
//	@Summary		Register user
//	@Description	Register a new user.
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		sqlc.CreateUserParams	true	"Registration details"
//	@Success		200		{object}	utils.EnvelopResponse[handlers.AuthResponse]
//	@Failure		400		{object}	utils.ErrResponse
//	@Failure		401		{object}	utils.ErrResponse
//	@Router			/v1/auth/register [post]
func RegisterUser(w http.ResponseWriter, r *http.Request) {
	var userRequest sqlc.CreateUserParams
	if err := json.NewDecoder(r.Body).Decode(&userRequest); err != nil {
		utils.SendErrorMessage(w, http.StatusBadRequest, "Invalid request format")
		return
	}
	hashedPassword, err := auth.HashPassword(userRequest.Password)
	if err != nil || len(userRequest.Password) < 8 {
		utils.SendErrorMessage(w, http.StatusBadRequest, "Invalid password")
		return
	}
	userRequest.Password = hashedPassword
	user, err := database.Q.CreateUser(r.Context(), userRequest)
	if err != nil {
		logger.Logger.Errorf("%v", err)
		utils.SendErrorMessage(w, http.StatusBadRequest, "Unable to create user")
		return
	}
	continueInviteWork := true
	userInInvited, err := database.Q.GetProjectInviteByEmail(r.Context(), user.Email)
	if err != nil && errors.Is(err, pgx.ErrNoRows) {
		logger.Logger.Errorf("Unable to load invite %v", err)
		continueInviteWork = false
	}
	if continueInviteWork {
		_, err = database.Q.AddProjectMember(r.Context(), sqlc.AddProjectMemberParams{
			ProjectID: userInInvited.ProjectID,
			UserID:    user.ID,
			Role:      ROLE_PROJECT_MEMBER,
		})
		if err != nil {
			logger.Logger.Errorf("Failed to add project member %v", err)
		}
		// Delete from invite table
		_, err = database.Q.DeleteProjectInviteByID(r.Context(), userInInvited.ID)
		if err != nil {
			logger.Logger.Errorf("Failed to delete member invite for id: %d -- %v", userInInvited.ID, err)
		}
	}

	accessToken, err := auth.CreateAccessToken(user.ID, user.Email, user.Role)
	if err != nil {
		logger.Logger.Errorf("%v", err)
		utils.SendErrorMessage(w, http.StatusBadRequest, "Fail to create token")
		return
	}

	refreshToken, err := auth.CreateRefreshToken(user.ID, user.Email, user.Role)
	if err != nil {
		logger.Logger.Errorf("%v", err)
		utils.SendErrorMessage(w, http.StatusBadRequest, "Fail to create token")
		return
	}

	response := AuthResponse{
		Message:      "User created successsfully",
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	utils.SendResponse(w, response, http.StatusOK)
}

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// SigninUser handles user login.
//
//	@Summary		User login
//	@Description	Authenticates a user with email and password, returns an access token.
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		handlers.LoginInput	true	"Login credentials"
//	@Success		200		{object}	utils.EnvelopResponse[handlers.AuthResponse]
//	@Failure		400		{object}	utils.ErrResponse
//	@Failure		401		{object}	utils.ErrResponse
//	@Failure		500		{object}	utils.ErrResponse
//	@Router			/v1/auth/signin [post]
func SigninUser(w http.ResponseWriter, r *http.Request) {
	var login LoginInput

	if err := json.NewDecoder(r.Body).Decode(&login); err != nil {
		utils.SendErrorMessage(
			w,
			http.StatusBadRequest,
			"invalid request body",
		)
		return
	}

	user, err := database.Q.GetUserByEmail(
		r.Context(),
		login.Email,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			utils.SendErrorMessage(
				w,
				http.StatusUnauthorized,
				"invalid email or password",
			)
			return
		}

		logger.Logger.Errorf(
			"failed to authenticate user: %v",
			err,
		)

		utils.SendErrorMessage(
			w,
			http.StatusInternalServerError,
			"failed to authenticate user",
		)
		return
	}

	if err := auth.CheckPassword(
		login.Password,
		user.Password,
	); err != nil {
		utils.SendErrorMessage(
			w,
			http.StatusUnauthorized,
			"invalid email or password",
		)
		return
	}

	accessToken, err := auth.CreateAccessToken(
		user.ID,
		user.Email,
		user.Role,
	)
	if err != nil {
		logger.Logger.Errorf(
			"failed to generate access token: %v",
			err,
		)

		utils.SendErrorMessage(
			w,
			http.StatusInternalServerError,
			"failed to generate access token",
		)
		return
	}

	refreshToken, err := auth.CreateRefreshToken(
		user.ID,
		user.Email,
		user.Role,
	)
	if err != nil {
		logger.Logger.Errorf(
			"failed to generate refresh token: %v",
			err,
		)

		utils.SendErrorMessage(
			w,
			http.StatusInternalServerError,
			"failed to generate refresh token",
		)
		return
	}

	utils.SendResponse(
		w,
		AuthResponse{
			Message:      "Login successful",
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		},
		http.StatusOK,
	)
}

// GetUser returns the authenticated user's profile.
//
//	@Summary		Get current user profile
//	@Description	Returns the profile of the authenticated user.
//	@Tags			auth
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	utils.EnvelopResponse[handlers.UserProfileResponse]
//	@Failure		401	{object}	utils.ErrResponse
//	@Failure		404	{object}	utils.ErrResponse
//	@Failure		500	{object}	utils.ErrResponse
//	@Router			/v1/auth/user [get]
func GetUser(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(
		middleware.UserClaimsKey,
	).(*auth.Payload)

	if !ok {
		utils.SendErrorMessage(
			w,
			http.StatusUnauthorized,
			"user authentication required",
		)
		return
	}

	user, err := database.Q.GetUserByID(
		r.Context(),
		int64(claims.Id),
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			utils.SendErrorMessage(
				w,
				http.StatusNotFound,
				"user not found",
			)
			return
		}

		logger.Logger.Errorf(
			"failed to get user by id %d: %v",
			claims.Id,
			err,
		)

		utils.SendErrorMessage(
			w,
			http.StatusInternalServerError,
			"failed to get user",
		)
		return
	}

	utils.SendResponse(
		w,
		UserProfileResponse{
			ID:    user.ID,
			Email: user.Email,
			Name:  user.Name,
		},
		http.StatusOK,
	)
}

// GetAccessToken handles access token refresh.
//
//	@Summary		Revalidate access token
//	@Description	Generates a new access token using a valid refresh token.
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	utils.EnvelopResponse[handlers.AuthResponse]
//	@Failure		400	{object}	utils.ErrResponse
//	@Failure		401	{object}	utils.ErrResponse
//	@Router			/v1/auth/revalidate-token [post]
func GetAccessToken(w http.ResponseWriter, r *http.Request) {
	authorizationHeader := r.Header.Get("Authorization")

	if len(authorizationHeader) == 0 ||
		len(strings.Split(authorizationHeader, " ")) != 2 {
		utils.SendErrorMessage(
			w,
			http.StatusUnauthorized,
			"Invalid token provided",
		)
		return
	}

	parts := strings.SplitN(authorizationHeader, " ", 2)

	if len(parts) != 2 || parts[0] != "Bearer" {
		utils.SendErrorMessage(
			w,
			http.StatusUnauthorized,
			"invalid authorization header",
		)
		return
	}

	accessToken, err := auth.GetAccessTokenFromRefreshToken(parts[1])
	if err != nil {
		utils.SendErrorMessage(
			w,
			http.StatusUnauthorized,
			"invalid refresh token",
		)
		return
	}

	utils.SendResponse(
		w,
		AuthResponse{
			Message:      "Revalidated token",
			AccessToken:  accessToken,
			RefreshToken: parts[1],
		},
		http.StatusOK,
	)
}
