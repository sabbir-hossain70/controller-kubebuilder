/*
Copyright 2024.

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

package controller

import (
	"context"
	"fmt"
	crdv1alpha1 "github.com/sabbir-hossain70/controller-kubebuilder/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"strings"
	_ "strings"
)

var (
	deployOwnerKey  = "customkind-sample"
	serviceOwnerKey = "customkind-sample"
	apiGVStr        = crdv1alpha1.GroupVersion.String()
	ourKind         = "Customkind"
)

// CustomkindReconciler reconciles a Customkind object
type CustomkindReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=crd.sabbir.com,resources=customkinds,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=crd.sabbir.com,resources=customkinds/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=crd.sabbir.com,resources=customkinds/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Customkind object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.17.3/pkg/reconcile
func (r *CustomkindReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)
	fmt.Println("ReqName ", req.Name, " ReqNameSpace ", req.Namespace)

	println("inside Reconcile ++++++")
	fmt.Println("req.Name:", req.Name)
	// TODO(user): your logic here

	var customkind crdv1alpha1.Customkind

	if err := r.Get(ctx, req.NamespacedName, &customkind); err != nil {
		log.Log.Info("customkind not found")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	var childDeploys appsv1.DeploymentList
	if err := r.List(ctx, &childDeploys, client.InNamespace(req.Namespace), client.MatchingFields{deployOwnerKey: req.Name}); err != nil {
		println("Req.Name:", req.Name)
		println("Req.Namespace:", req.Namespace)
		fmt.Println("Error:", err)
		log.Log.Info("childDeploys not found")

		return ctrl.Result{}, err
	}

	fmt.Println("len of childDeploys......+++... ", len(childDeploys.Items))

	newDeployment := func(customkind *crdv1alpha1.Customkind) *appsv1.Deployment {
		fmt.Println("New Deployment is called")
		fmt.Println("customkind", customkind.Name)
		labels := map[string]string{
			"controller": customkind.Name,
		}
		fmt.Println("labels:", labels)
		return &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      customkind.Name,
				Namespace: customkind.Namespace,
				OwnerReferences: []metav1.OwnerReference{
					*metav1.NewControllerRef(customkind, crdv1alpha1.GroupVersion.WithKind(ourKind)),
				},
			},
			Spec: appsv1.DeploymentSpec{
				Selector: &metav1.LabelSelector{
					MatchLabels: labels,
				},
				Replicas: customkind.Spec.Replicas,
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: labels,
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  "sabbir-container",
								Image: customkind.Spec.Container.Image,
								Ports: []corev1.ContainerPort{
									{
										ContainerPort: customkind.Spec.Container.Port,
									},
								},
							},
						},
					},
				},
			},
		}
	}
	fmt.Println("New Deployment call is finished.....++++----")

	if len(childDeploys.Items) == 0 || !depOwned(&childDeploys) {
		deploy := newDeployment(&customkind)
		if err := r.Create(ctx, deploy); err != nil {
			fmt.Println("Unable to create Deployment")
			return ctrl.Result{}, err
		}
		fmt.Println("Created Deployment")
	}

	var childServices corev1.ServiceList
	if err := r.List(ctx, &childServices, client.InNamespace(req.Namespace), client.MatchingFields{serviceOwnerKey: req.Name}); err != nil {
		fmt.Println("Unable to list child services")
		return ctrl.Result{}, err
	}
	fmt.Println("Len of childServices ", len(childServices.Items))

	println("Customkind.Spec.Container.Image ", customkind.Spec.Container.Image)

	//println("Customkind.Spec.Service.ServiceName ", customkind.Spec.Service.ServiceName)
	//println("Customkind.Name ", customkind.Name)

	getServiceType := func(s string) corev1.ServiceType {
		if s == "NodePort" {
			return corev1.ServiceTypeNodePort
		} else {
			return corev1.ServiceTypeClusterIP
		}
	}

	trimAppName := func(s string) string {
		name := strings.Split(s, "/")
		if len(name) == 1 {
			return name[0]
		}
		return name[1]
	}
	setName := func(s, suf string) string {
		if s == "" {
			s = customkind.Name + suf
		}
		return s
	}

	newService := func(customkind *crdv1alpha1.Customkind) *corev1.Service {
		fmt.Println("New Service is called")
		labels := map[string]string{
			"app":       trimAppName(customkind.Spec.Container.Image),
			"container": customkind.Name,
		}
		return &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      setName(customkind.Spec.Service.ServiceName, "-service"),
				Namespace: customkind.Namespace,
				OwnerReferences: []metav1.OwnerReference{
					*metav1.NewControllerRef(customkind, crdv1alpha1.GroupVersion.WithKind(ourKind)),
				},
			},
			Spec: corev1.ServiceSpec{
				Selector: labels,
				Type:     getServiceType(customkind.Spec.Service.ServiceName),
				Ports: []corev1.ServicePort{
					{
						Port:       customkind.Spec.Container.Port,
						NodePort:   customkind.Spec.Service.ServiceNodePort,
						TargetPort: intstr.FromInt32(int32(customkind.Spec.Container.Port)),
					},
				},
			},
		}
	}

	if len(childServices.Items) == 0 || !srvOwned(&childServices) {
		srvObj := newService(&customkind)
		if err := r.Create(ctx, srvObj); err != nil {
			fmt.Println("Unable to create Service")
		}
		fmt.Println("Created Services ", len(childServices.Items))
	}
	fmt.Println("<<<end of reconcile function>>>>")

	return ctrl.Result{}, nil
}

func srvOwned(srvs *corev1.ServiceList) bool {
	for i := 0; i < len(srvs.Items); i++ {
		ownerRef := srvs.Items[i].GetOwnerReferences()
		for j := 0; j < len(ownerRef); j++ {
			if ownerRef[j].Kind == ourKind && ownerRef[j].APIVersion == apiGVStr {
				return true
			}
		}
	}
	return false
}

func depOwned(deploys *appsv1.DeploymentList) bool {
	for i := 0; i < len(deploys.Items); i++ {
		ownerRef := deploys.Items[i].GetOwnerReferences()
		for j := 0; j < len(ownerRef); j++ {
			if ownerRef[j].Kind == ourKind && ownerRef[j].APIVersion == apiGVStr {
				return true
			}
		}
	}
	return false
}

// SetupWithManager sets up the controller with the Manager.
func (r *CustomkindReconciler) SetupWithManager(mgr ctrl.Manager) error {
	println("inside manager ++++")

	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &appsv1.Deployment{}, deployOwnerKey, func(rawObj client.Object) []string {

		fmt.Println("Inside SetupWithManager deploy +++++++++++++")

		deploy := rawObj.(*appsv1.Deployment)
		owner := metav1.GetControllerOf(deploy)

		//fmt.Println("deploy: ", deploy.Name)
		if owner == nil {
			fmt.Println("owner not found ")
			return nil
		}
		fmt.Println("owner:", owner.Name)
		if owner.APIVersion != apiGVStr || owner.Kind != ourKind {
			return nil
		}
		return []string{owner.Name}

	}); err != nil {
		return err
	}

	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &corev1.Service{}, serviceOwnerKey, func(rawObj client.Object) []string {

		fmt.Println("Inside SetupWithManager deploy +++++++++++++")

		service := rawObj.(*corev1.Service)
		owner := metav1.GetControllerOf(service)

		//fmt.Println("deploy: ", deploy.Name)
		if owner == nil {
			fmt.Println("owner not found ")
			return nil
		}
		fmt.Println("owner:", owner.Name)
		if owner.APIVersion != apiGVStr || owner.Kind != ourKind {
			return nil
		}
		return []string{owner.Name}

	}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&crdv1alpha1.Customkind{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Complete(r)
}
