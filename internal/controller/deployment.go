package controller

import (
	crdv1alpha1 "github.com/sabbir-hossain70/controller-kubebuilder/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	_ "strings"
)

func (r *BookserverReconciler) CheckDeployment() error {

	deploy := &appsv1.Deployment{}

	if err := r.Client.Get(r.ctx, types.NamespacedName{
		Name:      r.bookServer.DeploymentName(),
		Namespace: r.bookServer.Namespace,
	}, deploy); err != nil {
		if errors.IsNotFound(err) {
			r.Log.Info("Creating a new Deployment", "Namespace", r.bookServer.Namespace, "Name", r.bookServer.Name)
			deploy := r.NewDeployment()
			if err := r.Client.Create(r.ctx, deploy); err != nil {
				return err
			}
			r.Log.Info("Created Deployment", "Namespace", deploy.Namespace, "Name", deploy.Name)
			return nil
		}
		return err
	}
	if r.bookServer.Spec.Replicas != nil && *deploy.Spec.Replicas != *r.bookServer.Spec.Replicas {
		r.Log.Info("replica mismatch...")
		*deploy.Spec.Replicas = *r.bookServer.Spec.Replicas
		if err := r.Client.Update(r.ctx, deploy); err != nil {
			r.Log.Error(err, "Failed to update Deployment", "Namespace", deploy.Namespace, "Name", deploy.Name)
			return err
		}
	}

	return nil
}

func (r *BookserverReconciler) NewDeployment() *appsv1.Deployment {
	r.Log.Info("New Deployment is called")
	labels := map[string]string{
		"controller": r.bookServer.Name,
	}
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.bookServer.DeploymentName(),
			Namespace: r.bookServer.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(r.bookServer, crdv1alpha1.GroupVersion.WithKind(ourKind)),
			},
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Replicas: r.bookServer.Spec.Replicas,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "bookserver-container",
							Image: r.bookServer.Spec.Container.Image,
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: r.bookServer.Spec.Container.Port,
								},
							},
						},
					},
				},
			},
		},
	}
}
