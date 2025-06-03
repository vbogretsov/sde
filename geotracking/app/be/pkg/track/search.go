package track

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/uber/h3-go/v4"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

const geoClusterThreshold = 8

type Cluster struct {
	ID    int64     `bson:"_id"`
	Count int       `bson:"count"`
	MinTS time.Time `bson:"mints"`
}

func search(ctx context.Context, coll *mongo.Collection, filter FilterDTO) ([]PointDTO, error) {
	l := slog.With(slog.Any("filter", filter))
	l.Debug("Performing search")

	if filter.Zoom > geoClusterThreshold {
		res, err := searchLocations(ctx, coll, filter)
		if err != nil {
			l.Error("Failed to search locations", slog.Any("err", err))
			return nil, err
		}
		return mapSlice(res, locationToPoint), nil
	}

	res, err := searchClusters(ctx, coll, filter)
	if err != nil {
		slog.Error("Failed to search clusters", slog.Any("err", err))
		return nil, err
	}

	return mapSlice(res, clusterToPoint), nil
}

func mapSlice[I any, O any](s []I, f func(I) O) []O {
	o := make([]O, len(s))
	for i, v := range s {
		o[i] = f(v)
	}
	return o
}

func clusterToPoint(c Cluster) PointDTO {
	latlng, err := h3.Cell(c.ID).LatLng()
	if err != nil {
		slog.Error("Failed to get coordinates of H3 cell", "cell", c.ID)
	}

	return PointDTO{
		ID:        strconv.Itoa(int(c.ID)),
		UpdatedAt: c.MinTS,
		Location:  []float32{float32(latlng.Lat), float32(latlng.Lng)},
		Type:      pointTypeCluster,
		Vlaue:     c.Count,
	}
}

func locationToPoint(l LocationModel) PointDTO {
	return PointDTO{
		ID:        l.UserID,
		UpdatedAt: l.UpdatedAt,
		Location:  []float32{l.Lat, l.Lng},
		Type:      pointTypeLocation,
		Vlaue:     nil,
	}
}

func searchClusters(ctx context.Context, coll *mongo.Collection, filter FilterDTO) ([]Cluster, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"lat": bson.M{"$gte": filter.LatMin, "$lte": filter.LatMax},
			"lng": bson.M{"$gte": filter.LngMin, "$lte": filter.LngMax},
		}}},
		{{Key: "$group", Value: bson.M{
			"_id":   fmt.Sprintf("$h3_%d", filter.Zoom),
			"count": bson.M{"$sum": 1},
			"lat":   bson.M{"$avg": "lat"},
			"lng":   bson.M{"$avg": "lng"},
			"mints": bson.M{"$min": "$timestamp"},
		}}},
	}

	cursor, err := coll.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []Cluster
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	return results, nil
}

func searchLocations(ctx context.Context, coll *mongo.Collection, filter FilterDTO) ([]LocationModel, error) {
	query := bson.M{
		"lat": bson.M{
			"$gte": filter.LatMin,
			"$lte": filter.LatMax,
		},
		"lng": bson.M{
			"$gte": filter.LngMin,
			"$lte": filter.LngMax,
		},
	}

	cursor, err := coll.Find(ctx, query)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []LocationModel
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	return results, nil
}
