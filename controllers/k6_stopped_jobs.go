package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/go-logr/logr"
	"github.com/grafana/k6-operator/api/v1alpha1"
	k6api "go.k6.io/k6/api/v1"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func isJobRunning(log logr.Logger, service *v1.Service) bool {
	resp, err := http.Get(fmt.Sprintf("http://%v.%v.svc.cluster.local:6565/v1/status", service.ObjectMeta.Name, service.ObjectMeta.Namespace))
	if err != nil {
		return false
	}

	// Response has been received so assume the job is running.

	if resp.StatusCode >= 400 {
		log.Error(err, fmt.Sprintf("status from from runner job %v is %d", service.ObjectMeta.Name, resp.StatusCode))
		return true
	}

	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error(err, fmt.Sprintf("Error on reading status of the runner job %v", service.ObjectMeta.Name))
		return true
	}

	var status k6api.StatusJSONAPI
	if err := json.Unmarshal(data, &status); err != nil {
		log.Error(err, fmt.Sprintf("Error on parsing status of the runner job %v", service.ObjectMeta.Name))
		return true
	}

	return status.Status().Stopped
}

// StoppedJobs checks if the runners pods have stopped execution.
func StoppedJobs(ctx context.Context, log logr.Logger, k6 *v1alpha1.K6, r *K6Reconciler) (allStopped bool) {
	if len(k6.Status.TestRunID) > 0 {
		log = log.WithValues("testRunId", k6.Status.TestRunID)
	}

	log.Info("Waiting for pods to stop the test run")

	selector := labels.SelectorFromSet(map[string]string{
		"app":    "k6",
		"k6_cr":  k6.Name,
		"runner": "true",
	})

	opts := &client.ListOptions{LabelSelector: selector, Namespace: k6.Namespace}

	var hostnames []string
	sl := &v1.ServiceList{}

	if err := r.List(ctx, sl, opts); err != nil {
		log.Error(err, "Could not list services")
		return
	}

	var count int32
	for _, service := range sl.Items {
		hostnames = append(hostnames, service.Spec.ClusterIP)

		if isJobRunning(log, &service) {
			count++
		}
	}

	log.Info(fmt.Sprintf("%d/%d runners stopped execution", k6.Spec.Parallelism-count, k6.Spec.Parallelism))

	if count > 0 {
		return
	}

	allStopped = true
	return
}

func KillJobs(ctx context.Context, log logr.Logger, k6 *v1alpha1.K6, r *K6Reconciler) (err error) {
	if len(k6.Status.TestRunID) > 0 {
		log = log.WithValues("testRunId", k6.Status.TestRunID)
	}

	log.Info("Checking if all runner pods are finished")

	selector := labels.SelectorFromSet(map[string]string{
		"app":    "k6",
		"k6_cr":  k6.Name,
		"runner": "true",
	})

	opts := &client.ListOptions{LabelSelector: selector, Namespace: k6.Namespace}
	jl := &batchv1.JobList{}

	if err = r.List(ctx, jl, opts); err != nil {
		log.Error(err, "Could not list jobs")
		return
	}

	propagationPolicy := client.PropagationPolicy(metav1.DeletionPropagation(metav1.DeletePropagationBackground))
	for _, job := range jl.Items {
		if err = r.Delete(ctx, &job, propagationPolicy); err != nil {
			log.Error(err, fmt.Sprintf("Failed to delete runner job %s", job.Name))
			// do we need to retry here?
		}
	}

	return nil
}
