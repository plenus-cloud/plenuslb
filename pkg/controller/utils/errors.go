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

package utils

import "errors"

// ErrNoOperatorNodeAvailable is returned when an operator can not be found on a specific node
var ErrNoOperatorNodeAvailable = errors.New("No operator node available")

// ErrFailedToDialWithOperator is returned when the controller cannot talk with the operator
var ErrFailedToDialWithOperator = errors.New("Failed to dial with operator")
