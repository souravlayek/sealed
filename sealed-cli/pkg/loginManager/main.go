package loginmanager

import (
	"crypto/ed25519"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/souravlayek/sealed/pkg/keymanager"
	sealedservercomms "github.com/souravlayek/sealed/pkg/sealedServerComms"
)

func Login() {
	privateKey, err := keymanager.GetOrCreatePrivateKeyFromKeyring()
	if err != nil {
		log.Fatal(err)
	}
	publicKey := privateKey.Public()
	deviceName, err := os.Hostname()
	if err != nil {
		log.Fatal(err)
	}
	pubKeyPem, err := keymanager.ConvertPublicKeyToPem(publicKey.(ed25519.PublicKey))
	if err != nil {
		log.Fatal(err)
	}
	payload := sealedservercomms.DeviceAPIPayload{
		PublicKey:  pubKeyPem,
		DeviceName: deviceName,
	}
	result, err := sealedservercomms.GetLoginInformation(payload)

	fmt.Printf("Click the below url to continue with login \n %v \n And use this code: %v \n  ", result.Data.VerificationURL, result.Data.UserCode)

	// Wait for response
	interval := 4.0
	deviceCode := result.Data.DeviceCode
	pollRes, err := sealedservercomms.PollForTokenResponse(deviceCode)
	if err != nil {
		log.Fatal(err)
	}
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
		pollRes, err = sealedservercomms.PollForTokenResponse(deviceCode)
		if err != nil {
			log.Fatal(err)
		}
	}
	keymanager.UpdateKeyringToUseEmailAndDeviceID(pollRes.Data.Email, pollRes.Data.DeviceID)
	fmt.Println("\n Login successful as ", pollRes.Data.Email)
}
