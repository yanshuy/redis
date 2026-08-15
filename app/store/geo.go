package store

import (
	"fmt"
	"math"
)

const (
	MIN_LATITUDE  = -85.05112878
	MAX_LATITUDE  = 85.05112878
	MIN_LONGITUDE = -180.0
	MAX_LONGITUDE = 180.0

	LATITUDE_RANGE  = MAX_LATITUDE - MIN_LATITUDE
	LONGITUDE_RANGE = MAX_LONGITUDE - MIN_LONGITUDE
)

func spreadInt32ToInt64(v uint32) uint64 {
	result := uint64(v)
	result = (result | (result << 16)) & 0x0000FFFF0000FFFF
	result = (result | (result << 8)) & 0x00FF00FF00FF00FF
	result = (result | (result << 4)) & 0x0F0F0F0F0F0F0F0F
	result = (result | (result << 2)) & 0x3333333333333333
	result = (result | (result << 1)) & 0x5555555555555555
	return result
}

func interleave(x, y uint32) uint64 {
	xSpread := spreadInt32ToInt64(x)
	ySpread := spreadInt32ToInt64(y)

	return xSpread | (ySpread << 1)
}

func encode(longitude, latitude float64) uint64 {
	normalizedLatitude := math.Pow(2, 26) * (latitude - MIN_LATITUDE) / LATITUDE_RANGE
	normalizedLongitude := math.Pow(2, 26) * (longitude - MIN_LONGITUDE) / LONGITUDE_RANGE

	latInt := uint32(normalizedLatitude)
	lonInt := uint32(normalizedLongitude)

	return interleave(latInt, lonInt)
}

func compactInt64ToInt32(v uint64) uint32 {
	result := v
	result = (result | (result >> 1)) & 0x3333333333333333
	result = (result | (result >> 2)) & 0x0F0F0F0F0F0F0F0F
	result = (result | (result >> 4)) & 0x00FF00FF00FF00FF
	result = (result | (result >> 8)) & 0x0000FFFF0000FFFF
	result = (result | (result >> 16)) & 0x00000000FFFFFFFF
	return uint32(result)
}

func convertGridNumbersToCoordinates(gridLatitudeNumber, gridLongitudeNumber uint32) Coordinates {
	gridLatitudeMin := MIN_LATITUDE + LATITUDE_RANGE*(float64(gridLatitudeNumber)/math.Pow(2, 26))
	gridLatitudeMax := MIN_LATITUDE + LATITUDE_RANGE*(float64(gridLatitudeNumber+1)/math.Pow(2, 26))
	gridLongitudeMin := MIN_LONGITUDE + LONGITUDE_RANGE*(float64(gridLongitudeNumber)/math.Pow(2, 26))
	gridLongitudeMax := MIN_LONGITUDE + LONGITUDE_RANGE*(float64(gridLongitudeNumber+1)/math.Pow(2, 26))

	latitude := (gridLatitudeMin + gridLatitudeMax) / 2
	longitude := (gridLongitudeMin + gridLongitudeMax) / 2

	return Coordinates{Latitude: latitude, Longitude: longitude}
}

func decode(geoCode uint64) Coordinates {
	x := geoCode & 0x5555555555555555
	y := (geoCode >> 1) & 0x5555555555555555

	gridLatitudeNumber := compactInt64ToInt32(x)
	gridLongitudeNumber := compactInt64ToInt32(y)

	return convertGridNumbersToCoordinates(gridLatitudeNumber, gridLongitudeNumber)
}

type Coordinates struct {
	Longitude float64
	Latitude  float64
}

func inRange(long float64, lat float64) bool {
	return (long >= -180 && long <= 180) && (lat >= -85.05112878 && lat <= 85.05112878)
}

func (rs *RedisStore) Geoadd(key string, long float64, lat float64, member string) (int, error) {
	val, ok := rs.Look(key)
	if !ok {
		val = NewValue(ZSET, 0)
		rs.Store[key] = val
	}
	zset, err := As[Zset](val)
	if err != nil {
		return 0, err
	}

	if !inRange(long, lat) {
		return 0, fmt.Errorf("invalid longitude,latitude pair %f, %f", long, lat)
	}

	score := encode(long, lat)
	inserts := 0
	if zset.insert(Z{member: member, score: float64(score)}) {
		inserts++
	}

	rs.TouchWatchedKey(key)
	return inserts, nil
}

func (rs *RedisStore) Geopos(key string, members []string) ([]*Coordinates, error) {
	val, ok := rs.Look(key)
	if !ok {
		return make([]*Coordinates, len(members)), nil
	}

	zset, err := As[Zset](val)
	if err != nil {
		return nil, err
	}

	answers := make([]*Coordinates, 0, len(members))

	for _, member := range members {
		z, ok := zset.dict[member]
		if ok {
			geoCode := uint64(z.score)
			cords := decode(geoCode)
			if inRange(cords.Longitude, cords.Latitude) {
				answers = append(answers, &cords)
			} else {
				answers = append(answers, &Coordinates{})
			}
		} else {
			answers = append(answers, nil)
		}
	}

	return answers, nil
}

const EARTH_RADIUS_IN_METERS = 6372797.560856

var TO_RAD = math.Pi / 180.0

func haversineDist(c1 Coordinates, c2 Coordinates) float64 {
	lat1 := c1.Latitude * TO_RAD
	lon1 := c1.Longitude * TO_RAD
	lat2 := c2.Latitude * TO_RAD
	lon2 := c2.Longitude * TO_RAD

	dlat := lat2 - lat1
	dlon := lon2 - lon1

	u := math.Sin(dlat / 2)
	v := math.Sin(dlon / 2)
	a := u*u + math.Cos(lat1)*math.Cos(lat2)*v*v

	return 2 * EARTH_RADIUS_IN_METERS * math.Asin(math.Sqrt(a))
}

func (rs *RedisStore) Geodist(key string, place1 string, place2 string) (float64, error) {
	val, ok := rs.Look(key)
	if !ok {
		return 0, fmt.Errorf("Operation against a key holding the wrong kind of value")
	}

	zset, err := As[Zset](val)
	if err != nil {
		return 0, err
	}

	z1, ok1 := zset.dict[place1]
	z2, ok2 := zset.dict[place2]
	if !ok1 || !ok2 {
		return -1, fmt.Errorf("Operation against a key holding the wrong kind of value")
	}

	c1 := decode(uint64(z1.score))
	c2 := decode(uint64(z2.score))

	return haversineDist(c1, c2), nil
}
