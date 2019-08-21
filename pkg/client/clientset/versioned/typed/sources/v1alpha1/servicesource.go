/*
Copyright 2019 The Knative Authors

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
	v1alpha1 "github.com/n3wscott/sources/pkg/apis/sources/v1alpha1"
	scheme "github.com/n3wscott/sources/pkg/client/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// ServiceSourcesGetter has a method to return a ServiceSourceInterface.
// A group's client should implement this interface.
type ServiceSourcesGetter interface {
	ServiceSources(namespace string) ServiceSourceInterface
}

// ServiceSourceInterface has methods to work with ServiceSource resources.
type ServiceSourceInterface interface {
	Create(*v1alpha1.ServiceSource) (*v1alpha1.ServiceSource, error)
	Update(*v1alpha1.ServiceSource) (*v1alpha1.ServiceSource, error)
	UpdateStatus(*v1alpha1.ServiceSource) (*v1alpha1.ServiceSource, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1alpha1.ServiceSource, error)
	List(opts v1.ListOptions) (*v1alpha1.ServiceSourceList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.ServiceSource, err error)
	ServiceSourceExpansion
}

// serviceSources implements ServiceSourceInterface
type serviceSources struct {
	client rest.Interface
	ns     string
}

// newServiceSources returns a ServiceSources
func newServiceSources(c *SourcesV1alpha1Client, namespace string) *serviceSources {
	return &serviceSources{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the serviceSource, and returns the corresponding serviceSource object, and an error if there is any.
func (c *serviceSources) Get(name string, options v1.GetOptions) (result *v1alpha1.ServiceSource, err error) {
	result = &v1alpha1.ServiceSource{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("servicesources").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of ServiceSources that match those selectors.
func (c *serviceSources) List(opts v1.ListOptions) (result *v1alpha1.ServiceSourceList, err error) {
	result = &v1alpha1.ServiceSourceList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("servicesources").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested serviceSources.
func (c *serviceSources) Watch(opts v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("servicesources").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a serviceSource and creates it.  Returns the server's representation of the serviceSource, and an error, if there is any.
func (c *serviceSources) Create(serviceSource *v1alpha1.ServiceSource) (result *v1alpha1.ServiceSource, err error) {
	result = &v1alpha1.ServiceSource{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("servicesources").
		Body(serviceSource).
		Do().
		Into(result)
	return
}

// Update takes the representation of a serviceSource and updates it. Returns the server's representation of the serviceSource, and an error, if there is any.
func (c *serviceSources) Update(serviceSource *v1alpha1.ServiceSource) (result *v1alpha1.ServiceSource, err error) {
	result = &v1alpha1.ServiceSource{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("servicesources").
		Name(serviceSource.Name).
		Body(serviceSource).
		Do().
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().

func (c *serviceSources) UpdateStatus(serviceSource *v1alpha1.ServiceSource) (result *v1alpha1.ServiceSource, err error) {
	result = &v1alpha1.ServiceSource{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("servicesources").
		Name(serviceSource.Name).
		SubResource("status").
		Body(serviceSource).
		Do().
		Into(result)
	return
}

// Delete takes name of the serviceSource and deletes it. Returns an error if one occurs.
func (c *serviceSources) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("servicesources").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *serviceSources) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("servicesources").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched serviceSource.
func (c *serviceSources) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.ServiceSource, err error) {
	result = &v1alpha1.ServiceSource{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("servicesources").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}