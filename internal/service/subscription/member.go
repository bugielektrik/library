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

func (s *Service) ListMembers(ctx context.Context) (res []member.Response, err error) {
	logger := log.LoggerFromContext(ctx).Named("ListMembers")

	data, err := s.memberRepository.List(ctx)
	if err != nil {
		logger.Error("failed to select", zap.Error(err))
		return
	}
	res = member.ParseFromEntities(data)

	return
}

func (s *Service) CreateMember(ctx context.Context, req member.Request) (res member.Response, err error) {
	logger := log.LoggerFromContext(ctx).Named("CreateMember")

	data := member.Entity{
		FullName: &req.FullName,
		Books:    req.Books,
	}

	data.ID, err = s.memberRepository.Add(ctx, data)
	if err != nil {
		logger.Error("failed to create", zap.Error(err))
		return
	}
	res = member.ParseFromEntity(data)

	return
}

func (s *Service) GetMember(ctx context.Context, id string) (res member.Response, err error) {
	logger := log.LoggerFromContext(ctx).Named("GetMember").With(zap.String("id", id))

	data, err := s.memberRepository.Get(ctx, id)
	if err != nil && !errors.Is(err, store.ErrorNotFound) {
		logger.Error("failed to get by id", zap.Error(err))
		return
	}
	res = member.ParseFromEntity(data)

	return
}

func (s *Service) UpdateMember(ctx context.Context, id string, req member.Request) (err error) {
	logger := log.LoggerFromContext(ctx).Named("UpdateMember").With(zap.String("id", id))

	data := member.Entity{
		FullName: &req.FullName,
		Books:    req.Books,
	}

	err = s.memberRepository.Update(ctx, id, data)
	if err != nil && !errors.Is(err, store.ErrorNotFound) {
		logger.Error("failed to update by id", zap.Error(err))
		return
	}

	return
}

func (s *Service) DeleteMember(ctx context.Context, id string) (err error) {
	logger := log.LoggerFromContext(ctx).Named("DeleteMember").With(zap.String("id", id))

	err = s.memberRepository.Delete(ctx, id)
	if err != nil && !errors.Is(err, store.ErrorNotFound) {
		logger.Error("failed to delete by id", zap.Error(err))
		return
	}

	return
}

func (s *Service) ListMemberBooks(ctx context.Context, id string) (res []book.Response, err error) {
	logger := log.LoggerFromContext(ctx).Named("ListMemberBooks").With(zap.String("id", id))

	data, err := s.memberRepository.Get(ctx, id)
	if err != nil && !errors.Is(err, store.ErrorNotFound) {
		logger.Error("failed to get by id", zap.Error(err))
		return
	}
	res = make([]book.Response, len(data.Books))

	for i := 0; i < len(data.Books); i++ {
		res[i], err = s.libraryService.GetBook(ctx, data.Books[i])
		if err != nil && !errors.Is(err, store.ErrorNotFound) {
			logger.Error("failed to get book by id", zap.Error(err))
			return
		}
	}

	return
}
