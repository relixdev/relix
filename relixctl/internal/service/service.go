package service

// ServiceManager is a platform-agnostic interface for managing relixctl as a
// system service.
type ServiceManager interface {
	// Install writes the service unit/plist and enables+starts it.
	Install(binaryPath string) error
	// Uninstall stops+disables the service and removes the unit/plist.
	Uninstall() error
	// IsInstalled returns true if the service unit/plist file exists on disk.
	IsInstalled() bool
}
