/*
Copyright (c) 2014 VMware, Inc. All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package ipmi

// Client provides common high level functionality around the underlying transport
type Client struct {
	*Connection
	transport
}

// NewClient creates a new Client with the given Connection properties
func NewClient(c *Connection) (*Client, error) {
	t, err := newTransport(c)
	if err != nil {
		return nil, err
	}
	return &Client{
		Connection: c,
		transport:  t,
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

// DeviceID get the Device ID of the BMC
func (c *Client) DeviceID() (*DeviceIDResponse, error) {
	req := &Request{
		NetworkFunctionApp,
		CommandGetDeviceID,
		&DeviceIDRequest{},
	}
	res := &DeviceIDResponse{}
	return res, c.Send(req, res)
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
	return c.setBootDevice(dev, false)
}

// SetBootDeviceEFI is to support EFI boot option
func (c *Client) SetBootDeviceEFI(dev BootDevice) error {
	return c.setBootDevice(dev, true)
}

func (c *Client) setBootDevice(dev BootDevice, efi bool) error {
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

	flag := uint8(0x80)
	if efi {
		flag = flag | 0x20
	}
	err = c.setBootParam(BootParamBootFlags, flag, uint8(dev), 0x00, 0x00, 0x00)
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

func (c *Client) GetUserName(userID byte) (*GetUserNameResponse, error) {
	req := &Request{
		NetworkFunctionApp,
		CommandGetUserName,
		&GetUserNameRequest{
			UserID: userID,
		},
	}
	res := &GetUserNameResponse{}
	return res, c.Send(req, res)
}

func (c *Client) SetUserName(userID byte, username string) (*SetUserNameResponse, error) {
	req := &Request{
		NetworkFunctionApp,
		CommandSetUserName,
		&SetUserNameRequest{
			UserID:   userID,
			Username: username,
		},
	}
	res := &SetUserNameResponse{}
	return res, c.Send(req, res)
}
