package subscription

import (
	"context"
	"errors"

	"go.uber.org/zap"

	"library-service/internal/domain/book"
	"library-service/internal/domain/member"
	"library-service/pkg/log"
	"library-service/pkg/store"
)

func (s *Service) ListMembers(ctx context.Context) ([]member.Response, error) {
	logger := log.FromContext(ctx).Named("list_members")

	members, err := s.memberRepository.List(ctx)
	if err != nil {
		logger.Error("failed to list members", zap.Error(err))
		return nil, err
	}
	return member.ParseFromEntities(members), nil
}

func (s *Service) CreateMember(ctx context.Context, req member.Request) (member.Response, error) {
	logger := log.FromContext(ctx).Named("create_member").With(zap.Any("member", req))

	newMember := member.New(req)

	id, err := s.memberRepository.Add(ctx, newMember)
	if err != nil {
		logger.Error("failed to create member", zap.Error(err))
		return member.Response{}, err
	}
	newMember.ID = id

	return member.ParseFromEntity(newMember), nil
}

func (s *Service) GetMember(ctx context.Context, id string) (member.Response, error) {
	logger := log.FromContext(ctx).Named("get_member").With(zap.String("id", id))

	memberData, err := s.memberRepository.Get(ctx, id)
	if err != nil {
		if errors.Is(err, store.ErrorNotFound) {
			logger.Warn("member not found", zap.Error(err))
			return member.Response{}, err
		}
		logger.Error("failed to get member", zap.Error(err))
		return member.Response{}, err
	}
	return member.ParseFromEntity(memberData), nil
}

func (s *Service) UpdateMember(ctx context.Context, id string, req member.Request) error {
	logger := log.FromContext(ctx).Named("update_member").With(zap.String("id", id), zap.Any("member", req))

	updatedMember := member.New(req)

	err := s.memberRepository.Update(ctx, id, updatedMember)
	if err != nil {
		if errors.Is(err, store.ErrorNotFound) {
			logger.Warn("member not found", zap.Error(err))
			return err
		}
		logger.Error("failed to update member", zap.Error(err))
		return err
	}
	return nil
}

func (s *Service) DeleteMember(ctx context.Context, id string) error {
	logger := log.FromContext(ctx).Named("delete_member").With(zap.String("id", id))

	err := s.memberRepository.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, store.ErrorNotFound) {
			logger.Warn("member not found", zap.Error(err))
			return err
		}
		logger.Error("failed to delete member", zap.Error(err))
		return err
	}
	return nil
}

func (s *Service) ListMemberBooks(ctx context.Context, memberID string) ([]book.Response, error) {
	logger := log.FromContext(ctx).Named("list_member_books").With(zap.String("id", memberID))

	memberData, err := s.memberRepository.Get(ctx, memberID)
	if err != nil {
		if errors.Is(err, store.ErrorNotFound) {
			logger.Warn("member not found", zap.Error(err))
			return nil, err
		}
		logger.Error("failed to get member", zap.Error(err))
		return nil, err
	}

	bookResponses := make([]book.Response, 0, len(memberData.Books))
	for _, bookID := range memberData.Books {
		bookResponse, err := s.libraryService.GetBook(ctx, bookID)
		if err != nil {
			if errors.Is(err, store.ErrorNotFound) {
				logger.Warn("book not found", zap.Error(err))
				continue
			}
			logger.Error("failed to get book", zap.Error(err))
			return nil, err
		}
		bookResponses = append(bookResponses, bookResponse)
	}
	return bookResponses, nil
}
