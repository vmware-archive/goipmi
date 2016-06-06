package ipmi

const (
	MAX_MC_ID_STRING_LEN = 16
	DCMI_GROUP_EXTENSION_ID = 0xDC
)

// DcmiGetMcIdRequest per section 6.4.6.1
type DcmiGetMcIdRequest struct {
	GroupExtensionId  uint8
	Offset            uint8
	NumBytes          uint8
}

// DcmiGetMcIdResponse per section 6.4.6.1
type DcmiGetMcIdResponse struct {
	CompletionCode
	GroupExtensionId  uint8
	NumBytes          uint8
	Data              string
}

// DcmiSetMcIdRequest per section 6.4.6.2
type DcmiSetMcIdRequest struct {
	GroupExtensionId  uint8
	Offset            uint8
	NumBytes          uint8
	Data              string
}

// DcmiSetMcIdResponse per section 6.4.6.2
type DcmiSetMcIdResponse struct {
	CompletionCode
	GroupExtensionId  uint8
	LastOffsetWrriten uint8
}

// MarshalBinary implementation to handle variable length Data
func (r *DcmiGetMcIdRequest) MarshalBinary() ([]byte, error) {
	buf := make([]byte, 3)
	buf[0] = r.GroupExtensionId
	buf[1] = r.Offset
	buf[2] = r.NumBytes
	return buf, nil
}

// UnmarshalBinary implementation to handle variable length Data
func (r *DcmiGetMcIdRequest) UnmarshalBinary(buf []byte) error {
	if len(buf) < 3 {
		return ErrShortPacket
	}
	r.GroupExtensionId = buf[0]
	r.Offset = buf[1]
	r.NumBytes = buf[2]
	return nil
}

// MarshalBinary implementation to handle variable length Data
func (r *DcmiGetMcIdResponse) MarshalBinary() ([]byte, error) {
	buf := make([]byte, 3+len(r.Data))
	buf[0] = byte(r.CompletionCode)
	buf[1] = r.GroupExtensionId
	buf[2] = r.NumBytes
	copy(buf[3:], r.Data)
	return buf, nil
}

// UnmarshalBinary implementation to handle variable length Data
func (r *DcmiGetMcIdResponse) UnmarshalBinary(buf []byte) error {
	if len(buf) < 4 {
		return ErrShortPacket
	}
	r.CompletionCode = CompletionCode(buf[0])
	r.GroupExtensionId = buf[1]
	r.NumBytes = buf[2]
	r.Data = string(buf[3:3+r.NumBytes])
	return nil
}

// MarshalBinary implementation to handle variable length Data
func (r *DcmiSetMcIdRequest) MarshalBinary() ([]byte, error) {
	buf := make([]byte, 3+MAX_MC_ID_STRING_LEN)
	buf[0] = r.GroupExtensionId
	buf[1] = r.Offset
	buf[2] = MAX_MC_ID_STRING_LEN
	copy(buf[3:], r.Data)
	for i := 3 + len(r.Data); i < MAX_MC_ID_STRING_LEN; i++ {
		buf[i] = 0x00
	}
	return buf, nil
}

// UnmarshalBinary implementation to handle variable length Data
func (r *DcmiSetMcIdRequest) UnmarshalBinary(buf []byte) error {
	if len(buf) < 4 {
		return ErrShortPacket
	}
	r.GroupExtensionId = buf[0]
	r.Offset = buf[1]
	r.NumBytes = buf[2]
	r.Data = string(buf[3:3+r.NumBytes])
	return nil
}

// MarshalBinary implementation to handle variable length Data
func (r *DcmiSetMcIdResponse) MarshalBinary() ([]byte, error) {
	buf := make([]byte, 3)
	buf[0] = byte(r.CompletionCode)
	buf[1] = r.GroupExtensionId
	buf[2] = r.LastOffsetWrriten
	return buf, nil
}

// UnmarshalBinary implementation to handle variable length Data
func (r *DcmiSetMcIdResponse) UnmarshalBinary(buf []byte) error {
	if len(buf) < 3 {
		return ErrShortPacket
	}
	r.CompletionCode = CompletionCode(buf[0])
	r.GroupExtensionId = buf[1]
	r.LastOffsetWrriten = buf[2]
	return nil
}
