// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2020 Datadog, Inc.

package network

import (
	"bytes"
	"fmt"
	"net"
	"os/exec"
	"strings"
	"time"

	"go.uber.org/zap"
)

const tcPath = "/sbin/tc"

// TrafficController is an interface being able to interact with the host
// queueing discipline
type TrafficController interface {
	AddDelay(iface string, parent string, handle uint32, delay time.Duration) error
	AddPrio(iface string, parent string, handle uint32, bands uint32, priomap [16]uint32) error
	AddFilterDestIP(iface string, parent string, handle uint32, ip *net.IPNet, flowid string) error
	ClearQdisc(iface string) error
}

type tcExecuter interface {
	Run(args ...string) error
}

type defaultTcExecuter struct{}

// Run executes the given args using the tc command
// and returns a wrapped error containing both the error returned by the execution and
// the stderr content
func (e defaultTcExecuter) Run(args ...string) error {
	// parse args and execute
	stderr := &bytes.Buffer{}
	cmd := exec.Command(tcPath, args...)
	cmd.Stderr = stderr

	// run command
	err := cmd.Run()
	if err != nil {
		err = fmt.Errorf("%w: %s", err, stderr.String())
	}

	return err
}

type tc struct {
	log      *zap.SugaredLogger
	executer tcExecuter
}

// NewTrafficController creates a standard traffic controller using tc
// and being able to log
func NewTrafficController(log *zap.SugaredLogger) TrafficController {
	return tc{
		log:      log,
		executer: defaultTcExecuter{},
	}
}

func (t tc) AddDelay(iface string, parent string, handle uint32, delay time.Duration) error {
	return t.executer.Run(buildCmd("qdisc", iface, parent, handle, "netem", fmt.Sprintf("delay %s", delay))...)
}

func (t tc) AddPrio(iface string, parent string, handle uint32, bands uint32, priomap [16]uint32) error {
	priomapStr := ""
	for _, bit := range priomap {
		priomapStr += fmt.Sprintf(" %d", bit)
	}

	priomapStr = strings.TrimSpace(priomapStr)
	params := fmt.Sprintf("bands %d priomap %s", bands, priomapStr)

	return t.executer.Run(buildCmd("qdisc", iface, parent, handle, "prio", params)...)
}

func (t tc) ClearQdisc(iface string) error {
	return t.executer.Run(strings.Split(fmt.Sprintf("qdisc del dev %s root", iface), " ")...)
}

func (t tc) AddFilterDestIP(iface string, parent string, handle uint32, ip *net.IPNet, flowid string) error {
	params := fmt.Sprintf("match ip dst %s flowid %s", ip.String(), flowid)
	return t.executer.Run(buildCmd("filter", iface, parent, handle, "u32", params)...)
}

func buildCmd(module string, iface string, parent string, handle uint32, kind string, parameters string) []string {
	cmd := fmt.Sprintf("%s add dev %s", module, iface)

	// parent
	if parent == "root" {
		cmd += fmt.Sprintf(" root")
	} else {
		cmd += fmt.Sprintf(" parent %s", parent)
	}

	// handle
	if handle != 0 {
		cmd += fmt.Sprintf(" handle %d:", handle)
	}

	// kind
	cmd += fmt.Sprintf(" %s", kind)

	// parameters
	cmd += fmt.Sprintf(" %s", parameters)

	return strings.Split(cmd, " ")
}
