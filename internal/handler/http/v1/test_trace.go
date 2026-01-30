package v1

import (
	"context"
	"encoding/json"
	"math/rand"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"

	"library-service/pkg/log"
	"library-service/pkg/server/response"
)

type TestTraceHandler struct{}

func NewTestTraceHandler() *TestTraceHandler {
	return &TestTraceHandler{}
}

type TraceTestRequest struct {
	UserID string `json:"user_id"`
	Action string `json:"action"`
}

type TraceTestResponse struct {
	Success      bool                   `json:"success"`
	TraceID      string                 `json:"trace_id"`
	ProcessedBy  []string               `json:"processed_by"`
	ResponseTime string                 `json:"response_time"`
	Data         map[string]interface{} `json:"data"`
}

// Routes регистрирует маршруты для тестирования трейсинга
func (h *TestTraceHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/trace", h.TestDistributedTrace)
	return r
}

// TestDistributedTrace эмулирует distributed trace с вызовами нескольких сервисов
// @Summary Test distributed tracing
// @Description Тестовый эндпоинт для проверки distributed tracing с несколькими спанами
// @Tags test
// @Accept json
// @Produce json
// @Param request body TraceTestRequest true "Test request"
// @Success 200 {object} TraceTestResponse
// @Router /api/v1/
// test/trace [post]
func (h *TestTraceHandler) TestDistributedTrace(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := log.FromContext(ctx).Named("test_trace")

	tracer := otel.Tracer("library-service")
	ctx, span := tracer.Start(ctx, "test-trace-handler")
	defer span.End()

	start := time.Now()

	var req TraceTestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "invalid request")
		response.BadRequest(w, r, err, nil)
		return
	}

	span.SetAttributes(
		attribute.String("user.id", req.UserID),
		attribute.String("action", req.Action),
	)

	logger.Info("processing trace test",
		zap.String("user_id", req.UserID),
		zap.String("action", req.Action),
	)

	processedBy := []string{}

	// Шаг 1: Валидация пользователя (эмулируем вызов UserService)
	userValid, err := h.validateUser(ctx, req.UserID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "user validation failed")
		response.InternalServerError(w, r, err, nil)
		return
	}
	if !userValid {
		span.SetStatus(codes.Error, "user not valid")
		response.BadRequest(w, r, nil, nil)
		return
	}
	processedBy = append(processedBy, "UserService")

	// Шаг 2: Проверка прав доступа (эмулируем вызов AuthService)
	authorized, err := h.checkAuthorization(ctx, req.UserID, req.Action)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "authorization check failed")
		response.InternalServerError(w, r, err, nil)
		return
	}
	if !authorized {
		span.SetStatus(codes.Error, "unauthorized")
		response.Forbidden(w, r)
		return
	}
	processedBy = append(processedBy, "AuthService")

	// Шаг 3: Получение данных из базы (эмулируем вызов DatabaseService)
	data, err := h.fetchData(ctx, req.UserID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "data fetch failed")
		response.InternalServerError(w, r, err, nil)
		return
	}
	processedBy = append(processedBy, "DatabaseService")

	// Шаг 4: Обогащение данных (эмулируем вызов EnrichmentService)
	enrichedData, err := h.enrichData(ctx, data)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "data enrichment failed")
		response.InternalServerError(w, r, err, nil)
		return
	}
	processedBy = append(processedBy, "EnrichmentService")

	// Шаг 5: Отправка в кеш (эмулируем вызов CacheService)
	err = h.cacheResult(ctx, req.UserID, enrichedData)
	if err != nil {
		// Не критично, просто логируем
		logger.Warn("cache failed", zap.Error(err))
	} else {
		processedBy = append(processedBy, "CacheService")
	}

	// Шаг 6: Отправка уведомления (эмулируем вызов NotificationService)
	err = h.sendNotification(ctx, req.UserID, "Action completed successfully")
	if err != nil {
		// Не критично, просто логируем
		logger.Warn("notification failed", zap.Error(err))
	} else {
		processedBy = append(processedBy, "NotificationService")
	}

	duration := time.Since(start)

	span.SetAttributes(
		attribute.Int("services.count", len(processedBy)),
		attribute.StringSlice("services.names", processedBy),
		attribute.Int64("duration.ms", duration.Milliseconds()),
	)
	span.SetStatus(codes.Ok, "success")

	// Получаем trace ID из span context
	traceID := span.SpanContext().TraceID().String()

	resp := TraceTestResponse{
		Success:      true,
		TraceID:      traceID,
		ProcessedBy:  processedBy,
		ResponseTime: duration.String(),
		Data:         enrichedData,
	}

	response.OK(w, r, resp)
}

// validateUser эмулирует вызов внешнего UserService
func (h *TestTraceHandler) validateUser(ctx context.Context, userID string) (bool, error) {
	tracer := otel.Tracer("library-service")
	ctx, span := tracer.Start(ctx, "user-service.validate")
	defer span.End()

	span.SetAttributes(
		attribute.String("service.name", "UserService"),
		attribute.String("user.id", userID),
	)

	// Эмулируем сетевую задержку
	time.Sleep(time.Duration(50+rand.Intn(100)) * time.Millisecond)

	// Валидируем что user_id не пустой
	if userID == "" {
		span.SetStatus(codes.Error, "empty user id")
		return false, nil
	}

	span.SetStatus(codes.Ok, "user valid")
	return true, nil
}

// checkAuthorization эмулирует вызов внешнего AuthService
func (h *TestTraceHandler) checkAuthorization(ctx context.Context, userID, action string) (bool, error) {
	tracer := otel.Tracer("library-service")
	ctx, span := tracer.Start(ctx, "auth-service.check-permission")
	defer span.End()

	span.SetAttributes(
		attribute.String("service.name", "AuthService"),
		attribute.String("user.id", userID),
		attribute.String("action", action),
	)

	// Эмулируем сетевую задержку
	time.Sleep(time.Duration(30+rand.Intn(70)) * time.Millisecond)

	// Всегда возвращаем true для теста
	span.SetStatus(codes.Ok, "authorized")
	return true, nil
}

// fetchData эмулирует вызов DatabaseService
func (h *TestTraceHandler) fetchData(ctx context.Context, userID string) (map[string]interface{}, error) {
	tracer := otel.Tracer("library-service")
	ctx, span := tracer.Start(ctx, "database-service.query")
	defer span.End()

	span.SetAttributes(
		attribute.String("service.name", "DatabaseService"),
		attribute.String("db.operation", "SELECT"),
		attribute.String("user.id", userID),
	)

	// Эмулируем задержку БД
	time.Sleep(time.Duration(100+rand.Intn(150)) * time.Millisecond)

	data := map[string]interface{}{
		"user_id":    userID,
		"status":     "active",
		"last_login": time.Now().Add(-24 * time.Hour).Format(time.RFC3339),
		"books_read": rand.Intn(100),
	}

	span.SetAttributes(
		attribute.Int("db.rows_returned", 1),
	)
	span.SetStatus(codes.Ok, "data fetched")

	return data, nil
}

// enrichData эмулирует вызов EnrichmentService
func (h *TestTraceHandler) enrichData(ctx context.Context, data map[string]interface{}) (map[string]interface{}, error) {
	tracer := otel.Tracer("library-service")
	ctx, span := tracer.Start(ctx, "enrichment-service.enrich")
	defer span.End()

	span.SetAttributes(
		attribute.String("service.name", "EnrichmentService"),
	)

	// Эмулируем задержку обогащения
	time.Sleep(time.Duration(40+rand.Intn(80)) * time.Millisecond)

	// Добавляем дополнительные данные
	data["enriched_at"] = time.Now().Format(time.RFC3339)
	data["recommendations"] = []string{"Book A", "Book B", "Book C"}
	data["user_tier"] = "premium"

	span.SetAttributes(
		attribute.Int("enriched.fields_added", 3),
	)
	span.SetStatus(codes.Ok, "data enriched")

	return data, nil
}

// cacheResult эмулирует вызов CacheService
func (h *TestTraceHandler) cacheResult(ctx context.Context, key string, data map[string]interface{}) error {
	tracer := otel.Tracer("library-service")
	ctx, span := tracer.Start(ctx, "cache-service.set")
	defer span.End()

	span.SetAttributes(
		attribute.String("service.name", "CacheService"),
		attribute.String("cache.key", key),
		attribute.String("cache.operation", "SET"),
	)

	// Эмулируем задержку кеширования
	time.Sleep(time.Duration(10+rand.Intn(30)) * time.Millisecond)

	span.SetStatus(codes.Ok, "cached")
	return nil
}

// sendNotification эмулирует вызов NotificationService
func (h *TestTraceHandler) sendNotification(ctx context.Context, userID, message string) error {
	tracer := otel.Tracer("library-service")
	ctx, span := tracer.Start(ctx, "notification-service.send")
	defer span.End()

	span.SetAttributes(
		attribute.String("service.name", "NotificationService"),
		attribute.String("user.id", userID),
		attribute.String("notification.type", "email"),
	)

	// Эмулируем задержку отправки
	time.Sleep(time.Duration(60+rand.Intn(90)) * time.Millisecond)

	span.SetStatus(codes.Ok, "notification sent")
	return nil
}
