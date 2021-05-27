package backend

import (
	"context"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"

	"github.com/joyrex2001/kubedock/internal/model/types"
)

// DeleteAll will delete all resources that kubedock=true
func (in *instance) DeleteAll() error {
	if err := in.deleteServices("kubedock=true"); err != nil {
		klog.Errorf("error deleting services: %s", err)
	}
	return in.deleteDeployments("kubedock=true")
}

// DeleteWithKubedockID will delete all resources that have given kubedock.id
func (in *instance) DeleteWithKubedockID(id string) error {
	if err := in.deleteServices("kubedock.id=" + id); err != nil {
		klog.Errorf("error deleting services: %s", err)
	}
	return in.deleteDeployments("kubedock.id=" + id)
}

// DeleteContainer will delete given container object in kubernetes.
func (in *instance) DeleteContainer(tainr *types.Container) error {
	if err := in.deleteServices("kubedock.containerid=" + tainr.ShortID); err != nil {
		klog.Errorf("error deleting services: %s", err)
	}
	return in.deleteDeployments("kubedock.containerid=" + tainr.ShortID)
}

// DeleteContainersOlderThan will delete containers than are orchestrated
// by kubedock and are older than the given keepmax duration.
func (in *instance) DeleteContainersOlderThan(keepmax time.Duration) error {
	deps, err := in.cli.AppsV1().Deployments(in.namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: "kubedock=true",
	})
	if err != nil {
		return err
	}
	for _, dep := range deps.Items {
		if in.isOlderThan(dep.ObjectMeta, keepmax) {
			klog.V(3).Infof("deleting deployment: %s", dep.Name)
			if err := in.deleteServices("kubedock.containerid=" + dep.Name); err != nil {
				klog.Errorf("error deleting services: %s", err)
			}
			if err := in.cli.AppsV1().Deployments(dep.Namespace).Delete(context.TODO(), dep.Name, metav1.DeleteOptions{}); err != nil {
				return err
			}
		}
	}
	return nil
}

// DeleteServicesOlderThan will delete services than are orchestrated
// by kubedock and are older than the given keepmax duration.
func (in *instance) DeleteServicesOlderThan(keepmax time.Duration) error {
	svcs, err := in.cli.CoreV1().Services(in.namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: "kubedock=true",
	})
	if err != nil {
		return err
	}
	for _, svc := range svcs.Items {
		if in.isOlderThan(svc.ObjectMeta, keepmax) {
			klog.V(3).Infof("deleting service: %s", svc.Name)
			if err := in.cli.CoreV1().Services(svc.Namespace).Delete(context.TODO(), svc.Name, metav1.DeleteOptions{}); err != nil {
				return err
			}
		}
	}
	return nil
}

// isOlderThan will check if given resource metadata has an older timestamp
// compared to given keepmax duration
func (in *instance) isOlderThan(met metav1.ObjectMeta, keepmax time.Duration) bool {
	if met.DeletionTimestamp != nil {
		klog.V(3).Infof("ignoring %v, already in deleting state", met)
		return false
	}
	old := metav1.NewTime(time.Now().Add(-keepmax))
	return met.CreationTimestamp.Before(&old)
}

// deleteServices will delete k8s service resources which match the
// given label selector.
func (in *instance) deleteServices(selector string) error {
	svcs, err := in.cli.CoreV1().Services(in.namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: selector,
	})
	if err != nil {
		return err
	}
	for _, svc := range svcs.Items {
		if err := in.cli.CoreV1().Services(svc.Namespace).Delete(context.TODO(), svc.Name, metav1.DeleteOptions{}); err != nil {
			return err
		}
	}
	return nil
}

// deleteDeployments will delete k8s deployments resources which match the
// given label selector.
func (in *instance) deleteDeployments(selector string) error {
	deps, err := in.cli.AppsV1().Deployments(in.namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: selector,
	})
	if err != nil {
		return err
	}
	for _, svc := range deps.Items {
		if err := in.cli.AppsV1().Deployments(svc.Namespace).Delete(context.TODO(), svc.Name, metav1.DeleteOptions{}); err != nil {
			return err
		}
	}
	return nil
}