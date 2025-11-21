package subscription

import (
	"library-service/internal/domain/member"
	"library-service/internal/service/library"
)

type Service struct {
	memberRepository member.Repository
	libraryService   *library.Service
}

func New(memberRepository member.Repository, libraryService *library.Service) *Service {
	return &Service{
		memberRepository: memberRepository,
		libraryService:   libraryService,
	}
}
