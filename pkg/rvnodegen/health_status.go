package rvnodegen

import (
	"k8s.io/apimachinery/pkg/runtime"
)

// HealthStatusType is a health status type for an object.
type HealthStatusType string

const (
	// HealthStatusTypeHealthy is a healthy object.
	HealthStatusTypeHealthy HealthStatusType = "Healthy"
	// HealthStatusTypeDegraded is a degraded object.
	HealthStatusTypeDegraded HealthStatusType = "Degraded"
	// HealthStatusTypeFailure is a failed object.
	HealthStatusTypeFailure HealthStatusType = "Failure"
)

// HealthStatuserFactory is a factory that creates HealthStatusers.
type HealthStatuserFactory func(lister Lister) (HealthStatuser, error)

// HealthStatuser is an interface that wraps health status.
type HealthStatuser interface {
	// HealthStatus generates health status for an object.
	HealthStatus(object runtime.Object) (HealthStatusType, error)
}

// ClusterHealthStatus generates health status using the cluster.
type ClusterHealthStatus struct {
	lister Lister
}

var _ HealthStatuser = &ClusterHealthStatus{}

// NewClusterHealthStatus creates an instance of ClusterHealthStatus.
func NewClusterHealthStatus(lister Lister) *ClusterHealthStatus {
	hs := &ClusterHealthStatus{
		lister: lister,
	}
	return hs
}

// HealthStatus generates status for an object.
func (hs *ClusterHealthStatus) HealthStatus(object runtime.Object) (HealthStatusType, error) {
	return HealthStatusTypeHealthy, nil
}
