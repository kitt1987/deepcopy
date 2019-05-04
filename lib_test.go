package deepcopy_test

import (
	"github.com/kitt1987/deepcopy"
	"gotest.tools/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

type simpleStruct struct {
	FieldA string
	FieldB int
	FieldC float32
}

func TestSimpleStructWithSingleField(t *testing.T) {
	src := simpleStruct{
		FieldA: "FieldA",
		FieldB: 101,
		FieldC: 9.9,
	}

	var dst simpleStruct
	assert.Assert(t, deepcopy.Partial(&dst, &src, "FieldA"))

	t.Logf("%#v", dst)

	assert.Equal(t, dst.FieldA, src.FieldA)
	assert.Assert(t, dst.FieldB != src.FieldB)
	assert.Assert(t, dst.FieldC != src.FieldC)

	dupFieldA := simpleStruct{
		FieldA: src.FieldA,
		FieldB: 501,
		FieldC: 59.9,
	}

	assert.Assert(t, !deepcopy.OnChange(&dupFieldA, &src, "FieldA"))
	t.Logf("%#v", dupFieldA)
	assert.Assert(t, dupFieldA.FieldA == src.FieldA)
	assert.Assert(t, dupFieldA.FieldB == 501)
	assert.Assert(t, dupFieldA.FieldC == 59.9)
}

func TestNothingCopiedSimpleStructWithSingleField(t *testing.T) {
	src := simpleStruct{
		FieldB: 101,
		FieldC: 9.9,
	}

	var dst simpleStruct
	assert.Assert(t, !deepcopy.Partial(&dst, &src, "FieldA"))

	assert.Equal(t, dst.FieldA, "")
	assert.Assert(t, dst.FieldB == 0)
	assert.Assert(t, dst.FieldC == 0)
}

func TestSimpleStructWithMultipleFields(t *testing.T) {
	src := simpleStruct{
		FieldA: "FieldA",
		FieldB: 101,
		FieldC: 9.9,
	}

	var dst simpleStruct
	assert.Assert(t, deepcopy.Partial(&dst, &src, "FieldA", "FieldC"))

	t.Logf("%#v", dst)

	assert.Assert(t, dst.FieldA == src.FieldA)
	assert.Assert(t, dst.FieldC == src.FieldC)
	assert.Assert(t, dst.FieldB != src.FieldB)

	dupFieldAAndC := simpleStruct{
		FieldA: src.FieldA,
		FieldB: 501,
		FieldC: src.FieldC,
	}

	assert.Assert(t, !deepcopy.OnChange(&dupFieldAAndC, &src, "FieldA", "FieldC"))
	t.Logf("%#v", dupFieldAAndC)
	assert.Assert(t, dupFieldAAndC.FieldA == src.FieldA)
	assert.Assert(t, dupFieldAAndC.FieldB == 501)
	assert.Assert(t, dupFieldAAndC.FieldC == src.FieldC)
}

type structWithSliceOfPointers struct {
	StringA  string
	IntA     int
	SliceA   []*simpleStruct
	Replicas *int
}

func TestStructWithFieldsInSlice(t *testing.T) {
	srcReplica := 1
	src := structWithSliceOfPointers{
		StringA:  "StringA",
		IntA:     101,
		Replicas: &srcReplica,
		SliceA: []*simpleStruct{
			{
				FieldA: "SliceA",
				FieldB: 102,
				FieldC: 91.9,
			},
			{
				FieldA: "SliceB",
				FieldB: 101,
				FieldC: 9.9,
			},
		},
	}

	var dst structWithSliceOfPointers
	assert.Assert(t, deepcopy.Partial(&dst, &src, "SliceA.FieldA", "SliceA.FieldC", "IntA"))

	t.Logf("%#v", dst)

	assert.Assert(t, len(dst.SliceA) == len(src.SliceA))
	assert.Assert(t, dst.StringA != src.StringA)
	assert.Assert(t, dst.IntA == src.IntA)

	for i := range dst.SliceA {
		assert.Assert(t, dst.SliceA[i].FieldA == src.SliceA[i].FieldA)
		assert.Assert(t, dst.SliceA[i].FieldC == src.SliceA[i].FieldC)
	}

	withouSlice := structWithSliceOfPointers{
		StringA: "StringA",
		IntA:    101,
	}

	assert.Assert(t, deepcopy.OnChange(&withouSlice, &src, "SliceA"))
	t.Logf("%#v", withouSlice)
	assert.Assert(t, len(withouSlice.SliceA) == len(src.SliceA))

	dupSrc := structWithSliceOfPointers{
		StringA: "StringA",
		IntA:    101,
		SliceA: []*simpleStruct{
			{
				FieldA: "SliceA",
				FieldB: 102,
				FieldC: 91.9,
			},
			{
				FieldA: "SliceB",
				FieldB: 101,
				FieldC: 9.9,
			},
		},
	}

	assert.Assert(t, !deepcopy.OnChange(&dupSrc, &src, "SliceA"))

	dupReplica := 2
	dupSrc.Replicas = &dupReplica

	assert.Assert(t, deepcopy.OnChange(&dupSrc, &src, "Replicas"))
}

func TestStructWithSliceInSlice(t *testing.T) {
	justFalse := false
	justDigit := int64(10)
	src := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "podA",
			Namespace: "namespaceA",
			Labels: map[string]string{
				"A": "B",
			},
		},
		Spec: v1.PodSpec{
			InitContainers: []v1.Container{
				{
					Name: "containerA",
					LivenessProbe: &v1.Probe{
						InitialDelaySeconds: 121,
					},
					Ports: []v1.ContainerPort{
						{
							Name:     "port-81",
							HostIP:   "1.1.1.1",
							HostPort: 81,
						},
					},
				},
				{
					Name: "containerB",
					LivenessProbe: &v1.Probe{
						InitialDelaySeconds: 122,
					},
					Ports: []v1.ContainerPort{
						{
							Name:     "port-80",
							HostIP:   "2.2.2.2",
							HostPort: 80,
						},
						{
							Name:     "port-32767",
							HostIP:   "2.2.2.2",
							HostPort: 32767,
						},
					},
				},
				{
					Name: "containerC",
					LivenessProbe: &v1.Probe{
						InitialDelaySeconds: 125,
					},
					Ports: []v1.ContainerPort{
						{
							Name:     "port-82",
							HostIP:   "3.3.3.3",
							HostPort: 82,
						},
					},
				},
			},
			SecurityContext: &v1.PodSecurityContext{
				RunAsNonRoot: &justFalse,
				RunAsUser:    &justDigit,
			},
		},
	}

	var dst v1.Pod
	if !deepcopy.Partial(&dst, &src,
		"ObjectMeta.Name",
		"ObjectMeta.Labels",
		"Spec.SecurityContext.RunAsNonRoot",
		"Spec.InitContainers.Ports.Name",
		"Spec.InitContainers.Ports.HostPort") {
		t.Fail()
	}

	t.Logf("%#v", dst)

	assert.Assert(t, dst.ObjectMeta.Name == src.ObjectMeta.Name)
	assert.Assert(t, len(dst.ObjectMeta.Labels) == len(src.ObjectMeta.Labels))
	assert.Assert(t, dst.Spec.SecurityContext.RunAsNonRoot != src.Spec.SecurityContext.RunAsNonRoot)
	assert.Assert(t, *dst.Spec.SecurityContext.RunAsNonRoot == *src.Spec.SecurityContext.RunAsNonRoot)
	assert.Assert(t, len(dst.Spec.InitContainers) == len(src.Spec.InitContainers))

	for i := range dst.Spec.InitContainers {
		assert.Assert(t, len(dst.Spec.InitContainers[i].Ports) == len(src.Spec.InitContainers[i].Ports))
		assert.Assert(t, dst.Spec.InitContainers[i].Name != src.Spec.InitContainers[i].Name)
		for p := range dst.Spec.InitContainers[i].Ports {
			assert.Assert(t, dst.Spec.InitContainers[i].Ports[p].Name == src.Spec.InitContainers[i].Ports[p].Name)
			assert.Assert(t, dst.Spec.InitContainers[i].Ports[p].HostPort == src.Spec.InitContainers[i].Ports[p].HostPort)
			assert.Assert(t, dst.Spec.InitContainers[i].Ports[p].HostIP != src.Spec.InitContainers[i].Ports[p].HostIP)
		}
	}
}
