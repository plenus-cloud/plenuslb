package v1alpha1

func init() {
	// We only register manually written functions here. The registration of the
	// generated functions takes place in the generated files. The separation
	// makes the code compile even when the generated files are missing.
	SchemeBuilder.Register(addIPAllocationKnownTypes)

	SchemeBuilder.Register(addEphemeralIPPoolKnownTypes)

	SchemeBuilder.Register(addPersistentIPPoolKnownTypes)
}
