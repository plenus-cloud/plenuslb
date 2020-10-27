package utils

import "errors"

// ErrNoOperatorNodeAvailable is returned when an operator can not be found on a specific node
var ErrNoOperatorNodeAvailable = errors.New("No operator node available")

// ErrFailedToDialWithOperator is returned when the controller cannot talk with the operator
var ErrFailedToDialWithOperator = errors.New("Failed to dial with operator")
