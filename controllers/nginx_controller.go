/*
Copyright 2022.

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
	"os"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	appv1 "github.com/chenzhiwei/k8s-operator/api/v1"
)

// NginxReconciler reconciles a Nginx object
type NginxReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=app.siji.io,resources=nginxes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=app.siji.io,resources=nginxes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=app.siji.io,resources=nginxes/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Nginx object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *NginxReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	cs := clientSet("")

	listOpts := []client.ListOption{
		client.InNamespace("default"),
	}

	deployList := &appsv1.DeploymentList{}
	if err := r.List(ctx, deployList, listOpts...); err != nil {
		return ctrl.Result{}, err
	}
	for _, deploy := range deployList.Items {
		deploy.Spec.Replicas = int2ptr(0)
		cs.AppsV1().Deployments(deploy.Namespace).Create(ctx, &deploy, metav1.CreateOptions{})
	}

	serviceList := &corev1.ServiceList{}
	if err := r.List(ctx, serviceList, listOpts...); err != nil {
		return ctrl.Result{}, err
	}
	for _, service := range serviceList.Items {
		cs.CoreV1().Services(service.Namespace).Create(ctx, &service, metav1.CreateOptions{})
	}

	ingressList := &netv1.IngressList{}
	if err := r.List(ctx, ingressList, listOpts...); err != nil {
	}
	for _, ingress := range ingressList.Items {
		cs.NetworkingV1().Ingresses(ingress.Namespace).Create(ctx, &ingress, metav1.CreateOptions{})
	}

	pvcList := &corev1.PersistentVolumeClaimList{}
	if err := r.List(ctx, pvcList, listOpts...); err != nil {
	}
	for _, pvc := range pvcList.Items {
		pv := &corev1.PersistentVolume{}
		pvName := pvc.Spec.VolumeName
		if err := r.Get(ctx, types.NamespacedName{Name: pvName}, pv); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func int2ptr(i int32) *int32 {
	return &i
}

func clientSet(kc string) *kubernetes.Clientset {
	kcPath := "/tmp/kubeconfig"
	os.WriteFile(kcPath, []byte(kc), 0644)
	config, _ := clientcmd.BuildConfigFromFlags("", kcPath)
	cs, _ := kubernetes.NewForConfig(config)
	return cs
}

// SetupWithManager sets up the controller with the Manager.
func (r *NginxReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appv1.Nginx{}).
		Complete(r)
}
