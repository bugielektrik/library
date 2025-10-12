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

	ErrBookNotAvailable = &Error{
		Code:       "BOOK_NOT_AVAILABLE",
		Message:    "Book is not available for borrowing",
		HTTPStatus: http.StatusConflict,
	}

	ErrBookHasActiveLoans = &Error{
		Code:       "BOOK_HAS_ACTIVE_LOANS",
		Message:    "Book has active loans and cannot be deleted",
		HTTPStatus: http.StatusConflict,
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

	ErrMemberSuspended = &Error{
		Code:       "MEMBER_SUSPENDED",
		Message:    "Member account is suspended",
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

	ErrInvalidSubscription = &Error{
		Code:       "INVALID_SUBSCRIPTION",
		Message:    "Invalid subscription type or configuration",
		HTTPStatus: http.StatusBadRequest,
	}

	ErrSubscriptionExpired = &Error{
		Code:       "SUBSCRIPTION_EXPIRED",
		Message:    "Subscription has expired",
		HTTPStatus: http.StatusForbidden,
	}

	ErrSubscriptionNotActive = &Error{
		Code:       "SUBSCRIPTION_NOT_ACTIVE",
		Message:    "Member does not have an active subscription",
		HTTPStatus: http.StatusForbidden,
	}
)

// Payment errors
var (
	ErrPaymentNotFound = &Error{
		Code:       "PAYMENT_NOT_FOUND",
		Message:    "Payment not found",
		HTTPStatus: http.StatusNotFound,
	}

	ErrPaymentAlreadyProcessed = &Error{
		Code:       "PAYMENT_ALREADY_PROCESSED",
		Message:    "Payment has already been processed",
		HTTPStatus: http.StatusConflict,
	}

	ErrPaymentExpired = &Error{
		Code:       "PAYMENT_EXPIRED",
		Message:    "Payment has expired",
		HTTPStatus: http.StatusGone,
	}

	ErrPaymentGateway = &Error{
		Code:       "PAYMENT_GATEWAY_ERROR",
		Message:    "Payment provider error",
		HTTPStatus: http.StatusBadGateway,
	}

	ErrInvalidPaymentStatus = &Error{
		Code:       "INVALID_PAYMENT_STATUS",
		Message:    "Invalid payment status transition",
		HTTPStatus: http.StatusBadRequest,
	}

	ErrInvalidAmount = &Error{
		Code:       "INVALID_AMOUNT",
		Message:    "Invalid payment amount",
		HTTPStatus: http.StatusBadRequest,
	}

	ErrInsufficientFunds = &Error{
		Code:       "INSUFFICIENT_FUNDS",
		Message:    "Insufficient funds for this transaction",
		HTTPStatus: http.StatusPaymentRequired,
	}

	ErrRefundNotAllowed = &Error{
		Code:       "REFUND_NOT_ALLOWED",
		Message:    "Refund is not allowed for this payment",
		HTTPStatus: http.StatusConflict,
	}
)

// Reservation errors
var (
	ErrReservationNotFound = &Error{
		Code:       "RESERVATION_NOT_FOUND",
		Message:    "Reservation not found",
		HTTPStatus: http.StatusNotFound,
	}

	ErrReservationExpired = &Error{
		Code:       "RESERVATION_EXPIRED",
		Message:    "Reservation has expired",
		HTTPStatus: http.StatusGone,
	}

	ErrReservationAlreadyFulfilled = &Error{
		Code:       "RESERVATION_ALREADY_FULFILLED",
		Message:    "Reservation has already been fulfilled",
		HTTPStatus: http.StatusConflict,
	}

	ErrReservationAlreadyCancelled = &Error{
		Code:       "RESERVATION_ALREADY_CANCELLED",
		Message:    "Reservation has already been cancelled",
		HTTPStatus: http.StatusConflict,
	}

	ErrBookAlreadyReserved = &Error{
		Code:       "BOOK_ALREADY_RESERVED",
		Message:    "Member already has an active reservation for this book",
		HTTPStatus: http.StatusConflict,
	}

	ErrBookAlreadyBorrowed = &Error{
		Code:       "BOOK_ALREADY_BORROWED",
		Message:    "Member already has this book borrowed",
		HTTPStatus: http.StatusConflict,
	}

	ErrCannotCancelReservation = &Error{
		Code:       "CANNOT_CANCEL_RESERVATION",
		Message:    "Reservation cannot be cancelled in current status",
		HTTPStatus: http.StatusConflict,
	}
)
