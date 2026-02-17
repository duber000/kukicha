package kube

import (
	"context"
	"fmt"
	"time"

	ctxpkg "github.com/duber000/kukicha/stdlib/ctx"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// WatchPods streams pod events for the namespace and returns collected events until timeout.
// timeoutSeconds <= 0 defaults to 30 seconds.
func WatchPods(c Cluster, timeoutSeconds int64) ([]PodEvent, error) {
	if timeoutSeconds <= 0 {
		timeoutSeconds = 30
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSeconds)*time.Second)
	defer cancel()
	return watchPodsWithContext(ctx, c)
}

// WatchPodsCtx streams pod events until the provided context is canceled.
func WatchPodsCtx(c Cluster, h ctxpkg.Handle) ([]PodEvent, error) {
	ctx := ctxpkg.Value(h)
	return watchPodsWithContext(ctx, c)
}

// watchPodsWithContext streams pod events until the context is canceled.
// This function must remain in Go because Kukicha does not support the select statement.
func watchPodsWithContext(ctx context.Context, c Cluster) ([]PodEvent, error) {
	watcher, err := clientset(c).CoreV1().Pods(c.namespace).Watch(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("kube watch pods: %w", err)
	}
	defer watcher.Stop()

	events := []PodEvent{}
	for {
		select {
		case <-ctx.Done():
			return events, nil
		case event, ok := <-watcher.ResultChan():
			if !ok {
				return events, nil
			}
			p, ok := event.Object.(*corev1.Pod)
			if !ok {
				continue
			}
			ready := false
			for _, cond := range p.Status.Conditions {
				if cond.Type == corev1.PodReady {
					ready = cond.Status == corev1.ConditionTrue
					break
				}
			}
			events = append(events, PodEvent{
				eventType: string(event.Type),
				name:      p.Name,
				namespace: p.Namespace,
				phase:     string(p.Status.Phase),
				ready:     ready,
			})
		}
	}
}
