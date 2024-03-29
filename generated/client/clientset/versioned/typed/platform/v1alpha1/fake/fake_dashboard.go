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

// FakeDashboards implements DashboardInterface
type FakeDashboards struct {
	Fake *FakePlatformV1alpha1
	ns   string
}

var dashboardsResource = schema.GroupVersionResource{Group: "platform.plural.sh", Version: "v1alpha1", Resource: "dashboards"}

var dashboardsKind = schema.GroupVersionKind{Group: "platform.plural.sh", Version: "v1alpha1", Kind: "Dashboard"}

// Get takes name of the dashboard, and returns the corresponding dashboard object, and an error if there is any.
func (c *FakeDashboards) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.Dashboard, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(dashboardsResource, c.ns, name), &v1alpha1.Dashboard{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Dashboard), err
}

// List takes label and field selectors, and returns the list of Dashboards that match those selectors.
func (c *FakeDashboards) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.DashboardList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(dashboardsResource, dashboardsKind, c.ns, opts), &v1alpha1.DashboardList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.DashboardList{ListMeta: obj.(*v1alpha1.DashboardList).ListMeta}
	for _, item := range obj.(*v1alpha1.DashboardList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested dashboards.
func (c *FakeDashboards) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(dashboardsResource, c.ns, opts))

}

// Create takes the representation of a dashboard and creates it.  Returns the server's representation of the dashboard, and an error, if there is any.
func (c *FakeDashboards) Create(ctx context.Context, dashboard *v1alpha1.Dashboard, opts v1.CreateOptions) (result *v1alpha1.Dashboard, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(dashboardsResource, c.ns, dashboard), &v1alpha1.Dashboard{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Dashboard), err
}

// Update takes the representation of a dashboard and updates it. Returns the server's representation of the dashboard, and an error, if there is any.
func (c *FakeDashboards) Update(ctx context.Context, dashboard *v1alpha1.Dashboard, opts v1.UpdateOptions) (result *v1alpha1.Dashboard, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(dashboardsResource, c.ns, dashboard), &v1alpha1.Dashboard{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Dashboard), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeDashboards) UpdateStatus(ctx context.Context, dashboard *v1alpha1.Dashboard, opts v1.UpdateOptions) (*v1alpha1.Dashboard, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(dashboardsResource, "status", c.ns, dashboard), &v1alpha1.Dashboard{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Dashboard), err
}

// Delete takes name of the dashboard and deletes it. Returns an error if one occurs.
func (c *FakeDashboards) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteActionWithOptions(dashboardsResource, c.ns, name, opts), &v1alpha1.Dashboard{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeDashboards) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(dashboardsResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha1.DashboardList{})
	return err
}

// Patch applies the patch and returns the patched dashboard.
func (c *FakeDashboards) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.Dashboard, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(dashboardsResource, c.ns, name, pt, data, subresources...), &v1alpha1.Dashboard{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Dashboard), err
}
