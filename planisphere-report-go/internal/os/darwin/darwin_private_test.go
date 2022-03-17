package darwin

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExtractCrowdstrikeAID(t *testing.T) {
	got, err := extractCrowdstrikeAID([]byte(`=== agent_info ===
version: 6.36.14903.0
agentID: 1230BD01-7F0B-4A94-8BD1-97EA8C774511
customerID: EFB874A1-C37C-462A-9E45-AD5C00A96BF4
Sensor operational: true
`))
	require.NoError(t, err)
	require.Equal(t, "1230bd017f0b4a948bd197ea8c774511", got)
}
