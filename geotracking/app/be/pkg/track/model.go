package track

import "time"

type LocationModel struct {
	CorrID    string    `bson:"_cid"`
	UserID    string    `bson:"_id"`
	UpdatedAt time.Time `bson:"updated_at"`
	Lat       float32   `bson:"lat"`
	Lng       float32   `bson:"lng"`
	H3_1      int64     `bson:"h3_1"`
	H3_2      int64     `bson:"h3_2"`
	H3_3      int64     `bson:"h3_3"`
	H3_4      int64     `bson:"h3_4"`
	H3_5      int64     `bson:"h3_5"`
	H3_6      int64     `bson:"h3_6"`
	H3_7      int64     `bson:"h3_7"`
	H3_8      int64     `bson:"h3_8"`
	H3_9      int64     `bson:"h3_9"`
}
