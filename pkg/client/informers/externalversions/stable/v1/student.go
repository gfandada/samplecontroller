/*
Copyright The Kubernetes Authors.

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

package v1

import (
	time "time"

	stablev1 "github.com/gfandada/samplecontroller/pkg/apis/stable/v1"
	versioned "github.com/gfandada/samplecontroller/pkg/client/clientset/versioned"
	internalinterfaces "github.com/gfandada/samplecontroller/pkg/client/informers/externalversions/internalinterfaces"
	v1 "github.com/gfandada/samplecontroller/pkg/client/listers/stable/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// StudentInformer provides access to a shared informer and lister for
// Students.
type StudentInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1.StudentLister
}

type studentInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
	namespace        string
}

// NewStudentInformer constructs a new informer for Student type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewStudentInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredStudentInformer(client, namespace, resyncPeriod, indexers, nil)
}

// NewFilteredStudentInformer constructs a new informer for Student type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredStudentInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.StableV1().Students(namespace).List(options)
			},
			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.StableV1().Students(namespace).Watch(options)
			},
		},
		&stablev1.Student{},
		resyncPeriod,
		indexers,
	)
}

func (f *studentInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredStudentInformer(client, f.namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *studentInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&stablev1.Student{}, f.defaultInformer)
}

func (f *studentInformer) Lister() v1.StudentLister {
	return v1.NewStudentLister(f.Informer().GetIndexer())
}
