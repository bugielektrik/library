package subscription

import (
	"context"

	"library-service/internal/domain/book"
	"library-service/internal/domain/member"
)

func (s *Service) ListMembers(ctx context.Context) (res []member.Response, err error) {
	data, err := s.memberRepository.Select(ctx)
	if err != nil {
		return
	}
	res = member.ParseFromEntities(data)

	return
}

func (s *Service) AddMember(ctx context.Context, req member.Request) (res member.Response, err error) {
	data := member.Entity{
		FullName: &req.FullName,
		Books:    req.Books,
	}

	data.ID, err = s.memberRepository.Create(ctx, data)
	if err != nil {
		return
	}
	res = member.ParseFromEntity(data)

	return
}

func (s *Service) GetMember(ctx context.Context, id string) (res member.Response, err error) {
	data, err := s.memberRepository.Get(ctx, id)
	if err != nil {
		return
	}
	res = member.ParseFromEntity(data)

	return
}

func (s *Service) UpdateMember(ctx context.Context, id string, req member.Request) (err error) {
	data := member.Entity{
		FullName: &req.FullName,
		Books:    req.Books,
	}
	return s.memberRepository.Update(ctx, id, data)
}

func (s *Service) DeleteMember(ctx context.Context, id string) (err error) {
	return s.memberRepository.Delete(ctx, id)
}

func (s *Service) ListMemberBooks(ctx context.Context, id string) (res []book.Response, err error) {
	data, err := s.memberRepository.Get(ctx, id)
	if err != nil {
		return
	}
	res = make([]book.Response, len(data.Books))

	for i := 0; i < len(data.Books); i++ {
		res[i], err = s.libraryService.GetBook(ctx, data.Books[i])
		if err != nil {
			return
		}
	}

	return
}
