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

// SlashCommandLister helps list SlashCommands.
// All objects returned here must be treated as read-only.
type SlashCommandLister interface {
	// List lists all SlashCommands in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.SlashCommand, err error)
	// SlashCommands returns an object that can list and get SlashCommands.
	SlashCommands(namespace string) SlashCommandNamespaceLister
	SlashCommandListerExpansion
}

// slashCommandLister implements the SlashCommandLister interface.
type slashCommandLister struct {
	indexer cache.Indexer
}

// NewSlashCommandLister returns a new SlashCommandLister.
func NewSlashCommandLister(indexer cache.Indexer) SlashCommandLister {
	return &slashCommandLister{indexer: indexer}
}

// List lists all SlashCommands in the indexer.
func (s *slashCommandLister) List(selector labels.Selector) (ret []*v1alpha1.SlashCommand, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.SlashCommand))
	})
	return ret, err
}

// SlashCommands returns an object that can list and get SlashCommands.
func (s *slashCommandLister) SlashCommands(namespace string) SlashCommandNamespaceLister {
	return slashCommandNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// SlashCommandNamespaceLister helps list and get SlashCommands.
// All objects returned here must be treated as read-only.
type SlashCommandNamespaceLister interface {
	// List lists all SlashCommands in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.SlashCommand, err error)
	// Get retrieves the SlashCommand from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1alpha1.SlashCommand, error)
	SlashCommandNamespaceListerExpansion
}

// slashCommandNamespaceLister implements the SlashCommandNamespaceLister
// interface.
type slashCommandNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all SlashCommands in the indexer for a given namespace.
func (s slashCommandNamespaceLister) List(selector labels.Selector) (ret []*v1alpha1.SlashCommand, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.SlashCommand))
	})
	return ret, err
}

// Get retrieves the SlashCommand from the indexer for a given namespace and name.
func (s slashCommandNamespaceLister) Get(name string) (*v1alpha1.SlashCommand, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("slashcommand"), name)
	}
	return obj.(*v1alpha1.SlashCommand), nil
}
