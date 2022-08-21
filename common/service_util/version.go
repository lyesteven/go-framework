package service_util

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"strconv"
)

var (
	gitBranch  string
	gitTag     string
	gitCommit  string
	commitDate string
	buildDate  string
)

type Info struct {
	ServiceName string `json:"serviceName"`
	ServiceAddr string `json:"serviceAddr"`
	GitTag      string `json:"gitTag"`
	GitBranch   string `json:"gitBranch"`
	GitCommit   string `json:"gitCommit"`
	CommitDate  string `json:"commitDate"`
	BuildDate   string `json:"buildDate"`
	GoVersion   string `json:"goVersion"`
	Compiler    string `json:"compiler"`
	Platform    string `json:"platform"`
}

func (info Info) String() string {
	return info.GitCommit
}

func GetVersion() Info {
	return Info{
		ServiceName: suSingleton.serviceName,
		ServiceAddr: suSingleton.hostIP + ":" + strconv.Itoa(suSingleton.srvPort),
		GitTag:      gitTag,
		GitBranch:   gitBranch,
		GitCommit:   gitCommit,
		CommitDate:  commitDate,
		BuildDate:   buildDate,
		GoVersion:   runtime.Version(),
		Compiler:    runtime.Compiler,
		Platform:    fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}
}

func getVersionHandler(w http.ResponseWriter, r *http.Request) {
	bBuf, err := json.MarshalIndent(GetVersion(),"","\t")
	if err != nil {
		fmt.Fprint(w, "Get Version error!")
	}
	fmt.Fprint(w, string(bBuf[:]))
}
