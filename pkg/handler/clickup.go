package handler

import (
	"bytes"
	"context"
	"github.com/astreter/amqpwrapper/v2"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"net/http"
	"strconv"
	"x-qdo/jiraclick/pkg/contract"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"x-qdo/jiraclick/pkg/config"
	"x-qdo/jiraclick/pkg/model"
	"x-qdo/jiraclick/pkg/provider/clickup"
	"x-qdo/jiraclick/pkg/publisher"
)

type clickUpWebhooks struct {
	cfg       *config.Config
	logger    *logrus.Logger
	publisher *publisher.EventPublisher
	clickup   *clickup.ConnectorPool
	db        contract.Storage
}

func NewClickUpWebhooksHandler(
	cfg *config.Config,
	logger *logrus.Logger,
	queue *amqpwrapper.RabbitChannel,
	clickup *clickup.ConnectorPool,
	db contract.Storage,
) (*clickUpWebhooks, error) {
	p, err := publisher.NewEventPublisher(queue)
	if err != nil {
		return nil, err
	}
	return &clickUpWebhooks{
		cfg:       cfg,
		logger:    logger,
		publisher: p,
		clickup:   clickup,
		db:        db,
	}, nil
}

func (h *clickUpWebhooks) TaskEvent(ctx *gin.Context) {
	spanCtx, span := otel.Tracer("http handler").Start(ctx.Request.Context(), "TaskEvent")
	defer span.End()

	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(ctx.Request.Body); err != nil {
		err = errors.Wrap(err, "ClickUp webhook: body can't be read")
		span.RecordError(err)
		h.logger.Error(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	body := buf.String()
	accessed, tenant := h.checkWebhookSecret(spanCtx, ctx.Request.Header.Get("X-Signature"), body)
	if !accessed {
		err := errors.New("ClickUp webhook: signature is not valid")
		span.RecordError(err)
		h.logger.Error(err)
		ctx.Status(http.StatusForbidden)
		return
	}

	h.logger.Debug("Clickup Raw Event: ", body)
	event, err := clickup.ParseEvent(spanCtx, body)
	h.logger.Debug("Clickup Parsed Event: ", event)
	if err != nil {
		err = errors.Wrap(err, "ClickUp webhook: body can't be parsed")
		span.RecordError(err)
		h.logger.Error(err)
		ctx.Status(http.StatusInternalServerError)
		return
	} else if event == nil {
		msg := "ClickUp webhook: webhook is without changes data"
		span.AddEvent(msg)
		h.logger.Debug(msg)
		ctx.Status(http.StatusOK)
		return
	}

	err = h.doAction(spanCtx, event, tenant)
	if err != nil {
		span.RecordError(err)
		h.logger.Error(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.Status(http.StatusOK)
}

func (h *clickUpWebhooks) doAction(ctx context.Context, event *clickup.WebhookEvent, tenant string) error {
	var changes model.TaskChanges
	span := trace.SpanFromContext(ctx)

	task, err := h.clickup.GetInstance(tenant).GetTask(ctx, event.TaskID)
	if err != nil {
		err = errors.Wrap(err, "ClickUp webhook: can't get task")
		span.RecordError(err)
		return err
	}
	span.AddEvent("task retrieved from Clickup")

	if event.Type == clickup.TaskUpdated && !isEventActual(task.DateUpdated, event.Changes[0].Date) {
		span.AddEvent("task doesn't have `taskUpdated` status or update is not actual")
		return nil
	}

	slackChannel := task.GetSlackChannel()
	if slackChannel == "" {
		err = errors.Wrap(err, "ClickUp webhook: slackChannel is not defined")
		span.RecordError(err)
		return err
	}
	span.AddEvent("slackChannel retrieved from task")

	changes = generateTaskChangesByEvent(event, task)
	span.AddEvent("changes are defined")
	err = h.publisher.ClickUpTaskUpdated(ctx, changes, slackChannel)
	if err != nil {
		err = errors.Wrap(err, "ClickUp webhook: can't trigger changes event")
		span.RecordError(err)
		return err
	}

	return nil
}

func generateTaskChangesByEvent(event *clickup.WebhookEvent, task *clickup.Task) model.TaskChanges {
	changes := model.TaskChanges{
		Type:      string(event.Type),
		ClickupID: event.TaskID,
	}
	for _, historyItem := range event.Changes {
		var value interface{}
		switch event.Type {
		case clickup.TaskUpdated:
			value = historyItem.After
		case clickup.TaskStatusUpdated:
			if after, ok := historyItem.After.(map[string]interface{}); ok {
				value = after["status"]
			}
		case clickup.TaskPriorityUpdated:
			if after, ok := historyItem.After.(map[string]interface{}); ok {
				value = after["priority"]
			}
		case clickup.TaskAssigneeUpdated:
			value = task.Assignees
		}
		changes.AddChange(historyItem.Field, value)
		changes.Username = historyItem.User.Username
	}

	return changes
}

func isEventActual(taskDate, eventDate string) bool {
	taskTimestamp, err := strconv.Atoi(taskDate)
	if err != nil {
		return true
	}
	eventTimestamp, err := strconv.Atoi(eventDate)
	if err != nil {
		return true
	}

	return taskTimestamp <= eventTimestamp
}

func (h *clickUpWebhooks) checkWebhookSecret(ctx context.Context, signature, body string) (bool, string) {
	span := trace.SpanFromContext(ctx)
	clickUpAccounts, err := h.db.GetClickUpAccounts(ctx)
	if err != nil {
		span.RecordError(err)
		panic(err)
	}
	for tenant, acc := range clickUpAccounts {
		if clickup.CheckSignature(ctx, signature, body, acc.WebhookSecret) {
			span.AddEvent("signature checked", trace.WithAttributes(attribute.Bool("valid", true)))
			return true, tenant
		}
	}
	span.AddEvent("signature checked", trace.WithAttributes(attribute.Bool("valid", false)))
	return false, ""
}
