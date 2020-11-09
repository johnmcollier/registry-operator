/*


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

	"github.com/go-logr/logr"
	routev1 "github.com/openshift/api/route/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	registryv1alpha1 "github.com/devfile/registry-operator/api/v1alpha1"
	"github.com/devfile/registry-operator/pkg/cluster"
	"github.com/devfile/registry-operator/pkg/config"
	"github.com/devfile/registry-operator/pkg/registry"
)

// DevfileRegistryReconciler reconciles a DevfileRegistry object
type DevfileRegistryReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=registry.devfile.io,resources=devfileregistries,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=registry.devfile.io,resources=devfileregistries/status;devfileregistries/finalizers,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=services;persistentvolumeclaims,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;
// +kubebuilder:rbac:groups=extensions,resources=ingresses,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=route.openshift.io,resources=routes;routes/custom-host,verbs=get;list;watch;create;update;patch;delete

func (r *DevfileRegistryReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("devfileregistry", req.NamespacedName)

	// your logic here
	// Fetch the DevfileRegistry instance
	devfileRegistry := &registryv1alpha1.DevfileRegistry{}
	err := r.Get(ctx, req.NamespacedName, devfileRegistry)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Info("DevfileRegistry resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get DevfileRegistry")
		return ctrl.Result{}, err
	}

	// Check if the service already exists, if not create a new one
	result, err := r.ensureService(ctx, devfileRegistry)
	if result != nil {
		return *result, err
	}

	if registry.IsDevfileRegistryStorageEnabled(devfileRegistry) {
		// Check if the persistentvolumeclaim already exists, if not create a new one
		result, err = r.ensurePVC(ctx, devfileRegistry)
		if result != nil {
			return *result, err
		}
	}

	// Check if the deployment already exists, if not create a new one
	result, err = r.ensureDeployment(ctx, devfileRegistry)
	if result != nil {
		return *result, err
	}

	// Check if we're running on OpenShift
	// ToDo: Move to operator init in main.go so that we don't need to check on every reconcile
	/*isOS, err := cluster.IsOpenShift()
	if err != nil {
		return ctrl.Result{}, err
	}*/

	hostname := devfileRegistry.Spec.K8s.IngressDomain
	if config.ControllerCfg.IsOpenShift() {
		// Check if the route exposing the devfile index exists
		result, err = r.ensureDevfilesRoute(ctx, devfileRegistry, hostname)
		if result != nil {
			return *result, err
		}

		// If the route hostname was autodiscovered by OpenShift, need to retrieve the generated hostname.
		// This is so that we can re-use the hostname in the second route and allows us to expose both routes under the same hostname
		if hostname == "" {
			// Get the hostname of the devfiles route
			devfilesRoute := &routev1.Route{}
			err = r.Get(ctx, types.NamespacedName{Name: devfileRegistry.Name + "-devfiles", Namespace: devfileRegistry.Namespace}, devfilesRoute)
			if err != nil {
				// Log an error, but requeue, as the route may not have been registered yet in the Kube API
				// See https://github.com/operator-framework/operator-sdk/issues/4013#issuecomment-707267616 for an explanation on why we requeue rather than error out here
				log.Error(err, "Failed to get Route")
				return ctrl.Result{Requeue: true}, nil
			}
			hostname = devfilesRoute.Spec.Host
		}

		// Check if the route exposing the devfile index exists
		result, err = r.ensureOCIRoute(ctx, devfileRegistry, hostname)
		if result != nil {
			return *result, err
		}
	} else {
		// Check if the ingress already exists, if not create a new one
		hostname = registry.GetDevfileRegistryIngress(devfileRegistry)
		result, err = r.ensureIngress(ctx, devfileRegistry, hostname)
		if result != nil {
			return *result, err
		}
	}

	if devfileRegistry.Status.URL != hostname {
		devfileRegistry.Status.URL = hostname
		err := r.Status().Update(ctx, devfileRegistry)
		if err != nil {
			log.Error(err, "Failed to update DevfileRegistry status")
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *DevfileRegistryReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// Check if we're running on OpenShift
	isOS, err := cluster.IsOpenShift()
	if err != nil {
		return err
	}
	config.ControllerCfg.SetIsOpenShift(isOS)

	builder := ctrl.NewControllerManagedBy(mgr).
		For(&registryv1alpha1.DevfileRegistry{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.PersistentVolumeClaim{}).
		Owns(&v1beta1.Ingress{})

	// If on OpenShift, mark routes as owned by the controller
	if config.ControllerCfg.IsOpenShift() {
		builder.Owns(&routev1.Route{})
	}

	return builder.Complete(r)

}
