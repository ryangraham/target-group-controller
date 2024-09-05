package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// TargetGroupBindingSpec defines the desired state of TargetGroupBinding
type TargetGroupBindingSpec struct {
	TargetGroupARN string            `json:"targetGroupARN"`
	NodeSelector   map[string]string `json:"nodeSelector"`
}

// TargetGroupBindingStatus defines the observed state of TargetGroupBinding
type TargetGroupBindingStatus struct {
	CurrentNodes    []string    `json:"currentNodes,omitempty"`
	TotalNodes      int         `json:"totalNodes,omitempty"`
	RegisteredNodes int         `json:"registeredNodes,omitempty"`
	LastSyncTime    metav1.Time `json:"lastSyncTime,omitempty"`
}

// TargetGroupBinding is the Schema for the targetgroupbindings API
type TargetGroupBinding struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TargetGroupBindingSpec   `json:"spec,omitempty"`
	Status TargetGroupBindingStatus `json:"status,omitempty"`
}

// DeepCopyObject is required to implement runtime.Object
func (in *TargetGroupBinding) DeepCopyObject() runtime.Object {
	if in == nil {
		return nil
	}
	out := new(TargetGroupBinding)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto copies the receiver and writes into out. in must be non-nil.
func (in *TargetGroupBinding) DeepCopyInto(out *TargetGroupBinding) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ObjectMeta = *in.ObjectMeta.DeepCopy()
	out.Spec = in.Spec
	out.Status = in.Status
}

// TargetGroupBindingList contains a list of TargetGroupBinding
type TargetGroupBindingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TargetGroupBinding `json:"items"`
}

// DeepCopyObject is required to implement runtime.Object
func (in *TargetGroupBindingList) DeepCopyObject() runtime.Object {
	if in == nil {
		return nil
	}
	out := new(TargetGroupBindingList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto copies the receiver and writes into out. in must be non-nil.
func (in *TargetGroupBindingList) DeepCopyInto(out *TargetGroupBindingList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ListMeta = *in.ListMeta.DeepCopy()
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]TargetGroupBinding, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
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
	return nil
}
