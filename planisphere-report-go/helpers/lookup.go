package helpers

import (
	"net"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
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

func NewLookuper(overrides map[string]interface{}) (*Lookuper, error) {
	l := &Lookuper{
		Overrides: overrides,
	}
	// Here are some things that can only be set in the overrides section

	// Key - Not the API Key, but the instance key. This is the 'key' part of the post
	if instanceKey, ok := l.Overrides["instance_key"]; ok {
		l.Payload.Key = instanceKey.(string)
	}

	// Department Key
	if departmentKey, ok := l.Overrides["department_key"]; ok {
		l.Payload.Data.DepartmentKey = departmentKey.(string)
	}

	// Support Group ID
	if supportGroupID, ok := l.Overrides["support_group_id"]; ok {
		l.Payload.Data.SupportGroupId = uint64(supportGroupID.(int))
	}

	// Support Group Name
	if supportGroupName, ok := l.Overrides["support_group_name"]; ok {
		l.Payload.Data.SupportGroupName = supportGroupName.(string)
	}

	// Usage Type
	if usageType, ok := l.Overrides["usage_type"]; ok {
		l.Payload.Data.UsageType = usageType.(string)
	}

	// Status: deployed, rma, etc
	if status, ok := l.Overrides["status"]; ok {
		l.Payload.Data.Status = status.(string)
	}

	// Hostname Field
	if hostname, ok := l.Overrides["hostname"]; ok {
		l.Payload.Data.Hostname = hostname.(string)
	} else {
		hostname, err := os.Hostname()
		if err != nil {
			log.Warning("Could not detect hostname: ", err)
		}
		l.Payload.Data.Hostname = hostname
	}

	// Mac Addresses Field
	// "Most Dope"
	//   - Mac Miller ✌️
	if macAddresses, ok := l.Overrides["mac_addresses"]; ok {
		for _, item := range macAddresses.([]interface{}) {
			l.Payload.Data.MacAddresses = append(l.Payload.Data.MacAddresses, item.(string))
		}
	} else {
		macs, err := getMacAddr()
		if err != nil {
			log.Warning("Could not detect mac_addresses: ", err)
		}
		l.Payload.Data.MacAddresses = macs
	}

	// Extra data
	l.Payload.ExtraData = map[string]string{}
	if extraData, ok := l.Overrides["extra_data"]; ok {
		for k, v := range extraData.(map[string]interface{}) {
			log.Println(k, v)
			l.Payload.ExtraData[k] = v.(string)
		}
	}

	// Now
	l.Payload.LastActive = time.Now()

	err := ApplyPlatformDetections(l)
	if err != nil {
		return nil, err
	}

	// Return
	return l, nil
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
