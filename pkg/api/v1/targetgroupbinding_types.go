package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// ServiceRef defines the reference to the Kubernetes service whose endpoints should be registered in the target group
type ServiceRef struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

// DeepCopyInto copies the receiver and writes into out. in must be non-nil.
func (in *ServiceRef) DeepCopyInto(out *ServiceRef) {
	*out = *in
}

// DeepCopy creates a new deepcopy of the receiver.
func (in *ServiceRef) DeepCopy() *ServiceRef {
	if in == nil {
		return nil
	}
	out := new(ServiceRef)
	in.DeepCopyInto(out)
	return out
}

// TargetGroupSelector defines the tag-based selector to find the AWS target group
type TargetGroupSelector struct {
	Tags map[string]string `json:"tags"`
}

// DeepCopyInto copies the receiver and writes into out. in must be non-nil.
func (in *TargetGroupSelector) DeepCopyInto(out *TargetGroupSelector) {
	*out = *in
	if in.Tags != nil {
		in, out := &in.Tags, &out.Tags
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy creates a new deepcopy of the receiver.
func (in *TargetGroupSelector) DeepCopy() *TargetGroupSelector {
	if in == nil {
		return nil
	}
	out := new(TargetGroupSelector)
	in.DeepCopyInto(out)
	return out
}

// TargetGroupBindingSpec defines the desired state of TargetGroupBinding
type TargetGroupBindingSpec struct {
	ServiceRef          ServiceRef          `json:"serviceRef"`
	TargetGroupSelector TargetGroupSelector `json:"targetGroupSelector"`
}

// DeepCopyInto copies the receiver and writes into out. in must be non-nil.
func (in *TargetGroupBindingSpec) DeepCopyInto(out *TargetGroupBindingSpec) {
	*out = *in
	in.ServiceRef.DeepCopyInto(&out.ServiceRef)
	in.TargetGroupSelector.DeepCopyInto(&out.TargetGroupSelector)
}

// DeepCopy creates a new deepcopy of the receiver.
func (in *TargetGroupBindingSpec) DeepCopy() *TargetGroupBindingSpec {
	if in == nil {
		return nil
	}
	out := new(TargetGroupBindingSpec)
	in.DeepCopyInto(out)
	return out
}

// TargetGroupBindingStatus defines the observed state of TargetGroupBinding
type TargetGroupBindingStatus struct {
	TargetGroupARN string      `json:"targetGroupARN"`
	LastSyncTime   metav1.Time `json:"lastSyncTime"`
	RegisteredIPs  []string    `json:"registeredIPs"`
}

// DeepCopyInto copies the receiver and writes into out. in must be non-nil.
func (in *TargetGroupBindingStatus) DeepCopyInto(out *TargetGroupBindingStatus) {
	*out = *in
	in.LastSyncTime.DeepCopyInto(&out.LastSyncTime)
	if in.RegisteredIPs != nil {
		in, out := &in.RegisteredIPs, &out.RegisteredIPs
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy creates a new deepcopy of the receiver.
func (in *TargetGroupBindingStatus) DeepCopy() *TargetGroupBindingStatus {
	if in == nil {
		return nil
	}
	out := new(TargetGroupBindingStatus)
	in.DeepCopyInto(out)
	return out
}

// TargetGroupBinding is the Schema for the targetgroupbindings API
type TargetGroupBinding struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TargetGroupBindingSpec   `json:"spec,omitempty"`
	Status TargetGroupBindingStatus `json:"status,omitempty"`
}

// DeepCopyObject creates a new deepcopy of the receiver.
func (in *TargetGroupBinding) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto copies the receiver and writes into out. in must be non-nil.
func (in *TargetGroupBinding) DeepCopyInto(out *TargetGroupBinding) {
	*out = *in
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy creates a new deepcopy of the receiver.
func (in *TargetGroupBinding) DeepCopy() *TargetGroupBinding {
	if in == nil {
		return nil
	}
	out := new(TargetGroupBinding)
	in.DeepCopyInto(out)
	return out
}

// +kubebuilder:object:root=true

// TargetGroupBindingList contains a list of TargetGroupBinding
type TargetGroupBindingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TargetGroupBinding `json:"items"`
}

// DeepCopyObject creates a new deepcopy of the receiver.
func (in *TargetGroupBindingList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto copies the receiver and writes into out. in must be non-nil.
func (in *TargetGroupBindingList) DeepCopyInto(out *TargetGroupBindingList) {
	*out = *in
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]TargetGroupBinding, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy creates a new deepcopy of the receiver.
func (in *TargetGroupBindingList) DeepCopy() *TargetGroupBindingList {
	if in == nil {
		return nil
	}
	out := new(TargetGroupBindingList)
	in.DeepCopyInto(out)
	return out
}


// GroupVersion is the group and version used to register these objects.
var GroupVersion = schema.GroupVersion{
	Group:   "ryangraham.internal",
	Version: "v1",
}

// Register the TargetGroupBinding type into the scheme
var (
	// SchemeBuilder is used to add Go types to the GroupVersionKind scheme.
	SchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)

	// AddToScheme is a function to add this scheme to the given scheme.
	AddToScheme = SchemeBuilder.AddToScheme
)

// addKnownTypes adds our custom types to the API group version scheme.
func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(GroupVersion,
		&TargetGroupBinding{},
		&TargetGroupBindingList{},
	)
	// Add the metav1 types to the scheme (important for status subresources)
	metav1.AddToGroupVersion(scheme, GroupVersion)
	return nil
}
