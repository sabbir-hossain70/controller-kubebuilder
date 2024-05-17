package controller

import (
	"fmt"
	crdv1alpha1 "github.com/sabbir-hossain70/controller-kubebuilder/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func (r *BookserverReconciler) CheckSrv() error {
	srv := &corev1.Service{}
	if err := r.Client.Get(r.ctx, types.NamespacedName{
		Name:      r.bookServer.ServiceName(),
		Namespace: r.bookServer.Namespace,
	}, srv); err != nil {
		if errors.IsNotFound(err) {
			r.Log.Info("Bookserver service not found")
			if err := r.Client.Create(r.ctx, r.newService()); err != nil {
				r.Log.Error(err, "failed to create service")
				return err
			}
			r.Log.Info("created service")
			return nil
		}
		return err
	}
	return nil
}

func (r *BookserverReconciler) newService() *corev1.Service {
	fmt.Println("New Service is called")
	labels := map[string]string{
		"controller": r.bookServer.Name,
	}
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.bookServer.ServiceName(),
			Namespace: r.bookServer.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(r.bookServer, crdv1alpha1.GroupVersion.WithKind(ourKind)),
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Type:     getServiceType(r.bookServer.Spec.Service.ServiceName),
			Ports: []corev1.ServicePort{
				{
					Port:       r.bookServer.Spec.Container.Port,
					NodePort:   r.bookServer.Spec.Service.ServiceNodePort,
					TargetPort: intstr.FromInt32(r.bookServer.Spec.Container.Port),
				},
			},
		},
	}
}
