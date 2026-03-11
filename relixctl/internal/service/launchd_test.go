package service_test

import (
	"encoding/xml"
	"strings"
	"testing"

	"github.com/relixdev/relix/relixctl/internal/service"
)

func TestGeneratePlistContainsBinaryPath(t *testing.T) {
	binaryPath := "/usr/local/bin/relixctl"
	plist := service.GeneratePlist(binaryPath)

	if !strings.Contains(plist, binaryPath) {
		t.Errorf("plist does not contain binary path %q:\n%s", binaryPath, plist)
	}
}

func TestGeneratePlistContainsLabel(t *testing.T) {
	plist := service.GeneratePlist("/usr/local/bin/relixctl")

	if !strings.Contains(plist, "com.relix.agent") {
		t.Errorf("plist does not contain label 'com.relix.agent':\n%s", plist)
	}
}

func TestGeneratePlistIsValidXML(t *testing.T) {
	plist := service.GeneratePlist("/usr/local/bin/relixctl")

	// xml.Unmarshal into a generic structure to check well-formedness.
	var v interface{}
	if err := xml.Unmarshal([]byte(plist), &v); err != nil {
		t.Errorf("plist is not valid XML: %v\n%s", err, plist)
	}
}

func TestGeneratePlistContainsDaemonRunArg(t *testing.T) {
	plist := service.GeneratePlist("/usr/local/bin/relixctl")

	if !strings.Contains(plist, "daemon-run") {
		t.Errorf("plist does not contain 'daemon-run' argument:\n%s", plist)
	}
}

func TestGeneratePlistRunAtLoad(t *testing.T) {
	plist := service.GeneratePlist("/usr/local/bin/relixctl")

	if !strings.Contains(plist, "RunAtLoad") {
		t.Errorf("plist does not contain RunAtLoad key:\n%s", plist)
	}
}
