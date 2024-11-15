package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"

	"golang.org/x/term"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/yaml"

	"github.com/yokecd/yoke/pkg/flight"
)

type Config struct {
	Version     string            `json:"version"`
	Port        int               `json:"port"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	cfg := Config{
		Version: "latest",
		Port:    3000,
	}
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		if err := yaml.NewYAMLToJSONDecoder(os.Stdin).Decode(&cfg); err != nil && err != io.EOF {
			return fmt.Errorf("failed to decode stdin: %w", err)
		}
	}

	crd := apiextensionsv1.CustomResourceDefinition{}

	account := corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-atc-service-account", flight.Release()),
			Namespace: flight.Namespace(),
		},
	}

	binding := rbacv1.ClusterRoleBinding{}

	labels := map[string]string{}
	for k, v := range cfg.Labels {
		labels[k] = v
	}
	labels["yoke.cd/app"] = "atc"

	deploment := appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      flight.Release() + "-atc",
			Namespace: flight.Namespace(),
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: ptr[int32](1),
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      labels,
					Annotations: cfg.Annotations,
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: account.Name,
					Containers: []corev1.Container{
						{
							Name:  "yokecd-atc",
							Image: "davidmdm/atc:" + cfg.Version,
							Env:   []corev1.EnvVar{{Name: "PORT", Value: strconv.Itoa(cfg.Port)}},
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: int32(cfg.Port),
								},
							},
							LivenessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/live",
										Port: intstr.FromInt(cfg.Port),
									},
								},
								TimeoutSeconds: 5,
							},
							ReadinessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/ready",
										Port: intstr.FromInt(cfg.Port),
									},
								},
								TimeoutSeconds: 5,
							},
						},
					},
				},
			},
			Strategy: appsv1.DeploymentStrategy{Type: "Recreate"},
		},
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")

	return encoder.Encode([]any{crd, deploment, account, binding})
}

func ptr[T any](value T) *T { return &value }
