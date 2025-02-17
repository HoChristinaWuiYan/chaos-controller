// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2021 Datadog, Inc.

package v1beta1

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const (
	// FlowEgress is the string representation of network disruptions applied to outgoing packets
	FlowEgress = "egress"
	// FlowIngress is the string representation of network disruptions applied to incoming packets
	FlowIngress = "ingress"
)

// NetworkDisruptionSpec represents a network disruption injection
type NetworkDisruptionSpec struct {
	// +nullable
	Hosts []NetworkDisruptionHostSpec `json:"hosts,omitempty"`
	// +nullable
	AllowedHosts []NetworkDisruptionHostSpec `json:"allowedHosts,omitempty"`
	// +nullable
	Services []NetworkDisruptionServiceSpec `json:"services,omitempty"`
	// +kubebuilder:validation:Enum=egress;ingress
	// +ddmark:validation:Enum=egress;ingress
	Flow string `json:"flow,omitempty"`
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=100
	// +ddmark:validation:Minimum=0
	// +ddmark:validation:Maximum=100
	Drop int `json:"drop,omitempty"`
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=100
	// +ddmark:validation:Minimum=0
	// +ddmark:validation:Maximum=100
	Duplicate int `json:"duplicate,omitempty"`
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=100
	// +ddmark:validation:Minimum=0
	// +ddmark:validation:Maximum=100
	Corrupt int `json:"corrupt,omitempty"`
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=60000
	// +ddmark:validation:Minimum=0
	// +ddmark:validation:Maximum=60000
	Delay uint `json:"delay,omitempty"`
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=100
	// +ddmark:validation:Minimum=0
	// +ddmark:validation:Maximum=100
	DelayJitter uint `json:"delayJitter,omitempty"`
	// +kubebuilder:validation:Minimum=0
	// +ddmark:validation:Minimum=0
	BandwidthLimit int `json:"bandwidthLimit,omitempty"`
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=65535
	// +ddmark:validation:Minimum=0
	// +ddmark:validation:Maximum=65535
	// +nullable
	DeprecatedPort *int `json:"port,omitempty"`
}

type NetworkDisruptionHostSpec struct {
	Host string `json:"host,omitempty"`
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=65535
	// +ddmark:validation:Minimum=0
	// +ddmark:validation:Maximum=65535
	Port int `json:"port,omitempty"`
	// +kubebuilder:validation:Enum=tcp;udp;""
	// +ddmark:validation:Enum=tcp;udp;""
	Protocol string `json:"protocol,omitempty"`
}

type NetworkDisruptionServiceSpec struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

// Validate validates args for the given disruption
func (s *NetworkDisruptionSpec) Validate() error {
	// check that at least one network disruption is set
	if s.BandwidthLimit == 0 &&
		s.Drop == 0 &&
		s.Delay == 0 &&
		s.Corrupt == 0 &&
		s.Duplicate == 0 {
		return errors.New("the network disruption was selected, but no disruption type was specified. Please set at least one of: drop, delay, bandwidthLimit, corrupt, or duplicate. No injection will occur")
	}

	// ensure spec filters on something if ingress mode is enabled
	if s.Flow == FlowIngress {
		if len(s.Hosts) == 0 && len(s.Services) == 0 {
			return errors.New("the network disruption has ingress flow enabled but no hosts or services are provided, which is required for it to work")
		}
	}

	if k8sClient != nil {
		err := validateServices(k8sClient, s.Services)
		if err != nil {
			return err
		}
	}

	// ensure deprecated fields are not used
	if s.DeprecatedPort != nil {
		return fmt.Errorf("the port specification at the network disruption level is deprecated; apply to network disruption hosts instead")
	}

	return nil
}

// GenerateArgs generates injection or cleanup pod arguments for the given spec
func (s *NetworkDisruptionSpec) GenerateArgs() []string {
	args := []string{
		"network-disruption",
		"--corrupt",
		strconv.Itoa(s.Corrupt),
		"--drop",
		strconv.Itoa(s.Drop),
		"--duplicate",
		strconv.Itoa(s.Duplicate),
		"--delay",
		strconv.Itoa(int(s.Delay)),
		"--delay-jitter",
		strconv.Itoa(int(s.DelayJitter)),
		"--bandwidth-limit",
		strconv.Itoa(s.BandwidthLimit),
	}

	// append hosts
	for _, host := range s.Hosts {
		args = append(args, "--hosts", fmt.Sprintf("%s;%d;%s", host.Host, host.Port, host.Protocol))
	}

	// append allowed hosts
	for _, host := range s.AllowedHosts {
		args = append(args, "--allowed-hosts", fmt.Sprintf("%s;%d;%s", host.Host, host.Port, host.Protocol))
	}

	// append services
	for _, service := range s.Services {
		args = append(args, "--services", fmt.Sprintf("%s;%s", service.Name, service.Namespace))
	}

	// append flow
	if s.Flow != "" {
		args = append(args, "--flow", s.Flow)
	}

	return args
}

// NetworkDisruptionHostSpecFromString parses the given hosts to host specs
// The expected format for hosts is <host>;<port>;<protocol>
func NetworkDisruptionHostSpecFromString(hosts []string) ([]NetworkDisruptionHostSpec, error) {
	parsedHosts := []NetworkDisruptionHostSpec{}

	// parse given hosts
	for _, host := range hosts {
		// parse host with format <host>;<port>;<protocol>
		parsedHost := strings.SplitN(host, ";", 3)

		// cast port to int
		port, err := strconv.Atoi(parsedHost[1])
		if err != nil {
			return nil, fmt.Errorf("unexpected port parameter in %s: %v", host, err)
		}

		// generate host spec
		parsedHosts = append(parsedHosts, NetworkDisruptionHostSpec{
			Host:     parsedHost[0],
			Port:     port,
			Protocol: parsedHost[2],
		})
	}

	return parsedHosts, nil
}

// NetworkDisruptionServiceSpecFromString parses the given services to service specs
// The expected format for services is <serviceName>;<serviceNamespace>
func NetworkDisruptionServiceSpecFromString(services []string) ([]NetworkDisruptionServiceSpec, error) {
	parsedServices := []NetworkDisruptionServiceSpec{}

	// parse given services
	for _, service := range services {
		// parse service with format <name>;<namespace>
		parsedService := strings.Split(service, ";")
		if len(parsedService) != 2 {
			return nil, fmt.Errorf("unexpected service format: %s", service)
		}

		// generate service spec
		parsedServices = append(parsedServices, NetworkDisruptionServiceSpec{
			Name:      parsedService[0],
			Namespace: parsedService[1],
		})
	}

	return parsedServices, nil
}
