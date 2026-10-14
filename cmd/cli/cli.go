package main

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/99designs/keyring"
)

const (
	SERVICE  = "sealed"
	USER     = "default"
	KEY_NAME = "sealed-ed25519-private-key"
)

type DeviceAPIPayload struct {
	PublicKey  []byte `json:"public_key"`
	DeviceName string `json:"device_name"`
}
type EnvelopReponse[T any] struct {
	Data T `json:"data"`
}

func main() {
	cliArgs := os.Args[1:]
	primaryCommand := cliArgs[0]
	if primaryCommand == "login" {
		handeCliLogin()
	} else if primaryCommand == "load" {
		loadArguments := cliArgs[1:]
		if len(loadArguments) < 2 {
			printCommandHelp("load")
			log.Fatal("Invalid command")
			return
		}
		projectSlug := loadArguments[0]
		environment := loadArguments[1]
		result := createSignedHttpLoadRequest(projectSlug, environment)
		setEnvVariables(result)
	} else if primaryCommand == "run" {
		loadArguments := cliArgs[1:]
		if len(loadArguments) < 2 {
			printCommandHelp("load")
			log.Fatal("Invalid command")
			return
		}
		projectSlug := loadArguments[0]
		environment := loadArguments[1]
		result := createSignedHttpLoadRequest(projectSlug, environment)
		idx := slices.Index(os.Args, "--")
		if idx == -1 || idx == len(os.Args)-1 {
			log.Fatal("usage: yourcli run ... -- <command> [args...]")
		}
		cmdArgs := os.Args[idx+1:]
		err := runWithSecrets(result, cmdArgs[0], cmdArgs[1:]...)
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			os.Exit(exitErr.ExitCode())
		}
		if err != nil {
			log.Fatal(err)
		}
	} else {
		printHelp()
	}
}

func createSignedHttpLoadRequest(projectSlug string, env string) []KeyValuePair {
	ring, err := keyring.Open(keyring.Config{
		ServiceName: SERVICE,
	})
	if err != nil {
		log.Fatal(err)
	}
	keyringItem, found := getFromKeyring(ring)
	if !found {
		log.Fatal("Unable to find key try `sealed login` first")
	}
	privatekey, err := parsePrivateKey(string(keyringItem.Data))
	if err != nil {
		log.Fatal(err)
	}
	splittedLabel := strings.Split(keyringItem.Label, "|")
	if len(splittedLabel) != 2 {
		log.Fatal("Invalid key try `sealed login` first")
	}
	deviceId := splittedLabel[1]
	timestamp := time.Now().Unix()
	nonce := getNonce()
	pathWithQuery := fmt.Sprintf("/projects/%s/%s/secrets", projectSlug, env)
	requestURL := fmt.Sprintf("http://localhost:8080/api/v1%s", pathWithQuery)
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
		DeviceID:      deviceId,
	}, privatekey)
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

type KeyValuePair struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func setEnvVariables(pairs []KeyValuePair) {
	for _, kv := range pairs {
		fmt.Printf("export %s='%s'\n", kv.Key, kv.Value)
	}
}

func runWithSecrets(pairs []KeyValuePair, name string, args ...string) error {
	env := os.Environ()
	for _, kv := range pairs {
		env = append(env, kv.Key+"="+kv.Value)
	}

	cmd := exec.Command(name, args...)
	cmd.Env = env
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

type SignaturePayload struct {
	Method        string
	PathWithQuery string
	Timestamp     int64
	Nonce         string
	DeviceID      string
	Body          string
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

func getNonce() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		log.Fatal(err)
	}
	return base64.RawURLEncoding.EncodeToString(bytes)
}

func handeCliLogin() {
	privateKey, err := getOrCreatePrivateKeyFromKeyring()
	if err != nil {
		log.Fatal(err)
	}
	publicKey := privateKey.Public()
	deviceName, err := os.Hostname()
	if err != nil {
		log.Fatal(err)
	}
	pubKeyPem, err := convertPublicKeyToPem(publicKey.(ed25519.PublicKey))
	if err != nil {
		log.Fatal(err)
	}
	payload := DeviceAPIPayload{
		PublicKey:  pubKeyPem,
		DeviceName: deviceName,
	}
	jsonifyPayload, _ := json.Marshal(payload)
	buf := bytes.NewBuffer(jsonifyPayload)
	resp, err := http.Post("http://localhost:8080/api/v1/device/code", "application/json", buf)
	if err != nil {
		log.Fatalf("An Error Occured %v", err)
	}
	type DeviceCodeAPIResponse struct {
		DeviceCode      string `json:"device_code"`
		UserCode        string `json:"user_code"`
		VerificationURL string `json:"verification_url"`
	}
	var result EnvelopReponse[DeviceCodeAPIResponse]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Fatal(err)
	}
	// fmt.Println(result)

	fmt.Printf("Click the below url for login \n %v \n And use this code: %v \n [DEBUG] Device-code: %v \n ", result.Data.VerificationURL, result.Data.UserCode, result.Data.DeviceCode)

	// Wait for response
	interval := 4.0
	deviceCode := result.Data.DeviceCode
	pollRes := pollForTokenResponse(deviceCode)
	fmt.Println("Waiting for your authentication...")
	for true {
		if pollRes.Status == "slow_down" {
			interval *= 1.5
		} else if pollRes.Status == "access_denied" {
			log.Fatal("Unable to login")
			return
		} else if pollRes.Status == "expired_token" {
			log.Fatal("Your token expired")
			return
		} else if pollRes.Status == "done" {
			break
		}
		time.Sleep(time.Duration(interval) * time.Second)
		pollRes = pollForTokenResponse(deviceCode)
	}
	updateKeyringToUseEmailAndDeviceID(pollRes.Data.Email, pollRes.Data.DeviceID)
	fmt.Println("\n Login successful as ", pollRes.Data.Email)
}

type PollingSuccessData struct {
	DeviceID int64  `json:"device_id"`
	Email    string `json:"email"`
}

type PollingReponse struct {
	Status string             `json:"status"` // authorization_pending | slow_down | access_denied | expired_token | done
	Data   PollingSuccessData `json:"data"`
}

func pollForTokenResponse(deviceCode string) PollingReponse {
	// fmt.Println(deviceCode)
	payload := map[string]string{
		"device_code": deviceCode,
	}
	jsonifyPayload, err := json.Marshal(payload)
	buf := bytes.NewBuffer(jsonifyPayload)
	resp, err := http.Post("http://localhost:8080/api/v1/device/token", "application/json", buf)
	if err != nil {
		log.Fatalf("An Error Occured %v", err)
	}
	var result EnvelopReponse[PollingReponse]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Fatal(err)
	}
	return result.Data
}

type Keys struct {
	PrivateKey []byte
	PublicKey  []byte
}

func getOrCreatePrivateKeyFromKeyring() (ed25519.PrivateKey, error) {
	ring, err := keyring.Open(keyring.Config{
		ServiceName: SERVICE,
	})
	if err != nil {
		log.Fatal(err)
	}
	item, isThere := getFromKeyring(ring)
	if !isThere {
		keys, err := getED25519PrivateKeyAsPem()
		if err != nil {
			log.Fatal(err)
		}
		err = ring.Set(keyring.Item{
			Key:  KEY_NAME,
			Data: keys.PrivateKey,
		})
		if err != nil {
			log.Fatal(err)
		}
		item, isThere = getFromKeyring(ring)
		if !isThere {
			log.Fatal("Unable to store key something catastrophic happend")
		}
	}
	privateKey, err := parsePrivateKey(string(item.Data))
	if err != nil {
		log.Fatal(err)
	}
	return privateKey, nil
}

func updateKeyringToUseEmailAndDeviceID(email string, deviceID int64) {
	ring, err := keyring.Open(keyring.Config{
		ServiceName: SERVICE,
	})
	if err != nil {
		log.Fatal(err)
	}
	item, isThere := getFromKeyring(ring)
	if !isThere {
		log.Fatal("Unalbe to find the key")
	}
	item.Label = fmt.Sprintf("%s|%d", email, deviceID)
	err = ring.Set(item)
	if err != nil {
		log.Fatal(err)
	}
}

func getFromKeyring(ring keyring.Keyring) (keyring.Item, bool) {
	item, err := ring.Get(KEY_NAME)
	if errors.Is(err, keyring.ErrKeyNotFound) {
		return keyring.Item{}, false
	}
	if err != nil {
		log.Fatal(err)
	}
	return item, true
}

func getED25519PrivateKeyAsPem() (*Keys, error) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("Unable to create key %v", err)
	}
	privateKeyPem, err := convertPrivateKeyToPem(privateKey)
	if err != nil {
		return nil, fmt.Errorf("Unable to create key %v", err)
	}

	publicKeyPem, err := convertPublicKeyToPem(publicKey)
	if err != nil {
		return nil, fmt.Errorf("Unable to create key %v", err)
	}

	return &Keys{
		PrivateKey: privateKeyPem,
		PublicKey:  publicKeyPem,
	}, nil
}

func parsePrivateKey(privateKey string) (ed25519.PrivateKey, error) {
	block, _ := pem.Decode([]byte(privateKey))
	if block == nil || block.Type != "PRIVATE KEY" {
		return nil, fmt.Errorf("invalid PEM data found")
	}
	parsedKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	edKey, ok := parsedKey.(ed25519.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("not an ed25519 private key")
	}

	return edKey, nil
}

func convertPrivateKeyToPem(privateKey ed25519.PrivateKey) ([]byte, error) {
	derBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal private key: %v", err)
	}
	// 2. Create a pem.Block struct
	block := &pem.Block{
		Type:  "PRIVATE KEY", // PKCS#8 generic private key header
		Bytes: derBytes,
	}
	return pem.EncodeToMemory(block), nil
}
func convertPublicKeyToPem(publicKey ed25519.PublicKey) ([]byte, error) {
	publicDerBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal public key: %v", err)
	}
	// 2. Create a pem.Block struct
	publicBlock := &pem.Block{
		Type:  "PUBLIC KEY", // PKCS#8 generic private key header
		Bytes: publicDerBytes,
	}
	return pem.EncodeToMemory(publicBlock), nil
}

// --------------------------------------
func printHelp() {
	fmt.Println(`sealed - secrets management for developers

Usage:
  sealed <command> [arguments]

Commands:
  login                    Authenticate via OAuth device flow
  load <project> <env>     Load secrets for a project/environment
  run <project> <env> --   Run a command with secrets injected

Run 'sealed <command> --help' for more information on a command.`)
}

func printCommandHelp(cmd string) {
	switch cmd {
	case "login":
		fmt.Println(`sealed login - Authenticate via OAuth device flow

Usage:
  sealed login

Opens a browser to authorize this device. The Ed25519 private key is stored
in your OS keychain (macOS Keychain, Windows Credential Manager, or
libsecret on Linux). No tokens are written to disk.`)
	case "load":
		fmt.Println(`sealed load - Load secrets for a project/environment

Usage:
  sealed load <project> <env> [KEY...]

Arguments:
  project    Project name or ID
  env        Environment: dev, staging, or prod
  KEY...     Optional specific keys to load (default: all)

Flags:
  --allow-missing    Exit 0 even if some keys are missing (for CI)
  --format=FORMAT    Output format: export (default), json, dotenv

Output:
  Prints 'export KEY='value'' lines (POSIX shell compatible).
  Values with newlines work; single quotes escaped as '\''.
  NUL bytes are rejected at write time.

Exit codes:
  0  Success (or --allow-missing with missing keys)
  3  Missing keys (prints warning to stderr)
  1  Auth, network, or other error`)
	case "run":
		fmt.Println(`sealed run - Run a command with secrets injected into environment

Usage:
  sealed run <project> <env> -- <command> [args...]

Arguments:
  project    Project name or ID
  env        Environment: dev, staging, or prod
  --         Separator: everything after is the command to run

This is the preferred way to use secrets - they are injected into the
child process environment and never printed to stdout/stderr.

Example:
  sealed run myapp prod -- npm start
  sealed run myapp staging -- ./deploy.sh`)
	default:
		fmt.Printf("Unknown command: %s\n\n", cmd)
		printHelp()
	}
}
