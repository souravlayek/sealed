package sealedservercomms

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/souravlayek/sealed/pkg/keymanager"
	"github.com/souravlayek/sealed/utils"
)

type KeyValuePair struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type SignaturePayload struct {
	Method        string
	PathWithQuery string
	Timestamp     int64
	Nonce         string
	DeviceID      string
	Body          string
}
type EnvelopReponse[T any] struct {
	Data T `json:"data"`
}
type DeviceAPIPayload struct {
	PublicKey  []byte `json:"public_key"`
	DeviceName string `json:"device_name"`
}

type DeviceCodeAPIResponse struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationURL string `json:"verification_url"`
}

func GetLoginInformation(payload DeviceAPIPayload) (*EnvelopReponse[DeviceCodeAPIResponse], error) {
	jsonifyPayload, _ := json.Marshal(payload)
	buf := bytes.NewBuffer(jsonifyPayload)
	resp, err := http.Post("https://api-sealed.souravlayek.com/api/v1/device/code", "application/json", buf)
	if err != nil {
		return nil, fmt.Errorf("An Error Occured %v", err)
	}

	var result EnvelopReponse[DeviceCodeAPIResponse]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

func createRequestSignature(req *http.Request, payload SignaturePayload, privateKey ed25519.PrivateKey) {
	hash := sha256.Sum256([]byte(payload.Body))
	hashedBody := hex.EncodeToString(hash[:])
	canonicalString := fmt.Sprintf("%s\n%s\n%s\n%s\n%s", payload.Method, payload.PathWithQuery, payload.Timestamp, payload.Nonce, hashedBody)
	signature := ed25519.Sign(privateKey, []byte(canonicalString))
	req.Header.Add("Accept", "application/json")
	req.Header.Set("X-Sealed-Device", payload.DeviceID)
	req.Header.Set("X-Sealed-Timestamp", strconv.Itoa(int(payload.Timestamp)))
	req.Header.Set("X-Sealed-Nonce", payload.Nonce)
	req.Header.Set("X-Sealed-Signature", base64.StdEncoding.EncodeToString(signature))

}

func CreateSignedHttpLoadRequest(projectSlug string, env string) []KeyValuePair {
	keyAndDeviceID := keymanager.GetDeviceIDAndPrivateKey()
	timestamp := time.Now().Unix()
	nonce := utils.GetNonce()
	pathWithQuery := fmt.Sprintf("/projects/%s/%s/secrets", projectSlug, env)
	requestURL := fmt.Sprintf("https://api-sealed.souravlayek.com/api/v1%s", pathWithQuery)
	req, err := http.NewRequest(http.MethodGet, requestURL, nil)
	if err != nil {
		log.Fatal(err)
	}
	createRequestSignature(req, SignaturePayload{
		Method:        "GET",
		PathWithQuery: pathWithQuery,
		Nonce:         nonce,
		Timestamp:     timestamp,
		Body:          "",
		DeviceID:      keyAndDeviceID.DeviceID,
	}, keyAndDeviceID.PrivateKey)
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	var result EnvelopReponse[[]KeyValuePair]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Fatal(err)
	}
	return result.Data
}
