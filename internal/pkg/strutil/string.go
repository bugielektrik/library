// Package strutil provides string utility functions.
package strutil

// SafeString returns the value of a string pointer or empty string if nil.
// This is useful when working with domain entities that use pointer fields
// for optional values.
//
// Example:
//
//	var name *string
//	fmt.Println(strutil.SafeString(name)) // ""
//
//	str := "hello"
//	fmt.Println(strutil.SafeString(&str)) // "hello"
func SafeString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// SafeStringPtr returns a pointer to the given string.
// This is useful when constructing domain entities that use pointer fields.
//
// Example:
//
//	book := book.Book{
//	    Name: strutil.SafeStringPtr("Clean Code"),
//	}
func SafeStringPtr(s string) *string {
	return &s
}
