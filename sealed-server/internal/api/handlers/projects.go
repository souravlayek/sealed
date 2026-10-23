package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/souravlayek/sealed-server/internal/database"
	"github.com/souravlayek/sealed-server/internal/middleware"
	"github.com/souravlayek/sealed-server/pkg/auth"
	"github.com/souravlayek/sealed-server/pkg/encryption"
	"github.com/souravlayek/sealed-server/pkg/logger"
	"github.com/souravlayek/sealed-server/pkg/resend"
	"github.com/souravlayek/sealed-server/pkg/sqlc"
	"github.com/souravlayek/sealed-server/utils"
)

const (
	ROLE_PROJECT_MEMBER = "PROJECT_MEMBER"
	ROLE_PROJECT_ADMIN  = "PROJECT_ADMIN"
)

// @Summary		Create project
// @Description	Create a new project.
// @Tags			projects
// @Accept			json
// @Produce		json
// @Security		BearerAuth
// @Param			request	body		sqlc.CreateProjectParams	true	"Project details"
// @Success		200		{object}	utils.EnvelopResponse[handlers.AuthResponse]
// @Failure		400		{object}	utils.ErrResponse
// @Failure		401		{object}	utils.ErrResponse
// @Router			/v1/projects/ [post]
func CreateProject(w http.ResponseWriter, r *http.Request) {
	var projectInput sqlc.CreateProjectParams
	if err := json.NewDecoder(r.Body).Decode(&projectInput); err != nil {
		utils.SendErrorMessage(w, http.StatusBadRequest, "Invalid request format")
		return
	}
	if len(projectInput.Name) < 3 {
		utils.SendErrorMessage(w, http.StatusBadRequest, "Invalid project name")
		return
	}
	// Start transaction
	tx, err := database.Pool.Begin(r.Context())
	if err != nil {
		logger.Logger.Errorf(
			"failed to start transaction: %v",
			err,
		)

		utils.SendErrorMessage(
			w,
			http.StatusInternalServerError,
			"failed to create account",
		)
		return
	}

	// Rollback if the transaction has not been committed.
	defer tx.Rollback(r.Context())

	// Create SQLC queries using the transaction.
	qtx := database.Q.WithTx(tx)
	project, err := qtx.CreateProject(r.Context(), projectInput)
	if err != nil {
		logger.Logger.Errorf("%v", err)
		utils.SendErrorMessage(w, http.StatusBadRequest, "Unable to create project")
		return
	}
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
	// Append to Project Member
	_, err = qtx.AddProjectMember(r.Context(), sqlc.AddProjectMemberParams{
		ProjectID: project.ID,
		UserID:    claims.Id,
		Role:      ROLE_PROJECT_ADMIN,
	})
	if err != nil {
		logger.Logger.Errorf("%v", err)
		utils.SendErrorMessage(w, http.StatusBadRequest, "Unable to create project")
		return
	}
	// Create Project Key
	dek, kek_version, err := encryption.GetDekAndKeyVersion(r.Context(), project.ID)
	if err != nil {
		logger.Logger.Errorf("%v", err)
		utils.SendErrorMessage(w, http.StatusBadRequest, "Unable to create project")
		return
	}

	_, err = qtx.CreateProjectKey(r.Context(), sqlc.CreateProjectKeyParams{
		ProjectID:  project.ID,
		DekWrapped: dek,
		KekVersion: kek_version,
	})
	if err != nil {
		logger.Logger.Errorf("%v", err)
		utils.SendErrorMessage(w, http.StatusBadRequest, "Unable to create project")
		return
	}
	// TODO: Create audit event

	// Commit both operations.
	if err := tx.Commit(r.Context()); err != nil {
		logger.Logger.Errorf(
			"failed to commit account creation: %v",
			err,
		)

		utils.SendErrorMessage(
			w,
			http.StatusInternalServerError,
			"failed to create account",
		)
		return
	}

	utils.SendResponse(w, project, http.StatusOK)
}

// @Summary		Get all projects
// @Description	Get all projects.
// @Tags			projects
// @Accept			json
// @Produce		json
// @Security		BearerAuth
// @Success		200	{object}	utils.EnvelopResponse[[]sqlc.ListProjectsRow]
// @Failure		400	{object}	utils.ErrResponse
// @Failure		404	{object}	utils.ErrResponse
// @Failure		401	{object}	utils.ErrResponse
// @Router			/v1/projects/ [get]
func GetAllProjects(w http.ResponseWriter, r *http.Request) {
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

	projects, err := database.Q.ListProjects(r.Context(), claims.Id)
	if err != nil {
		logger.Logger.Error(err)
		utils.SendErrorMessage(w, http.StatusNotFound, "Unable to find projects")
		return
	}
	utils.SendResponse(w, projects, http.StatusOK)
}

// @Summary		Get project details
// @Description	Get project with secrets.
// @Tags			projects
// @Accept			json
// @Produce		json
// @Security		BearerAuth
// @Param			slug	path		string	true	"Project Slug"
// @Success		200		{object}	utils.EnvelopResponse[sqlc.GetProjectByIDRow]
// @Failure		400		{object}	utils.ErrResponse
// @Failure		404		{object}	utils.ErrResponse
// @Failure		401		{object}	utils.ErrResponse
// @Router			/v1/projects/{slug} [get]
func GetProjectDetails(w http.ResponseWriter, r *http.Request) {
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
	slug := r.PathValue("slug")
	if strings.Trim(slug, " ") == "" {
		utils.SendErrorMessage(
			w,
			http.StatusBadRequest,
			"invalid project slug",
		)
		return
	}
	project, err := database.Q.GetProjectBySlug(r.Context(), sqlc.GetProjectBySlugParams{
		UserID: claims.Id,
		Slug:   slug,
	})
	if err != nil {
		logger.Logger.Error(err)
		utils.SendErrorMessage(
			w,
			http.StatusNotFound,
			"Failed to get project",
		)
		return
	}
	utils.SendResponse(w, project, http.StatusOK)
}

type KeyValuePair struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// @Summary		Get project secrets
// @Description	Get all project secrets that belongs to env.
// @Tags			projects
// @Accept			json
// @Produce		json
// @Param			projectSlug	path		string	true	"Project Slug"
// @Param			environment	path		string	true	"Environment"
// @Success		200			{object}	utils.EnvelopResponse[[]KeyValuePair]
// @Failure		400			{object}	utils.ErrResponse
// @Failure		404			{object}	utils.ErrResponse
// @Failure		401			{object}	utils.ErrResponse
// @Router			/v1/projects/{projectSlug}/{environment}/secrets [get]
func GetAllProjectSecretsForEnv(w http.ResponseWriter, r *http.Request) {
	projectSlug := r.PathValue("projectSlug")
	env := r.PathValue("environment")
	if strings.Trim(projectSlug, " ") == "" || strings.Trim(env, " ") == "" {
		utils.SendErrorMessage(w, http.StatusBadRequest, "Invalid project or environment")
		return
	}
	userId := r.Context().Value("user_id").(int64)
	secrets, err := database.Q.ListSecretsByProjectSlugAndEnvironment(r.Context(), sqlc.ListSecretsByProjectSlugAndEnvironmentParams{
		Slug:        projectSlug,
		UserID:      userId,
		Environment: env,
	})
	if err != nil {
		logger.Logger.Error(err)
		utils.SendErrorMessage(w, http.StatusBadRequest, "Invalid project or environment")
		return
	}
	if len(secrets) == 0 {
		utils.SendResponse(w, secrets, http.StatusOK)
		return
	}
	projectKeys, err := database.Q.GetProjectKeyByProjectID(r.Context(), secrets[0].ProjectID)
	if err != nil {
		logger.Logger.Error(err)
		utils.SendErrorMessage(w, http.StatusBadRequest, "Invalid project or environment")
		return
	}
	keyAndSecret := make([]KeyValuePair, 0)
	for _, secret := range secrets {
		item := KeyValuePair{}
		item.Key = secret.KeyName
		decryptedValue, err := encryption.DecryptSecret(r.Context(), secret.ProjectID, *secret.Nonce, projectKeys.DekWrapped, *secret.Ciphertext)
		if err != nil {
			utils.SendErrorMessage(w, http.StatusBadRequest, fmt.Sprintf("Found broken secret for %s", secret.KeyName))
			return
		}
		item.Value = decryptedValue
		keyAndSecret = append(keyAndSecret, item)
	}
	utils.SendResponse(w, keyAndSecret, http.StatusOK)
}

// @Summary		Delete project
// @Description	Delete project.
// @Tags			projects
// @Accept			json
// @Produce		json
// @Security		BearerAuth
// @Param			id	path		int	true	"Project ID"
// @Success		200	{object}	utils.EnvelopResponse[string]
// @Failure		400	{object}	utils.ErrResponse
// @Failure		404	{object}	utils.ErrResponse
// @Failure		401	{object}	utils.ErrResponse
// @Router			/v1/projects/{id} [delete]
func DeleteProject(w http.ResponseWriter, r *http.Request) {
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
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		utils.SendErrorMessage(
			w,
			http.StatusBadRequest,
			"invalid project id",
		)
		return
	}
	project, err := database.Q.GetProjectByID(r.Context(), sqlc.GetProjectByIDParams{
		UserID: claims.Id,
		ID:     int64(id),
	})
	if err != nil {
		logger.Logger.Error(err)
		utils.SendErrorMessage(
			w,
			http.StatusNotFound,
			"Failed to get project",
		)
		return
	}
	if project.MyRole != ROLE_PROJECT_ADMIN {
		utils.SendErrorMessage(
			w,
			http.StatusBadRequest,
			"You're not allowed to delete the project",
		)
		return
	}

	_, err = database.Q.SoftDeleteProject(r.Context(), project.ID)
	if err != nil {
		logger.Logger.Error(err)
		utils.SendErrorMessage(
			w,
			http.StatusNotFound,
			"Failed to get project",
		)
		return
	}
	utils.SendResponse(w, "Project deleted successfully", http.StatusOK)
}

type CreateProjectSecretParams struct {
	Environment  string `json:"environment"`
	KeyName      string `json:"key_name"`
	Secret       string `json:"secret"`
	RotationDays int64  `json:"rotation_days"`
}

// @Summary		Create project secret
// @Description	Create project secret.
// @Tags			projects
// @Accept			json
// @Produce		json
// @Security		BearerAuth
// @Param			id		path		int									true	"Project ID"
// @Param			request	body		handlers.CreateProjectSecretParams	true	"Secret to create"
// @Success		200		{object}	utils.EnvelopResponse[string]
// @Failure		400		{object}	utils.ErrResponse
// @Failure		404		{object}	utils.ErrResponse
// @Failure		401		{object}	utils.ErrResponse
// @Router			/v1/projects/{id}/secret [post]
func CreateProjectSecret(w http.ResponseWriter, r *http.Request) {
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
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		utils.SendErrorMessage(
			w,
			http.StatusBadRequest,
			"invalid project id",
		)
		return
	}
	var payload CreateProjectSecretParams
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		utils.SendErrorMessage(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	if strings.Contains(payload.KeyName, " ") {
		utils.SendErrorMessage(w, http.StatusBadRequest, "Invalid key name, There should no space in key")
		return
	}
	projectKeys, err := database.Q.GetProjectKey(r.Context(), sqlc.GetProjectKeyParams{
		ProjectID: int64(id),
		UserID:    claims.Id,
	})

	if err != nil {
		utils.SendErrorMessage(w, http.StatusNotFound, "Unable to find project")
		return
	}
	if !slices.Contains(projectKeys.Environments, payload.Environment) {
		utils.SendErrorMessage(w, http.StatusBadRequest, "Invalid environment")
		return
	}

	result, err := encryption.EncryptSecret(r.Context(), int64(id), payload.Secret, projectKeys.DekWrapped)
	if err != nil {
		logger.Logger.Error(err)
		utils.SendErrorMessage(w, http.StatusNotAcceptable, "Unable to create secret")
		return
	}
	changeTime := time.Now()
	secretParams := sqlc.CreateSecretParams{
		ProjectID:      int64(id),
		Environment:    payload.Environment,
		RotationDays:   payload.RotationDays,
		KeyName:        payload.KeyName,
		KekVersion:     &projectKeys.KekVersion,
		Ciphertext:     &result.Ciphertext,
		Nonce:          &result.Nonce,
		ValueChangedAt: &changeTime,
	}
	_, err = database.Q.CreateSecret(r.Context(), secretParams)
	if err != nil {
		logger.Logger.Error(err)
		utils.SendErrorMessage(w, http.StatusBadRequest, "Unable to create secret")
		return
	}
	utils.SendResponse(w, "Secret created successfully", http.StatusOK)
}

// @Summary		Update project secret
// @Description	Update project secret.
// @Tags			projects
// @Accept			json
// @Produce		json
// @Security		BearerAuth
// @Param			id		path		int									true	"Project ID"
// @Param			request	body		handlers.CreateProjectSecretParams	true	"Secret to create"
// @Success		200		{object}	utils.EnvelopResponse[string]
// @Failure		400		{object}	utils.ErrResponse
// @Failure		404		{object}	utils.ErrResponse
// @Failure		401		{object}	utils.ErrResponse
// @Router			/v1/projects/{id}/secret/{secretId} [put]
func UpdateProjectSecret(w http.ResponseWriter, r *http.Request) {
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
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		utils.SendErrorMessage(
			w,
			http.StatusBadRequest,
			"invalid project id",
		)
		return
	}
	secretId, err := strconv.Atoi(r.PathValue("secretId"))
	if err != nil {
		utils.SendErrorMessage(
			w,
			http.StatusBadRequest,
			"invalid secret id",
		)
		return
	}
	var payload CreateProjectSecretParams
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		utils.SendErrorMessage(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	if strings.Contains(payload.KeyName, " ") {
		utils.SendErrorMessage(w, http.StatusBadRequest, "Invalid key name, There should no space in key")
		return
	}
	projectKeys, err := database.Q.GetProjectKey(r.Context(), sqlc.GetProjectKeyParams{
		ProjectID: int64(id),
		UserID:    claims.Id,
	})

	if err != nil {
		utils.SendErrorMessage(w, http.StatusNotFound, "Unable to find project")
		return
	}
	if !slices.Contains(projectKeys.Environments, payload.Environment) {
		utils.SendErrorMessage(w, http.StatusBadRequest, "Invalid environment")
		return
	}

	result, err := encryption.EncryptSecret(r.Context(), int64(id), payload.Secret, projectKeys.DekWrapped)
	if err != nil {
		logger.Logger.Error(err)
		utils.SendErrorMessage(w, http.StatusNotAcceptable, "Unable to create secret")
		return
	}
	changeTime := time.Now()
	secretParams := sqlc.UpdateSecretParams{
		ID:             int64(secretId),
		RotationDays:   payload.RotationDays,
		KekVersion:     &projectKeys.KekVersion,
		Ciphertext:     &result.Ciphertext,
		Nonce:          &result.Nonce,
		ValueChangedAt: &changeTime,
	}
	_, err = database.Q.UpdateSecret(r.Context(), secretParams)
	if err != nil {
		logger.Logger.Error(err)
		utils.SendErrorMessage(w, http.StatusBadRequest, "Unable to create secret")
		return
	}
	utils.SendResponse(w, "Secret updated successfully", http.StatusOK)
}

type AddEnvironmentToProjectPayload struct {
	Environment string `json:"environment"`
}

// @Summary		Add environment to project
// @Description	Add environment to project.
// @Tags			projects
// @Accept			json
// @Produce		json
// @Security		BearerAuth
// @Param			id		path		int								true	"Project ID"
// @Param			request	body		AddEnvironmentToProjectPayload	true	"Environemnt"
// @Success		200		{object}	utils.EnvelopResponse[string]
// @Failure		400		{object}	utils.ErrResponse
// @Failure		404		{object}	utils.ErrResponse
// @Failure		401		{object}	utils.ErrResponse
// @Router			/v1/projects/{id}/environment [put]
func AddEnvironmentToProject(w http.ResponseWriter, r *http.Request) {
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
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		utils.SendErrorMessage(
			w,
			http.StatusBadRequest,
			"invalid project id",
		)
		return
	}
	var payload AddEnvironmentToProjectPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		utils.SendErrorMessage(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	if strings.Trim(payload.Environment, " ") == "" {
		utils.SendErrorMessage(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	project, err := database.Q.GetProjectByID(r.Context(), sqlc.GetProjectByIDParams{
		UserID: claims.Id,
		ID:     int64(id),
	})
	if err != nil {
		utils.SendErrorMessage(w, http.StatusNotFound, "Unable to found project")
		return
	}
	if slices.Contains(project.Environments, payload.Environment) {
		utils.SendErrorMessage(w, http.StatusBadRequest, "Environment already exist")
		return
	}
	_, err = database.Q.UpdateProjectEnvironments(r.Context(), sqlc.UpdateProjectEnvironmentsParams{
		ID:           int64(id),
		Environments: append(project.Environments, payload.Environment),
	})
	if err != nil {
		utils.SendErrorMessage(w, http.StatusBadRequest, "Unable to add environment")
		return
	}
	utils.SendResponse(w, "Successfully added environment", http.StatusOK)
}

// @Summary		Delete environment from project
// @Description	Deltete environment from project.
// @Tags			projects
// @Accept			json
// @Produce		json
// @Security		BearerAuth
// @Param			id		path		int								true	"Project ID"
// @Param			request	body		AddEnvironmentToProjectPayload	true	"Environemnt"
// @Success		200		{object}	utils.EnvelopResponse[string]
// @Failure		400		{object}	utils.ErrResponse
// @Failure		404		{object}	utils.ErrResponse
// @Failure		401		{object}	utils.ErrResponse
// @Router			/v1/projects/{id}/environment/{environment} [delete]
func DeleteEnvironment(w http.ResponseWriter, r *http.Request) {
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
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		utils.SendErrorMessage(
			w,
			http.StatusBadRequest,
			"invalid project id",
		)
		return
	}
	env := r.PathValue("environment")
	if strings.Trim(env, " ") == "" {
		utils.SendErrorMessage(w, http.StatusBadRequest, "Invalid environment")
		return
	}
	project, err := database.Q.GetProjectByID(r.Context(), sqlc.GetProjectByIDParams{
		UserID: claims.Id,
		ID:     int64(id),
	})
	if err != nil {
		utils.SendErrorMessage(w, http.StatusNotFound, "Unable to found project")
		return
	}
	if !slices.Contains(project.Environments, env) {
		utils.SendErrorMessage(w, http.StatusBadRequest, "Environment doesn't exist")
		return
	}
	_, err = database.Q.UpdateProjectEnvironments(r.Context(), sqlc.UpdateProjectEnvironmentsParams{
		ID: int64(id),
		Environments: slices.DeleteFunc(project.Environments, func(environment string) bool {
			return environment == env
		}),
	})
	if err != nil {
		utils.SendErrorMessage(w, http.StatusBadRequest, "Unable to delete environment")
		return
	}
	database.Q.DeleteSecretByEnvironment(r.Context(), sqlc.DeleteSecretByEnvironmentParams{
		ProjectID:   project.ID,
		Environment: env,
	})
	utils.SendResponse(w, "Successfully deleted environment", http.StatusOK)
}

type InviteProjectMemberPayload struct {
	Email string `json:"email"`
}

// @Summary		Invite a new member
// @Description	Invite a project member
// @Tags			projects
// @Accept			json
// @Produce		json
// @Security		BearerAuth
// @Param			id		path		int									true	"Project ID"
// @Param			request	body		handlers.InviteProjectMemberPayload	true	"Environemnt"
// @Success		200		{object}	utils.EnvelopResponse[string]
// @Failure		400		{object}	utils.ErrResponse
// @Failure		404		{object}	utils.ErrResponse
// @Failure		401		{object}	utils.ErrResponse
// @Router			/v1/projects/{id}/invite-member [post]
func InviteProjectMember(w http.ResponseWriter, r *http.Request) {
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
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		utils.SendErrorMessage(
			w,
			http.StatusBadRequest,
			"invalid project id",
		)
		return
	}
	var payload InviteProjectMemberPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		utils.SendErrorMessage(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	if strings.Trim(payload.Email, " ") == "" {
		utils.SendErrorMessage(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	project, err := database.Q.GetProjectByID(r.Context(), sqlc.GetProjectByIDParams{
		UserID: claims.Id,
		ID:     int64(id),
	})
	if err != nil {
		utils.SendErrorMessage(w, http.StatusNotFound, "Unable to found project")
		return
	}
	if project.MyRole != ROLE_PROJECT_ADMIN {
		utils.SendErrorMessage(w, http.StatusBadRequest, "You don't have the permission.")
		return
	}
	// If the user exist then add to Project and send invite email
	user, err := database.Q.GetUserByEmail(r.Context(), payload.Email)
	fmt.Println(err)
	if errors.Is(err, pgx.ErrNoRows) {
		logger.Logger.Infoln("User doesn't exist hance inviting the user")
		// If the user doesn't exist add to in-transit db and invite user, if the user signups and signup check in transit db if the user is there then add to project
		// TODO: WIP
		signupUrl := os.Getenv("FE_SIGNUP_URL")
		if signupUrl == "" {
			logger.Logger.Error("FE_SIGNUP_URL is not presend in env")
			utils.SendErrorMessage(w, http.StatusInternalServerError, "ENV not found")
			return
		}
		to := []string{payload.Email}
		subject := "You've been invited to join a project"
		html := fmt.Sprintf(`
		<!DOCTYPE html>
		<html>
  <body style="font-family: Arial, sans-serif; color: #111; line-height: 1.5;">
    <h2>You're invited</h2>
    <p>You've been invited to join <strong>%s</strong>.</p>
    <p>Create your account to accept the invite:</p>
    <p>
      <a href="%s"
         style="display:inline-block;padding:10px 18px;background:#2563eb;color:#fff;text-decoration:none;border-radius:6px;">
        Sign up
      </a>
    </p>
    <p>If the button doesn't work, copy this link into your browser:<br>
      <a href="%s">%s</a>
    </p>
    <p style="color:#666;font-size:12px;">If you weren't expecting this invite, you can ignore this email.</p>
  </body>
		</html>`, project.Name, signupUrl, signupUrl, signupUrl)

		err := resend.SendEmailRequest(to, subject, html)
		_, err = database.Q.CreateProjectInvite(r.Context(), sqlc.CreateProjectInviteParams{
			InvitedEmail: payload.Email,
			ProjectID:    project.ID,
		})
		if err != nil {
			logger.Logger.Errorf("Failed to create project invite, %v", err)
			utils.SendErrorMessage(w, http.StatusBadRequest, "Unable to invite")
			return
		}
		utils.SendResponse(w, "Member invited successfully", http.StatusOK)
		return
	}
	if err != nil {
		logger.Logger.Errorf("Failed to find user, %v", err)
		utils.SendErrorMessage(w, http.StatusBadRequest, "Unable to invite")
		return
	}
	_, err = database.Q.AddProjectMember(r.Context(), sqlc.AddProjectMemberParams{
		ProjectID: project.ID,
		UserID:    user.ID,
		Role:      ROLE_PROJECT_MEMBER,
	})
	if err != nil {
		logger.Logger.Errorf("Failed to Add member, %v", err)
		utils.SendErrorMessage(w, http.StatusBadRequest, "Unable to invite")
		return
	}
	utils.SendResponse(w, "Member added successfully", http.StatusOK)
}

// @Summary		Delete project secret
// @Description	Delete project secret.
// @Tags			projects
// @Accept			json
// @Produce		json
// @Security		BearerAuth
// @Param			id			path		int									true	"Project ID"
// @Param			secretKey	path		string								true	"Secret ID"
// @Param			environment	path		string								true	"Environment"
// @Param			request		body		handlers.CreateProjectSecretParams	true	"Secret to create"
// @Success		200			{object}	utils.EnvelopResponse[string]
// @Failure		400			{object}	utils.ErrResponse
// @Failure		404			{object}	utils.ErrResponse
// @Failure		401			{object}	utils.ErrResponse
// @Router			/v1/projects/{id}/secret/{environment}/{secretKey} [delete]
func DeleteProjectSecret(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		utils.SendErrorMessage(
			w,
			http.StatusBadRequest,
			"invalid project id",
		)
		return
	}
	secretKey := r.PathValue("secretKey")
	if strings.Trim(secretKey, " ") == "" {
		utils.SendErrorMessage(
			w,
			http.StatusBadRequest,
			"secret Key can't be empty",
		)
		return
	}
	environment := r.PathValue("environment")
	if strings.Trim(environment, " ") == "" {
		utils.SendErrorMessage(
			w,
			http.StatusBadRequest,
			"environment can't be empty",
		)
		return
	}
	_, err = database.Q.DeleteSecretByKey(r.Context(), sqlc.DeleteSecretByKeyParams{
		ProjectID:   int64(id),
		KeyName:     secretKey,
		Environment: environment,
	})
	if err != nil {
		logger.Logger.Error(err)
		utils.SendErrorMessage(w, http.StatusBadRequest, "Unable to delete secret")
		return
	}
	utils.SendResponse(w, "Secret deleted successfully", http.StatusOK)
}
