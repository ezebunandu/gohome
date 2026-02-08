package scheduler

import (
	"context"
	"fmt"

	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Scheduler manages Kubernetes CronJob operations
type Scheduler struct {
	client    *kubernetes.Clientset
	namespace string
}

// NewScheduler creates a new Scheduler configured with in-cluster config and the target namespace for CronJob updates.
func NewScheduler(namespace string) (*Scheduler, error) {
	cfg, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get in-cluster config: %w", err)
	}
	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %w", err)
	}
	return &Scheduler{client: clientset, namespace: namespace}, nil
}

// ListCronJobs lists all CronJobs in the specified namespace
func (s *Scheduler) ListCronJobs(ctx context.Context, ns string) (*batchv1.CronJobList, error) {
	return s.client.BatchV1().
		CronJobs(ns).
		List(ctx, metav1.ListOptions{})
}

// ModifyCronJobExecution modifies the execution time (schedule) of a CronJob in the Scheduler's namespace.
func (s *Scheduler) ModifyCronJobExecution(ctx context.Context, name, schedule string) error {
	// Get the current CronJob
	cronJob, err := s.client.BatchV1().CronJobs(s.namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get cronjob: %w", err)
	}

	// Update the schedule
	cronJob.Spec.Schedule = schedule

	// Update the CronJob
	_, err = s.client.BatchV1().CronJobs(s.namespace).Update(ctx, cronJob, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update cronjob: %w", err)
	}

	return nil
}
