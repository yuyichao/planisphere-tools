package helpers

import (
	"net"

	"gitlab.oit.duke.edu/devil-ops/planisphere-sdk/planisphere"
)

/*
Lookuper will do the more advanced lookups. Using a custom struct for this so we
don't have to make duplicate system calls to look at the system_profiler
*/
type Lookuper struct {
	Overrides map[string]interface{}
	Payload   planisphere.SelfReportPayload
}

/*
This is pretty OS agnostic, but if we need to later, we can break it out in to
OS specific functions
*/
func getMacAddr() ([]string, error) {
	ifas, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	var as []string
	for _, ifa := range ifas {
		a := ifa.HardwareAddr.String()
		if a != "" {
			as = append(as, a)
		}
	}
	return as, nil
}
