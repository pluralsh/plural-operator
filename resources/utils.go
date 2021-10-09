package resources

func BoolPtr(val bool) *bool {
	v := val
	return &v
}

func StringPtr(val string) *string {
	v := val
	return &v
}
