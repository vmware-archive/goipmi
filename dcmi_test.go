package ipmi

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestGetMcIdStringRequest(t *testing.T) {
	req := &Request{
		NetworkFunctionDcmi,
		CommandGetMcIdString,
		&DcmiGetMcIdRequest{DCMI_GROUP_EXTENSION_ID, 0, MAX_MC_ID_STRING_LEN},
	}
	raw := requestToStrings(req)
	assert.Equal(t, []string{"0x2c", "0x09", "0xdc", "0x00", "0x10"}, raw)
}

func TestGetMcIdStringResponse(t *testing.T) {
	status := &DcmiGetMcIdResponse{}
	err := responseFromString("dc 0c 61 62 63 64 65 66 67 68 69 6a 6b 6c 00 00 00 00", status)
	assert.NoError(t, err)
	assert.Equal(t, status.GroupExtensionId, uint8(DCMI_GROUP_EXTENSION_ID))
	assert.Equal(t, status.NumBytes, uint8(12))
	assert.Equal(t, status.Data, "abcdefghijkl")
}

func TestSetMcIdStringRequest(t *testing.T) {
	testMcId := "abcdefghijkl"
	req := &Request{
		NetworkFunctionDcmi,
		CommandSetMcIdString,
		&DcmiSetMcIdRequest{DCMI_GROUP_EXTENSION_ID, 0, uint8(len(testMcId)), testMcId},
	}
	raw := requestToStrings(req)
	assert.Equal(t, []string{"0x2c", "0x0a", "0xdc", "0x00", "0x10", "0x61", "0x62", "0x63", "0x64", "0x65", "0x66", "0x67", "0x68", "0x69", "0x6a", "0x6b", "0x6c", "0x00", "0x00", "0x00", "0x00"}, raw)
}

func TestSetMcIdStringResponse(t *testing.T) {
	status := &DcmiSetMcIdResponse{}
	err := responseFromString("dc 0c 61 62 63 64 65 66 67 68 69 6a 6b 6c 00 00 00 00", status)
	assert.NoError(t, err)
	assert.Equal(t, status.GroupExtensionId, uint8(DCMI_GROUP_EXTENSION_ID))
	assert.Equal(t, status.LastOffsetWrriten, uint8(0x0c))
}
