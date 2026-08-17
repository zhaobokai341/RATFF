package shared

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractIPInfo_IPAPI(t *testing.T) {
	rawData := map[string]interface{}{
		"status":      "success",
		"continent":   "Asia",
		"country":     "China",
		"countryCode": "CN",
		"regionName":  "Shanxi",
		"region":      "SX",
		"city":        "Taiyuan",
		"zip":         "030000",
		"lat":         37.8706,
		"lon":         112.551,
		"timezone":    "Asia/Shanghai",
		"isp":         "CNC Group",
		"org":         "",
		"as":          "AS4837",
		"query":       "118.81.81.81",
	}

	result := ExtractIPInfo(rawData, "ip-api.com")

	assert.Equal(t, "118.81.81.81", result.IP)
	assert.Equal(t, "Asia", result.Continent)
	assert.Equal(t, "China", result.Country)
	assert.Equal(t, "CN", result.CountryCode)
	assert.Equal(t, "Shanxi", result.RegionName)
	assert.Equal(t, "Taiyuan", result.City)
	assert.Equal(t, "030000", result.Zip)
	assert.Equal(t, 37.8706, result.Latitude)
	assert.Equal(t, 112.551, result.Longitude)
	assert.Equal(t, "Asia/Shanghai", result.Timezone)
	assert.Equal(t, "CNC Group", result.ISP)
	assert.Equal(t, "AS4837", result.AS)
}

func TestExtractIPInfo_IPInfo(t *testing.T) {
	rawData := map[string]interface{}{
		"ip":       "118.81.81.81",
		"hostname": "81.81.81.118.adsl-pool.sx.cn",
		"city":     "Taiyuan",
		"region":   "Shanxi",
		"country":  "CN",
		"loc":      "37.8694,112.5603",
		"org":      "AS4837 CHINA UNICOM",
		"postal":   "030000",
		"timezone": "Asia/Shanghai",
	}

	result := ExtractIPInfo(rawData, "ipinfo.io")

	assert.Equal(t, "118.81.81.81", result.IP)
	assert.Equal(t, "CN", result.Country)
	assert.Equal(t, "CN", result.CountryCode)
	assert.Equal(t, "Shanxi", result.Region)
	assert.Equal(t, "Taiyuan", result.City)
	assert.Equal(t, "030000", result.Zip)
	assert.Equal(t, 37.8694, result.Latitude)
	assert.Equal(t, 112.5603, result.Longitude)
	assert.Equal(t, "Asia/Shanghai", result.Timezone)
	assert.Equal(t, "AS4837 CHINA UNICOM", result.ISP)
	assert.Equal(t, "AS4837 CHINA UNICOM", result.Org)
}

func TestExtractIPInfo_HTTPBin(t *testing.T) {
	rawData := map[string]interface{}{
		"origin": "118.81.81.81",
	}

	result := ExtractIPInfo(rawData, "httpbin.org")

	assert.Equal(t, "118.81.81.81", result.IP)
	assert.Equal(t, "", result.Country)
	assert.Equal(t, 0.0, result.Latitude)
}

func TestExtractIPInfo_UnknownAPI(t *testing.T) {
	rawData := map[string]interface{}{
		"ip":      "1.2.3.4",
		"country": "Test",
		"city":    "TestCity",
	}

	result := ExtractIPInfo(rawData, "unknown-api.com")

	assert.Equal(t, "1.2.3.4", result.IP)
	assert.Equal(t, "Test", result.Country)
	assert.Equal(t, "TestCity", result.City)
}

func TestParseLocation(t *testing.T) {
	t.Run("valid location", func(t *testing.T) {
		lat, lon := parseLocation("37.8694,112.5603")
		assert.Equal(t, 37.8694, lat)
		assert.Equal(t, 112.5603, lon)
	})

	t.Run("invalid location", func(t *testing.T) {
		lat, lon := parseLocation("invalid")
		assert.Equal(t, 0.0, lat)
		assert.Equal(t, 0.0, lon)
	})

	t.Run("empty location", func(t *testing.T) {
		lat, lon := parseLocation("")
		assert.Equal(t, 0.0, lat)
		assert.Equal(t, 0.0, lon)
	})
}

func TestToString(t *testing.T) {
	assert.Equal(t, "test", toString("test"))
	assert.Equal(t, "", toString(123))
	assert.Equal(t, "", toString(nil))
}

func TestToFloat64(t *testing.T) {
	assert.Equal(t, 123.45, toFloat64(123.45))
	assert.Equal(t, 0.0, toFloat64("test"))
	assert.Equal(t, 0.0, toFloat64(nil))
}
