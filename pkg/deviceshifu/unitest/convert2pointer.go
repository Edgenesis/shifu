package unitest

// SO FAR ONLY FOR UNIT TESTING USAGE
// DO NOT USE IN NONE UNIT TESTING CODE
// if there are common use case, let's move it to an much common package or even another code repo

func BoolPointer(b bool) *bool {
	return &b
}

func StrPointer(s string) *string {
	return &s
}

func IntPointer(i int) *int {
	return &i
}

func Int64Pointer(i int64) *int64 {
	return &i
}
