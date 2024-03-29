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

// FakeRunbooks implements RunbookInterface
type FakeRunbooks struct {
	Fake *FakePlatformV1alpha1
	ns   string
}

var runbooksResource = schema.GroupVersionResource{Group: "platform.plural.sh", Version: "v1alpha1", Resource: "runbooks"}

var runbooksKind = schema.GroupVersionKind{Group: "platform.plural.sh", Version: "v1alpha1", Kind: "Runbook"}

// Get takes name of the runbook, and returns the corresponding runbook object, and an error if there is any.
func (c *FakeRunbooks) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.Runbook, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(runbooksResource, c.ns, name), &v1alpha1.Runbook{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Runbook), err
}

// List takes label and field selectors, and returns the list of Runbooks that match those selectors.
func (c *FakeRunbooks) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.RunbookList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(runbooksResource, runbooksKind, c.ns, opts), &v1alpha1.RunbookList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.RunbookList{ListMeta: obj.(*v1alpha1.RunbookList).ListMeta}
	for _, item := range obj.(*v1alpha1.RunbookList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested runbooks.
func (c *FakeRunbooks) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(runbooksResource, c.ns, opts))

}

// Create takes the representation of a runbook and creates it.  Returns the server's representation of the runbook, and an error, if there is any.
func (c *FakeRunbooks) Create(ctx context.Context, runbook *v1alpha1.Runbook, opts v1.CreateOptions) (result *v1alpha1.Runbook, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(runbooksResource, c.ns, runbook), &v1alpha1.Runbook{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Runbook), err
}

// Update takes the representation of a runbook and updates it. Returns the server's representation of the runbook, and an error, if there is any.
func (c *FakeRunbooks) Update(ctx context.Context, runbook *v1alpha1.Runbook, opts v1.UpdateOptions) (result *v1alpha1.Runbook, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(runbooksResource, c.ns, runbook), &v1alpha1.Runbook{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Runbook), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeRunbooks) UpdateStatus(ctx context.Context, runbook *v1alpha1.Runbook, opts v1.UpdateOptions) (*v1alpha1.Runbook, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(runbooksResource, "status", c.ns, runbook), &v1alpha1.Runbook{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Runbook), err
}

// Delete takes name of the runbook and deletes it. Returns an error if one occurs.
func (c *FakeRunbooks) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteActionWithOptions(runbooksResource, c.ns, name, opts), &v1alpha1.Runbook{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeRunbooks) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(runbooksResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha1.RunbookList{})
	return err
}

// Patch applies the patch and returns the patched runbook.
func (c *FakeRunbooks) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.Runbook, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(runbooksResource, c.ns, name, pt, data, subresources...), &v1alpha1.Runbook{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Runbook), err
}
