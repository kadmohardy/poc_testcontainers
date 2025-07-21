package poc_testcontainers_test

import (
	"context"
	"fmt"
	"net/http"
	"poc_testcontainers"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
)

func TestIntegrationNginxLatestReturn(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()

	nginxC, err := poc_testcontainers.StartContainer(ctx)
	testcontainers.CleanupContainer(t, nginxC)
	require.NoError(t, err)
	print("____________________________")
	fmt.Print(nginxC)
	fmt.Print(nginxC.URI)
	print("____________________________")
	resp, err := http.Get(nginxC.URI)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
}
