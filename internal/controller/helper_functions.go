package controller

import (
	corev1 "k8s.io/api/core/v1"
)

func getServiceType(s string) corev1.ServiceType {
	if s == "NodePort" {
		return corev1.ServiceTypeNodePort
	} else {
		return corev1.ServiceTypeClusterIP
	}
}
