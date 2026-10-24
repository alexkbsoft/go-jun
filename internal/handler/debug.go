package handler

import (
	"fmt"
	"runtime/debug"
	"net/http"
)

func HandleServiceInfo(w http.ResponseWriter, req *http.Request) {
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
