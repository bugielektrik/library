package subscription

import (
	"library-service/internal/domain/member"
	"library-service/internal/service/library"
)

// Service is an implementation of the Service
type Service struct {
	memberRepository member.Repository
	libraryService   *library.Service
}

// New creates a new instance of the Service with the provided member repository and library service.
func New(memberRepository member.Repository, libraryService *library.Service) *Service {
	return &Service{
		memberRepository: memberRepository,
		libraryService:   libraryService,
	}
}
