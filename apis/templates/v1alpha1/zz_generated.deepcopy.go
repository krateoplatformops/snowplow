//go:build !ignore_autogenerated

/*
Copyright 2024 Krateo SRL.

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

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *API) DeepCopyInto(out *API) {
	*out = *in
	if in.Path != nil {
		in, out := &in.Path, &out.Path
		*out = new(string)
		**out = **in
	}
	if in.Verb != nil {
		in, out := &in.Verb, &out.Verb
		*out = new(string)
		**out = **in
	}
	if in.Headers != nil {
		in, out := &in.Headers, &out.Headers
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Payload != nil {
		in, out := &in.Payload, &out.Payload
		*out = new(string)
		**out = **in
	}
	if in.EndpointRef != nil {
		in, out := &in.EndpointRef, &out.EndpointRef
		*out = new(Reference)
		**out = **in
	}
	if in.DependOn != nil {
		in, out := &in.DependOn, &out.DependOn
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new API.
func (in *API) DeepCopy() *API {
	if in == nil {
		return nil
	}
	out := new(API)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Action) DeepCopyInto(out *Action) {
	*out = *in
	if in.Template != nil {
		in, out := &in.Template, &out.Template
		*out = new(ActionTemplate)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Action.
func (in *Action) DeepCopy() *Action {
	if in == nil {
		return nil
	}
	out := new(Action)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ActionResult) DeepCopyInto(out *ActionResult) {
	*out = *in
	if in.PayloadToOverride != nil {
		in, out := &in.PayloadToOverride, &out.PayloadToOverride
		*out = make([]Data, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.Payload != nil {
		in, out := &in.Payload, &out.Payload
		*out = new(ActionResultPayload)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ActionResult.
func (in *ActionResult) DeepCopy() *ActionResult {
	if in == nil {
		return nil
	}
	out := new(ActionResult)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ActionResultPayload) DeepCopyInto(out *ActionResultPayload) {
	*out = *in
	if in.MetaData != nil {
		in, out := &in.MetaData, &out.MetaData
		*out = new(Reference)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ActionResultPayload.
func (in *ActionResultPayload) DeepCopy() *ActionResultPayload {
	if in == nil {
		return nil
	}
	out := new(ActionResultPayload)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ActionResultTemplate) DeepCopyInto(out *ActionResultTemplate) {
	*out = *in
	if in.Template != nil {
		in, out := &in.Template, &out.Template
		*out = new(ActionResult)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ActionResultTemplate.
func (in *ActionResultTemplate) DeepCopy() *ActionResultTemplate {
	if in == nil {
		return nil
	}
	out := new(ActionResultTemplate)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ActionTemplate) DeepCopyInto(out *ActionTemplate) {
	*out = *in
	if in.PayloadToOverride != nil {
		in, out := &in.PayloadToOverride, &out.PayloadToOverride
		*out = make([]Data, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ActionTemplate.
func (in *ActionTemplate) DeepCopy() *ActionTemplate {
	if in == nil {
		return nil
	}
	out := new(ActionTemplate)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CustomForm) DeepCopyInto(out *CustomForm) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CustomForm.
func (in *CustomForm) DeepCopy() *CustomForm {
	if in == nil {
		return nil
	}
	out := new(CustomForm)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *CustomForm) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CustomFormApp) DeepCopyInto(out *CustomFormApp) {
	*out = *in
	if in.Template != nil {
		in, out := &in.Template, &out.Template
		*out = new(CustomFormAppTemplate)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CustomFormApp.
func (in *CustomFormApp) DeepCopy() *CustomFormApp {
	if in == nil {
		return nil
	}
	out := new(CustomFormApp)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CustomFormAppTemplate) DeepCopyInto(out *CustomFormAppTemplate) {
	*out = *in
	if in.PropertiesToHide != nil {
		in, out := &in.PropertiesToHide, &out.PropertiesToHide
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.PropertiesToOverride != nil {
		in, out := &in.PropertiesToOverride, &out.PropertiesToOverride
		*out = make([]Data, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CustomFormAppTemplate.
func (in *CustomFormAppTemplate) DeepCopy() *CustomFormAppTemplate {
	if in == nil {
		return nil
	}
	out := new(CustomFormAppTemplate)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CustomFormList) DeepCopyInto(out *CustomFormList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]CustomForm, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CustomFormList.
func (in *CustomFormList) DeepCopy() *CustomFormList {
	if in == nil {
		return nil
	}
	out := new(CustomFormList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *CustomFormList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CustomFormSpec) DeepCopyInto(out *CustomFormSpec) {
	*out = *in
	if in.PropsRef != nil {
		in, out := &in.PropsRef, &out.PropsRef
		*out = new(Reference)
		**out = **in
	}
	if in.Actions != nil {
		in, out := &in.Actions, &out.Actions
		*out = make([]*Action, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(Action)
				(*in).DeepCopyInto(*out)
			}
		}
	}
	if in.App != nil {
		in, out := &in.App, &out.App
		*out = new(CustomFormApp)
		(*in).DeepCopyInto(*out)
	}
	if in.API != nil {
		in, out := &in.API, &out.API
		*out = make([]*API, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(API)
				(*in).DeepCopyInto(*out)
			}
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CustomFormSpec.
func (in *CustomFormSpec) DeepCopy() *CustomFormSpec {
	if in == nil {
		return nil
	}
	out := new(CustomFormSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CustomFormStatus) DeepCopyInto(out *CustomFormStatus) {
	*out = *in
	if in.UID != nil {
		in, out := &in.UID, &out.UID
		*out = new(string)
		**out = **in
	}
	if in.Props != nil {
		in, out := &in.Props, &out.Props
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.Content != nil {
		in, out := &in.Content, &out.Content
		*out = new(CustomFormStatusContent)
		(*in).DeepCopyInto(*out)
	}
	if in.Actions != nil {
		in, out := &in.Actions, &out.Actions
		*out = make([]*ActionResultTemplate, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(ActionResultTemplate)
				(*in).DeepCopyInto(*out)
			}
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CustomFormStatus.
func (in *CustomFormStatus) DeepCopy() *CustomFormStatus {
	if in == nil {
		return nil
	}
	out := new(CustomFormStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CustomFormStatusContent) DeepCopyInto(out *CustomFormStatusContent) {
	*out = *in
	if in.Schema != nil {
		in, out := &in.Schema, &out.Schema
		*out = new(runtime.RawExtension)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CustomFormStatusContent.
func (in *CustomFormStatusContent) DeepCopy() *CustomFormStatusContent {
	if in == nil {
		return nil
	}
	out := new(CustomFormStatusContent)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Data) DeepCopyInto(out *Data) {
	*out = *in
	if in.AsString != nil {
		in, out := &in.AsString, &out.AsString
		*out = new(bool)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Data.
func (in *Data) DeepCopy() *Data {
	if in == nil {
		return nil
	}
	out := new(Data)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Form) DeepCopyInto(out *Form) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Form.
func (in *Form) DeepCopy() *Form {
	if in == nil {
		return nil
	}
	out := new(Form)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Form) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FormList) DeepCopyInto(out *FormList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Form, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FormList.
func (in *FormList) DeepCopy() *FormList {
	if in == nil {
		return nil
	}
	out := new(FormList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *FormList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FormSpec) DeepCopyInto(out *FormSpec) {
	*out = *in
	if in.SchemaDefinitionRef != nil {
		in, out := &in.SchemaDefinitionRef, &out.SchemaDefinitionRef
		*out = new(Reference)
		**out = **in
	}
	if in.CompositionDefinitionRef != nil {
		in, out := &in.CompositionDefinitionRef, &out.CompositionDefinitionRef
		*out = new(Reference)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FormSpec.
func (in *FormSpec) DeepCopy() *FormSpec {
	if in == nil {
		return nil
	}
	out := new(FormSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FormStatus) DeepCopyInto(out *FormStatus) {
	*out = *in
	if in.Content != nil {
		in, out := &in.Content, &out.Content
		*out = new(FormStatusContent)
		(*in).DeepCopyInto(*out)
	}
	if in.Actions != nil {
		in, out := &in.Actions, &out.Actions
		*out = make([]*Action, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(Action)
				(*in).DeepCopyInto(*out)
			}
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FormStatus.
func (in *FormStatus) DeepCopy() *FormStatus {
	if in == nil {
		return nil
	}
	out := new(FormStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FormStatusContent) DeepCopyInto(out *FormStatusContent) {
	*out = *in
	if in.Schema != nil {
		in, out := &in.Schema, &out.Schema
		*out = new(runtime.RawExtension)
		(*in).DeepCopyInto(*out)
	}
	if in.Instance != nil {
		in, out := &in.Instance, &out.Instance
		*out = new(runtime.RawExtension)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FormStatusContent.
func (in *FormStatusContent) DeepCopy() *FormStatusContent {
	if in == nil {
		return nil
	}
	out := new(FormStatusContent)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ObjectReference) DeepCopyInto(out *ObjectReference) {
	*out = *in
	out.Reference = in.Reference
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ObjectReference.
func (in *ObjectReference) DeepCopy() *ObjectReference {
	if in == nil {
		return nil
	}
	out := new(ObjectReference)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Reference) DeepCopyInto(out *Reference) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Reference.
func (in *Reference) DeepCopy() *Reference {
	if in == nil {
		return nil
	}
	out := new(Reference)
	in.DeepCopyInto(out)
	return out
}
