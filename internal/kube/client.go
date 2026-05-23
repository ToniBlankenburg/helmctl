package kube

import (
	"context"
	"fmt"
	"log"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type KubeClient interface {
	EnsureNamespace(ctx context.Context, namespace string) error
}

type kubeClient struct {
	clientset kubernetes.Interface
	logger    *log.Logger
}

func NewKubeClient(logger *log.Logger) (KubeClient, error) {
	rules := clientcmd.NewDefaultClientConfigLoadingRules()
	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules, nil).ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load kubeconfig: %w", err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %w", err)
	}
	return &kubeClient{clientset: clientset, logger: logger}, nil
}

func (k *kubeClient) EnsureNamespace(ctx context.Context, namespace string) error {
	_, err := k.clientset.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
	if err == nil {
		return nil
	}
	if errors.IsNotFound(err) {
		k.logger.Printf("namespace %q not found, creating it", namespace)
		_, err = k.clientset.CoreV1().Namespaces().Create(ctx, &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{Name: namespace},
		}, metav1.CreateOptions{})
		return err
	}
	return err
}
