package model

// DriverLocation 司机实时位置 (MongoDB driver_place 集合)
type DriverLocation struct {
	DriverId  int64   `bson:"driver_id"`
	Lat       float64 `bson:"lat"`
	Lng       float64 `bson:"lng"`
	Heading   float64 `bson:"heading"`
	Speed     float64 `bson:"speed"`
	Status    int8    `bson:"status"` // 1-空车 2-有客 3-离线
	CityId    int64   `bson:"city_id"`
	UpdatedAt int64   `bson:"updated_at"` // unix timestamp
}
