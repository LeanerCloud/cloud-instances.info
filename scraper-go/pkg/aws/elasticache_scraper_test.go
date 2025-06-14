package aws

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestElastiCacheScraper(t *testing.T) {
	ctx := context.Background()
	scraper, err := NewElastiCacheScraper(ctx)
	require.NoError(t, err)

	t.Run("GetServiceName", func(t *testing.T) {
		assert.Equal(t, "ElastiCache", scraper.GetServiceName())
	})

	t.Run("LocationToRegion", func(t *testing.T) {
		testCases := []struct {
			location string
			expected string
		}{
			{"US East (N. Virginia)", "us-east-1"},
			{"US East (Ohio)", "us-east-2"},
			{"Europe (Ireland)", "eu-west-1"},
			{"Asia Pacific (Tokyo)", "ap-northeast-1"},
			{"Unknown Location", ""},
		}

		for _, tc := range testCases {
			result := scraper.locationToRegion(tc.location)
			assert.Equal(t, tc.expected, result, "Region for location: %s", tc.location)
		}
	})

	// Note: Testing private methods is not possible, but these are covered by integration tests
}