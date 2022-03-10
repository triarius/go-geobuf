package encode

import (
	"github.com/triarius/go-geobuf/pkg/geojson"
	proto "go.buf.build/grpc/go/qwant/geobuf/geobufproto"
)

func EncodeFeatureCollection(collection *geojson.FeatureCollection, opts *EncodingConfig) (*proto.Data_FeatureCollection, error) {
	features := make([]*proto.Data_Feature, len(collection.Features))

	for i, feature := range collection.Features {
		encoded, err := EncodeFeature(feature, opts)
		if err != nil {
			return nil, err
		}
		features[i] = encoded
	}

	return &proto.Data_FeatureCollection{
		Features: features,
	}, nil
}
