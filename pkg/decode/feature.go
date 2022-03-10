package decode

import (
	"github.com/triarius/go-geobuf/pkg/geojson"
	"github.com/triarius/go-geobuf/pkg/geometry"
	"github.com/triarius/go-geobuf/proto"
)

func DecodeFeature(msg *proto.Data, feature *proto.Data_Feature, precision, dimension uint32) *geojson.Feature {
	geo := feature.Geometry
	decodedGeo := DecodeGeometry(geo, msg.Precision, msg.Dimensions)
	var geoFeature *geojson.Feature
	switch decodedGeo.Type {
	case geojson.GeometryCollectionType:
		collection := make(geometry.Collection, len(decodedGeo.Geometries))
		for i, child := range decodedGeo.Geometries {
			collection[i] = child.Coordinates
		}
		geoFeature = geojson.NewFeature(collection)
	default:
		geoFeature = geojson.NewFeature(decodedGeo.Coordinates)
	}

	for i := 0; i < len(feature.Properties); i = i + 2 {
		keyIdx := feature.Properties[i]
		valIdx := feature.Properties[i+1]
		val := feature.Values[valIdx]
		switch actualVal := val.ValueType.(type) {
		case *proto.Data_Value_BoolValue:
			geoFeature.Properties[msg.Keys[keyIdx]] = actualVal.BoolValue
		case *proto.Data_Value_DoubleValue:
			geoFeature.Properties[msg.Keys[keyIdx]] = actualVal.DoubleValue
		case *proto.Data_Value_StringValue:
			geoFeature.Properties[msg.Keys[keyIdx]] = actualVal.StringValue
		case *proto.Data_Value_PosIntValue:
			geoFeature.Properties[msg.Keys[keyIdx]] = uint(actualVal.PosIntValue)
		case *proto.Data_Value_NegIntValue:
			geoFeature.Properties[msg.Keys[keyIdx]] = int(actualVal.NegIntValue) * -1
		case *proto.Data_Value_JsonValue:
			geoFeature.Properties[msg.Keys[keyIdx]] = actualVal.JsonValue
		}
	}
	switch id := feature.IdType.(type) {
	case *proto.Data_Feature_Id:
		geoFeature.ID = id.Id
	case *proto.Data_Feature_IntId:
		geoFeature.ID = id.IntId
	}
	return geoFeature
}
