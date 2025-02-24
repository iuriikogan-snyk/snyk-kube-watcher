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

func TestRunInformer(t *testing.T) {
	taskCh := make(chan tasks.ImageTask, 1)
	defer close(taskCh)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	client := fake.NewSimpleClientset()
	_, err := client.CoreV1().Pods("default").Create(ctx, &v1.Pod{