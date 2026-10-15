package main

import (
	"errors"
	"log"
	"os"
	"os/exec"
	"slices"

	loginmanager "github.com/souravlayek/sealed/pkg/loginManager"
	sealedservercomms "github.com/souravlayek/sealed/pkg/sealedServerComms"
	secretsmanager "github.com/souravlayek/sealed/pkg/secretsManager"
	"github.com/souravlayek/sealed/utils"
)

func main() {
	cliArgs := os.Args[1:]
	primaryCommand := cliArgs[0]
	if primaryCommand == "version" {
		utils.RunVersion()
		return
	}
	if primaryCommand == "login" {
		loginmanager.Login()
		return
	}
	if primaryCommand == "load" {
		loadArguments := cliArgs[1:]
		if len(loadArguments) < 2 {
			utils.PrintCommandHelp("load")
			log.Fatal("Invalid command")
			return
		}
		projectSlug := loadArguments[0]
		environment := loadArguments[1]
		result := sealedservercomms.CreateSignedHttpLoadRequest(projectSlug, environment)
		secretsmanager.SetEnvVariables(result)
		return
	}
	if primaryCommand == "run" {
		loadArguments := cliArgs[1:]
		if len(loadArguments) < 2 {
			utils.PrintCommandHelp("load")
			log.Fatal("Invalid command")
			return
		}
		projectSlug := loadArguments[0]
		environment := loadArguments[1]
		result := sealedservercomms.CreateSignedHttpLoadRequest(projectSlug, environment)
		idx := slices.Index(os.Args, "--")
		if idx == -1 || idx == len(os.Args)-1 {
			log.Fatal("usage: yourcli run ... -- <command> [args...]")
		}
		cmdArgs := os.Args[idx+1:]
		err := secretsmanager.RunWithSecrets(result, cmdArgs[0], cmdArgs[1:]...)
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			os.Exit(exitErr.ExitCode())
		}
		if err != nil {
			log.Fatal(err)
		}
		return
	}
	utils.PrintHelp()

}
