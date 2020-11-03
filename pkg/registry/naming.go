package registry

// DeploymentName returns the name of the deployment object associated with the DevfileRegistry CR
// Just returns the CR name right now, but extracting to a function to avoid relying on that assumption
func DeploymentName(devfileRegistryName string) string {
	return devfileRegistryName
}

// ServiceName returns the name of the service object associated with the DevfileRegistry CR
// Just returns the CR name right now, but extracting to a function to avoid relying on that assumption
func ServiceName(devfileRegistryName string) string {
	return devfileRegistryName
}

// PVCName returns the name of the PVC object associated with the DevfileRegistry CR
// Just returns the CR name right now, but extracting to a function to avoid relying on that assumption
func PVCName(devfileRegistryName string) string {
	return devfileRegistryName
}

// DevfilesRouteName returns the name of the route object associated with the devfile index route
func DevfilesRouteName(devfileRegistryName string) string {
	return devfileRegistryName + "-devfiles"
}

// OCIRouteName returns the name of the route object associated with the OCI registry route
func OCIRouteName(devfileRegistryName string) string {
	return devfileRegistryName + "-oci"
}
