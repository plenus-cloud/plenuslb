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

package v1alpha1

import (
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	// EphemeralIPPoolCRDPlural is the plural name of the EphemeralIPPoolCRD
	EphemeralIPPoolCRDPlural string = "ephemeralippools"
	// FullEphemeralIPPoolCRDName is the full name of the EphemeralIPPoolCRD
	FullEphemeralIPPoolCRDName string = EphemeralIPPoolCRDPlural + "." + CRDGroup
)

func addEphemeralIPPoolKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion,
		&EphemeralIPPool{},
		&EphemeralIPPoolList{},
	)
	meta_v1.AddToGroupVersion(scheme, SchemeGroupVersion)
	return nil
}
