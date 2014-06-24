// Copyright (c) 2014 VMware, Inc. All Rights Reserved.

package ipmi

// Connection properties for a Client
type Connection struct {
	Path      string
	Hostname  string
	Port      int
	Username  string
	Password  string
	Interface string
}

// Client provides common high level functionality around the underlying transport
type Client struct {
	transport
}

// NewClient creates a new Client with the given Connection properties
func NewClient(c *Connection) (*Client, error) {
	t, err := newTransport(c)
	if err != nil {
		return nil, err
	}
	return &Client{
		transport: t,
	}, nil
}

// Open a new IPMI session
func (c *Client) Open() error {
	// TODO: auto-select transport based on BMC capabilities
	return c.open()
}

// Close the IPMI session
func (c *Client) Close() error {
	return c.close()
}

// Send a Request and unmarshal to given Response type
func (c *Client) Send(req *Request, res Response) error {
	// TODO: handle retry, timeouts, etc.
	return c.send(req, res)
}

func (c *Client) setBootParam(param uint8, data ...uint8) error {
	r := &Request{
		NetworkFunctionChassis,
		CommandSetSystemBootOptions,
		&SetSystemBootOptionsRequest{
			Param: param,
			Data:  data,
		},
	}
	return c.Send(r, &SetSystemBootOptionsResponse{})
}

// SetBootDevice is a wrapper around SetSystemBootOptionsRequest to configure the BootDevice
// per section 28.12 - table 28
func (c *Client) SetBootDevice(dev BootDevice) error {
	useProgress := true
	// set set-in-progress flag
	err := c.setBootParam(BootParamSetInProgress, 0x01)
	if err != nil {
		useProgress = false
	}

	err = c.setBootParam(BootParamInfoAck, 0x01, 0x01)
	if err != nil {
		if useProgress {
			// set-in-progress = set-complete
			_ = c.setBootParam(BootParamSetInProgress, 0x00)
		}
		return err
	}

	err = c.setBootParam(BootParamBootFlags, 0x80, uint8(dev), 0x00, 0x00, 0x00)
	if err == nil {
		if useProgress {
			// set-in-progress = commit-write
			_ = c.setBootParam(BootParamSetInProgress, 0x02)
		}
	}

	if useProgress {
		// set-in-progress = set-complete
		_ = c.setBootParam(BootParamSetInProgress, 0x00)
	}

	return err
}

// Control sends a chassis power control command
func (c *Client) Control(ctl ChassisControl) error {
	r := &Request{
		NetworkFunctionChassis,
		CommandChassisControl,
		&ChassisControlRequest{ctl},
	}
	return c.Send(r, &ChassisControlResponse{})
}
