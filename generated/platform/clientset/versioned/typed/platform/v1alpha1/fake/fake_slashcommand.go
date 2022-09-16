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

// FakeSlashCommands implements SlashCommandInterface
type FakeSlashCommands struct {
	Fake *FakePlatformV1alpha1
	ns   string
}

var slashcommandsResource = schema.GroupVersionResource{Group: "platform.plural.sh", Version: "v1alpha1", Resource: "slashcommands"}

var slashcommandsKind = schema.GroupVersionKind{Group: "platform.plural.sh", Version: "v1alpha1", Kind: "SlashCommand"}

// Get takes name of the slashCommand, and returns the corresponding slashCommand object, and an error if there is any.
func (c *FakeSlashCommands) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.SlashCommand, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(slashcommandsResource, c.ns, name), &v1alpha1.SlashCommand{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.SlashCommand), err
}

// List takes label and field selectors, and returns the list of SlashCommands that match those selectors.
func (c *FakeSlashCommands) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.SlashCommandList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(slashcommandsResource, slashcommandsKind, c.ns, opts), &v1alpha1.SlashCommandList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.SlashCommandList{ListMeta: obj.(*v1alpha1.SlashCommandList).ListMeta}
	for _, item := range obj.(*v1alpha1.SlashCommandList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested slashCommands.
func (c *FakeSlashCommands) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(slashcommandsResource, c.ns, opts))

}

// Create takes the representation of a slashCommand and creates it.  Returns the server's representation of the slashCommand, and an error, if there is any.
func (c *FakeSlashCommands) Create(ctx context.Context, slashCommand *v1alpha1.SlashCommand, opts v1.CreateOptions) (result *v1alpha1.SlashCommand, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(slashcommandsResource, c.ns, slashCommand), &v1alpha1.SlashCommand{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.SlashCommand), err
}

// Update takes the representation of a slashCommand and updates it. Returns the server's representation of the slashCommand, and an error, if there is any.
func (c *FakeSlashCommands) Update(ctx context.Context, slashCommand *v1alpha1.SlashCommand, opts v1.UpdateOptions) (result *v1alpha1.SlashCommand, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(slashcommandsResource, c.ns, slashCommand), &v1alpha1.SlashCommand{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.SlashCommand), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeSlashCommands) UpdateStatus(ctx context.Context, slashCommand *v1alpha1.SlashCommand, opts v1.UpdateOptions) (*v1alpha1.SlashCommand, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(slashcommandsResource, "status", c.ns, slashCommand), &v1alpha1.SlashCommand{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.SlashCommand), err
}

// Delete takes name of the slashCommand and deletes it. Returns an error if one occurs.
func (c *FakeSlashCommands) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(slashcommandsResource, c.ns, name), &v1alpha1.SlashCommand{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeSlashCommands) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(slashcommandsResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha1.SlashCommandList{})
	return err
}

// Patch applies the patch and returns the patched slashCommand.
func (c *FakeSlashCommands) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.SlashCommand, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(slashcommandsResource, c.ns, name, pt, data, subresources...), &v1alpha1.SlashCommand{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.SlashCommand), err
}
