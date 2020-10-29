/*
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package crddeployer

import (
	"errors"
	"strconv"
	"time"

	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"

	"plenus.io/plenuslb/pkg/controller/clients"
	plwait "plenus.io/plenuslb/pkg/controller/wait"
)

// CreateOrUpdateCRD create or opdate the given CRD
func CreateOrUpdateCRD(crdName string, v1Definition *apiextv1.CustomResourceDefinition) error {
	versionInfo, err := clients.GetK8sClient().Discovery().ServerVersion()
	if err != nil {
		klog.Error(err)
		return err
	}
	klog.Infof("Deploying %s CRD, server version is %s.%s", crdName, versionInfo.Major, versionInfo.Minor)

	extClient := clients.GetExtensionClient()
	existingResourceV1, err := extClient.ApiextensionsV1().CustomResourceDefinitions().Get(crdName, metav1.GetOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		klog.Error(err)
		return err
	} else if err == nil {
		return updateCRDV1(existingResourceV1, v1Definition)
	}

	minorNum, err := strconv.Atoi(versionInfo.Minor)
	if err != nil {
		klog.Error(err)
		return err
	}

	if minorNum >= 16 {
		return createCRDV1(v1Definition)
	}
	return errors.New("Cluster version ot supported")
}

func createCRDV1(crd *apiextv1.CustomResourceDefinition) error {
	klog.Infof("CRD %s already exists, creating using apiextension v1", crd.GetName())
	deployedCRD, err := clients.GetExtensionClient().ApiextensionsV1().CustomResourceDefinitions().Create(crd)
	if err != nil {
		klog.Error(err)
		return err
	}
	return plwait.ForCRD(2*time.Minute, *deployedCRD)
}

func updateCRDV1(existingCRD *apiextv1.CustomResourceDefinition, newCrd *apiextv1.CustomResourceDefinition) error {
	klog.Infof("CRD %s already exists, updating using apiextension v1", existingCRD.GetName())
	// ResourceVersion in mandatory when updatin a crd
	newCrd.ObjectMeta.ResourceVersion = existingCRD.ObjectMeta.ResourceVersion
	deployedCRD, err := clients.GetExtensionClient().ApiextensionsV1().CustomResourceDefinitions().Update(newCrd)
	if err != nil {
		klog.Error(err)
		return err
	}
	return plwait.ForCRD(2*time.Minute, *deployedCRD)
}
