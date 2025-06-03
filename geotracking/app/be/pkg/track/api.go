package track

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/segmentio/kafka-go"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"app/pkg/config"
)

type API struct {
	producer  *kafka.Writer
	routes    *mongo.Collection
	locations *mongo.Collection
}

func New(
	cfg config.Config,
	producer *kafka.Writer,
	mongocli *mongo.Client,
) *API {
	return &API{
		producer:  producer,
		routes:    mongocli.Database(cfg.Mongo.Database).Collection(CollectionRoutes),
		locations: mongocli.Database(cfg.Mongo.Database).Collection(CollectionLocations),
	}
}

func (t *API) LocationPost(c echo.Context) error {
	var dto LocationDTO
	if err := c.Bind(&dto); err != nil {
		return err
	}

	if err := c.Validate(&dto); err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	dto.CorrID = c.Response().Header().Get(echo.HeaderXRequestID)
	slog.Debug("Received location", "payload", dto)

	if err := onReceive(c.Request().Context(), dto, t.producer); err != nil {
		return fmt.Errorf("Failed to produce location: %w", err)
	}

	return c.NoContent(http.StatusNoContent)
}

func (t *API) RoutePost(c echo.Context) error {
	var dto RouteDTO
	if err := c.Bind(&dto); err != nil {
		return err
	}

	if err := c.Validate(&dto); err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	dto.CorrID = c.Response().Header().Get(echo.HeaderXRequestID)
	slog.Debug("Received route", "payload", dto)

	if err := setRoute(c.Request().Context(), &dto, t.routes); err != nil {
		return fmt.Errorf("Failed to update route: %w", err)
	}

	return c.NoContent(http.StatusNoContent)
}

func (t *API) CourierList(c echo.Context) error {
	dto := FilterDTO{
		LatMin: -90,
		LatMax: 90,
		LngMin: -180,
		LngMax: 180,
		Zoom:   1,
	}

	if err := c.Bind(&dto); err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	if err := c.Validate(&dto); err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	res, err := search(c.Request().Context(), t.locations, dto)
	if err != nil {
		return fmt.Errorf("Failed to search: %w", err)
	}

	return c.JSON(http.StatusOK, ListDTO[PointDTO]{Items: res})
}
