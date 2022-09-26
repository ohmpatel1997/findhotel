package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsIpv4Regex(t *testing.T) {
	cases := []struct {
		Name         string
		ExpectedResp bool
		Req          string
	}{
		{
			Name:         "valid",
			ExpectedResp: true,
			Req:          "192.0.2.146",
		},
		{
			Name:         "invalid due to space",
			ExpectedResp: false,
			Req:          "192. 0.2.146",
		},
		{
			Name:         "invalid due to wrong format",
			ExpectedResp: false,
			Req:          "192.0.2.146.123",
		},
	}

	for _, tt := range cases {
		tt := tt
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()
			assert := assert.New(t)
			assert.Equal(tt.ExpectedResp, IsIpv4Regex(tt.Req))
		})
	}
}
