package logging

import (
	corev1 "k8s.io/api/core/v1"
)

func promtailServiceAccount() (*corev1.ServiceAccount, error) {
	sa := &corev1.ServiceAccount{}
	return sa, nil
}

func createOrUpdateServiceAccount(sa *corev1.ServiceAccount) error {
	return nil
}
