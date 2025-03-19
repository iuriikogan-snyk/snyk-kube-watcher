package watcher

import (
	"context"
	"testing"
	"time"

	"github.com/iuriikogan-snyk/snyk-kube-watcher/internal/config"
	"github.com/iuriikogan-snyk/snyk-kube-watcher/internal/tasks"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestEnqueue(t *testing.T) {
	taskCh := make(chan tasks.ImageTask, 1)
	defer close(taskCh)
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "test-namespace",
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:  "test-container",
					Image: "test-image",
				},
			},
		},
	}
	enqueue(pod, "test-org", taskCh)
	select {
	case task := <-taskCh:
		if task.Image != "test-image" {
			t.Errorf("Expected image to be test-image, got %s", task.Image)
		}
		if task.OrgID != "test-org" {
			t.Errorf("Expected orgID to be test-org, got %s", task.OrgID)
		}
	case <-time.After(100 * time.Millisecond):
		t.Errorf("Expected task to be enqueued, but no task was received")
	}
}

func TestRunInformerAdd(t *testing.T) {
	taskCh := make(chan tasks.ImageTask, 1)
	defer close(taskCh)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	client := fake.NewSimpleClientset()
	cl := config.ClusterConfig{Kubeconfig: "test-kubeconfig", Context: "test-context"}
	go func() {
		err := runInformer(ctx, cl, "test-org", taskCh)
		if err != nil {
			t.Errorf("runInformer returned an error: %v", err)
		}
	}()
	time.Sleep(100 * time.Millisecond)
	_, err := client.CoreV1().Pods("default").Create(ctx, &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "default",
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:  "test-container",
					Image: "test-image",
				},
			},
		},
	})
	if err != nil {
		t.Errorf("Failed to create pod: %v", err)
	}
	select {
	case task := <-taskCh:
		if task.Image != "test-image" {
			t.Errorf("Expected image to be test-image, got %s", task.Image)
		}
		if task.OrgID != "test-org" {
			t.Errorf("Expected orgID to be test-org, got %s", task.OrgID)
		}
	case <-time.After(5 * time.Second):
		t.Errorf("Expected task to be enqueued, but no task was received")
	}
}

func TestRunInformerUpdateAndStop(t *testing.T) {
	taskCh := make(chan tasks.ImageTask, 2)
	defer close(taskCh)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	client := fake.NewSimpleClientset()
	cl := config.ClusterConfig{Kubeconfig: "test-kubeconfig", Context: "test-context"}
	go func() {
		err := runInformer(ctx, cl, "test-org", taskCh)
		if err != nil {
			t.Errorf("runInformer returned an error: %v", err)
		}
	}()
	time.Sleep(100 * time.Millisecond)
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "default",
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:  "test-container",
					Image: "test-image",
				},
			},
		},
	}
	_, err := client.CoreV1().Pods("default").Create(ctx, pod)
	if err != nil {
		t.Errorf("Failed to create pod: %v", err)
	}
	select {
	case task := <-taskCh:
		if task.Image != "test-image" {
			t.Errorf("Expected image to be test-image, got %s", task.Image)
		}
		if task.OrgID != "test-org" {
			t.Errorf("Expected orgID to be test-org, got %s", task.OrgID)
		}
	case <-time.After(5 * time.Second):
		t.Errorf("Expected task to be enqueued, but no task was received")
	}
	pod.Spec.Containers[0].Image = "test-image-updated"
	_, err = client.CoreV1().Pods("default").Update(ctx, pod)
	if err != nil {
		t.Errorf("Failed to update pod: %v", err)
	}
	select {
	case task := <-taskCh:
		if task.Image != "test-image-updated" {
			t.Errorf("Expected image to be test-image-updated, got %s", task.Image)
		}
		if task.OrgID != "test-org" {
			t.Errorf("Expected orgID to be test-org, got %s", task.OrgID)
		}
	case <-time.After(5 * time.Second):
		t.Errorf("Expected task to be enqueued, but no task was received")
	}
	cancel()
}
