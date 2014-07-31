// Copyright (c) 2014 VMware, Inc. All Rights Reserved.

package supermicro

import (
	"github.com/vmware/goipmi"
	"github.com/vmware/goipmi/media"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type sim struct {
	*ipmi.Simulator
	c *ipmi.Connection
	l net.Listener
	h *http.ServeMux
	m media.DeviceMap
	t *testing.T
	r map[string]int

	calledSetBoot map[ipmi.BootDevice]bool
	calledControl bool
}

func (s *sim) Run() error {
	s.r = make(map[string]int)
	s.calledSetBoot = make(map[ipmi.BootDevice]bool)
	s.Simulator = ipmi.NewSimulator(net.UDPAddr{})
	if err := s.Simulator.Run(); err != nil {
		return err
	}

	s.SetHandler(ipmi.NetworkFunctionApp, ipmi.CommandGetDeviceID, func(*ipmi.Message) ipmi.Response {
		return &ipmi.DeviceIDResponse{
			CompletionCode: ipmi.CommandCompleted,
			ManufacturerID: ipmi.OemSupermicro,
		}
	})
	s.SetHandler(ipmi.NetworkFunctionChassis, ipmi.CommandSetSystemBootOptions, func(m *ipmi.Message) ipmi.Response {
		if m.Data[0] == ipmi.BootParamBootFlags {
			s.calledSetBoot[ipmi.BootDevice(m.Data[2])] = true
		}
		return ipmi.CommandCompleted
	})
	s.SetHandler(ipmi.NetworkFunctionChassis, ipmi.CommandChassisControl, func(*ipmi.Message) ipmi.Response {
		s.calledControl = true
		return ipmi.CommandCompleted
	})

	s.c = s.NewConnection()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return err
	}
	smcScheme = "http"
	smcPort = strconv.Itoa(l.Addr().(*net.TCPAddr).Port)
	s.l = l
	s.h = http.NewServeMux()
	go (&http.Server{Handler: s.h}).Serve(s.l)

	return nil
}

func (s *sim) Stop() {
	s.Simulator.Stop()
	s.l.Close()
}

func (s *sim) login(w http.ResponseWriter, req *http.Request) {
	user := req.PostFormValue("name")
	pass := req.PostFormValue("pwd")
	if user != "ADMIN" || pass != "ADMIN" {
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:    "SID",
		Expires: time.Date(1970, 1, 1, 1, 0, 0, 0, time.UTC),
	})
	http.SetCookie(w, &http.Cookie{
		Name:  "SID",
		Path:  "/",
		Value: "t0k3n",
	})
}

func (s *sim) authenticate(w http.ResponseWriter, req *http.Request) bool {
	c, err := req.Cookie("SID")
	if err == nil && c.Value == "t0k3n" {
		return true
	}
	assert.Fail(s.t, "authenticate")
	return false
}

func (s *sim) insertIMG(w http.ResponseWriter, req *http.Request) {
	if !s.authenticate(w, req) {
		return
	}

	// chunked encoding not supported by the BMC,
	// so test that we send a Content-Length
	assert.NotEmpty(s.t, req.Header.Get("Content-Length"))
	assert.Empty(s.t, req.Header.Get("Transfer-Encoding"))

	path := s.m[media.IMG].Path
	f, h, err := req.FormFile("file")
	assert.NoError(s.t, err)
	assert.Equal(s.t, path, h.Filename)
	assert.Equal(s.t, []string{"application/octet-stream"}, h.Header["Content-Type"])

	actual, err := ioutil.ReadFile(path)
	assert.NoError(s.t, err)
	given := make([]byte, len(actual))

	_, err = f.Read(given)
	assert.NoError(s.t, err)
	assert.Equal(s.t, actual, given)
}

func (s *sim) ejectIMG(w http.ResponseWriter, req *http.Request) {
	if !s.authenticate(w, req) {
		return
	}
}

func (s *sim) configISO(w http.ResponseWriter, req *http.Request) {
	if !s.authenticate(w, req) {
		return
	}
	if err := req.ParseForm(); err != nil {
		s.t.Error(err)
	}
	v := req.Form
	assert.Equal(s.t, 4, len(v))
	assert.Equal(s.t, []string{"127.0.0.1"}, v["host"])
	assert.Equal(s.t, []string{"\\images\\driver.go"}, v["path"])
	assert.Equal(s.t, []string{"guest"}, v["user"])
	assert.Equal(s.t, []string{""}, v["pwd"])
}

func (s *sim) insertISO(w http.ResponseWriter, req *http.Request) {
	if !s.authenticate(w, req) {
		return
	}
}

func (s *sim) ejectISO(w http.ResponseWriter, req *http.Request) {
	if !s.authenticate(w, req) {
		return
	}
}

func TestSMC(t *testing.T) {
	s := &sim{t: t}
	err := s.Run()
	if err != nil {
		t.Fatal(err)
	}
	defer s.Stop()

	handlers := map[string]http.HandlerFunc{
		smcLogin:     s.login,
		smcInsertIMG: s.insertIMG,
		smcEjectIMG:  s.ejectIMG,
		smcConfigISO: s.configISO,
		smcInsertISO: s.insertISO,
		smcEjectISO:  s.ejectISO,
	}
	for path := range handlers {
		s.h.HandleFunc(path, func(w http.ResponseWriter, req *http.Request) {
			p := req.URL.Path
			s.r[p]++
			handlers[p](w, req)
		})
	}

	calledHandler := false
	bootHandler := func(*ipmi.Client) error {
		calledHandler = true
		return nil
	}
	s.m = media.DeviceMap{
		media.ISO: &media.Device{
			Path: "/images/driver.go",
			Boot: true,
		},
		media.IMG: &media.Device{
			Path: "driver_test.go",
		},
	}

	err = media.Boot(s.c, s.m, bootHandler)
	assert.Error(t, err)
	assert.Equal(t, http.ErrNoCookie, err) // auth failed
	assert.Equal(t, 1, len(s.r))
	assert.False(t, calledHandler)
	assert.False(t, s.calledControl)

	s.c.Username = "ADMIN"
	s.c.Password = "ADMIN"

	err = media.Boot(s.c, s.m, bootHandler)
	assert.NoError(t, err)

	assert.Equal(t, len(handlers), len(s.r))
	assert.True(t, s.calledControl)
	assert.True(t, calledHandler)
	assert.Equal(t, 1, len(s.calledSetBoot))
	assert.True(t, s.calledSetBoot[ipmi.BootDeviceCdrom])

	for path, handler := range handlers {
		handlers[path] = http.NotFound
		err = media.Boot(s.c, s.m, bootHandler)

		if strings.HasSuffix(path, "pout.cgi") {
			// Boot just logs unmount errors
			assert.NoError(t, err, path)
		} else {
			assert.Error(t, err, path)
			nf := http.StatusText(http.StatusNotFound)
			assert.True(t, strings.Contains(err.Error(), nf), err.Error())
		}

		handlers[path] = handler
	}
}
