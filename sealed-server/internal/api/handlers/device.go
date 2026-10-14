package handlers

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/souravlayek/sealed-server/internal/database"
	"github.com/souravlayek/sealed-server/internal/middleware"
	"github.com/souravlayek/sealed-server/pkg/auth"
	"github.com/souravlayek/sealed-server/pkg/encryption"
	"github.com/souravlayek/sealed-server/pkg/logger"
	"github.com/souravlayek/sealed-server/pkg/sqlc"
	"github.com/souravlayek/sealed-server/utils"
)

type DeviceCodeAPIPayload struct {
	PublicKey  []byte `json:"public_key"`
	DeviceName string `json:"device_name"`
}
type DeviceCodeAPIResponse struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationURL string `json:"verification_url"`
}

// @Summary		Create device code
// @Description	Create device code for login.
// @Tags			device
// @Accept			json
// @Produce		json
// @Param			request	body		DeviceCodeAPIPayload	true	"Device key"
// @Success		200		{object}	utils.EnvelopResponse[handlers.DeviceCodeAPIResponse]
// @Failure		400		{object}	utils.ErrResponse
// @Failure		401		{object}	utils.ErrResponse
// @Router			/v1/device/code [post]
func CreateDeviceCode(w http.ResponseWriter, r *http.Request) {
	var payload DeviceCodeAPIPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		utils.SendErrorMessage(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	deviceCode, err := encryption.GetRandomDeviceCode()
	if err != nil {
		logger.Logger.Errorln(err)
		utils.SendErrorMessage(w, http.StatusBadRequest, "Unable to process your request")
		return
	}

	userCode := utils.GenerateDeviceUserCode()
	_, err = database.Q.CreateDeviceEntryInTransit(r.Context(), sqlc.CreateDeviceEntryInTransitParams{
		DeviceCode: deviceCode,
		UserCode:   userCode,
		DeviceName: payload.DeviceName,
		PubKey:     hex.EncodeToString(payload.PublicKey),
		ExpiresAt:  time.Now().Add(10 * time.Minute),
	})
	if err != nil {
		logger.Logger.Errorln(err)
		utils.SendErrorMessage(w, http.StatusBadRequest, "Unable to process your request")
		return
	}
	response := DeviceCodeAPIResponse{
		DeviceCode:      deviceCode,
		UserCode:        userCode,
		VerificationURL: fmt.Sprintf("%s/verify?code=%s", os.Getenv("FE_HOSTNAME"), deviceCode),
	}
	utils.SendResponse(w, response, http.StatusOK)
}

type PollingSuccessData struct {
	DeviceID int64  `json:"device_id"`
	Email    string `json:"email"`
}

type PollingResponse struct {
	Status string             `json:"status"` // authorization_pending | slow_down | access_denied | expired_token | done
	Data   PollingSuccessData `json:"data"`
}
type ValidateTokenPayload struct {
	DeviceCode string `json:"device_code"`
}

// @Summary		Poll for device authorization
// @Description	Poll for device authorization status using device code.
// @Tags			device
// @Accept			json
// @Produce		json
// @Param			request	body		ValidateTokenPayload	true	"Device code"
// @Success		200		{object}	utils.EnvelopResponse[handlers.PollingResponse]
// @Failure		400		{object}	utils.ErrResponse
// @Failure		401		{object}	utils.ErrResponse
// @Router			/v1/device/token [post]
func ValidateToken(w http.ResponseWriter, r *http.Request) {
	var payload ValidateTokenPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		utils.SendErrorMessage(w, http.StatusBadRequest, "Invalid request format")
		return
	}
	result, found := database.Cache.Get(payload.DeviceCode)
	if !found {
		dbResult, err := database.Q.GetDeviceEntryInTransitByDeviceCode(r.Context(), payload.DeviceCode)
		if errors.Is(err, pgx.ErrNoRows) {
			deviceResult, deviceQueryErr := database.Q.GetDeviceByDeviceCode(r.Context(), &payload.DeviceCode)
			if deviceQueryErr != nil {
				utils.SendResponse(w, PollingResponse{
					Status: "access_denied",
					Data:   PollingSuccessData{},
				}, http.StatusOK)
				return
			}
			userResponse, userQueryErr := database.Q.GetUserByID(r.Context(), deviceResult.UserID)
			if userQueryErr != nil {
				utils.SendResponse(w, PollingResponse{
					Status: "access_denied",
					Data:   PollingSuccessData{},
				}, http.StatusOK)
				return
			}
			deviceId := deviceResult.ID
			utils.SendResponse(w, PollingResponse{
				Status: "done",
				Data: PollingSuccessData{
					Email:    userResponse.Email,
					DeviceID: deviceId,
				},
			}, http.StatusOK)
		}
		if err != nil {
			logger.Logger.Error(err)
			utils.SendResponse(w, PollingResponse{
				Status: "access_denied",
				Data:   PollingSuccessData{},
			}, http.StatusOK)
			return
		}
		database.Cache.Set(dbResult.DeviceCode, time.Now(), 10*time.Minute)
		utils.SendResponse(w, PollingResponse{
			Status: "authorization_pending",
			Data:   PollingSuccessData{},
		}, http.StatusOK)
		return
	}
	lastPolled := result.(time.Time)
	duration := time.Since(lastPolled)
	if duration.Seconds() < 10 {
		utils.SendResponse(w, PollingResponse{
			Status: "slow_down",
			Data:   PollingSuccessData{},
		}, http.StatusOK)
		return
	}
	utils.SendResponse(w, PollingResponse{
		Status: "authorization_pending",
		Data:   PollingSuccessData{},
	}, http.StatusOK)
}

type ValidateAuthenticationPayload struct {
	DeviceCode string `json:"device_code"`
	UserCode   string `json:"user_code"`
}

// @Summary		Validate device authentication
// @Description	Validate user code and register device for authenticated user.
// @Tags			device
// @Accept			json
// @Produce		json
// @Param			request	body		ValidateAuthenticationPayload	true	"Device code and user code"
// @Success		200		{object}	utils.EnvelopResponse[string]
// @Failure		400		{object}	utils.ErrResponse
// @Failure		401		{object}	utils.ErrResponse
// @Router			/v1/device/validate [post]
func ValidateAuthentication(w http.ResponseWriter, r *http.Request) {
	var payload ValidateAuthenticationPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		utils.SendErrorMessage(w, http.StatusBadRequest, "Invalid request format")
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
	inTransit, err := database.Q.GetDeviceEntryInTransitByDeviceCode(r.Context(), payload.DeviceCode)
	if err != nil || inTransit.UserCode != payload.UserCode {
		logger.Logger.Errorln(err)
		utils.SendErrorMessage(w, http.StatusBadRequest, "Unable to process your request")
		return
	}
	createDeviceParams := sqlc.CreateDeviceParams{
		UserID:        claims.Id,
		Ed25519Pubkey: inTransit.PubKey,
		DeviceName:    &inTransit.DeviceName,
		DeviceCode:    &inTransit.DeviceCode,
	}
	_, err = database.Q.CreateDevice(r.Context(), createDeviceParams)
	if err != nil || inTransit.UserCode != payload.UserCode {
		logger.Logger.Errorln(err)
		utils.SendErrorMessage(w, http.StatusBadRequest, "Unable to process your request")
		return
	}
	database.Cache.Delete(inTransit.DeviceCode)
	_, err = database.Q.DeleteDeviceEntryInTransitByDeviceCode(r.Context(), inTransit.DeviceCode)
	if err != nil || inTransit.UserCode != payload.UserCode {
		logger.Logger.Errorln(err)
		utils.SendErrorMessage(w, http.StatusBadRequest, "Unable to process your request")
		return
	}
	utils.SendResponse(w, "Successfully authenticated", http.StatusOK)
}

// @Summary		Get all devices
// @Description	Get all devices belonging to the authenticated user.
// @Tags			device
// @Accept			json
// @Produce		json
// @Success		200	{object}	utils.EnvelopResponse[[]sqlc.Device]
// @Failure		401	{object}	utils.ErrResponse
// @Router			/v1/device [get]
func GetAllDevices(w http.ResponseWriter, r *http.Request) {
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

	devices, err := database.Q.ListUserDevices(r.Context(), claims.Id)
	if err != nil {
		logger.Logger.Errorln(err)
		utils.SendErrorMessage(w, http.StatusBadRequest, "Unable to fetch devices")
		return
	}

	utils.SendResponse(w, devices, http.StatusOK)
}

type RevokeDevicePayload struct {
	DeviceID int64 `json:"device_id"`
}

// @Summary		Revoke a device
// @Description	Revoke a device belonging to the authenticated user.
// @Tags			device
// @Accept			json
// @Produce		json
// @Param			request	body		RevokeDevicePayload	true	"Device ID to revoke"
// @Success		200		{object}	utils.EnvelopResponse[sqlc.Device]
// @Failure		400		{object}	utils.ErrResponse
// @Failure		401		{object}	utils.ErrResponse
// @Failure		404		{object}	utils.ErrResponse
// @Router			/v1/device/revoke [post]
func RevokeDevice(w http.ResponseWriter, r *http.Request) {
	var payload RevokeDevicePayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		utils.SendErrorMessage(w, http.StatusBadRequest, "Invalid request format")
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

	// First verify the device belongs to the user
	device, err := database.Q.GetDeviceById(r.Context(), payload.DeviceID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			utils.SendErrorMessage(w, http.StatusNotFound, "Device not found")
			return
		}
		logger.Logger.Errorln(err)
		utils.SendErrorMessage(w, http.StatusBadRequest, "Unable to fetch device")
		return
	}

	if device.UserID != claims.Id {
		utils.SendErrorMessage(w, http.StatusNotFound, "Device not found")
		return
	}

	now := time.Now()
	revokedDevice, err := database.Q.RevokeDevice(r.Context(), sqlc.RevokeDeviceParams{
		ID:        payload.DeviceID,
		RevokedAt: &now,
	})
	if err != nil {
		logger.Logger.Errorln(err)
		utils.SendErrorMessage(w, http.StatusBadRequest, "Unable to revoke device")
		return
	}

	utils.SendResponse(w, revokedDevice, http.StatusOK)
}
