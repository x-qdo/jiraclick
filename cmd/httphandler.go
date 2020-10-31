package cmd

import (
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"x-qdo/jiraclick/pkg/config"
)

func NewHttpHandlerCmd(cfg *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "http-handler",
		Short: "Runs HTTP handler",
		Long:  `Runs HTTP server to handle API requests.`,
		Run: func(cmd *cobra.Command, args []string) {
			var router = gin.New()

			if cfg.Debug {
				gin.SetMode(gin.DebugMode)
			}

			router.Use(gin.LoggerWithWriter(gin.DefaultWriter, "/health-check"))
			router.Use(gin.Recovery())
			router.GET("/health-check", func(c *gin.Context) {
				c.JSON(200, gin.H{
					"message": "ok",
				})
			})

			//APIRouter := router.Group("/api")
			//{
			//	APIRouter.Use(middleware.Authentication())
			//	var customersController, addressesController, documentsController, consentsController contract.ApiSource
			//
			//	amqpProvider, err := amqpwrapper.NewRabbitChannel(cfg.RabbitMQ.URL)
			//	if err != nil {
			//		panic(err)
			//	}
			//	err = amqpProvider.DeclareExchange(contract.CustomersExchange)
			//	if err != nil {
			//		panic(err)
			//	}
			//	customerPublisher := publisher.NewCustomerPublisher(amqpProvider)
			//	redisProvider, err := provider.NewRedisClient(cfg)
			//	if err != nil {
			//		panic(err)
			//	}
			//
			//	requestsController := controller.NewRequestsController(redisProvider)
			//	APIRouter.GET("/requests/:request_id", requestsController.Get)
			//}

			go func() {
				if err := router.Run(":" + cfg.HttpHandler.Port); err != nil {
					panic(err)
				}
			}()
		},
	}
}
