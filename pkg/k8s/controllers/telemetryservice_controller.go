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

	"github.com/edgenesis/shifu/pkg/k8s/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// TelemetryServiceReconciler reconciles a TelemetryService object
type TelemetryServiceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

var tsNamespacedName = types.NamespacedName{
	Namespace: "shifu-service",
	Name:      "telemetryservice",
}

// image of telemetryservice deployment
const IMAGE = "edgehub/telemetryservice:v0.75.0-rc1"

//+kubebuilder:rbac:groups=shifu.edgenesis.io,resources=telemetryservices,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=shifu.edgenesis.io,resources=telemetryservices/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=shifu.edgenesis.io,resources=telemetryservices/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the TelemetryService object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.1/pkg/reconcile
func (r *TelemetryServiceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	rlog := log.FromContext(ctx)
	ts := &v1alpha1.TelemetryService{}
	if err := r.Get(ctx, req.NamespacedName, ts); err != nil {
		rlog.Error(err, "Unable to fetch TelemetryService")
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	deploy := &appsv1.Deployment{}
	if err := r.Get(ctx, tsNamespacedName, deploy); err != nil {
		if errors.IsNotFound(err) {
			if err := CreateTelemetryServiceDeployment(ctx, r, ts, req); err != nil {
				rlog.Error(err, "Failed to create TelemetryService deployment")
				return ctrl.Result{}, err
			}
		}
	}
	service := &corev1.Service{}
	if err := r.Get(ctx, tsNamespacedName, service); err != nil {
		if err := CreateTelemetryServiceService(ctx, r, ts, req); err != nil {
			rlog.Error(err, "Failed to create TelemetryService service")
			return ctrl.Result{}, err
		}
	}
	rlog.Info("Reconciling TelemetryService")
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *TelemetryServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.TelemetryService{}).
		Complete(r)
}

func CreateTelemetryServiceService(ctx context.Context, r *TelemetryServiceReconciler, ts *v1alpha1.TelemetryService, req ctrl.Request) error {
	rlog := log.FromContext(ctx)
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      tsNamespacedName.Name,
			Namespace: tsNamespacedName.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{"app": tsNamespacedName.Name},
			Ports: []corev1.ServicePort{
				{
					Port:       80,
					Protocol:   corev1.ProtocolTCP,
					TargetPort: intstr.IntOrString{IntVal: 8080},
				},
			},
			Type: corev1.ServiceTypeLoadBalancer,
		},
	}
	rlog.Info("Start creating TelemetryService service")
	if err := r.Create(ctx, svc); err != nil {
		rlog.Error(err, "Failed to create a new TelemetryService service", "Service.Namespace", svc.Namespace, "Service.Name", svc.Name)
		return err
	}
	return nil
}

func CreateTelemetryServiceDeployment(ctx context.Context, r *TelemetryServiceReconciler, ts *v1alpha1.TelemetryService, req ctrl.Request) error {
	rlog := log.FromContext(ctx)
	deploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      tsNamespacedName.Name,
			Namespace: tsNamespacedName.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": tsNamespacedName.Name},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": tsNamespacedName.Name},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  tsNamespacedName.Name,
							Image: IMAGE,
							Ports: []corev1.ContainerPort{
								{ContainerPort: 8080},
							},
							Env: []corev1.EnvVar{
								{
									Name:  "SERVER_LISTEN_PORT",
									Value: ":8080",
								},
								{
									Name:  "EDGEDEVICE_NAMESPACE",
									Value: "devices",
								},
							},
						},
					},
					ServiceAccountName: "telemetry-service-sa",
				},
			},
		},
	}
	rlog.Info("Start creating TelemetryService deployment")
	if err := r.Create(ctx, deploy); err != nil {
		rlog.Error(err, "Failed to create a new TelemetryService deployment", "Deployment.Namespace", deploy.Namespace, "Deployment.Name", deploy.Name)
		return err
	}
	return nil
}
