package secretsmanager

import (
	"fmt"
	"os"
	"os/exec"

	sealedservercomms "github.com/souravlayek/sealed/pkg/sealedServerComms"
)

func SetEnvVariables(pairs []sealedservercomms.KeyValuePair) {
	for _, kv := range pairs {
		fmt.Printf("export %s='%s'\n", kv.Key, kv.Value)
	}
}

func RunWithSecrets(pairs []sealedservercomms.KeyValuePair, name string, args ...string) error {
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
