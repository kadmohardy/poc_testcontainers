package poc_testcontainers

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// DynamodbLocalContainer represents the a DynamoDB Local container - https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/DynamoDBLocal.html
type DynamodbLocalContainer struct {
	Container    testcontainers.Container
	DynamoClient *dynamodb.Client
}

var (
	onceContainer sync.Once
	testContainer *DynamodbLocalContainer
)

const (
	image         = "amazon/dynamodb-local:2.2.1"
	port          = nat.Port("8000/tcp")
	containerName = "dynamodb_local"
)

func SetupTestContainer(ctx context.Context) *DynamodbLocalContainer {
	onceContainer.Do(func() {
		var err error
		testContainer, err = RunContainer(ctx)
		if err != nil {
			panic("erro ao iniciar container: " + err.Error())
		}
	})
	return testContainer
}

// RunContainer creates an instance of the dynamodb container type
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*DynamodbLocalContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        image,
		ExposedPorts: []string{string(port)},
		WaitingFor:   wait.ForLog("Initializing DynamoDB Local").WithStartupTimeout(30 * time.Second),
		HostConfigModifier: func(config *container.HostConfig) {
			config.AutoRemove = true
		},

		Name: fmt.Sprintf("%s-%d", containerName, time.Now().UnixNano()),
	}

	genericContainerReq := testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	}

	for _, opt := range opts {
		if err := opt.Customize(&genericContainerReq); err != nil {
			return nil, fmt.Errorf("failed to apply container option: %w", err)
		}
	}

	container, err := testcontainers.GenericContainer(ctx, genericContainerReq)
	if err != nil {
		return nil, err
	}

	localContainer := &DynamodbLocalContainer{Container: container}
	client, err := localContainer.GetDynamoDBClient(ctx)
	if err != nil {

		return nil, err
	}
	print("TESTANDO TUDO AGORA_______________")
	fmt.Print(client)
	err = SetupTables(client)
	if err != nil {
		return nil, err
	}

	localContainer.DynamoClient = client
	return localContainer, nil
}

func (c *DynamodbLocalContainer) GetDynamoDBClient(ctx context.Context) (*dynamodb.Client, error) {
	hostAndPort, err := c.ConnectionString(ctx)
	if err != nil {
		return nil, err
	}
	fmt.Println("HOST AND PORT:", hostAndPort)
	cfg, err := config.LoadDefaultConfig(ctx, config.WithCredentialsProvider(credentials.StaticCredentialsProvider{
		Value: aws.Credentials{
			AccessKeyID:     "DUMMYIDEXAMPLE",
			SecretAccessKey: "DUMMYEXAMPLEKEY",
		},
	}))
	if err != nil {
		return nil, err
	}

	return dynamodb.NewFromConfig(cfg, dynamodb.WithEndpointResolverV2(&DynamoDBLocalResolver{hostAndPort: hostAndPort})), nil
}

// ConnectionString returns DynamoDB local endpoint host and port in <host>:<port> format
func (c *DynamodbLocalContainer) ConnectionString(ctx context.Context) (string, error) {
	mappedPort, err := c.Container.MappedPort(ctx, port)
	if err != nil {
		return "", err
	}
	fmt.Println("PORT:", mappedPort.Port())
	hostIP, err := c.Container.Host(ctx)
	if err != nil {
		return "", err
	}
	fmt.Println("HOST:", hostIP)

	uri := fmt.Sprintf("%s:%s", hostIP, mappedPort.Port())
	fmt.Println("URL DO ENDPOINT:", uri)
	return uri, nil
}

func SetupTables(c *dynamodb.Client) error {
	tableNames := []string{"dev_pix", "dev_ledger", "dev_idempotency"}

	for _, tableName := range tableNames {
		err := CreateTable(c, tableName)
		if err != nil {
			return fmt.Errorf("falha ao criar a tabela %s: %w", tableName, err)
		}
	}

	return nil
}

func CreateTable(c *dynamodb.Client, tableName string) error {
	print("CREATING NEW TABLE_________")
	fmt.Print(c)
	_, err := c.CreateTable(context.TODO(), &dynamodb.CreateTableInput{
		TableName:   aws.String(tableName),
		BillingMode: types.BillingModePayPerRequest,
		AttributeDefinitions: []types.AttributeDefinition{
			{AttributeName: aws.String("PK"), AttributeType: types.ScalarAttributeTypeS},
			{AttributeName: aws.String("SK"), AttributeType: types.ScalarAttributeTypeS},
			{AttributeName: aws.String("GSI1PK"), AttributeType: types.ScalarAttributeTypeS},
			{AttributeName: aws.String("GSI1SK"), AttributeType: types.ScalarAttributeTypeS},
			{AttributeName: aws.String("schema"), AttributeType: types.ScalarAttributeTypeS},
			{AttributeName: aws.String("schema_version"), AttributeType: types.ScalarAttributeTypeN},
		},
		KeySchema: []types.KeySchemaElement{
			{AttributeName: aws.String("PK"), KeyType: types.KeyTypeHash},
			{AttributeName: aws.String("SK"), KeyType: types.KeyTypeRange},
		},
		GlobalSecondaryIndexes: []types.GlobalSecondaryIndex{
			{
				IndexName: aws.String("GSI1"),
				KeySchema: []types.KeySchemaElement{
					{AttributeName: aws.String("GSI1PK"), KeyType: types.KeyTypeHash},
					{AttributeName: aws.String("GSI1SK"), KeyType: types.KeyTypeRange},
				},
				Projection: &types.Projection{
					ProjectionType: types.ProjectionTypeAll,
				},
			},
			{
				IndexName: aws.String("SchemaIndex"),
				KeySchema: []types.KeySchemaElement{
					{AttributeName: aws.String("schema"), KeyType: types.KeyTypeHash},
					{AttributeName: aws.String("schema_version"), KeyType: types.KeyTypeRange},
				},
				Projection: &types.Projection{
					ProjectionType: types.ProjectionTypeAll,
				},
			},
		},
		StreamSpecification: &types.StreamSpecification{
			StreamEnabled:  aws.Bool(true),
			StreamViewType: types.StreamViewTypeNewAndOldImages,
		},
	})
	if err != nil {
		var inUseErr *types.ResourceInUseException
		if errors.As(err, &inUseErr) {
			fmt.Printf("Tabela %s já existe\n", tableName)
			return nil
		}
	}

	return err
}
