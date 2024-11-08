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

package controllers

import (
	"context"
	"testing"

	"github.com/edgenesis/shifu/pkg/k8s/api/v1alpha1"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func setupScheme(scheme *runtime.Scheme) {
	_ = v1alpha1.AddToScheme(scheme)
	_ = appsv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)
}

func TestCreateServiceIfNotExists(t *testing.T) {
	// Setup the test environment
	scheme := runtime.NewScheme()
	setupScheme(scheme)
	// Create a fake client for the reconciler
	fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()

	// Create a TelemetryService instance
	ts := &v1alpha1.TelemetryService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "telemetryservice",
			Namespace: "shifu-service",
		},
	}

	// Create the reconciler
	reconciler := &TelemetryServiceReconciler{
		Client: fakeClient,
		Scheme: scheme,
	}

	// Call CreateServiceIfNotExists function
	err := CreateServiceIfNotExists(context.Background(), reconciler, ts, reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      "telemetryservice",
			Namespace: "shifu-service",
		},
	})

	// Assert that there were no errors
	assert.NoError(t, err)

	// Assert that the Service was created
	service := &corev1.Service{}
	err = fakeClient.Get(context.Background(), types.NamespacedName{Name: "telemetryservice", Namespace: "shifu-service"}, service)
	assert.NoError(t, err)
	assert.NotNil(t, service)
}

func TestCreateDeploymentIfNotExists(t *testing.T) {
	// Setup the test environment
	scheme := runtime.NewScheme()
	setupScheme(scheme)

	// Create a fake client for the reconciler
	fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()

	// Create a TelemetryService instance
	ts := &v1alpha1.TelemetryService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "telemetryservice",
			Namespace: "shifu-service",
		},
	}

	// Create the reconciler
	reconciler := &TelemetryServiceReconciler{
		Client: fakeClient,
		Scheme: scheme,
	}

	// Call CreateDeploymentIfNotExists function
	err := CreateDeploymentIfNotExists(context.Background(), reconciler, ts, reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      "telemetryservice",
			Namespace: "shifu-service",
		},
	})

	// Assert that there were no errors
	assert.NoError(t, err)

	// Assert that the Deployment was created
	deployment := &appsv1.Deployment{}
	err = fakeClient.Get(context.Background(), types.NamespacedName{Name: "telemetryservice", Namespace: "shifu-service"}, deployment)
	assert.NoError(t, err)
	assert.NotNil(t, deployment)
}

func TestTelemetryServiceReconcile(t *testing.T) {
	// Setup the test environment
	scheme := runtime.NewScheme()
	setupScheme(scheme)

	// Create a fake client for the reconciler
	fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()

	// Create a TelemetryService instance
	ts := &v1alpha1.TelemetryService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "telemetryservice",
			Namespace: "shifu-service",
		},
	}
	// Add the above objects to the fake client
	err := fakeClient.Create(context.Background(), ts)
	assert.NoError(t, err)
	// Create a reconcile.Request
	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      "telemetryservice",
			Namespace: "shifu-service",
		},
	}

	// Create the reconciler
	reconciler := &TelemetryServiceReconciler{
		Client: fakeClient,
		Scheme: scheme,
	}

	// Call the Reconcile function
	result, err := reconciler.Reconcile(context.Background(), req)
	// Assert that there were no errors
	assert.NoError(t, err)
	// Assert that the result does not requeue
	assert.False(t, result.Requeue)

	// Check if Deployment and Service instances were created
	deployment := &appsv1.Deployment{}
	err = fakeClient.Get(context.Background(), req.NamespacedName, deployment)
	assert.NoError(t, err)
	assert.NotNil(t, deployment)

	service := &corev1.Service{}
	err = fakeClient.Get(context.Background(), req.NamespacedName, service)
	assert.NoError(t, err)
	assert.NotNil(t, service)
}
