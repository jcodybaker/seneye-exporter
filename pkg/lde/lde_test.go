package lde

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestFromRequestBody(t *testing.T) {
	goodMsg := []byte("eyJhbGciOiJIUzI1NiJ9.eyJ2ZXJzaW9uIjoiMS4wLjAiLCJTVUQiOnsiaWQiOiJBQUFBQUJCQkJCQ0NDQ0NEREREREVFRUVFRkZGRkYwMCIsIm5hbWUiOiJPZmZpY2UiLCJ0eXBlIjoxLCJUUyI6MTYwOTU2MTIyMiwiZGF0YSI6eyJTIjp7IlciOjEsIlQiOjAsIlAiOjAsIk4iOjAsIlMiOjB9LCJUIjoyMS4xMjUsIlAiOjcuOTQsIk4iOjAuMDAxfX19.PHNLWLYWY7t4c9MYY--KPGxoHkIFgapk0gnPa9wJH2Q")
	goodSecret := []byte("AAAAAAAA")

	expected := &LDE{
		Version: "1.0.0",
		SUD: SUD{
			ID:        "AAAAABBBBBCCCCCDDDDDEEEEEFFFFF00",
			Name:      "Office",
			Type:      HomeSUD,
			Timestamp: 1609561222,
			Data: Data{
				Status: SUDStatus{
					Water:       1,
					Temperature: 0,
					PH:          0,
					NH3:         0,
					Slide:       0,
					Kelvin:      0,
				},
				Temperature: 21.125,
				PH:          7.94,
				NH3:         0.001,
				Kelvin:      0,
				Lux:         0,
				PAR:         0,
			},
		},
	}

	tcs := []struct {
		name      string
		msg       []byte
		secrets   map[string][]byte
		expectErr string
	}{
		{
			name: "happy path w/ default secret",
			msg:  goodMsg,
			secrets: map[string][]byte{
				"": goodSecret,
			},
		},
		{
			name: "happy path w/ device id secret",
			msg:  goodMsg,
			secrets: map[string][]byte{
				"AAAAABBBBBCCCCCDDDDDEEEEEFFFFF00": goodSecret,
			},
		},
		{
			name:      "secret not found",
			msg:       goodMsg,
			secrets:   nil,
			expectErr: `Unknown Seneye device ID: "AAAAABBBBBCCCCCDDDDDEEEEEFFFFF00"`,
		},
		{
			name: "bad secret",
			msg:  goodMsg,
			secrets: map[string][]byte{
				"": []byte("BADBADBA"),
			},
			expectErr: `signature is invalid`,
		},
		{
			name: "corrupt message",
			msg:  []byte("eyJhbGciOiJIUzI1NiJ9.eyJ2ZXJzaW9uIjoiMS4wLjAiLCJTVUQiOnsiaWQiOiJBQUFBQUJCQkJCQ0NDQ0NEREREREVFRUVFRkZGRkYwMCIsIm5hbWUiOiJOb3BlIiwidHlwZSI6MSwiVFMiOjE2MDk1NjEyMjIsImRhdGEiOnsiUyI6eyJXIjoxLCJUIjowLCJQIjowLCJOIjowLCJTIjowfSwiVCI6MjEuMTI1LCJQIjo3Ljk0LCJOIjowLjAwMX19fQ.PHNLWLYWY7t4c9MYY--KPGxoHkIFgapk0gnPa9wJH2Q"),
			secrets: map[string][]byte{
				"": goodSecret,
			},
			expectErr: `signature is invalid`,
		},
	}
	for _, tc := range tcs {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			lde, err := FromRequestBody(tc.msg, tc.secrets)
			if tc.expectErr == "" {
				require.NoError(t, err)
				assert.Equal(t, expected, lde)
				return
			}
			assert.EqualError(t, err, tc.expectErr)
		})
	}

}
