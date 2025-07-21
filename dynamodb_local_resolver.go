package poc_testcontainers

import (
	"context"
	"net/url"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	smithyendpoints "github.com/aws/smithy-go/endpoints"
)

type DynamoDBLocalResolver struct {
	hostAndPort string
}

func (r *DynamoDBLocalResolver) ResolveEndpoint(ctx context.Context, params dynamodb.EndpointParameters) (endpoint smithyendpoints.Endpoint, err error) {
	u := url.URL{
		Scheme: "http",
		Host:   r.hostAndPort,
	}

	return smithyendpoints.Endpoint{
		URI: u,
	}, nil

}
