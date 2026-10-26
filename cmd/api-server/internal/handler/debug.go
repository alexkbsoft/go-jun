package handler

import (
	"fmt"
	"net/http"
	"runtime/debug"
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

	rev := vcsRevision
	if len(rev) > 8 {
		rev = rev[:8]
	}

	fmt.Fprintln(w,
		"Go Version:", info.GoVersion,
		"Build Version:", rev,
		"Build Time:", vcsTime)

}
