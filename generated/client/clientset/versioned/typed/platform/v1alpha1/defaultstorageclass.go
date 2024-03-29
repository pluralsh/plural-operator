/*
Copyright 2021.

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
// Code generated by client-gen. DO NOT EDIT.

package v1alpha1

import (
	"context"
	"time"

	v1alpha1 "github.com/pluralsh/plural-operator/apis/platform/v1alpha1"
	scheme "github.com/pluralsh/plural-operator/generated/client/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// DefaultStorageClassesGetter has a method to return a DefaultStorageClassInterface.
// A group's client should implement this interface.
type DefaultStorageClassesGetter interface {
	DefaultStorageClasses(namespace string) DefaultStorageClassInterface
}

// DefaultStorageClassInterface has methods to work with DefaultStorageClass resources.
type DefaultStorageClassInterface interface {
	Create(ctx context.Context, defaultStorageClass *v1alpha1.DefaultStorageClass, opts v1.CreateOptions) (*v1alpha1.DefaultStorageClass, error)
	Update(ctx context.Context, defaultStorageClass *v1alpha1.DefaultStorageClass, opts v1.UpdateOptions) (*v1alpha1.DefaultStorageClass, error)
	UpdateStatus(ctx context.Context, defaultStorageClass *v1alpha1.DefaultStorageClass, opts v1.UpdateOptions) (*v1alpha1.DefaultStorageClass, error)
	Delete(ctx context.Context, name string, opts v1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error
	Get(ctx context.Context, name string, opts v1.GetOptions) (*v1alpha1.DefaultStorageClass, error)
	List(ctx context.Context, opts v1.ListOptions) (*v1alpha1.DefaultStorageClassList, error)
	Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.DefaultStorageClass, err error)
	DefaultStorageClassExpansion
}

// defaultStorageClasses implements DefaultStorageClassInterface
type defaultStorageClasses struct {
	client rest.Interface
	ns     string
}

// newDefaultStorageClasses returns a DefaultStorageClasses
func newDefaultStorageClasses(c *PlatformV1alpha1Client, namespace string) *defaultStorageClasses {
	return &defaultStorageClasses{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the defaultStorageClass, and returns the corresponding defaultStorageClass object, and an error if there is any.
func (c *defaultStorageClasses) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.DefaultStorageClass, err error) {
	result = &v1alpha1.DefaultStorageClass{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("defaultstorageclasses").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of DefaultStorageClasses that match those selectors.
func (c *defaultStorageClasses) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.DefaultStorageClassList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1alpha1.DefaultStorageClassList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("defaultstorageclasses").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested defaultStorageClasses.
func (c *defaultStorageClasses) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("defaultstorageclasses").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch(ctx)
}

// Create takes the representation of a defaultStorageClass and creates it.  Returns the server's representation of the defaultStorageClass, and an error, if there is any.
func (c *defaultStorageClasses) Create(ctx context.Context, defaultStorageClass *v1alpha1.DefaultStorageClass, opts v1.CreateOptions) (result *v1alpha1.DefaultStorageClass, err error) {
	result = &v1alpha1.DefaultStorageClass{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("defaultstorageclasses").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(defaultStorageClass).
		Do(ctx).
		Into(result)
	return
}

// Update takes the representation of a defaultStorageClass and updates it. Returns the server's representation of the defaultStorageClass, and an error, if there is any.
func (c *defaultStorageClasses) Update(ctx context.Context, defaultStorageClass *v1alpha1.DefaultStorageClass, opts v1.UpdateOptions) (result *v1alpha1.DefaultStorageClass, err error) {
	result = &v1alpha1.DefaultStorageClass{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("defaultstorageclasses").
		Name(defaultStorageClass.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(defaultStorageClass).
		Do(ctx).
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *defaultStorageClasses) UpdateStatus(ctx context.Context, defaultStorageClass *v1alpha1.DefaultStorageClass, opts v1.UpdateOptions) (result *v1alpha1.DefaultStorageClass, err error) {
	result = &v1alpha1.DefaultStorageClass{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("defaultstorageclasses").
		Name(defaultStorageClass.Name).
		SubResource("status").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(defaultStorageClass).
		Do(ctx).
		Into(result)
	return
}

// Delete takes name of the defaultStorageClass and deletes it. Returns an error if one occurs.
func (c *defaultStorageClasses) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("defaultstorageclasses").
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *defaultStorageClasses) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	var timeout time.Duration
	if listOpts.TimeoutSeconds != nil {
		timeout = time.Duration(*listOpts.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Namespace(c.ns).
		Resource("defaultstorageclasses").
		VersionedParams(&listOpts, scheme.ParameterCodec).
		Timeout(timeout).
		Body(&opts).
		Do(ctx).
		Error()
}

// Patch applies the patch and returns the patched defaultStorageClass.
func (c *defaultStorageClasses) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.DefaultStorageClass, err error) {
	result = &v1alpha1.DefaultStorageClass{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("defaultstorageclasses").
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}
