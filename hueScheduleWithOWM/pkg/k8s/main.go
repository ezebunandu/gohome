package main

import (
	"context"
	"fmt"
	"os"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func main() {
	ns := getenv("NAMESPACE", "default")

	cfg, err := rest.InClusterConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: failed to get in-cluster config: %v\n", err)
		os.Exit(1)
	}
	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: failed to create kubernetes client: %v\n", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// List all CronJobs in the namespace
	cronJobs, err := listCronJobs(ctx, clientset, ns)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: failed to list cronjobs: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Found %d CronJob(s) in namespace %s:\n", len(cronJobs.Items), ns)
	for _, cj := range cronJobs.Items {
		schedule := cj.Spec.Schedule
		if schedule == "" {
			schedule = "(no schedule)"
		}
		fmt.Printf("  - %s (schedule: %s)\n", cj.Name, schedule)
	}
}

func listCronJobs(ctx context.Context, clientset *kubernetes.Clientset, ns string) (*batchv1.CronJobList, error) {
	return clientset.BatchV1().
		CronJobs(ns).
		List(ctx, metav1.ListOptions{})
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
