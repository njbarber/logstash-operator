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
	// "strconv"
	"fmt"
	// "strings"
	// "html"
)

var log = logf.Log.WithName("controller_logstash")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Logstash Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
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

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
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

// Reconcile reads that state of the cluster for a Logstash object and makes changes based on the state read
// and what is in the Logstash.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
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

  	// Define a new Pod object
	// pod := newPodForCR(instance)

	// Define a new ConfigMap object
  	configMap := newConfigMapForCR(instance)

	// Set Logstash instance as the owner and controller
	// if err := controllerutil.SetControllerReference(instance, pod, r.scheme); err != nil {
	// 	return reconcile.Result{}, err
	// }
	
	// Set Logstash instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, configMap, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this Pod already exists
	// found := &corev1.Pod{}
	// err = r.client.Get(context.TODO(), types.NamespacedName{Name: pod.Name, Namespace: pod.Namespace}, found)
	// if err != nil && errors.IsNotFound(err) {
	// 	reqLogger.Info("Creating a new Pod", "Pod.Namespace", pod.Namespace, "Pod.Name", pod.Name)
	// 	err = r.client.Create(context.TODO(), pod)
	// 	if err != nil {
	// 		return reconcile.Result{}, err
	// 	}

	// 	// Pod created successfully - don't requeue
	// 	return reconcile.Result{}, nil
	// } else if err != nil {
	// 	return reconcile.Result{}, err
	// }

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

	// Update the ConfigMap Data (Just overwrite, no status currently - will eventually be based off of)
	m := found.Data
	if m == nil {
		m = make(map[string]string)
	}
	var b bytes.Buffer
	for _, app := range instance.Spec.Applications {
		t.Execute(&b, app)
		m[app.Name] = b.String()
		b.Reset()
	}
	found.Data = m
	r.client.Update(context.TODO(), found)

	// // Pod already exists - don't requeue
	// reqLogger.Info("Skip reconcile: Pod already exists", "Pod.Namespace", found.Namespace, "Pod.Name", found.Name)
	// return reconcile.Result{}, nil

	// ConfigMap already exists - don't requeue
	reqLogger.Info("Skip reconcile: ConfigMap already exists", "ConfigMap.Namespace", configMap.Namespace, "ConfigMap.Name", configMap.Name)
	return reconcile.Result{}, nil
}

// newPodForCR returns a busybox pod with the same name/namespace as the cr
// func newPodForCR(cr *loggingv1alpha1.Logstash) *corev1.Pod {
// 	labels := map[string]string{
// 		"app": cr.Name,
// 	}
// 	return &corev1.Pod{
// 		ObjectMeta: metav1.ObjectMeta{
// 			Name:      cr.Name + "-pod",
// 			Namespace: cr.Namespace,
// 			Labels:    labels,
// 		},
// 		Spec: corev1.PodSpec{
// 			Containers: []corev1.Container{
// 				{
// 					Name:    "busybox",
// 					Image:   "busybox",
// 					Command: []string{"sleep", "3600"},
// 				},
// 			},
// 		},
// 	}
// }

const tmpl = `
filter {
	if [app_name] == "{{ .Name }}" {
		grok {
			pattern_definitions => {
				{{ range $k, $v := .Patterns -}}
				"{{ $k }}" => "{{ $v }}"
				{{- end }}
			}
			match => {
				{{ range $k, $v := .Matchers -}}
				"message" => {
					{{ $v }}
				}
				{{- end }}
			}
		}
	}
}
`

const temp = `filter {
  if [app_name] == "{{ .Name }}" {
    grok {
      pattern_definitions => {
      {{ range $k, $v := .Patterns }}  "{{ $k }}" => "{{ $v }}"
      {{ end -}}
      }
      match => {
      {{ range $k, $v := .Matchers }}  "message" => "{{ $v }}"
      {{ end -}}
      }
    }
  }
}
`

var t, _  = template.New("").Parse(temp)

func newConfigMapForCR(cr *loggingv1alpha1.Logstash) *corev1.ConfigMap {
	// patterns  := ""	
	// for k, v := range cr.Spec.Pattern {
	// 	patterns += k + " " + v + "\n"
	// }

	m := make(map[string]string)
	var b bytes.Buffer
	// var b strings.Builder

	// t, _ := template.New("").Parse(temp)
	
	for _, app := range cr.Spec.Applications {
		t.Execute(&b, app)
		// s := b.String()
		s := fmt.Sprint(&b)

		// s, _ = strconv.Unquote(s)
		// s = strings.Replace(s, `\n`, "\n", -1)
		// html.UnescapeString(s)
		// s = fmt.Sprint(s)

		m[app.Name] = s	
		b.Reset()
	}
	
	// t.Execute()

	// for _,  a := range cr.Spec.Applications {
	// 	result := "filter {\n"
	// 	result += "  grok {\n"
	// 	result += "    pattern_definition => {\n"
	// 	for k, v := range a.Patterns {
	// 		result += "      \"" + k + "\"" + " => " + "\"" + v + "\"\n"
	// 	}
	// 	result += "    }\n"
	// 	for _, v := range a.Matchers {
	// 		// result += "match => { \"message\" => " + v + "}\n"
	// 		result += "    " + v
	// 	}
	// 	result += "\n  }\n"
	// 	result += "}"
	// 	m["filter_" + a.Name] = result
 	// }

	// result := ""
	// for _, a := range cr.Spec.Applications {
	// 	// result += "Name:" + a.Name + "\n"
	// 	result += "Patterns\n"
	// 	for k, v := range a.Patterns {
	// 		result += k + ": " + v
	// 	}
	// 	result += "\n"
	// 	for k, v := range a.Matchers {
	// 		result += k + ": " + v
	// 	}
	// 	result += "\n"
	// }
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			// Name: cr.Name,
			// Namespace: cr.Namespace,
			Name: "gsp-logstash-pipeline",
			Namespace: "grayskull-logs",
			},	
		// Data: map[string]string{"Pattern": patterns},
		// Data: map[string]string{"Result": result},
		Data: m,
	}
}