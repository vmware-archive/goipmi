// Copyright (c) 2014 VMware, Inc. All Rights Reserved.

package ipmi

import (
	"bytes"

	"hash/adler32"
	"log"
	"net"
	"sync"
)

// Handler function
type Handler func(*Message) Response

// Simulator for IPMI
type Simulator struct {
	wg       sync.WaitGroup
	addr     net.UDPAddr
	conn     *net.UDPConn
	handlers map[NetworkFunction]map[Command]Handler
	ids      map[uint32]string
}

// NewSimulator constructs a Simulator with the given addr
func NewSimulator(addr net.UDPAddr) *Simulator {
	s := &Simulator{
		addr: addr,
		ids:  map[uint32]string{},
		handlers: map[NetworkFunction]map[Command]Handler{
			NetworkFunctionChassis: map[Command]Handler{},
		},
	}

	// Built-in handlers for session management
	s.handlers[NetworkFunctionApp] = map[Command]Handler{
		CommandGetDeviceID:              s.deviceID,
		CommandGetAuthCapabilities:      s.authCapabilities,
		CommandGetSessionChallenge:      s.sessionChallenge,
		CommandActivateSession:          s.sessionActivate,
		CommandSetSessionPrivilegeLevel: s.sessionPrivilege,
		CommandCloseSession:             s.sessionClose,
	}

	return s
}

// SetHandler sets the command handler for the given netfn and command
func (s *Simulator) SetHandler(netfn NetworkFunction, command Command, handler Handler) {
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

func (s *Simulator) deviceID(*Message) Response {
	return &DeviceIDResponse{
		CompletionCode: CommandCompleted,
		IPMIVersion:    0x51, // 1.5
	}
}

func (s *Simulator) authCapabilities(*Message) Response {
	return &AuthCapabilitiesResponse{
		CompletionCode:  CommandCompleted,
		ChannelNumber:   0x01,
		AuthTypeSupport: AuthTypeNone | AuthTypeMD5 | AuthTypePassword,
	}
}

func (s *Simulator) sessionChallenge(m *Message) Response {
	// Convert username to a uint32 and use as the SessionID.
	// The SessionID will be propagated such that all requests
	// for this session include the ID, which can be used to
	// dispatch requests.
	username := bytes.TrimRight(m.Data[1:], "\000")
	hash := adler32.New()
	_, err := hash.Write(username)
	if err != nil {
		panic(err)
	}
	id := hash.Sum32()

	s.ids[id] = string(username)

	return &SessionChallengeResponse{
		CompletionCode:     CommandCompleted,
		TemporarySessionID: id,
	}
}

func (s *Simulator) sessionActivate(m *Message) Response {
	return &ActivateSessionResponse{
		CompletionCode: CommandCompleted,
		AuthType:       m.AuthType,
		SessionID:      m.SessionID,
		InboundSeq:     m.Sequence,
		MaxPriv:        0x04, // Admin
	}
}

func (s *Simulator) sessionPrivilege(m *Message) Response {
	return &SessionPrivilegeLevelResponse{
		CompletionCode:    CommandCompleted,
		NewPrivilegeLevel: m.Data[0],
	}
}

func (s *Simulator) sessionClose(*Message) Response {
	return CommandCompleted
}

func (s *Simulator) ipmiCommand(m *Message) []byte {
	response := Response(InvalidCommand)

	if commands, ok := s.handlers[m.NetFn()]; ok {
		if handler, ok := commands[m.Command]; ok {
			m.RequestID = s.ids[m.SessionID]
			response = handler(m)
		}
	}

	return m.toBytes(response)
}

func (s *Simulator) asfCommand(m *asfMessage) []byte {
	if m.MessageType != asfMessageTypePing {
		log.Print(m.unsupportedMessageType())
		return []byte{} // TODO: general ASF error code?
	}

	m.MessageType = asfMessageTypePong
	response := asfPong{
		IANAEnterpriseNumber: asfIANA,
		SupportedEntities:    0x81, // IPMI
	}

	return m.toBytes(&response)
}

func (s *Simulator) serve() error {
	buf := make([]byte, ipmiBufSize)

	for {
		var response []byte
		var err error

		n, addr, err := s.conn.ReadFrom(buf)
		if err != nil {
			return err // conn closed
		}

		header, err := rmcpHeaderFromBytes(buf)
		if err != nil {
			log.Print(err)
			continue
		}

		switch header.Class {
		case rmcpClassASF:
			m, err := asfMessageFromBytes(buf)
			if err != nil {
				log.Print(err)
				continue
			}

			response = s.asfCommand(m)
		case rmcpClassIPMI:
			m, err := messageFromBytes(buf[:n])
			if err != nil {
				log.Print(err)
				continue
			}

			response = s.ipmiCommand(m)
		default:
			log.Print(header.unsupportedClass())
			continue
		}

		_, err = s.conn.WriteTo(response, addr)
		if err != nil {
			return err // conn closed
		}
	}
}
