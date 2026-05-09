package container

import (
	"example.com/tinder/internal/gateway"
	"example.com/tinder/internal/repository"
	"example.com/tinder/internal/service"
	"log/slog"
	"net/http"
)

type Container struct {
	Logger  *slog.Logger
	Repo    repository.PersonRepository
	Service service.MatchingService
	Gateway *gateway.HTTPGateway
}

func New(logger *slog.Logger) *Container {
	repo := repository.NewInMemoryPersonRepository()
	svc := service.NewMatchingService(repo)
	gw := gateway.NewHTTPGateway(logger, svc)

	return &Container{
		Logger:  logger,
		Repo:    repo,
		Service: svc,
		Gateway: gw,
	}
}

func (c *Container) Handler() http.Handler {
	return c.Gateway.Handler()
}
