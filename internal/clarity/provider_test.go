package clarity

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

var createSnapshot = `
{
  "name": "hey",
  "info": {
    "type": "aws",
    "account_id": "01234576",
    "role": "role-name",
    "region": "us-east-1"
  }
}
`

func TestSerialization(t *testing.T) {
	var p Provider
	err := json.Unmarshal([]byte(createSnapshot), &p)
	require.NoError(t, err)
	require.Equal(t, "hey", p.Name)
	expected := ProviderInfo{}
	expected.TypeSwitch = AWSProviderType
	expected.AWS = &AWS{
		AccountID:           "01234576",
		AdditionalAccountID: nil,
		Role:                "role-name",
		Region:              "us-east-1",
	}
	require.Equal(t, expected, p.Info)
}
