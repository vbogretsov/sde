package track

import (
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/uber/h3-go/v4"
)

const (
	pointTypeCluster  = iota
	pointTypeLocation = iota
)

var validate = validator.New()

type ListDTO[T any] struct {
	Items []T `json:"items"`
}

type PointDTO struct {
	ID        string    `json:"id"`
	Location  []float32 `json:"loc"`
	UpdatedAt time.Time `json:"ts"`
	Type      int       `json:"type"`
	Vlaue     any       `json:"value"`
}

type FilterDTO struct {
	LatMin float32 `json:"latMin" query:"latMin" validate:"gte=-90,lte=90"`
	LatMax float32 `json:"latMax" query:"latMax" validate:"gte=-90,lte=90"`
	LngMin float32 `json:"lngMin" query:"lngMin" validate:"gte=-180,lte=180"`
	LngMax float32 `json:"lngMax" query:"lngMax" validate:"gte=-180,lte=180"`
	Zoom   int     `json:"zoom" query:"zoom" validate:"gte=1,lte=15"`
}

type LocationDTO struct {
	CorrID    string    `avro:"_cid"`
	UserID    string    `json:"uid" validate:"required" avro:"uid"`
	Location  []float32 `json:"loc" validate:"required,len=2,dive,required" avro:"loc"`
	UpdatedAt time.Time `avro:"updated_at"`
}

func (l *LocationDTO) Validate() map[string]string {
	errs := make(map[string]string)

	if err := validate.Struct(l); err != nil {
		for _, e := range err.(validator.ValidationErrors) {
			errs[e.Field()] = fmt.Sprintf("failed %s validation", e.Tag())
		}
	}

	lat, lng := l.Location[0], l.Location[1]
	if lat < -90 || lat > 90 {
		errs["loc"] = "latitude must be between -90 and 90"
	}
	if lng < -180 || lng > 180 {
		errs["loc"] = "longitude must be between -180 and 180"
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}

func h3index(latlng h3.LatLng, resolution int) int64 {
	c, _ := h3.LatLngToCell(latlng, resolution)
	// NOTE: We ignore error because we did validation before
	return int64(c)
}

func (l *LocationDTO) toModel() LocationModel {
	latlng := h3.LatLng{Lat: float64(l.Location[0]), Lng: float64(l.Location[1])}

	return LocationModel{
		CorrID:    l.CorrID,
		UserID:    l.UserID,
		Lat:       l.Location[0],
		Lng:       l.Location[1],
		UpdatedAt: l.UpdatedAt,
		H3_1:      h3index(latlng, 1),
		H3_2:      h3index(latlng, 2),
		H3_3:      h3index(latlng, 4),
		H3_4:      h3index(latlng, 4),
		H3_5:      h3index(latlng, 5),
		H3_6:      h3index(latlng, 6),
		H3_7:      h3index(latlng, 7),
		H3_8:      h3index(latlng, 8),
		H3_9:      h3index(latlng, 9),
	}
}

type RouteDTO struct {
	CorrID    string      `bson:"_cid"`
	UserID    string      `json:"uid" validate:"required" bson:"_id"`
	Route     [][]float32 `json:"route" validate:"required,min=2,max=10000,dive,required" bson:"points"`
	UpdatedAt time.Time   `bson:"updated_at"`
}

func (r *RouteDTO) Validate() map[string]string {
	errs := make(map[string]string)

	if err := validate.Struct(r); err != nil {
		for _, e := range err.(validator.ValidationErrors) {
			errs[e.Field()] = fmt.Sprintf("failed '%s' validation", e.Tag())
		}
	}

	for i, point := range r.Route {
		if len(point) != 2 {
			errs[fmt.Sprintf("route[%d]", i)] = "each point must contain exactly 2 coordinates"
			continue
		}
		lat := point[0]
		lng := point[1]

		if lat < -90 || lat > 90 {
			errs[fmt.Sprintf("route[%d]", i)] = "latitude must be between -90 and 90"
		}

		if lng < -180 || lng > 180 {
			key := fmt.Sprintf("route[%d]", i)
			if _, exists := errs[key]; !exists {
				errs[key] = "longitude must be between -180 and 180"
			}
		}
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}
