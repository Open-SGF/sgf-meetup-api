package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"sgf-meetup-api/pkg/shared/appconfig"
	"sgf-meetup-api/pkg/shared/resource"
)

func main() {
	config, err := appconfig.NewCommonConfig(context.Background())
	if err != nil {
		log.Fatal("unable to create config")
	}

	cmds := []*exec.Cmd{
		exec.Command("npm", "run", "synth"),
	}

	flag.Parse()
	args := flag.Args()
	if len(args) < 1 {
		log.Fatal("Usage: go run main.go [api|importer]")
	}

	subcommand := args[0]
	samArgs := args[1:]

	switch subcommand {
	case "api":
		cmds = append(cmds, startAPI(config.AppEnv, samArgs...))
	case "importer":
		cmds = append(cmds, invokeImporter(config.AppEnv, samArgs...))
	default:
		log.Fatal("Unknown command. Use 'api' or 'importer'")
	}

	for _, cmd := range cmds {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			log.Fatalf("Command failed: %v", err)
		}
	}
}

func startAPI(appEnv string, samArgs ...string) *exec.Cmd {
	templateNamer := resource.NewNamer(appEnv, "SgfMeetupApi.template.json")

	templatePath := filepath.Join(
		"./cdk.out",
		templateNamer.FullName(),
	)

	args := []string{
		"local", "start-api",
		"-t", templatePath,
		"--docker-network", "sgf-meetup-api",
		"--env-vars", ".lambda-env.json",
	}

	args = append(args, samArgs...)

	return exec.Command("sam", args...)
}

func invokeImporter(appEnv string, samArgs ...string) *exec.Cmd {
	templateNamer := resource.NewNamer(appEnv, "SgfMeetupApi.template.json")

	templatePath := filepath.Join(
		"./cdk.out",
		templateNamer.FullName(),
	)

	functionNamer := resource.NewNamer(appEnv, "Importer")

	args := []string{
		"local", "invoke",
		"-t", templatePath,
		"--docker-network", "sgf-meetup-api",
		"--env-vars", ".lambda-env.json",
		functionNamer.FullName(),
	}

	args = append(args, samArgs...)

	return exec.Command("sam", args...)
}
