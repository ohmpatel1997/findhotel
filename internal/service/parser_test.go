package service

import (
	"fmt"
	"os"
	"testing"

	"github.com/ohmpatel1997/findhotel/internal/common"
	"github.com/ohmpatel1997/findhotel/internal/model"
	"github.com/ohmpatel1997/findhotel/internal/model/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestParseAndStore(t *testing.T) {
	cases := []struct {
		Name         string
		fileName     string
		ValidCount   int64
		InvalidCount int64
	}{
		{
			Name:         "csv1",
			fileName:     "test1.csv",
			ValidCount:   3,
			InvalidCount: 2,
		},
	}

	locationManager := new(mocks.GeoLocationManager)
	locationManager.On("BulkInsert", mock.Anything, mock.Anything).Return(nil)

	for _, tt := range cases {
		tt := tt
		t.Run(tt.Name, func(t *testing.T) {
			assert := assert.New(t)
			t.Parallel()
			f, err := os.Open(fmt.Sprintf("./test_data/%s", tt.fileName))
			if err != nil {
				assert.Fail("error opening file", err)
			}
			_, invalid, valid, err := NewParser(f, locationManager).ParseAndStore()
			if err != nil {
				assert.Fail("error parsing file", err)
			}

			assert.Equal(tt.InvalidCount, invalid)
			assert.Equal(tt.ValidCount, valid)
		})
	}
}

func TestIsValidLine(t *testing.T) {
	cases := []struct {
		Name          string
		visitedIP     map[string]bool
		Text          string
		ExpectedValid bool
		ExpectedResp  *model.Geolocation
	}{
		{
			Name:          "missing ip in line",
			visitedIP:     map[string]bool{},
			Text:          ",SI,Nepal,DuBuquemouth,-84.87503094689836,7.206435933364332,7823011346",
			ExpectedValid: false,
			ExpectedResp:  nil,
		},
		{
			Name:          "extra field",
			visitedIP:     map[string]bool{},
			Text:          "70.95.73.73,SI,Nepal,DuBuquemouth,-84.87503094689836,7.206435933364332,7823011346,70.95.73.73",
			ExpectedValid: false,
			ExpectedResp:  nil,
		},
		{
			Name:          "invalid ip field",
			visitedIP:     map[string]bool{},
			Text:          "7012.95.73.73,SI,Nepal,DuBuquemouth,-84.87503094689836,7.206435933364332,7823011346",
			ExpectedValid: false,
			ExpectedResp:  nil,
		},
		{
			Name: "already visited ip field",
			visitedIP: map[string]bool{
				"70.95.73.73": true,
			},
			Text:          "70.95.73.73,SI,Nepal,DuBuquemouth,-84.87503094689836,7.206435933364332,7823011346",
			ExpectedValid: false,
			ExpectedResp:  nil,
		},
		{
			Name:          "already longitude field",
			visitedIP:     map[string]bool{},
			Text:          "7012.95.73.73,SI,Nepal,DuBuquemouth,-84.87503094689836,1027.206435933364332,7823011346",
			ExpectedValid: false,
			ExpectedResp:  nil,
		},
		{
			Name:          "already latitude field",
			visitedIP:     map[string]bool{},
			Text:          "7012.95.73.73,SI,Nepal,DuBuquemouth,-154.87503094689836,150.206435933364332,7823011346",
			ExpectedValid: false,
			ExpectedResp:  nil,
		},
		{
			Name:          "valid line",
			visitedIP:     map[string]bool{},
			Text:          "70.95.73.73,SI,Nepal,DuBuquemouth,-84.87503094689836,7.206435933364332,7823011346",
			ExpectedValid: true,
			ExpectedResp: &model.Geolocation{
				IP:           "70.95.73.73",
				Country:      "Nepal",
				CountryCode:  "SI",
				City:         "DuBuquemouth",
				Latitude:     "-84.87503094689836",
				Longitude:    "7.206435933364332",
				MysteryValue: "7823011346",
			},
		},
	}

	positions := map[int]string{
		0: common.IP,
		1: common.CountryCode,
		2: common.Country,
		3: common.City,
		4: common.Latitude,
		5: common.Longitude,
		6: common.MysteryValue,
	}
	for _, tt := range cases {
		tt := tt
		t.Run(tt.Name, func(t *testing.T) {
			assert := assert.New(t)
			t.Parallel()

			resp, isValid := isValidLine(positions, tt.Text, tt.visitedIP)
			assert.Equal(resp, tt.ExpectedResp)
			assert.Equal(isValid, tt.ExpectedValid)
		})
	}
}
