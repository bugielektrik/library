// Package email provides email delivery adapter implementations.
//
// This package implements outbound adapters for sending transactional emails
// to members. Supports SMTP and potential cloud email services (SendGrid, SES).
//
// Email types:
//   - Welcome email: Sent on member registration
//   - Reservation ready: Book is available for pickup
//   - Reservation expiring: Reminder before 48h deadline
//   - Subscription expiring: Reminder before renewal date
//   - Payment receipt: Confirmation of successful payment
//   - Password reset: Token for password reset flow
//
// Email service interface:
//
//	type Service interface {
//	    SendEmail(ctx context.Context, to string, template EmailTemplate, data map[string]interface{}) error
//	    SendBulkEmail(ctx context.Context, recipients []string, template EmailTemplate, data map[string]interface{}) error
//	}
//
// Template system:
//   - HTML templates with embedded CSS
//   - Text fallback for email clients without HTML support
//   - Variable substitution for personalization
//   - Templates stored in templates/email/
//
// SMTP configuration:
//   - SMTP host and port via environment variables
//   - TLS encryption required
//   - Authentication with username/password
//   - Connection pooling for performance
//
// Example usage:
//
//	emailService.SendEmail(ctx, member.Email, emailTemplateWelcome, map[string]interface{}{
//	    "MemberName": member.FullName,
//	    "LoginURL":   "https://library.example.com/login",
//	})
//
// Error handling:
//   - Email delivery failures logged but not blocking
//   - Retry logic for transient failures (max 3 attempts)
//   - Failed emails queued for later retry
//   - Bounced emails tracked to pause further sends
//
// Rate limiting:
//   - Max 100 emails per minute
//   - Bulk emails sent in batches
//   - Exponential backoff on rate limit errors
//
// Testing:
//   - Mock email service for unit tests
//   - Development mode: Log emails instead of sending
//   - Test mode: Capture emails in memory for assertions
//
// Configuration:
//   - SMTP_HOST, SMTP_PORT, SMTP_USERNAME, SMTP_PASSWORD
//   - EMAIL_FROM: Sender address
//   - EMAIL_FROM_NAME: Sender display name
//   - EMAIL_ENABLED: Toggle email sending (dev/test)
package email
