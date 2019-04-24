package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"runtime"
	"strings"
	"testing"
)

//const TFVARS_BASE_PATH = "examples/terraform.tfvars"

func assert(o bool) {
	if !o {
		fmt.Printf("\n%c[35m%s%c[0m\n\n", 27, _GetRecentLine(), 27)
		os.Exit(1)
	}
}

func _GetRecentLine() string {
	_, file, line, _ := runtime.Caller(2)
	buf, _ := ioutil.ReadFile(file)
	code := strings.TrimSpace(strings.Split(string(buf), "\n")[line-1])
	return fmt.Sprintf("%v:%d\n%s", path.Base(file), line, code)
}

func TestMain(m *testing.M) {

	assert(true)
}

func TestVersionInfo(t *testing.T) {
	assert("unknown" == versionInfo())

	buildVersion = "vXYZ"
	assert("vXYZ" == versionInfo())
}
