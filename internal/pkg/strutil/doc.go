// Package strutil provides common string manipulation utilities.
//
// This package contains helper functions for string operations commonly
// needed across the application, particularly when working with pointer
// string fields in domain entities.
//
// Utilities:
//   - SafeString: Dereference string pointer with nil check
//   - SafeStringPtr: Create string pointer from value
//
// Example usage:
//
//	import "library-service/internal/infrastructure/pkg/strutil"
//
//	// Converting from domain entity (pointers) to DTO (values)
//	dto := BookDTO{
//	    Name: strutil.SafeString(entity.Name),
//	}
//
//	// Converting from DTO (values) to domain entity (pointers)
//	entity := book.Book{
//	    Name: strutil.SafeStringPtr(dto.Name),
//	}
package strutil
