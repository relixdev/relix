package service_test

import (
	"strings"
	"testing"

	"github.com/relixdev/relix/relixctl/internal/service"
)

func TestGenerateUnitContainsBinaryPath(t *testing.T) {
	binaryPath := "/usr/local/bin/relixctl"
	unit := service.GenerateUnit(binaryPath)

	if !strings.Contains(unit, binaryPath) {
		t.Errorf("unit file does not contain binary path %q:\n%s", binaryPath, unit)
	}
}

func TestGenerateUnitContainsDaemonRunArg(t *testing.T) {
	unit := service.GenerateUnit("/usr/local/bin/relixctl")

	if !strings.Contains(unit, "daemon-run") {
		t.Errorf("unit file does not contain 'daemon-run':\n%s", unit)
	}
}

func TestGenerateUnitContainsServiceSection(t *testing.T) {
	unit := service.GenerateUnit("/usr/local/bin/relixctl")

	if !strings.Contains(unit, "[Service]") {
		t.Errorf("unit file missing [Service] section:\n%s", unit)
	}
	if !strings.Contains(unit, "[Unit]") {
		t.Errorf("unit file missing [Unit] section:\n%s", unit)
	}
	if !strings.Contains(unit, "[Install]") {
		t.Errorf("unit file missing [Install] section:\n%s", unit)
	}
}

func TestGenerateUnitContainsRestart(t *testing.T) {
	unit := service.GenerateUnit("/usr/local/bin/relixctl")

	if !strings.Contains(unit, "Restart=") {
		t.Errorf("unit file does not contain Restart directive:\n%s", unit)
	}
}

func TestGenerateUnitContainsWantedBy(t *testing.T) {
	unit := service.GenerateUnit("/usr/local/bin/relixctl")

	if !strings.Contains(unit, "WantedBy=") {
		t.Errorf("unit file does not contain WantedBy directive:\n%s", unit)
	}
}
