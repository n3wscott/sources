// +build !ignore_autogenerated

/*
Copyright 2019 The Knative Authors

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

// Code generated by deepcopy-gen. DO NOT EDIT.

package v1alpha1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
	v1beta1 "knative.dev/pkg/apis/duck/v1beta1"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *BaseSourceSpec) DeepCopyInto(out *BaseSourceSpec) {
	*out = *in
	in.Sink.DeepCopyInto(&out.Sink)
	if in.CloudEventOverrides != nil {
		in, out := &in.CloudEventOverrides, &out.CloudEventOverrides
		*out = new(v1beta1.CloudEventOverrides)
		(*in).DeepCopyInto(*out)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new BaseSourceSpec.
func (in *BaseSourceSpec) DeepCopy() *BaseSourceSpec {
	if in == nil {
		return nil
	}
	out := new(BaseSourceSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *BaseSourceStatus) DeepCopyInto(out *BaseSourceStatus) {
	*out = *in
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new BaseSourceStatus.
func (in *BaseSourceStatus) DeepCopy() *BaseSourceStatus {
	if in == nil {
		return nil
	}
	out := new(BaseSourceStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *JobSource) DeepCopyInto(out *JobSource) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new JobSource.
func (in *JobSource) DeepCopy() *JobSource {
	if in == nil {
		return nil
	}
	out := new(JobSource)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *JobSource) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *JobSourceList) DeepCopyInto(out *JobSourceList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]JobSource, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new JobSourceList.
func (in *JobSourceList) DeepCopy() *JobSourceList {
	if in == nil {
		return nil
	}
	out := new(JobSourceList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *JobSourceList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *JobSourceSpec) DeepCopyInto(out *JobSourceSpec) {
	*out = *in
	in.BaseSourceSpec.DeepCopyInto(&out.BaseSourceSpec)
	in.JobSpec.DeepCopyInto(&out.JobSpec)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new JobSourceSpec.
func (in *JobSourceSpec) DeepCopy() *JobSourceSpec {
	if in == nil {
		return nil
	}
	out := new(JobSourceSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *JobSourceStatus) DeepCopyInto(out *JobSourceStatus) {
	*out = *in
	in.BaseSourceStatus.DeepCopyInto(&out.BaseSourceStatus)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new JobSourceStatus.
func (in *JobSourceStatus) DeepCopy() *JobSourceStatus {
	if in == nil {
		return nil
	}
	out := new(JobSourceStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ServiceSource) DeepCopyInto(out *ServiceSource) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ServiceSource.
func (in *ServiceSource) DeepCopy() *ServiceSource {
	if in == nil {
		return nil
	}
	out := new(ServiceSource)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ServiceSource) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ServiceSourceList) DeepCopyInto(out *ServiceSourceList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]ServiceSource, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ServiceSourceList.
func (in *ServiceSourceList) DeepCopy() *ServiceSourceList {
	if in == nil {
		return nil
	}
	out := new(ServiceSourceList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ServiceSourceList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ServiceSourceSpec) DeepCopyInto(out *ServiceSourceSpec) {
	*out = *in
	in.BaseSourceSpec.DeepCopyInto(&out.BaseSourceSpec)
	in.ServiceSpec.DeepCopyInto(&out.ServiceSpec)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ServiceSourceSpec.
func (in *ServiceSourceSpec) DeepCopy() *ServiceSourceSpec {
	if in == nil {
		return nil
	}
	out := new(ServiceSourceSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ServiceSourceStatus) DeepCopyInto(out *ServiceSourceStatus) {
	*out = *in
	in.BaseSourceStatus.DeepCopyInto(&out.BaseSourceStatus)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ServiceSourceStatus.
func (in *ServiceSourceStatus) DeepCopy() *ServiceSourceStatus {
	if in == nil {
		return nil
	}
	out := new(ServiceSourceStatus)
	in.DeepCopyInto(out)
	return out
}
