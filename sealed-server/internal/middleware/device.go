package middleware

import (
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/souravlayek/sealed-server/internal/database"
	"github.com/souravlayek/sealed-server/pkg/logger"
	"github.com/souravlayek/sealed-server/utils"
)

func ValidateDeviceRequest(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		deviceIdStr := r.Header.Get("X-Sealed-Device")
		deviceId, err := strconv.Atoi(deviceIdStr)
		if err != nil {
			logger.Logger.Error(err)
			utils.SendErrorMessage(w, http.StatusBadRequest, "Invalid request")
			return
		}
		// IF not present return 401
		device, err := database.Q.GetDeviceById(r.Context(), int64(deviceId))
		if err != nil {
			logger.Logger.Error(err)
			utils.SendErrorMessage(w, http.StatusUnauthorized, "Unauthorized user")
			return
		}
		if device.RevokedAt != nil {
			logger.Logger.Error("Device revoked")
			utils.SendErrorMessage(w, http.StatusUnauthorized, "Unauthorized user")
			return
		}
		// Lookup device if revoked return 401
		ts := r.Header.Get("X-Sealed-Timestamp")
		signature := r.Header.Get("X-Sealed-Signature")
		// ts should be under 60 sec of server time or else return gateway timeout
		tsInt, err := strconv.ParseInt(ts, 10, 64)
		if err != nil {
			logger.Logger.Error(err)
			utils.SendErrorMessage(w, http.StatusBadRequest, "Invalid request")
			return
		}
		duration := time.Since(time.Unix(int64(tsInt), 0)).Seconds()
		if duration >= 60 {
			logger.Logger.Error("Duration exceeds")
			utils.SendErrorMessage(w, http.StatusGatewayTimeout, "Invalid request")
			return
		}
		nonce := r.Header.Get("X-Sealed-Nonce")

		// Rebuilds the canonical string and use pub key to validate signature
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			logger.Logger.Error(err)
			utils.SendErrorMessage(w, http.StatusBadRequest, "Invalid request")
			return
		}
		// Always close the original body
		r.Body.Close()
		hash := sha256.Sum256(bodyBytes)
		hashedBody := hex.EncodeToString(hash[:])
		canonicalString := fmt.Sprintf("%s\n%s\n%s\n%s\n%s", r.Method, r.URL.Path, tsInt, nonce, hashedBody)
		pubKey, err := parsePublicKey(device.Ed25519Pubkey)
		if err != nil {
			logger.Logger.Error(err)
			utils.SendErrorMessage(w, http.StatusUnauthorized, "Invlid key")
			return
		}
		signatureBytes, err := base64.StdEncoding.DecodeString(signature)
		if err != nil {
			logger.Logger.Error(err)
			utils.SendErrorMessage(w, http.StatusBadRequest, "Invalid request")
			return
		}
		isVerified := ed25519.Verify(pubKey, []byte(canonicalString), signatureBytes)
		if !isVerified {
			utils.SendErrorMessage(w, http.StatusBadRequest, "Invalid request")
			return
		}
		if exists := setNonceIfNotExist(nonce); exists {
			utils.SendErrorMessage(w, http.StatusBadRequest, "Invalid request")
			return
		}
		// Add userid too
		ctx := context.WithValue(
			r.Context(),
			"user_id",
			device.UserID,
		)

		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(fn)
}

func parsePublicKey(publicKey string) (ed25519.PublicKey, error) {
	pubBytes, err := hex.DecodeString(publicKey)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(pubBytes)
	if block == nil || block.Type != "PUBLIC KEY" {
		return nil, fmt.Errorf("invalid PEM data found")
	}
	parsedKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	edKey, ok := parsedKey.(ed25519.PublicKey)
	if !ok {
		return nil, fmt.Errorf("not an ed25519 private key")
	}

	return edKey, nil
}

func setNonceIfNotExist(nonce string) bool {
	key_prefix := "api:nonce:"
	key := key_prefix + nonce
	_, isExisting := database.Cache.Get(key)
	if isExisting {
		return true
	}
	database.Cache.Set(key, nonce, 10*time.Minute)
	return false
}
