package main

import (
	"github.com/ThanvirXo/jira-auto-bug-solver/handlers"
	"github.com/ThanvirXo/jira-auto-bug-solver/middlewares"
	"github.com/ThanvirXo/jira-auto-bug-solver/services"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)


type App struct {
	Middlewares *middlewares.Middleware
	Router      *gin.Engine
	Handler     *handlers.Handler
}

func NewApp() *App{
	r:=gin.New()

	middleware:=middlewares.Middleware{}

	service:=services.Service{}

	handler:=handlers.Handler{
		Services: &service,
	}
	return &App{
		Router: r,
		Handler: &handler,
		Middlewares: &middleware,
	}
}


func (a *App) SetupMiddleware() *App{
	a.Router.Use(a.Middlewares.RequestMiddleware())
	a.Router.Use(gin.Recovery())
	a.Router.Use(gin.Logger())
	return a
}

func (a *App) Listen() error {
	logrus.Infof("Starting server on port %s", "8000")
	return a.Router.Run(":8000")
}

func (a *App) SetupRoutes() *App{
	router:=a.Router

	handler:=a.Handler
	jiraAuth:=router.Group("/jira")
	jiraAuth.Use(a.Middlewares.AuthMiddleware())
	{
		jiraAuth.POST("/webhook", handler.JiraWebhook)
	}
	router.GET("/health", handler.HealthCheck)
	return a
}