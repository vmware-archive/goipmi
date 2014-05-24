// Copyright (c) 2014 VMware, Inc. All Rights Reserved.

package ipmi

import (
	"bytes"
	"encoding/binary"
	"hash/adler32"
	"io"
	"log"
	"net"
	"sync"
)

type rmcpHeader struct {
	Version            uint8
	Reserved           uint8
	RMCPSequenceNumber uint8
	Class              uint8
}

type asfHeader struct {
	IANAEnterpriseNumber uint32
	MessageType          uint8
	MessageTag           uint8
	Reserved             uint8
	DataLength           uint8
}

type asfMessage struct {
	rmcpHeader
	asfHeader
	Data []byte
}

type ipmiSession struct {
	AuthType  uint8
	Sequence  uint32
	SessionID uint32
}

type ipmiHeader struct {
	MsgLen     uint8
	RsAddr     uint8
	NetFnRsLUN uint8
	Checksum   uint8
	RqAddr     uint8
	RqSeq      uint8
	Command
}

// Message encapsulates an IPMI message
type Message struct {
	rmcpHeader
	ipmiSession
	AuthCode [16]byte
	ipmiHeader
	Data      []byte
	RequestID string
}

// NetworkFunction identifies the functional class of an IPMI message
type NetworkFunction uint8

// Network Function Codes (section 5.1)
var (
	NetworkFunctionChassis = NetworkFunction(0x00)
	NetworkFunctionApp     = NetworkFunction(0x06)
)

// Command fields on an IPMI message
type Command uint8

// Command Number Assignments (table G-1)
var (
	CommandGetAuthenticationCapabilities = Command(0x38)
	CommandGetSessionChallenge           = Command(0x39)
	CommandActivateSession               = Command(0x3a)
	CommandSetSessionPrivilegeLevel      = Command(0x3b)
	CommandCloseSession                  = Command(0x3c)
	CommandChassisControl                = Command(0x02)
	CommandSetSystemBootOptions          = Command(0x08)
)

// CompletionCode is the first byte in the data field of all IPMI responses
type CompletionCode uint8

// Code returns the CompletionCode as uint8
func (c CompletionCode) Code() uint8 {
	return uint8(c)
}

// Completion Codes (section 5.2)
var (
	CommandCompleted       = CompletionCode(0x00)
	InvalidCommand         = CompletionCode(0xc1)
	DestinationUnavailable = CompletionCode(0xd3)
	UnspecifiedError       = CompletionCode(0xff)
)

// Request handler
type Request func(*Message) Response

// Response to an IPMI request must include at least a CompletionCode
type Response interface {
	Code() uint8
}

// Simulator for IPMI
type Simulator struct {
	wg       sync.WaitGroup
	addr     net.UDPAddr
	conn     *net.UDPConn
	handlers map[NetworkFunction]map[Command]Request
	ids      map[uint32]string
}

// NewSimulator constructs a Simulator with the given addr
func NewSimulator(addr net.UDPAddr) *Simulator {
	s := &Simulator{
		addr: addr,
		ids:  map[uint32]string{},
		handlers: map[NetworkFunction]map[Command]Request{
			NetworkFunctionChassis: map[Command]Request{},
		},
	}

	// Built-in handlers for session management
	s.handlers[NetworkFunctionApp] = map[Command]Request{
		CommandGetAuthenticationCapabilities: s.authenticationCapabilities,
		CommandGetSessionChallenge:           s.sessionChallenge,
		CommandActivateSession:               s.sessionActivate,
		CommandSetSessionPrivilegeLevel:      s.sessionPrivilege,
		CommandCloseSession:                  s.sessionClose,
	}

	return s
}

// SetHandler sets the command handler for the given netfn and command
func (s *Simulator) SetHandler(netfn NetworkFunction, command Command, handler Request) {
	s.handlers[netfn][command] = handler
}

// NewConnection to this Simulator instance
func (s *Simulator) NewConnection() *Connection {
	addr := s.LocalAddr()
	return &Connection{
		Hostname:  addr.IP.String(),
		Port:      addr.Port,
		Interface: "lan",
	}
}

// LocalAddr returns the address the server is bound to.
func (s *Simulator) LocalAddr() *net.UDPAddr {
	if s.conn != nil {
		return s.conn.LocalAddr().(*net.UDPAddr)
	}
	return nil
}

// NetFn returns the NetworkFunction portion of the NetFn/RsLUN field
func (m *Message) NetFn() NetworkFunction {
	return NetworkFunction(m.NetFnRsLUN >> 2)
}

// Run the Simulator.
func (s *Simulator) Run() error {
	var err error
	s.conn, err = net.ListenUDP("udp4", &s.addr)
	if err != nil {
		return err
	}

	s.wg.Add(1)

	go func() {
		_ = s.serve()
		s.wg.Done()
	}()

	return nil
}

// Stop the Simulator.
func (s *Simulator) Stop() {
	_ = s.conn.Close()
	s.wg.Wait()
}

func (s *Simulator) authenticationCapabilities(*Message) Response {
	const (
		authNone = (1 << iota)
		authMD2
		authMD5
		authReserved
		authPassword
		authOEM
	)

	return struct {
		CompletionCode
		ChannelNumber             uint8
		AuthenticationTypeSupport uint8
		Status                    uint8
		Reserved                  uint8
		OEMID                     uint16
		OEMAux                    uint8
	}{
		CompletionCode:            CommandCompleted,
		ChannelNumber:             0x01,
		AuthenticationTypeSupport: authNone | authMD5 | authPassword,
	}
}

func (s *Simulator) sessionChallenge(m *Message) Response {
	// Convert username to a uint32 and use as the SessionID.
	// The SessionID will be propagated such that all requests
	// for this session include the ID, which can be used to
	// dispatch requests.
	username := bytes.TrimRight(m.Data[1:], "\000")
	hash := adler32.New()
	hash.Sum(username)
	id := hash.Sum32()
	s.ids[id] = string(username)

	return struct {
		CompletionCode
		TemporarySessionID uint32
		Challenge          [15]byte
	}{
		CompletionCode:     CommandCompleted,
		TemporarySessionID: id,
	}
}

func (s *Simulator) sessionActivate(m *Message) Response {
	return struct {
		CompletionCode
		AuthType   uint8
		SessionID  uint32
		InboundSeq uint32
		MaxPriv    uint8
	}{
		CompletionCode: CommandCompleted,
		AuthType:       m.AuthType,
		SessionID:      m.SessionID,
		InboundSeq:     m.Sequence,
		MaxPriv:        0x04, // Admin
	}
}

func (s *Simulator) sessionPrivilege(m *Message) Response {
	return struct {
		CompletionCode
		NewPrivilegeLevel uint8
	}{
		CompletionCode:    CommandCompleted,
		NewPrivilegeLevel: m.Data[0],
	}
}

func (s *Simulator) sessionClose(*Message) Response {
	return CommandCompleted
}

func (s *Simulator) write(writer io.Writer, data interface{}) {
	err := binary.Write(writer, binary.BigEndian, data)
	if err != nil {
		// shouldn't happen to a bytes.Buffer
		panic(err)
	}
}

func (s *Simulator) read(reader io.Reader, data interface{}) {
	err := binary.Read(reader, binary.BigEndian, data)
	if err != nil {
		// in this case, client gets no response or InvalidCommand
		log.Printf("binary.Read error: %s", err)
	}
}

func (s *Simulator) ipmiCommand(m *Message) []byte {
	response := Response(InvalidCommand)

	if commands, ok := s.handlers[m.NetFn()]; ok {
		if handler, ok := commands[m.Command]; ok {
			m.RequestID = s.ids[m.SessionID]
			response = handler(m)
		}
	}

	buf := new(bytes.Buffer)
	s.write(buf, &m.rmcpHeader)
	s.write(buf, &m.ipmiSession)
	if m.AuthType != 0 {
		s.write(buf, m.AuthCode)
	}
	s.write(buf, &m.ipmiHeader)
	s.write(buf, response)

	return buf.Bytes()
}

func (s *Simulator) asfCommand(m *asfMessage) []byte {
	if m.MessageType != 0x80 {
		log.Panicf("ASF message type not supported: %d", m.MessageType)
	}

	response := struct {
		IANAEnterpriseNumber  uint32
		OEM                   uint32
		SupportedEntities     uint8
		SupportedInteractions uint8
		Reserved              [6]uint8
	}{
		IANAEnterpriseNumber: m.IANAEnterpriseNumber,
		SupportedEntities:    0x81, // IPMI
	}

	buf := new(bytes.Buffer)
	s.write(buf, &m.rmcpHeader)
	s.write(buf, &m.asfHeader)
	s.write(buf, &response)

	return buf.Bytes()
}

func (s *Simulator) serve() error {
	buf := make([]byte, 1024)
	ipmiHeaderSize := binary.Size(ipmiHeader{})

	for {
		var header rmcpHeader
		var response []byte
		var err error

		n, addr, err := s.conn.ReadFrom(buf)
		if err != nil {
			return err // conn closed
		}

		reader := bytes.NewReader(buf[:n])

		s.read(reader, &header)

		switch header.Class {
		case 6:
			m := &asfMessage{
				rmcpHeader: header,
			}

			s.read(reader, &m.asfHeader)

			response = s.asfCommand(m)
		case 7:
			m := &Message{
				rmcpHeader: header,
			}

			s.read(reader, &m.ipmiSession)
			if m.AuthType != 0 {
				s.read(reader, &m.AuthCode)
			}
			s.read(reader, &m.ipmiHeader)

			dataLen := int(m.MsgLen) - ipmiHeaderSize
			m.Data = make([]byte, dataLen)
			_, _ = reader.Read(m.Data)

			response = s.ipmiCommand(m)
		default:
			log.Printf("Unsupported Class: %d", header.Class)
			continue
		}

		_, err = s.conn.WriteTo(response, addr)
		if err != nil {
			return err // conn closed
		}
	}
}
