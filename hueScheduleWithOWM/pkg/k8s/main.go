package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Scheduler manages Kubernetes CronJob operations
type Scheduler struct {
	client *kubernetes.Clientset
}

// NewScheduler creates a new Scheduler configured with in-cluster config
func NewScheduler() (*Scheduler, error) {
	cfg, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get in-cluster config: %w", err)
	}
	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %w", err)
	}
	return &Scheduler{client: clientset}, nil
}

// ListCronJobs lists all CronJobs in the specified namespace
func (s *Scheduler) ListCronJobs(ctx context.Context, ns string) (*batchv1.CronJobList, error) {
	return s.client.BatchV1().
		CronJobs(ns).
		List(ctx, metav1.ListOptions{})
}

// ModifyCronJobExecution modifies the execution time (schedule) of a CronJob
func (s *Scheduler) ModifyCronJobExecution(ctx context.Context, ns, name, schedule string) error {
	// Get the current CronJob
	cronJob, err := s.client.BatchV1().CronJobs(ns).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get cronjob: %w", err)
	}

	// Update the schedule
	cronJob.Spec.Schedule = schedule

	// Update the CronJob
	_, err = s.client.BatchV1().CronJobs(ns).Update(ctx, cronJob, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update cronjob: %w", err)
	}

	return nil
}

func main() {
	ns := getenv("NAMESPACE", "default")

	scheduler, err := NewScheduler()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// List all CronJobs in the namespace
	cronJobs, err := scheduler.ListCronJobs(ctx, ns)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: failed to list cronjobs: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Found %d CronJob(s) in namespace %s:\n", len(cronJobs.Items), ns)

	// Iterate through cronjobs and modify "cronjob-lister" if found
	for _, cj := range cronJobs.Items {
		schedule := cj.Spec.Schedule
		if schedule == "" {
			schedule = "(no schedule)"
		}
		fmt.Printf("  - %s (schedule: %s)\n", cj.Name, schedule)

		// If this is "cronjob-lister", modify its execution time
		if cj.Name == "cronjob-lister" {
			// Generate random number between 1 and 5
			rand.Seed(time.Now().UnixNano())
			x := rand.Intn(5) + 1 // rand.Intn(5) gives 0-4, so +1 gives 1-5

			// Create cron schedule for every x minutes: "*/x * * * *"
			newSchedule := fmt.Sprintf("*/%d * * * *", x)

			fmt.Printf("Modifying cronjob-lister schedule to every %d minute(s): %s\n", x, newSchedule)
			if err := scheduler.ModifyCronJobExecution(ctx, ns, cj.Name, newSchedule); err != nil {
				fmt.Fprintf(os.Stderr, "error: failed to modify cronjob-lister: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Successfully updated cronjob-lister schedule\n")
		}
	}
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
