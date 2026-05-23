package kube

import (
	"context"
	"errors"
	"io"
	"log"
	"testing"

	k8sfake "k8s.io/client-go/kubernetes/fake"
)

var discardLogger = log.New(io.Discard, "", 0)

type fakeKubeClient struct {
	err          error
	called       bool
	gotNamespace string
}

func (f *fakeKubeClient) EnsureNamespace(_ context.Context, namespace string) error {
	f.called = true
	f.gotNamespace = namespace
	return f.err
}

func TestEnsureNamespace(t *testing.T) {
	tests := []struct {
		name      string
		namespace string
		err       error
		wantErr   bool
	}{
		{
			name:      "creates namespace successfully",
			namespace: "my-namespace",
		},
		{
			name:      "propagates error",
			namespace: "my-namespace",
			err:       errors.New("cluster unreachable"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeKubeClient{err: tt.err}

			err := fake.EnsureNamespace(context.Background(), tt.namespace)

			if (err != nil) != tt.wantErr {
				t.Errorf("EnsureNamespace() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !fake.called {
				t.Error("expected EnsureNamespace to be called")
			}
			if fake.gotNamespace != tt.namespace {
				t.Errorf("got namespace %q, want %q", fake.gotNamespace, tt.namespace)
			}
		})
	}
}

func TestKubeClientEnsureNamespace(t *testing.T) {
	t.Run("creates namespace when it does not exist", func(t *testing.T) {
		clientset := k8sfake.NewSimpleClientset()
		client := &kubeClient{clientset: clientset, logger: discardLogger}

		err := client.EnsureNamespace(context.Background(), "my-namespace")

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("does nothing when namespace already exists", func(t *testing.T) {
		clientset := k8sfake.NewSimpleClientset()
		client := &kubeClient{clientset: clientset, logger: discardLogger}

		_ = client.EnsureNamespace(context.Background(), "my-namespace")
		err := client.EnsureNamespace(context.Background(), "my-namespace")

		if err != nil {
			t.Fatalf("expected no error on second call, got %v", err)
		}
	})
}
