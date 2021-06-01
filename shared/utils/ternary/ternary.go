package ternary

// IfNotEmptyOr returns a if not empty, otherwise b
func IfNotEmptyOr(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

// If returns a if con is true, b otherwise
func If(con bool, a, b interface{}) interface{} {
	if con {
		return a
	}
	return b
}
