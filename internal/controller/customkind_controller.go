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

/*
import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
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
	"time"
)

var (
	deployOwnerKey  = "bookserver-sample"
	serviceOwnerKey = "bookserver-sample"
	apiGVStr        = crdv1alpha1.GroupVersion.String()
	ourKind         = "Bookserver"
)

// BookserverReconciler reconciles a Bookserver object
type BookserverReconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	Log        logr.Logger
	ctx        context.Context
	bookServer *crdv1alpha1.Bookserver
}

//+kubebuilder:rbac:groups=crd.sabbir.com,resources=bookservers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=crd.sabbir.com,resources=bookservers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=crd.sabbir.com,resources=bookservers/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Bookserver object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.17.3/pkg/reconcile
func (r *BookserverReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	fmt.Println("ReqKind ", req.NamespacedName)
	fmt.Println("ctx ", ctx)
	time.Sleep(time.Second * 5)

	// TODO(user): your logic here

	var bookserver crdv1alpha1.Bookserver

	if err := r.Get(ctx, req.NamespacedName, &bookserver); err != nil {
		log.Log.Info("bookserver not found")
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

	NewDeployment := func(bookserver *crdv1alpha1.Bookserver) *appsv1.Deployment {
		fmt.Println("New Deployment is called")
		fmt.Println("bookserver", bookserver.Name)
		labels := map[string]string{
			"controller": bookserver.Name,
		}
		fmt.Println("labels:", labels)
		return &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      bookserver.Name,
				Namespace: bookserver.Namespace,
				OwnerReferences: []metav1.OwnerReference{
					*metav1.NewControllerRef(bookserver, crdv1alpha1.GroupVersion.WithKind(ourKind)),
				},
			},
			Spec: appsv1.DeploymentSpec{
				Selector: &metav1.LabelSelector{
					MatchLabels: labels,
				},
				Replicas: bookserver.Spec.Replicas,
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: labels,
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  "sabbir-container",
								Image: bookserver.Spec.Container.Image,
								Ports: []corev1.ContainerPort{
									{
										ContainerPort: bookserver.Spec.Container.Port,
									},
								},
							},
						},
					},
				},
			},
		}
	}
	if len(childDeploys.Items) == 0 || !depOwned(&childDeploys, &bookserver, r) {
		deploy := NewDeployment(&bookserver)
		if err := r.Create(ctx, deploy); err != nil {
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

	println("Bookserver.Spec.Container.Image ", bookserver.Spec.Container.Image)

	//println("Bookserver.Spec.Service.ServiceName ", bookserver.Spec.Service.ServiceName)
	//println("Bookserver.Name ", bookserver.Name)

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
			s = bookserver.Name + suf
		}
		return s
	}

	newService := func(bookserver *crdv1alpha1.Bookserver) *corev1.Service {
		fmt.Println("New Service is called")
		labels := map[string]string{
			"app":       trimAppName(bookserver.Spec.Container.Image),
			"container": bookserver.Name,
		}
		return &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      setName(bookserver.Spec.Service.ServiceName, "-service"),
				Namespace: bookserver.Namespace,
				OwnerReferences: []metav1.OwnerReference{
					*metav1.NewControllerRef(bookserver, crdv1alpha1.GroupVersion.WithKind(ourKind)),
				},
			},
			Spec: corev1.ServiceSpec{
				Selector: labels,
				Type:     getServiceType(bookserver.Spec.Service.ServiceName),
				Ports: []corev1.ServicePort{
					{
						Port:       bookserver.Spec.Container.Port,
						NodePort:   bookserver.Spec.Service.ServiceNodePort,
						TargetPort: intstr.FromInt32(int32(bookserver.Spec.Container.Port)),
					},
				},
			},
		}
	}

	if len(childServices.Items) == 0 || !srvOwned(&childServices) {
		srvObj := newService(&bookserver)
		if err := r.Create(ctx, srvObj); err != nil {
			fmt.Println("Unable to create Service")
		}
		fmt.Println("Created Services ", len(childServices.Items))
	}
	fmt.Println("<<<end of reconcile function>>>>")

	return ctrl.Result{}, nil
}

func checkDep(dep *appsv1.Deployment, bookserver *crdv1alpha1.Bookserver, r *BookserverReconciler) {

	fmt.Println(bookserver)
	if dep.Spec.Replicas != bookserver.Spec.Replicas {
		*dep.Spec.Replicas = *bookserver.Spec.Replicas
		fmt.Println("*dep.Spec.replicas", *dep.Spec.Replicas)
		fmt.Println("*bookserver.Spec.Replicas", *bookserver.Spec.Replicas)
		err := r.Client.Update(context.TODO(), dep)
		if err != nil {
			fmt.Println("Unable to update bookserver inside checkDep", err)
		}
		fmt.Println("Replicas changed!!!! ")
	}

}

func srvOwned(srvs *corev1.ServiceList) bool {
	flag := false
	for i := 0; i < len(srvs.Items); i++ {
		ownerRef := srvs.Items[i].GetOwnerReferences()
		for j := 0; j < len(ownerRef); j++ {
			if ownerRef[j].Kind == ourKind && ownerRef[j].APIVersion == apiGVStr {
				flag = true
			}
		}
	}
	return flag
}

func depOwned(deploys *appsv1.DeploymentList, bookserver *crdv1alpha1.Bookserver, r *BookserverReconciler) bool {
	flag := false
	for i := 0; i < len(deploys.Items); i++ {
		ownerRef := deploys.Items[i].GetOwnerReferences()
		fmt.Println(deploys.Items[i].Name, " ", *deploys.Items[i].Spec.Replicas)
		for j := 0; j < len(ownerRef); j++ {
			if ownerRef[j].Kind == ourKind && ownerRef[j].APIVersion == apiGVStr {
				fmt.Println("before update: ", *deploys.Items[i].Spec.Replicas)
				checkDep(&deploys.Items[i], bookserver, r)
				fmt.Println("after update: ", *deploys.Items[i].Spec.Replicas)
				flag = true
			}
		}
	}
	return flag
}

// SetupWithManager sets up the controller with the Manager.
func (r *BookserverReconciler) SetupWithManager(mgr ctrl.Manager) error {
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
		For(&crdv1alpha1.Bookserver{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Complete(r)
}
*/
