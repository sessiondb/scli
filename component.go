package main

// Component identifies which service(s) to operate on: api (backend), ui (frontend server), or all.
const (
	ComponentAPI = "api"
	ComponentUI  = "ui"
	ComponentAll = "all"
)

// systemd unit and local file names per component
const (
	systemdAPIServiceName = "sessiondb.service"
	systemdUIServiceName   = "sessiondb-ui.service"
	pidFileAPI            = "sessiondb.pid"
	pidFileUI             = "sessiondb-ui.pid"
	runLogFileAPI         = "sessiondb.log"
	runLogFileUI          = "sessiondb-ui.log"
)

// uiBinaryName is the name of the UI server binary under versions/<tag>/ui/
const uiBinaryName = "sessiondb-ui"

// validComponent returns true if c is api, ui, or all.
func validComponent(c string) bool {
	return c == ComponentAPI || c == ComponentUI || c == ComponentAll
}

// systemdUnitName returns the systemd unit for the component (api or ui). For "all" returns "".
func systemdUnitName(component string) string {
	switch component {
	case ComponentAPI:
		return systemdAPIServiceName
	case ComponentUI:
		return systemdUIServiceName
	default:
		return ""
	}
}
