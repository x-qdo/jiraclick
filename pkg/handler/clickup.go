package handler

import (
	"bytes"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"x-qdo/jiraclick/pkg/config"
	"x-qdo/jiraclick/pkg/model"
	"x-qdo/jiraclick/pkg/provider"
	"x-qdo/jiraclick/pkg/provider/clickup"
	"x-qdo/jiraclick/pkg/publisher"
)

type clickUpWebhooks struct {
	cfg       *config.Config
	logger    *logrus.Logger
	publisher *publisher.EventPublisher
	clickup   *clickup.ConnectorPool
}

func NewClickUpWebhooksHandler(
	cfg *config.Config,
	logger *logrus.Logger,
	queue *provider.RabbitChannel,
	clickup *clickup.ConnectorPool,
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
	}, nil
}

func (h *clickUpWebhooks) TaskEvent(ctx *gin.Context) {
	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(ctx.Request.Body); err != nil {
		h.logger.Error(errors.Wrap(err, "ClickUp webhook: body can't be read"))
		ctx.Status(http.StatusInternalServerError)
		return
	}

	body := buf.String()
	accessed, tenant := h.checkWebhookSecret(ctx, body)
	if !accessed {
		h.logger.Error("ClickUp webhook: signature is not valid")
		ctx.Status(http.StatusForbidden)
		return
	}

	h.logger.Debug("Clickup Raw Event: ", body)
	event, err := clickup.ParseEvent(body)
	h.logger.Debug("Clickup Parsed Event: ", event)
	if err != nil {
		h.logger.Error(errors.Wrap(err, "ClickUp webhook: body can't be parsed"))
		ctx.Status(http.StatusInternalServerError)
		return
	} else if event == nil {
		h.logger.Debug("ClickUp webhook: webhook is without changes data")
		ctx.Status(http.StatusOK)
		return
	}

	err = h.doAction(event, tenant)
	if err != nil {
		h.logger.Error(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.Status(http.StatusOK)
}

func (h *clickUpWebhooks) doAction(event *clickup.WebhookEvent, tenant string) error {
	var changes model.TaskChanges

	task, err := h.clickup.GetInstance(tenant).GetTask(event.TaskID)
	if err != nil {
		return errors.Wrap(err, "ClickUp webhook: can't get task")
	}

	if event.Type == clickup.TaskUpdated && !isEventActual(task.DateUpdated, event.Changes[0].Date) {
		return nil
	}

	slackChannel := task.GetSlackChannel()
	if slackChannel == "" {
		return errors.Wrap(err, "ClickUp webhook: slackChannel is not defined")
	}

	changes = generateTaskChangesByEvent(event)
	err = h.publisher.ClickUpTaskUpdated(changes, slackChannel)
	if err != nil {
		return errors.Wrap(err, "ClickUp webhook: can't trigger changes event")
	}

	return nil
}

func generateTaskChangesByEvent(event *clickup.WebhookEvent) model.TaskChanges {
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
			if after, ok := historyItem.After.(map[string]interface{}); ok {
				value = after["username"]
			} else {
				value = nil
			}
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

func (h *clickUpWebhooks) checkWebhookSecret(ctx *gin.Context, body string) (bool, string) {
	for tenant, clickUpCfg := range h.cfg.ClickUp {
		if clickup.CheckSignature(ctx.Request.Header.Get("X-Signature"), body, clickUpCfg.WebhookSecret) {
			return true, tenant
		}
	}
	return false, ""
}
