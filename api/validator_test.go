// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2021 Datadog, Inc.

package api_test

import (
	chaostypes "github.com/DataDog/chaos-controller/types"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/util/intstr"

	. "github.com/DataDog/chaos-controller/api"
	"github.com/DataDog/chaos-controller/api/v1beta1"
)

var _ = Describe("Validator", func() {
	var (
		err       error
		validator DisruptionKind
	)

	JustBeforeEach(func() {
		err = validator.Validate()
	})

	Describe("validating network spec", func() {
		var spec *v1beta1.NetworkDisruptionSpec

		BeforeEach(func() {
			spec = &v1beta1.NetworkDisruptionSpec{}
			validator = spec
		})

		Context("with an empty disruption", func() {
			It("should not validate", func() {
				Expect(err).To(Not(BeNil()))
			})
		})

		Context("with a non-empty disruption", func() {
			BeforeEach(func() {
				spec.Drop = 100
				spec.BandwidthLimit = 100
				spec.Delay = 100
				spec.Corrupt = 100
				spec.Duplicate = 100
			})

			It("should validate", func() {
				Expect(err).To(BeNil())
			})
		})
	})

	Describe("validating disk pressure spec", func() {
		var spec *v1beta1.DiskPressureSpec

		BeforeEach(func() {
			spec = &v1beta1.DiskPressureSpec{}
			validator = spec
		})

		Context("with an empty disruption", func() {
			It("should not validate", func() {
				Expect(err).To(Not(BeNil()))
			})
		})

		Context("with a non-empty disruption", func() {
			BeforeEach(func() {
				spec.Throttling.WriteBytesPerSec = func(i int) *int { return &i }(1024)
				spec.Throttling.ReadBytesPerSec = func(i int) *int { return &i }(2048)
			})

			It("should validate", func() {
				Expect(err).To(BeNil())
			})
		})
	})
})

var _ = Describe("Validator", func() {
	var (
		err       error
		validator *v1beta1.DisruptionSpec
	)

	JustBeforeEach(func() {
		err = validator.Validate()
	})
	Describe("validating container failure spec", func() {
		var spec *v1beta1.DisruptionSpec

		BeforeEach(func() {
			spec = &v1beta1.DisruptionSpec{
				Count:            &intstr.IntOrString{Type: intstr.String, StrVal: "100%"},
				ContainerFailure: &v1beta1.ContainerFailureSpec{},
				Selector:         map[string]string{"foo": "bar"},
			}
			validator = spec
		})

		Context("with level set to node", func() {
			BeforeEach(func() {
				spec.Level = chaostypes.DisruptionLevelNode
			})
			It("should not validate", func() {
				Expect(err).To(Not(BeNil()))
			})
		})

		Context("with level set to pod", func() {
			BeforeEach(func() {
				spec.Level = chaostypes.DisruptionLevelPod
			})
			It("should validate", func() {
				Expect(err).To(BeNil())
			})
		})
	})
})
