package encryption

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	openbao "github.com/openbao/openbao/api/v2"
)

func create_dek() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

const (
	TRANSIT_PATH = "transit"
	KEK_KEY_NAME = "my-kek"
)

// GenerateDeviceCode creates a secure, human-readable user code (e.g., ABCD-EFGH)
func GetRandomDeviceCode() (string, error) {
	bytes := make([]byte, 12)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func create_wrapped_dek(ctx context.Context, project_id int64) (string, error) {
	dek, err := create_dek()
	if err != nil {
		return "", err
	}
	client, err := get_open_bao()
	if err != nil {
		return "", err
	}

	// Endpoint: transit/encrypt/<key-name>
	encryptPath := fmt.Sprintf("%s/encrypt/%s", TRANSIT_PATH, KEK_KEY_NAME)
	base64dek := base64.StdEncoding.EncodeToString([]byte(dek))
	contextBase64 := base64.StdEncoding.EncodeToString([]byte(strconv.Itoa(int(project_id))))
	encryptResponse, err := client.Logical().WriteWithContext(ctx, encryptPath, map[string]interface{}{
		"plaintext": base64dek,
		"context":   contextBase64,
	})
	if err != nil {
		return "", err
	}
	ciphertext, ok := encryptResponse.Data["ciphertext"].(string)
	if !ok {
		return "", errors.New("Ciphertext missing from response")
	}
	return ciphertext, nil
}

func get_unwrapped_dek(ctx context.Context, ciphertext string, project_id int64) (string, error) {
	client, err := get_open_bao()
	if err != nil {
		return "", err
	}
	// Endpoint: transit/decrypt/<key-name>
	decryptPath := fmt.Sprintf("%s/decrypt/%s", TRANSIT_PATH, KEK_KEY_NAME)
	contextBase64 := base64.StdEncoding.EncodeToString([]byte(strconv.Itoa(int(project_id))))
	decryptResponse, err := client.Logical().WriteWithContext(ctx, decryptPath, map[string]interface{}{
		"ciphertext": ciphertext,
		"context":    contextBase64,
	})
	if err != nil {
		return "", err
	}
	// Extract the base64-encoded plaintext string from OpenBao
	base64Result, ok := decryptResponse.Data["plaintext"].(string)
	if !ok {
		return "", errors.New("Plaintext missing from response")
	}

	// Decode Base64 back to original string
	decodedBytes, err := base64.StdEncoding.DecodeString(base64Result)
	if err != nil {
		return "", err
	}

	return string(decodedBytes), nil
}

func get_open_bao() (*openbao.Client, error) {
	token := os.Getenv("BAO_TOKEN")
	address := os.Getenv("BAO_ADDR")
	if token == "" || address == "" {
		return nil, errors.New("Unable to read encryptor details")
	}
	config := openbao.DefaultConfig()
	config.Address = address

	client, err := openbao.NewClient(config)
	if err != nil {
		return nil, err
	}
	client.SetToken(token)
	return client, nil
}

func GetDekAndKeyVersion(ctx context.Context, project_id int64) (string, string, error) {
	wrapped_dek, err := create_wrapped_dek(ctx, project_id)
	if err != nil {
		return "", "", err
	}
	parts := strings.Split(wrapped_dek, ":")
	if len(parts) < 2 {
		return "", "", errors.New("Invalid ciphertext")
	}
	versionStr := parts[1]
	return wrapped_dek, versionStr, nil
}

type EncryptSecretResult struct {
	Ciphertext string `json:"ciphertext"`
	Nonce      string `json:"nonce"`
}

func EncryptSecret(ctx context.Context, project_id int64, secret string, dek string) (*EncryptSecretResult, error) {
	dek, err := get_unwrapped_dek(ctx, dek, project_id)
	if err != nil {
		return nil, err
	}
	key, err := hex.DecodeString(dek)
	if err != nil {
		return nil, err
	}
	secretByte := []byte(secret)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, aesgcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	ciphertext := aesgcm.Seal(nil, nonce, secretByte, nil)
	result := &EncryptSecretResult{
		Ciphertext: hex.EncodeToString(ciphertext),
		Nonce:      hex.EncodeToString(nonce),
	}
	return result, nil
}

func DecryptSecret(ctx context.Context, project_id int64, nonce string, dek string, cypherText string) (string, error) {
	decryptedDek, err := get_unwrapped_dek(ctx, dek, project_id)
	if err != nil {
		return "", err
	}
	key, err := hex.DecodeString(decryptedDek)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher block: %w", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}
	nonceByte, err := hex.DecodeString(nonce)
	if err != nil {
		return "", err
	}
	cypherTextByte, err := hex.DecodeString(cypherText)
	if err != nil {
		return "", err
	}
	plaintext, err := aesGCM.Open(nil, nonceByte, cypherTextByte, nil)
	if err != nil {
		return "", fmt.Errorf("decryption failed (tag mismatch or corrupted data): %w", err)
	}
	fmt.Println("HEre")
	return string(plaintext), nil
}
