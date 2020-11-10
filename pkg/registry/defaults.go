//
// Copyright (c) 2020 Red Hat, Inc.
// This program and the accompanying materials are made
// available under the terms of the Eclipse Public License 2.0
// which is available at https://www.eclipse.org/legal/epl-2.0/
//
// SPDX-License-Identifier: EPL-2.0
//
// Contributors:
//   Red Hat, Inc. - initial API and implementation

package registry

import (
	registryv1alpha1 "github.com/devfile/registry-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

const (
	// Default image:tags
	DefaultDevfileIndexImage = "quay.io/devfile/metadata-server:next"
	DefaultOCIRegistryImage  = "registry:2.7.1"

	// Defaults/constants for devfile registry storages
	DefaultDevfileRegistryVolumeSize = "1Gi"
	DevfileRegistryVolumeEnabled     = true
	DevfileRegistryVolumeName        = "devfile-registry-storage"

	// Defaults/constants for devfile registry services
	DevfileIndexPortName = "devfile-registry-metadata"
	DevfileIndexPort     = 8080
	OCIRegistryPortName  = "oci-registry"
	OCIRegistryPort      = 5000
)

func getOCIRegistryImage(cr *registryv1alpha1.DevfileRegistry) string {
	if cr.Spec.OciRegistryImage != "" {
		return cr.Spec.OciRegistryImage
	}
	return DefaultOCIRegistryImage
}

func getDevfileRegistryVolumeSize(cr *registryv1alpha1.DevfileRegistry) string {
	if cr.Spec.Storage.RegistryVolumeSize != "" {
		return cr.Spec.Storage.RegistryVolumeSize
	}
	return DefaultDevfileRegistryVolumeSize
}

func getDevfileRegistryVolumeSource(cr *registryv1alpha1.DevfileRegistry) corev1.VolumeSource {
	if IsDevfileRegistryStorageEnabled(cr) {
		return corev1.VolumeSource{
			PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
				ClaimName: PVCName(cr.Name),
			},
		}
	}
	// If persistence is not enabled, return an empty dir volume source
	return corev1.VolumeSource{}
}

// IsDevfileRegistryStorageEnabled returns true if storage.Enabled is set in the DevfileRegistry CR
// If it's not set, it returns true by default.
func IsDevfileRegistryStorageEnabled(cr *registryv1alpha1.DevfileRegistry) bool {
	if cr.Spec.Storage.Enabled != nil {
		return *cr.Spec.Storage.Enabled
	}
	return DevfileRegistryVolumeEnabled
}
