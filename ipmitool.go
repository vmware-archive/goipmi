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

type rawRequest interface {
	request() []byte // input for impitool raw
	parse([]byte)    // parse output of ipmitool raw
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

	// ipmitool raw wasn't happy with hex.Encode
	for i := 0; i < n; i++ {
		buf = append(buf, fmt.Sprintf("%#x", data[i:i+1]))
	}

	return buf
}

func (c *Connection) raw(r rawRequest) error {
	// ipmitool ... raw .. .. ..
	args := []string{"raw"}
	args = append(args, rawEncode(r.request())...)

	output, err := c.run(args...)
	if err != nil {
		return err
	}

	r.parse(rawDecode(strings.TrimSpace(output)))

	return nil
}

func (c *Connection) ChassisStatus() (*ChassisStatus, error) {
	// ipmitool ... chassis status
	s := &ChassisStatus{}
	return s, c.raw(s)
}

func (c *Connection) SetBootDevice(device BootDevice) error {
	// impitool ... chassis bootdev pxe|floppy|etc
	_, err := c.run("chassis", "bootdev", device.String())
	return err
}

func (c *Connection) GetBootFlags() (*BootFlags, error) {
	// ipmitool ... chassis bootparam get 5
	flags := &BootFlags{}
	return flags, c.raw(flags)
}

func (c *Connection) ChassisControl(ctl ChassisControl) error {
	// impitool ... power up|down|cycle|etc
	_, err := c.run("power", ctl.String())
	return err
}
