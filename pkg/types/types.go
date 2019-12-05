package types

import (
	"helm.sh/helm/v3/pkg/release"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// resourceImpl wrapper type
type resourceImpl struct {
	RO runtime.Object
	MO metav1.Object
}

// Resource wrapper type
type Resource interface {
	Kind() string
	GroupVersion() string
	Name() string
	KindName() string
}

// ToResource convert an unstructured object ta a resource
func ToResource(yaml map[string]interface{}) Resource {

	var us interface{} = &unstructured.Unstructured{
		Object: yaml,
	}

	ro, ok := us.(runtime.Object)
	if !ok {
		return nil
	}

	meta, ok := us.(metav1.Object)
	if !ok {
		return nil
	}

	return &resourceImpl{
		RO: ro,
		MO: meta,
	}
}

// Kind get the kind of the resource
func (r *resourceImpl) Kind() string {
	return r.RO.GetObjectKind().GroupVersionKind().Kind
}

// GroupVersion get the group and version of the resource
func (r *resourceImpl) GroupVersion() string {
	return r.RO.GetObjectKind().GroupVersionKind().GroupVersion().String()
}

// Name get the name of the resource
func (r *resourceImpl) Name() string {
	return r.MO.GetName()
}

// KindName get the kind name of the resource
func (r *resourceImpl) KindName() string {
	return r.Kind() + "/" + r.Name()
}

// Options basic options
type Options struct {
	DryRun      bool
	ReleaseName string
	Revision    int
}

func (o Options) Filter() func(rel *release.Release) bool {
	return func(rel *release.Release) bool {
		if rel == nil || rel.Name == "" || rel.Name != o.ReleaseName {
			return false
		}

		if o.Revision > 0 {
			return rel.Version == o.Revision
		}
		return true
	}
}
