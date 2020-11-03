package registry

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

func generateObjectMeta(name string, namespace string, labels map[string]string) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:      name,
		Namespace: namespace,
		Labels:    labels,
	}
}

// LabelsForDevfileRegistry returns the labels for selecting the resources
// belonging to the given devfileregistry CR name.
func LabelsForDevfileRegistry(name string) map[string]string {
	return map[string]string{"app": "devfileregistry", "devfileregistry_cr": name}
}
