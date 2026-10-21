package main

import (
	"fmt"
	"net/http"
	"runtime/debug"
)

func hello(w http.ResponseWriter, req *http.Request) {

	fmt.Fprintf(w, "hello\n")
}

func headers(w http.ResponseWriter, req *http.Request) {

	for name, headers := range req.Header {
		for _, h := range headers {
			fmt.Fprintf(w, "%v: %v\n", name, h)
		}
	}
}

const PORT = ":8090"

func main() {
	http.HandleFunc("/debug/info", handleServiceInfo)
	http.HandleFunc("/hello", hello)
	http.HandleFunc("/headers", headers)

	fmt.Println("Server started on", PORT)
	http.ListenAndServe(PORT, nil)
}

func handleServiceInfo(w http.ResponseWriter, req *http.Request) {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		fmt.Fprintf(w, "Не удалось прочитать информацию о сборке")
		return
	}

	var version string
	found := false
	for _, setting := range info.Settings {
		if setting.Key == "vcs.revision" {
			version = setting.Value
			found = true
			break
		}
	}
	if !found {
		fmt.Fprintf(w, "Данные Git (VCS) отсутствуют. Соберите бинарник через 'go build'\n")
	} else {
		fmt.Fprintf(w, "Commit SHA1: %s\n", version)
		return
	}

	fmt.Fprintf(w, "Service Info:\n")
	fmt.Fprintf(w, "Go version: %s\n", info.GoVersion)
	fmt.Fprintf(w, "Method: %s\n", req.Method)
}
