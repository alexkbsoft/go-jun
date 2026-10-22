package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"sort"
	"strings"
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

	switch os.Args[1] {
	case "versions":
		return runVersionsCommand()

	case "is-deployed":
		return runIsDeployed()

	default:
		return errUnknownCommandName
	}

	return nil
}

func parseServices(input string, fs *flag.FlagSet) ([]string, error) {
	// Разбиваем строку и очищаем элементы от случайных пробелов
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

type serviceState struct {
	name    string
	unknown bool
	err     string
	meta    string
}

func runVersionsCommand() error {
	versionsCmd := flag.NewFlagSet("versions", flag.ExitOnError)
	var versionsServices string
	versionsCmd.StringVar(&versionsServices, "s", "", "Список сервисов через запятую")

	versionsCmd.Parse(os.Args[2:])
	services, err := parseServices(versionsServices, versionsCmd)

	if err != nil {
		return err
	}

	if services == nil {
		services = allServices()
	}

	serviceStates := make([]serviceState, len(services))
	for i, s := range services {
		serviceStates[i] = fetchServiceState(s)
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

var servicePorts = map[string]int{
	"api-server": 8090,
}

func allServices() []string {
	result := make([]string, 0, len(servicePorts))
	for name := range servicePorts {
		result = append(result, name)
	}
	sort.Strings(result)

	return result
}

func fetchServiceState(name string) serviceState {
	state := serviceState{name: name}
	port, ok := servicePorts[name]
	if !ok {
		state.unknown = true
		return state
	}

	url := fmt.Sprintf("http://127.0.0.1:%d/debug/info", port)
	resp, err := http.Get(url)
	if err != nil {
		state.err = fmt.Sprintf("get service info %v: ", err)
		return state
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		state.err = fmt.Sprintf("returned non-ok status code: %d", resp.StatusCode)
		return state
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		state.err = fmt.Sprintf("read response body: %v", err)
		return state
	}

	state.meta = string(body)

	return state
}

func runIsDeployed() error {
	versionsCmd := flag.NewFlagSet("versions", flag.ExitOnError)
	var versionsServices string
	var commit string
	versionsCmd.StringVar(&versionsServices, "s", "", "Список сервисов через запятую")
	versionsCmd.StringVar(&commit, "c", "", "commit hash")

	versionsCmd.Parse(os.Args[2:])
	services, err := parseServices(versionsServices, versionsCmd)

	if err != nil {
		return err
	}

	if services == nil {
		services = allServices()
	}

	if commit == "" {
		return fmt.Errorf("%w: commit hash", errMissedRequiredArg)
	}
	commitMsg, err := commitMessage(commit)
	if err != nil {
		return fmt.Errorf("get commit message: %w", err)
	}

	type state struct {
		serviceState
		deployed bool
	}
	states := make([]state, len(services))

	for i, service := range services {
		state := state{serviceState: fetchServiceState(service)}
		serviceCommit := getCommitFromMeta(state.meta)
		if serviceCommit != "" {
			deployed, err := isAncestor(commit, serviceCommit)
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

func getCommitFromMeta(meta string) string {
	n := strings.Index(meta, "Build Version: ")
	if n == -1 {
		return ""
	}

	commit, _, _ := strings.Cut(meta[n+len("Build Version: "):], " ")
	return commit
}

func isAncestor(a, b string) (bool, error) {
	args := []string{"merge-base", "--is-ancestor", a, b}

	cmd := exec.Command("git", args...)
	errBuf := &bytes.Buffer{}
	cmd.Stderr = errBuf

	if err := cmd.Run(); err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) && ee.ExitCode() == 1 {
			return false, nil
		}

		return false, fmt.Errorf("%w (%s)", err, errBuf.Bytes())
	}

	return true, nil
}

func commitMessage(commit string) (string, error) {
	args := []string{"log", "--format=%B", "-n", "1", commit}

	cmd := exec.Command("git", args...)
	outBuf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	cmd.Stdout = outBuf
	cmd.Stderr = errBuf

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("%w (%s)", err, errBuf.Bytes())
	}

	return strings.TrimSpace(outBuf.String()), nil
}
