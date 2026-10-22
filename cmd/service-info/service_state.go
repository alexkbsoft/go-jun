package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type serviceState struct {
	name    string
	unknown bool
	err     string
	meta    string
}

func (s serviceState) Commit() string {
	n := strings.Index(s.meta, "Build Version: ")
	if n == -1 {
		return ""
	}

	commit, _, _ := strings.Cut(s.meta[n+len("Build Version: "):], " ")
	return commit
}

func fetchServiceState(ctx context.Context, name string) serviceState {
	state := serviceState{name: name}
	port, ok := servicePorts[name]
	if !ok {
		state.unknown = true
		return state
	}

	url := fmt.Sprintf("http://127.0.0.1:%d/debug/info", port)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)

	if err != nil {
		state.err = fmt.Sprintf("get service info %v: ", err)
		return state
	}
	resp, err := http.DefaultClient.Do(req)

	if resp.StatusCode != http.StatusOK {
		state.err = fmt.Sprintf("returned non-ok status code: %d", resp.StatusCode)
		return state
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		state.err = fmt.Sprintf("read response body: %v", err)
		return state
	}

	state.meta = string(body)

	return state
}