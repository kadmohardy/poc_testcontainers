package nginx

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
)

func TestIntegrationNginxLatestReturn(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()

	nginxC, err := startContainer(ctx)
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
