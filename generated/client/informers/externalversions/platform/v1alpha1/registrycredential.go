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

// RegistryCredentialInformer provides access to a shared informer and lister for
// RegistryCredentials.
type RegistryCredentialInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1alpha1.RegistryCredentialLister
}

type registryCredentialInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
	namespace        string
}

// NewRegistryCredentialInformer constructs a new informer for RegistryCredential type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewRegistryCredentialInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredRegistryCredentialInformer(client, namespace, resyncPeriod, indexers, nil)
}

// NewFilteredRegistryCredentialInformer constructs a new informer for RegistryCredential type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredRegistryCredentialInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.PlatformV1alpha1().RegistryCredentials(namespace).List(context.TODO(), options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.PlatformV1alpha1().RegistryCredentials(namespace).Watch(context.TODO(), options)
			},
		},
		&platformv1alpha1.RegistryCredential{},
		resyncPeriod,
		indexers,
	)
}

func (f *registryCredentialInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredRegistryCredentialInformer(client, f.namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *registryCredentialInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&platformv1alpha1.RegistryCredential{}, f.defaultInformer)
}

func (f *registryCredentialInformer) Lister() v1alpha1.RegistryCredentialLister {
	return v1alpha1.NewRegistryCredentialLister(f.Informer().GetIndexer())
}