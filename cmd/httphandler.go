package cmd

import (
	"github.com/astreter/amqpwrapper/v2"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"x-qdo/jiraclick/pkg/contract"

	"x-qdo/jiraclick/pkg/config"
	"x-qdo/jiraclick/pkg/handler"
	"x-qdo/jiraclick/pkg/provider/clickup"
)

func NewHTTPHandlerCmd(
	cfg *config.Config,
	logger *logrus.Logger,
	queue *amqpwrapper.RabbitChannel,
	clickup *clickup.ConnectorPool,
	db contract.Storage,
) *cobra.Command {
	return &cobra.Command{
		Use:   "http-handler",
		Short: "Runs HTTP handler",
		Long:  `Runs HTTP server to handle API requests and webhooks`,
		Run: func(cmd *cobra.Command, args []string) {
			var router = gin.New()

			if cfg.Debug {
				gin.SetMode(gin.DebugMode)
			}

			clickUpHandler, err := handler.NewClickUpWebhooksHandler(cfg, logger, queue, clickup, db)
			if err != nil {
				panic(err)
			}

			router.Use(gin.LoggerWithWriter(gin.DefaultWriter, "/health-check"))
			router.Use(gin.Recovery())
			router.GET("/health-check", func(c *gin.Context) {
				c.JSON(200, gin.H{
					"message": "ok",
				})
			})
			router.Use(otelgin.Middleware(config.ServiceName))

			router.POST("webhooks/clickup", clickUpHandler.TaskEvent)

			go func() {
				if err := router.Run(":" + cfg.HTTPHandler.Port); err != nil {
					panic(err)
				}
			}()
		},
	}
}
