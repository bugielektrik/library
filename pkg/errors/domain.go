package errors

import "net/http"

// Domain-specific errors for the library service

// Author errors
var (
	ErrAuthorNotFound = &Error{
		Code:       "AUTHOR_NOT_FOUND",
		Message:    "Author not found",
		HTTPStatus: http.StatusNotFound,
	}

	ErrAuthorAlreadyExists = &Error{
		Code:       "AUTHOR_ALREADY_EXISTS",
		Message:    "Author with this name already exists",
		HTTPStatus: http.StatusConflict,
	}

	ErrInvalidAuthorData = &Error{
		Code:       "INVALID_AUTHOR_DATA",
		Message:    "Invalid author data provided",
		HTTPStatus: http.StatusBadRequest,
	}
)

// Book errors
var (
	ErrBookNotFound = &Error{
		Code:       "BOOK_NOT_FOUND",
		Message:    "Book not found",
		HTTPStatus: http.StatusNotFound,
	}

	ErrBookAlreadyExists = &Error{
		Code:       "BOOK_ALREADY_EXISTS",
		Message:    "Book with this ISBN already exists",
		HTTPStatus: http.StatusConflict,
	}

	ErrInvalidBookData = &Error{
		Code:       "INVALID_BOOK_DATA",
		Message:    "Invalid book data provided",
		HTTPStatus: http.StatusBadRequest,
	}

	ErrInvalidISBN = &Error{
		Code:       "INVALID_ISBN",
		Message:    "Invalid ISBN format",
		HTTPStatus: http.StatusBadRequest,
	}
)

// Member errors
var (
	ErrMemberNotFound = &Error{
		Code:       "MEMBER_NOT_FOUND",
		Message:    "Member not found",
		HTTPStatus: http.StatusNotFound,
	}

	ErrMemberAlreadyExists = &Error{
		Code:       "MEMBER_ALREADY_EXISTS",
		Message:    "Member with this email already exists",
		HTTPStatus: http.StatusConflict,
	}

	ErrInvalidMemberData = &Error{
		Code:       "INVALID_MEMBER_DATA",
		Message:    "Invalid member data provided",
		HTTPStatus: http.StatusBadRequest,
	}

	ErrMembershipExpired = &Error{
		Code:       "MEMBERSHIP_EXPIRED",
		Message:    "Member's subscription has expired",
		HTTPStatus: http.StatusForbidden,
	}
)

// Subscription errors
var (
	ErrSubscriptionNotFound = &Error{
		Code:       "SUBSCRIPTION_NOT_FOUND",
		Message:    "Subscription not found",
		HTTPStatus: http.StatusNotFound,
	}

	ErrSubscriptionActive = &Error{
		Code:       "SUBSCRIPTION_ALREADY_ACTIVE",
		Message:    "Member already has an active subscription",
		HTTPStatus: http.StatusConflict,
	}

	ErrCannotCancelSubscription = &Error{
		Code:       "CANNOT_CANCEL_SUBSCRIPTION",
		Message:    "Cannot cancel subscription in current state",
		HTTPStatus: http.StatusUnprocessableEntity,
	}
)
