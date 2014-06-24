// Copyright (c) 2014 VMware, Inc. All Rights Reserved.

package ipmi

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// Connection wraps the ipmitool program arguments
type Connection struct {
	Path      string
	Hostname  string
	Port      int
	Username  string
	Password  string
	Interface string
}

type tool struct {
	*Connection
}

func newToolTransport(c *Connection) transport {
	return &tool{Connection: c}
}

func (t *tool) open() error {
	return nil
}

func (t *tool) close() error {
	return nil
}

func (t *tool) send(req *Request, res Response) error {
	// ipmitool ... raw .. .. ..
	args := append([]string{"raw"}, requestToStrings(req)...)

	output, err := t.run(args...)
	if err != nil {
		// TODO: parse CompletionCode from stderr
		return err
	}

	return responseFromString(output, res)
}

func requestToBytes(r *Request) []byte {
	data := messageDataToBytes(r.Data)
	msg := make([]byte, 2+len(data))
	msg[0] = uint8(r.NetworkFunction)
	msg[1] = uint8(r.Command)
	copy(msg[2:], data)
	return msg
}

func requestToStrings(r *Request) []string {
	msg := requestToBytes(r)
	return rawEncode(msg)
}

func responseFromBytes(msg []byte, r Response) error {
	buf := new(bytes.Buffer)
	buf.WriteByte(uint8(CommandCompleted))
	buf.Write(msg)
	return messageDataFromBytes(buf.Bytes(), r)
}

func responseFromString(s string, r Response) error {
	msg := rawDecode(strings.TrimSpace(s))
	return responseFromBytes(msg, r)
}

func (c *Connection) options() []string {
	intf := c.Interface
	if intf == "" {
		intf = "lanplus"
	}

	options := []string{
		"-H", c.Hostname,
		"-U", c.Username,
		"-P", c.Password,
		"-I", intf,
	}

	if c.Port != 0 {
		options = append(options, "-p", strconv.Itoa(c.Port))
	}

	return options
}

func (c *Connection) run(args ...string) (string, error) {
	path := c.Path
	opts := append(c.options(), args...)

	if path == "" {
		path = "ipmitool"
	}

	cmd := exec.Command(path, opts...)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("run %s %s: %s (%s)",
			path, strings.Join(opts, " "), stderr.String(), err)
	}

	return stdout.String(), err
}

func rawDecode(data string) []byte {
	var buf bytes.Buffer

	for _, s := range strings.Split(data, " ") {
		b, err := hex.DecodeString(s)
		if err != nil {
			panic(err)
		}

		_, err = buf.Write(b)
		if err != nil {
			panic(err)
		}
	}

	return buf.Bytes()
}

func rawEncode(data []byte) []string {
	n := len(data)
	buf := make([]string, 0, n)

	// ipmitool needs every byte to be a separate argument
	for i := 0; i < n; i++ {
		buf = append(buf, "0x"+hex.EncodeToString(data[i:i+1]))
	}

	return buf
}
