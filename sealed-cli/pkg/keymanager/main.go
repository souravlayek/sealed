package keymanager

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/99designs/keyring"
	"github.com/souravlayek/sealed/constants"
)

type Keys struct {
	PrivateKey []byte
	PublicKey  []byte
}

func getED25519PrivateKeyAsPem() (*Keys, error) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("Unable to create key %v", err)
	}
	privateKeyPem, err := ConvertPrivateKeyToPem(privateKey)
	if err != nil {
		return nil, fmt.Errorf("Unable to create key %v", err)
	}

	publicKeyPem, err := ConvertPublicKeyToPem(publicKey)
	if err != nil {
		return nil, fmt.Errorf("Unable to create key %v", err)
	}

	return &Keys{
		PrivateKey: privateKeyPem,
		PublicKey:  publicKeyPem,
	}, nil
}

func ConvertPublicKeyToPem(publicKey ed25519.PublicKey) ([]byte, error) {
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

func ConvertPrivateKeyToPem(privateKey ed25519.PrivateKey) ([]byte, error) {
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

func GetOrCreatePrivateKeyFromKeyring() (ed25519.PrivateKey, error) {
	ring, err := keyring.Open(keyring.Config{
		ServiceName: constants.SERVICE,
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
			Key:  constants.KEY_NAME,
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

func UpdateKeyringToUseEmailAndDeviceID(email string, deviceID int64) {
	ring, err := keyring.Open(keyring.Config{
		ServiceName: constants.SERVICE,
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
	item, err := ring.Get(constants.KEY_NAME)
	if errors.Is(err, keyring.ErrKeyNotFound) {
		return keyring.Item{}, false
	}
	if err != nil {
		log.Fatal(err)
	}
	return item, true
}

type DeviceIDAndPrivateKeyReponse struct {
	DeviceID   string
	PrivateKey ed25519.PrivateKey
}

func GetDeviceIDAndPrivateKey() DeviceIDAndPrivateKeyReponse {
	ring, err := keyring.Open(keyring.Config{
		ServiceName: constants.SERVICE,
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
	return DeviceIDAndPrivateKeyReponse{
		DeviceID:   deviceId,
		PrivateKey: privatekey,
	}
}
