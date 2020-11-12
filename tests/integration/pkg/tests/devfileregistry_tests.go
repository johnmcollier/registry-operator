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
//

package tests

import (
	"fmt"
	"time"

	"github.com/devfile/registry-operator/pkg/util"
	"github.com/devfile/registry-operator/tests/integration/pkg/client"
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
)

// Integration/e2e test logic based on https://github.com/devfile/devworkspace-operator/tree/master/test/e2e

var _ = ginkgo.Describe("[Create Devfile Registry resource]", func() {
	ginkgo.It("Should deploy a devfile registry on to the cluster", func() {
		crName := "devfileregistry"
		label := "devfileregistry_cr=" + crName
		k8sClient, err := client.NewK8sClient()
		if err != nil {
			ginkgo.Fail("Failed to create k8s client: " + err.Error())
			return
		}

		// Deploy the devfileregistry resource for this test case and wait for the pod to be running
		err = k8sClient.KubectlApplyResource("tests/integration/samples/devfileregistry.yaml")
		if err != nil {
			ginkgo.Fail("Failed to create devfileregistry instance: " + err.Error())
			return
		}
		deploy, err := k8sClient.WaitForPodRunningByLabel(label)
		if !deploy {
			fmt.Println("Devfile Registry didn't start properly")
		}
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		// Wait for the registry instance to become ready
		err = k8sClient.WaitForRegistryInstance(crName, 30*time.Second)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		// Retrieve the registry URL and verify the server is up and running
		registry, err := k8sClient.GetRegistryInstance(crName)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		err = util.WaitForServer("http://"+registry.Status.URL, 30*time.Second)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
	})
})
