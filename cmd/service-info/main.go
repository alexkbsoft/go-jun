package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

var (
	errMissedCommandName  = errors.New("missed command name")
	errUnknownCommandName = errors.New("unknown command name")
	errMissedRequiredArg  = errors.New("missed require argument")
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	if len(os.Args) < 2 {
		return errMissedCommandName
	}

	ctx, cancel := signal.NotifyContext(
		context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	switch os.Args[1] {
	case "versions":
		return runVersionsCommand(ctx)

	case "is-deployed":
		return runIsDeployed(ctx)

	default:
		return errUnknownCommandName
	}
}

func parseServices(input string) ([]string, error) {
	parts := strings.Split(input, ",")
	var result []string
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result, nil
}

func runVersionsCommand(ctx context.Context) error {
	versionsCmd := flag.NewFlagSet("versions", flag.ExitOnError)
	var versionsServices string
	versionsCmd.StringVar(&versionsServices, "s", "", "Список сервисов через запятую")

	versionsCmd.Parse(os.Args[2:])
	services, err := parseServices(versionsServices)

	if err != nil {
		return err
	}

	if services == nil {
		services = allServices()
	}

	serviceStates := make([]serviceState, len(services))
	for i, s := range services {
		serviceStates[i] = fetchServiceState(ctx, s)
	}

	for _, s := range serviceStates {
		if s.unknown {
			fmt.Printf("\033[1;33m%s  [UNKNOWN]\033[0m\n", s.name)
			continue
		}
		if s.err != "" {
			fmt.Printf("\033[0;31m%s  [ERROR] %s\033[0m\n", s.name, s.err)
			continue
		}
		fmt.Printf("\033[0;32m%s  [OK] %s\033[0m\n", s.name, s.meta)
	}

	return nil
}

func runIsDeployed(ctx context.Context) error {
	versionsCmd := flag.NewFlagSet("versions", flag.ExitOnError)
	var versionsServices string
	var commit string
	versionsCmd.StringVar(&versionsServices, "s", "", "Список сервисов через запятую")
	versionsCmd.StringVar(&commit, "c", "", "commit hash")

	versionsCmd.Parse(os.Args[2:])
	services, err := parseServices(versionsServices)

	if err != nil {
		return err
	}

	if services == nil {
		services = allServices()
	}

	if commit == "" {
		return fmt.Errorf("%w: commit hash", errMissedRequiredArg)
	}
	commitMsg, err := commitMessage(ctx, commit)
	if err != nil {
		return fmt.Errorf("get commit message: %w", err)
	}

	type state struct {
		serviceState
		deployed bool
	}
	states := make([]state, len(services))

	for i, service := range services {
		state := state{serviceState: fetchServiceState(ctx, service)}
		serviceCommit := state.Commit()
		if serviceCommit != "" {
			deployed, err := isAncestor(ctx, commit, serviceCommit)
			if err != nil {
				state.err = "check is deployed: " + err.Error()
			} else {
				state.deployed = deployed
			}
		}

		states[i] = state
	}

	fmt.Printf("Looking up for commit %s:\n%s\n", commit, commitMsg)
	fmt.Println("---------------")

	for _, state := range states {
		if state.err != "" {
			fmt.Printf("\033[0;31m%s  [ERROR] (%v)\033[0m\n", state.name, state.err)
		} else if state.unknown {
			fmt.Printf("\033[1;33m%s  [UNKNOWN]\033[0m\n", state.name)
		} else if state.deployed {
			fmt.Printf("\033[0;32m%s  [DEPLOYED]\033[0m\n", state.name)
		} else {
			fmt.Printf("\033[0;37m%s  [NOT DEPLOYED] Current=%+v\033[0m\n", state.name, state.meta)
		}
	}

	return nil
}
