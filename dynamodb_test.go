package poc_testcontainers_test

import (
	"context"
	"poc_testcontainers"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/stretchr/testify/require"
)

func TestSetupTestContainer(t *testing.T) {
	ctx := context.Background()

	container := poc_testcontainers.SetupTestContainer(ctx)
	require.NotNil(t, container, "container deve ser inicializado")
	require.NotNil(t, container.DynamoClient, "cliente do dynamodb não pode ser nil")

	// Verifica se uma tabela existe, por exemplo "dev_pix"
	output, err := container.DynamoClient.DescribeTable(ctx, &dynamodb.DescribeTableInput{
		TableName: ptr("dev_pix"),
	})
	require.NoError(t, err, "esperado sucesso ao descrever tabela")
	require.Equal(t, "dev_pix", *output.Table.TableName)
}

func ptr[T any](v T) *T {
	return &v
}
