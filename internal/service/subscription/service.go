package subscription

import (
	"library/internal/domain/member"
	"library/internal/service/library"
)

type Service struct {
	libraryService   library.Service
	memberRepository member.Repository
}

func New(l library.Service, m member.Repository) Service {
	return Service{
		libraryService:   l,
		memberRepository: m,
	}
}
