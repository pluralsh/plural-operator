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
// Code generated by lister-gen. DO NOT EDIT.

package v1alpha1

import (
	v1alpha1 "github.com/pluralsh/plural-operator/apis/platform/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// SecretSyncLister helps list SecretSyncs.
// All objects returned here must be treated as read-only.
type SecretSyncLister interface {
	// List lists all SecretSyncs in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.SecretSync, err error)
	// SecretSyncs returns an object that can list and get SecretSyncs.
	SecretSyncs(namespace string) SecretSyncNamespaceLister
	SecretSyncListerExpansion
}

// secretSyncLister implements the SecretSyncLister interface.
type secretSyncLister struct {
	indexer cache.Indexer
}

// NewSecretSyncLister returns a new SecretSyncLister.
func NewSecretSyncLister(indexer cache.Indexer) SecretSyncLister {
	return &secretSyncLister{indexer: indexer}
}

// List lists all SecretSyncs in the indexer.
func (s *secretSyncLister) List(selector labels.Selector) (ret []*v1alpha1.SecretSync, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.SecretSync))
	})
	return ret, err
}

// SecretSyncs returns an object that can list and get SecretSyncs.
func (s *secretSyncLister) SecretSyncs(namespace string) SecretSyncNamespaceLister {
	return secretSyncNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// SecretSyncNamespaceLister helps list and get SecretSyncs.
// All objects returned here must be treated as read-only.
type SecretSyncNamespaceLister interface {
	// List lists all SecretSyncs in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.SecretSync, err error)
	// Get retrieves the SecretSync from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1alpha1.SecretSync, error)
	SecretSyncNamespaceListerExpansion
}

// secretSyncNamespaceLister implements the SecretSyncNamespaceLister
// interface.
type secretSyncNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all SecretSyncs in the indexer for a given namespace.
func (s secretSyncNamespaceLister) List(selector labels.Selector) (ret []*v1alpha1.SecretSync, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.SecretSync))
	})
	return ret, err
}

// Get retrieves the SecretSync from the indexer for a given namespace and name.
func (s secretSyncNamespaceLister) Get(name string) (*v1alpha1.SecretSync, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("secretsync"), name)
	}
	return obj.(*v1alpha1.SecretSync), nil
}