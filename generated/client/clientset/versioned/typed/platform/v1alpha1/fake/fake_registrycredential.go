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

package fake

import (
	"context"

	v1alpha1 "github.com/pluralsh/plural-operator/apis/platform/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeRegistryCredentials implements RegistryCredentialInterface
type FakeRegistryCredentials struct {
	Fake *FakePlatformV1alpha1
	ns   string
}

var registrycredentialsResource = schema.GroupVersionResource{Group: "platform.plural.sh", Version: "v1alpha1", Resource: "registrycredentials"}

var registrycredentialsKind = schema.GroupVersionKind{Group: "platform.plural.sh", Version: "v1alpha1", Kind: "RegistryCredential"}

// Get takes name of the registryCredential, and returns the corresponding registryCredential object, and an error if there is any.
func (c *FakeRegistryCredentials) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.RegistryCredential, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(registrycredentialsResource, c.ns, name), &v1alpha1.RegistryCredential{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.RegistryCredential), err
}

// List takes label and field selectors, and returns the list of RegistryCredentials that match those selectors.
func (c *FakeRegistryCredentials) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.RegistryCredentialList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(registrycredentialsResource, registrycredentialsKind, c.ns, opts), &v1alpha1.RegistryCredentialList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.RegistryCredentialList{ListMeta: obj.(*v1alpha1.RegistryCredentialList).ListMeta}
	for _, item := range obj.(*v1alpha1.RegistryCredentialList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested registryCredentials.
func (c *FakeRegistryCredentials) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(registrycredentialsResource, c.ns, opts))

}

// Create takes the representation of a registryCredential and creates it.  Returns the server's representation of the registryCredential, and an error, if there is any.
func (c *FakeRegistryCredentials) Create(ctx context.Context, registryCredential *v1alpha1.RegistryCredential, opts v1.CreateOptions) (result *v1alpha1.RegistryCredential, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(registrycredentialsResource, c.ns, registryCredential), &v1alpha1.RegistryCredential{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.RegistryCredential), err
}

// Update takes the representation of a registryCredential and updates it. Returns the server's representation of the registryCredential, and an error, if there is any.
func (c *FakeRegistryCredentials) Update(ctx context.Context, registryCredential *v1alpha1.RegistryCredential, opts v1.UpdateOptions) (result *v1alpha1.RegistryCredential, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(registrycredentialsResource, c.ns, registryCredential), &v1alpha1.RegistryCredential{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.RegistryCredential), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeRegistryCredentials) UpdateStatus(ctx context.Context, registryCredential *v1alpha1.RegistryCredential, opts v1.UpdateOptions) (*v1alpha1.RegistryCredential, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(registrycredentialsResource, "status", c.ns, registryCredential), &v1alpha1.RegistryCredential{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.RegistryCredential), err
}

// Delete takes name of the registryCredential and deletes it. Returns an error if one occurs.
func (c *FakeRegistryCredentials) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteActionWithOptions(registrycredentialsResource, c.ns, name, opts), &v1alpha1.RegistryCredential{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeRegistryCredentials) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(registrycredentialsResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha1.RegistryCredentialList{})
	return err
}

// Patch applies the patch and returns the patched registryCredential.
func (c *FakeRegistryCredentials) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.RegistryCredential, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(registrycredentialsResource, c.ns, name, pt, data, subresources...), &v1alpha1.RegistryCredential{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.RegistryCredential), err
}
