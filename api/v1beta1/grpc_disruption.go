// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2021 Datadog, Inc.

package v1beta1

import (
	"errors"
	"fmt"
	"strings"
)

// GRPCDisruptionSpec represents a gRPC disruption
type GRPCDisruptionSpec []EndpointAlteration

// EndpointAlteration represents an endpoint to disrupt and the corresponding error to return
type EndpointAlteration struct {
	TargetEndpoint string `json:"endpoint,omitempty"`
	// +kubebuilder:validation:Enum=OK;CANCELED_CODE;UNKNOWN_CODE;INVALID_ARGUMENT;DEADLINE_EXCEEDED;NOT_FOUND;ALREADY_EXISTS;PERMISSION_DENIED;RESOURCE_EXHAUSTED;FAILED_PRECONDITION;ABORTED_CODE;OUT_OF_RANGE;UNIMPLEMENTED_CODE;INTERNAL_CODE;UNAVAILABLE_CODE;DATALOSS_CODE;UNAUTHENTICATED_CODE
	ErrorToReturn string `json:"error,omitempty"`
	// +kubebuilder:validation:Enum={}
	OverrideToReturn string `json:"override,omitempty"`
}

// Validate validates that there are no missing hostnames or records for the given grpc disruption spec
func (s GRPCDisruptionSpec) Validate() error {

	if len(s) == 0 {
		return errors.New("the gRPC disruption was selected with no endpoints specified, but endpoints must be specified")
	}

	for _, pair := range s {
		if pair.TargetEndpoint == "" {
			return errors.New("some list items in gRPC disruption are missing endpoints; specify an endpoint for each item in the list")
		} else if (pair.ErrorToReturn != "" && pair.OverrideToReturn != "") || (pair.ErrorToReturn == "" && pair.OverrideToReturn == "") {
			return fmt.Errorf("the gRPC disruption can either return an error or override; specify exactly one for endpoint %s", pair.TargetEndpoint)
		}
	}
	return nil
}

// GenerateArgs generates injection pod arguments for the given spec
func (s GRPCDisruptionSpec) GenerateArgs() []string {
	args := []string{
		"grpc-disruption",
	}

	endpointAlterationArgs := []string{}

	for _, pair := range s {
		var alterationType, alterationValue string
		if pair.ErrorToReturn != "" {
			alterationType = "error"
			alterationValue = pair.ErrorToReturn
		}
		if pair.OverrideToReturn != "" {
			alterationType = "override"
			alterationValue = pair.OverrideToReturn
		}
		arg := fmt.Sprintf("%s;%s;%s", pair.TargetEndpoint, alterationType, alterationValue)

		endpointAlterationArgs = append(endpointAlterationArgs, arg)
	}

	args = append(args, "--endpoint-alterations")

	// Each value passed to --host-record-pairs should be of the form `endpoint;alteration_type;alteration_value`, e.g.
	// `/chaos_dogfood.ChaosDogfood/order;error;ALREADY_EXISTS`
	// `/chaos_dogfood.ChaosDogfood/order;override;{}`
	args = append(args, strings.Split(strings.Join(endpointAlterationArgs, " --endpoint-alterations "), " ")...)

	return args
}
