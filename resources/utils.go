package resources

func boolPtr(val bool) *bool {
	v := val
	return &v
}

func stringPtr(val string) *string {
	v := val
	return &v
}
