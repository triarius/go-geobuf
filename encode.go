package geobuf

import (
	"math"

	"github.com/cairnapp/go-geobuf/proto"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geojson"
)

const MaxPrecision = 100000000

func Encode(obj interface{}) *proto.Data {
	builder := newBuilder()
	builder.Analyze(obj)
	return builder.Build(obj)
}

type protoBuilder struct {
	keys      []string
	precision uint32
	dimension uint32
}

func newBuilder() *protoBuilder {
	pb := &protoBuilder{
		keys:      []string{},
		precision: 1,
		// Since Orb forces us into a 2 dimensional point, we'll have to use other ways to encode elevation + time
		// int(math.Max(float64(b.dimension), float64(len(point))))
		dimension: 2,
	}
	return pb
}

func (b *protoBuilder) Build(obj interface{}) *proto.Data {
	precision := math.Ceil(math.Log(float64(b.precision)) / math.Ln10)
	pbf := proto.Data{
		Keys:       b.keys,
		Dimensions: b.dimension,
		Precision:  uint32(precision),
	}

	switch t := obj.(type) {
	case geojson.MultiPoint:
		pbf.DataType = &proto.Data_Geometry_{
			Geometry: &proto.Data_Geometry{
				Type:   proto.Data_Geometry_MULTIPOINT,
				Coords: translateLine(b.precision, b.dimension, t, false),
			},
		}
	case *geojson.Geometry:
		switch t.Type {
		case "Point":
			p := t.Coordinates.(orb.Point)
			pbf.DataType = &proto.Data_Geometry_{
				Geometry: &proto.Data_Geometry{
					Type:   proto.Data_Geometry_POINT,
					Coords: translateCoords(b.precision, p[:]),
				},
			}
			return &pbf
		}
	case geojson.Point:
		pbf.DataType = &proto.Data_Geometry_{
			Geometry: &proto.Data_Geometry{
				Type:   proto.Data_Geometry_POINT,
				Coords: translateCoords(b.precision, t[:]),
			},
		}
	}
	return &pbf
}

func (b *protoBuilder) Analyze(obj interface{}) {
	switch t := obj.(type) {
	case geojson.FeatureCollection:
		for _, feature := range t.Features {
			b.Analyze(feature)
		}
	case geojson.Feature:
		b.Analyze(t.Geometry)
		for key, _ := range t.Properties {
			b.keys = append(b.keys, key)
		}
	case *geojson.Geometry:
		switch t.Type {
		case "Point":
			b.updatePrecision(t.Coordinates.(orb.Point))
		}
	case geojson.Point:
		b.updatePrecision(orb.Point(t))
	case geojson.MultiPoint:
		for _, coord := range t {
			b.updatePrecision(coord)
		}
	// case geojson.GeometryCollection:
	case geojson.LineString:
		for _, coord := range t {
			b.updatePrecision(coord)
		}
	case geojson.Polygon:
		for _, line := range t {
			for _, coord := range line {
				b.updatePrecision(coord)
			}
		}
	case geojson.MultiPolygon:
		for _, polygon := range t {
			for _, line := range polygon {
				for _, coord := range line {
					b.updatePrecision(coord)
				}
			}
		}
	}
}

func (b *protoBuilder) updatePrecision(point orb.Point) {
	e := getPrecision([2]float64(point))
	if e > b.precision {
		b.precision = e
	}
}

func translateLine(e uint32, dim uint32, points geojson.MultiPoint, isClosed bool) []int64 {
	sums := make([]int64, dim)
	ret := make([]int64, len([]orb.Point(points))*int(dim))
	// TODO: Skip last one if isClosed
	for i, point := range points {
		for j, p := range point {
			n := doTheMaths(e, p) - sums[j]
			ret[(int(dim)*i)+j] = n
			sums[j] = sums[j] + n
		}
	}
	if isClosed {
		return ret[:len(ret)-1]
	}
	return ret
}

func translateCoords(e uint32, point []float64) []int64 {
	ret := make([]int64, len(point))
	for i, p := range point {
		ret[i] = doTheMaths(e, p)
	}
	return ret
}

func doTheMaths(e uint32, p float64) int64 {
	return int64(math.Round(p * float64(e)))
}

func getPrecision(point [2]float64) uint32 {
	var e uint32 = 1
	for _, val := range point {
		for {
			base := math.Round(float64(val * float64(e)))
			if (base/float64(e)) != val && float64(e) < MaxPrecision {
				e = e * 10
			} else {
				break
			}
		}
	}
	return e
}
