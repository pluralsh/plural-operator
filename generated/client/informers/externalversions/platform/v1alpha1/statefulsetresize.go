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
// Code generated by informer-gen. DO NOT EDIT.

package v1alpha1

import (
	"context"
	time "time"

	platformv1alpha1 "github.com/pluralsh/plural-operator/apis/platform/v1alpha1"
	versioned "github.com/pluralsh/plural-operator/generated/client/clientset/versioned"
	internalinterfaces "github.com/pluralsh/plural-operator/generated/client/informers/externalversions/internalinterfaces"
	v1alpha1 "github.com/pluralsh/plural-operator/generated/client/listers/platform/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// StatefulSetResizeInformer provides access to a shared informer and lister for
// StatefulSetResizes.
type StatefulSetResizeInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1alpha1.StatefulSetResizeLister
}

type statefulSetResizeInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
	namespace        string
}

// NewStatefulSetResizeInformer constructs a new informer for StatefulSetResize type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewStatefulSetResizeInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredStatefulSetResizeInformer(client, namespace, resyncPeriod, indexers, nil)
}

// NewFilteredStatefulSetResizeInformer constructs a new informer for StatefulSetResize type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredStatefulSetResizeInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.PlatformV1alpha1().StatefulSetResizes(namespace).List(context.TODO(), options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.PlatformV1alpha1().StatefulSetResizes(namespace).Watch(context.TODO(), options)
			},
		},
		&platformv1alpha1.StatefulSetResize{},
		resyncPeriod,
		indexers,
	)
}

func (f *statefulSetResizeInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredStatefulSetResizeInformer(client, f.namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *statefulSetResizeInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&platformv1alpha1.StatefulSetResize{}, f.defaultInformer)
}

func (f *statefulSetResizeInformer) Lister() v1alpha1.StatefulSetResizeLister {
	return v1alpha1.NewStatefulSetResizeLister(f.Informer().GetIndexer())
}
