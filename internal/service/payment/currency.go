package payment

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"go.uber.org/zap"

	"library-service/internal/provider/currency"
	"library-service/pkg/log"
)

func (s *Service) GetCurrencyRatesByDate(ctx context.Context, date time.Time) (dest []currency.Rate, err error) {
	logger := log.LoggerFromContext(ctx).Named("GetCurrencyRatesByDate")

	dest, err = s.currencyClient.GetRatesByDate(ctx, date)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		logger.Error("failed to get currency rates by date", zap.Error(err), zap.Any("date", date))
		return
	}

	return
}

func (s *Service) GetCurrencyRateByID(ctx context.Context, id string, date time.Time) (dest currency.Rate, err error) {
	logger := log.LoggerFromContext(ctx).Named("GetCurrencyRateByID")

	dest, err = s.currencyClient.GetRateByID(ctx, id, date)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		logger.Error("failed to get currency rate by id", zap.Error(err), zap.Any("id", id), zap.Any("date", date))
		return
	}

	return
}

func (s *Service) GetCurrencyRateFromCacheByID(ctx context.Context, id string) (dest currency.Rate, err error) {
	logger := log.LoggerFromContext(ctx).Named("GetCurrencyRateFromCacheByID")

	dest, err = s.currencyClient.GetRateFromCacheByID(id)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		logger.Error("failed to get currency rate from cache by id", zap.Error(err), zap.Any("id", id))
		return
	}

	return
}
