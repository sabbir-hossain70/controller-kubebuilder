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
	"github.com/go-logr/logr"
	crdv1alpha1 "github.com/sabbir-hossain70/controller-kubebuilder/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	_ "strings"
)

const (
	ourKind = "Bookserver"
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

func (r *BookserverReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.Log = ctrl.Log.WithValues("Bookserver", req.NamespacedName)
	r.ctx = ctx

	// TODO(user): your logic here

	var bookserver crdv1alpha1.Bookserver

	if err := r.Get(ctx, req.NamespacedName, &bookserver); err != nil {
		r.Log.Info("bookserver not found")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	r.bookServer = &bookserver

	if err := r.CheckDeployment(); err != nil {
		return ctrl.Result{}, err
	}
	if err := r.CheckService(); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}
