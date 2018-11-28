package deployment

import (
	"context"
	"log"

	integreatlyv1alpha1 "github.com/integr8ly/deployment-operator/pkg/apis/integreatly/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"github.com/integr8ly/operator-sdk-openshift-utils/pkg/api/template"
	"k8s.io/client-go/rest"
	"github.com/gobuffalo/packr"
	"github.com/integr8ly/operator-sdk-openshift-utils/pkg/api/kubernetes"
	"k8s.io/apimachinery/pkg/util/yaml"
	"github.com/openshift/api/apps/v1"
	"k8s.io/apimachinery/pkg/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new TDeployment Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileDeployment{
		client: mgr.GetClient(),
		scheme: mgr.GetScheme(),
		config: mgr.GetConfig(),
		box: packr.NewBox("../../../res"),
	}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("deployment-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource TDeployment
	err = c.Watch(&source.Kind{Type: &integreatlyv1alpha1.TDeployment{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner TDeployment
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &integreatlyv1alpha1.TDeployment{},
	})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileDeployment{}

// ReconcileDeployment reconciles a TDeployment object
type ReconcileDeployment struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
	config *rest.Config
	box packr.Box
	tmpl *template.Tmpl
}

// Reconcile reads that state of the cluster for a TDeployment object and makes changes based on the state read
// and what is in the TDeployment.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileDeployment) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	log.Printf("Reconciling TDeployment %s/%s\n", request.Namespace, request.Name)

	var err error

	// Fetch the TDeployment instance
	instance := &integreatlyv1alpha1.TDeployment{}
	err = r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		log.Printf("%v", err)
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	if r.tmpl == nil {
		err = r.Bootstrap(instance)
		if err != nil {
			return reconcile.Result{}, err
		}
	}

	if instance.Status.Phase == integreatlyv1alpha1.NoPhase {
		err = r.DeployTemplate(instance)
		if err != nil {
			log.Printf("Template error : %v", err)
			instance.Status.Phase = integreatlyv1alpha1.ErrorPhase

			updateErr := r.client.Update(context.TODO(), instance)
			if updateErr != nil {
				log.Printf("Update error : %v", updateErr)
			}

			return reconcile.Result{}, err
		}

		instance.Status.Phase = integreatlyv1alpha1.ProvisionPhase
		err = r.client.Update(context.TODO(), instance)
		if err != nil {
			log.Printf("Update error : %v", err)
			return reconcile.Result{}, err
		}
	}

	if instance.Status.Phase == integreatlyv1alpha1.ProvisionPhase {
		isProvisioned, err := r.IsProvisioningFinished(instance)
		if err != nil {
			return reconcile.Result{}, err
		}

		if isProvisioned {
			instance.Status.Phase = integreatlyv1alpha1.ReadyPhase
			err = r.client.Update(context.TODO(), instance)
			if err != nil {
				return reconcile.Result{}, err
			}
		}
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileDeployment) Bootstrap(cr *integreatlyv1alpha1.TDeployment) error {
	templateData, err := r.box.Find(cr.Spec.Template.Path)
	if err != nil {
		return err
	}

	jsonData, err := yaml.ToJSON(templateData)
	if err != nil {
		return err
	}

	tmpl, err := template.New(r.config, jsonData)
	if err != nil {
		return err
	}

	r.tmpl = tmpl

	return nil
}

func (r *ReconcileDeployment) DeployTemplate(cr *integreatlyv1alpha1.TDeployment) error {
	var err error

	err = r.tmpl.Process(cr.Spec.Template.Parameters, cr.Namespace)
	if err != nil {
		return err
	}

	objects := make([]runtime.Object, 0)
	r.tmpl.CopyObjects(template.NoFilterFn, &objects)

	for _, obj := range objects {
		uo, _ := kubernetes.UnstructuredFromRuntimeObject(obj)
		uo.SetNamespace(cr.Namespace)

		err = r.client.Create(context.TODO(), uo.DeepCopyObject())
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *ReconcileDeployment) IsProvisioningFinished(cr *integreatlyv1alpha1.TDeployment) (bool, error) {
	deploymentName := "tutorial-web-app"
	dc := &v1.DeploymentConfig{
		TypeMeta: metav1.TypeMeta{
			Kind: "DeploymentConfig",
			APIVersion: "apps.openshift.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: deploymentName,
			Namespace: cr.Namespace,
		},
	}
	key := types.NamespacedName{
		Namespace: cr.Namespace,
		Name: deploymentName,
	}

	err := r.client.Get(context.TODO(), key, dc)
	if err != nil {
		return false, err
	}

	if dc.Status.ReadyReplicas == dc.Status.Replicas {
		return true, nil
	}

	return false, nil
}
