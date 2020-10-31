package consumer

import (
	"github.com/streadway/amqp"
	"x-qdo/jiraclick/pkg/config"
	"x-qdo/jiraclick/pkg/contract"
	"x-qdo/jiraclick/pkg/provider"
)

type TaskCreateClickupAction struct {
	client *provider.ClickUpAPIClient
}

func NewTaskCreateClickupAction(cfg *config.Config) (contract.Action, error) {
	return &TaskCreateClickupAction{}, nil
}

func (a *TaskCreateClickupAction) ProcessAction(delivery amqp.Delivery) error {

	//var tenantTag, sourceTag string
	//var inputBody interface{}
	//
	//routingKeyParts := regexp.MustCompile(`:`).Split(delivery.RoutingKey, 3)
	//if len(routingKeyParts) < 2 {
	//	return fmt.Errorf("routing key can't be parsed")
	//}
	//tenantTag = routingKeyParts[1]
	//sourceTag = routingKeyParts[2]
	//
	//logrus.Debug("Message processing t:", tenantTag, " tag:", sourceTag)
	//
	//err := json.Unmarshal(delivery.Body, &inputBody)
	//if err != nil {
	//	logrus.Error("Can't unmarshall rbt body", err)
	//	return err
	//}
	//
	//rules, err := (&repository.ActionWorkflowRuleRepository{}).GetByTenantAndSource(tenantTag, sourceTag)
	//if err != nil {
	//	return err
	//}
	//
	//doableList := rules.FilterDoable(inputBody)
	//if len(doableList) == 0 {
	//	logrus.Debug("Nothing to execute")
	//}
	//
	//doableList.Run(inputBody, l.publisher)

	return nil
}
