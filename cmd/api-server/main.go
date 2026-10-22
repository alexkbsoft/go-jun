package main

import (
	"fmt"
	"net/http"
	"runtime/debug"
)

const PORT = ":8090"

func main() {
	http.HandleFunc("/debug/info", handleServiceInfo)

	fmt.Println("Server started on", PORT)
	http.ListenAndServe(PORT, nil)
}

func handleServiceInfo(w http.ResponseWriter, req *http.Request) {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		fmt.Fprintf(w, "Не удалось прочитать информацию о сборке")
		return
	}

	var vcsRevision, vcsTime string
	for _, setting := range info.Settings {
		switch setting.Key {
		case "vcs.revision":
			vcsRevision = setting.Value
		case "vcs.time":
			vcsTime = setting.Value
		}
	}

	fmt.Fprintln(w,
		"Go Version:", info.GoVersion,
		"Build Version:", vcsRevision[:8],
		"Build Time:", vcsTime)

}
