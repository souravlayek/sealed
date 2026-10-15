package sealedservercomms

import (
	"bytes"
	"encoding/json"
	"net/http"
)

type PollingSuccessData struct {
	DeviceID int64  `json:"device_id"`
	Email    string `json:"email"`
}

type PollingReponse struct {
	Status string             `json:"status"` // authorization_pending | slow_down | access_denied | expired_token | done
	Data   PollingSuccessData `json:"data"`
}

func PollForTokenResponse(deviceCode string) (PollingReponse, error) {
	payload := map[string]string{
		"device_code": deviceCode,
	}
	jsonifyPayload, err := json.Marshal(payload)
	buf := bytes.NewBuffer(jsonifyPayload)
	resp, err := http.Post("https://api-sealed.souravlayek.com/api/v1/device/token", "application/json", buf)
	if err != nil {
		return PollingReponse{}, err
	}
	var result EnvelopReponse[PollingReponse]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return PollingReponse{}, err
	}
	return result.Data, nil
}
