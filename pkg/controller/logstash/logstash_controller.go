package logstash

import (
	"context"

	loggingv1alpha1 "github.com/njbarber/logstash-operator/pkg/apis/logging/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"text/template"
	"bytes"
)

var log = logf.Log.WithName("controller_logstash")

// Add creates a new Logstash Controller and adds it to the Manager. The Manager will set fields on the Controller
// and start it when the Manager is started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileLogstash{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("logstash-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Logstash
	err = c.Watch(&source.Kind{Type: &loggingv1alpha1.Logstash{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource ConfigMap and requeue the owner Logstash
	err = c.Watch(&source.Kind{Type: &corev1.ConfigMap{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &loggingv1alpha1.Logstash{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileLogstash implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileLogstash{}

// ReconcileLogstash reconciles a Logstash object
type ReconcileLogstash struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Template to be used when generating Logstash configuration entries
  var t, _  = template.New("").Parse(`filter {
	if [app_name] == "{{ .Name }}" {
	  grok {
		pattern_definitions => {
		{{ range $k, $v := .Patterns }}  "{{ $k }}" => "{{ $v }}"
		{{ end -}}
		}
		match => {
		  "message" => [
		  {{ range $i, $v := .Matchers }}{{ if $i }},
		  {{ end }}  "{{ $v }}"{{ end }}
		  ]
		}
	  }
	}
  }
  `)

// Reconcile reads that state of the cluster for a Logstash object and makes changes based on the state read
// and what is in the Logstash.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileLogstash) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Logstash")

	// Fetch the Logstash instance
	instance := &loggingv1alpha1.Logstash{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// Define a new ConfigMap object
  	configMap := newConfigMapForCR(instance)
	
	// Set Logstash instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, configMap, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this ConfigMap already exists
	found := &corev1.ConfigMap{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: configMap.Name, Namespace: configMap.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new ConfigMap", "ConfigMap.Namespace", configMap.Namespace, "ConfigMap.Name", configMap.Name)
		err = r.client.Create(context.TODO(), configMap)
		if err != nil {
			return reconcile.Result{}, err
		}
	} else if err != nil {
		return reconcile.Result{}, err
	}

	// Update the ConfigMap Data
	m := found.Data
	if m == nil {
		m = make(map[string]string)
	}
	var b bytes.Buffer
	for _, app := range instance.Spec.Applications {
		t.Execute(&b, app)
		m[instance.Namespace + "_" + app.Name] = b.String()
		b.Reset()
	}
	found.Data = m
	r.client.Update(context.TODO(), found)

	// ConfigMap already exists - don't requeue
	reqLogger.Info("Skip reconcile: ConfigMap already exists", "ConfigMap.Namespace", configMap.Namespace, "ConfigMap.Name", configMap.Name)
	return reconcile.Result{}, nil
}

func newConfigMapForCR(cr *loggingv1alpha1.Logstash) *corev1.ConfigMap {
	m := make(map[string]string)
	var b bytes.Buffer
	
	for _, app := range cr.Spec.Applications {
		t.Execute(&b, app)
		m[cr.Namespace + "_" + app.Name] = b.String()
		b.Reset()
	}

	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: "gsp-logstash-pipeline",
			Namespace: "grayskull-logs",
			},	
		Data: m,
	}
}