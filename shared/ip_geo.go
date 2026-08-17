package shared

import (
	"strconv"
	"strings"
)

// IPGeoStandard holds normalized IP geolocation information.
type IPGeoStandard struct {
	IP          string  `json:"ip"`
	Continent   string  `json:"continent"`
	Country     string  `json:"country"`
	CountryCode string  `json:"country_code"`
	Region      string  `json:"region"`
	RegionName  string  `json:"region_name"`
	City        string  `json:"city"`
	District    string  `json:"district"`
	Zip         string  `json:"zip"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	Timezone    string  `json:"timezone"`
	ISP         string  `json:"isp"`
	Org         string  `json:"org"`
	AS          string  `json:"as"`
}

// apiFieldMappings defines field mappings for known IP geolocation APIs.
var apiFieldMappings = map[string]map[string]string{
	"ip-api.com": {
		"ip":           "query",
		"continent":    "continent",
		"country":      "country",
		"country_code": "countryCode",
		"region":       "region",
		"region_name":  "regionName",
		"city":         "city",
		"district":     "district",
		"zip":          "zip",
		"latitude":     "lat",
		"longitude":    "lon",
		"timezone":     "timezone",
		"isp":          "isp",
		"org":          "org",
		"as":           "as",
	},
	"ipinfo.io": {
		"ip":           "ip",
		"country":      "country",
		"country_code": "country",
		"region":       "region",
		"city":         "city",
		"zip":          "postal",
		"latitude":     "loc",
		"longitude":    "loc",
		"timezone":     "timezone",
		"isp":          "org",
		"org":          "org",
	},
	"httpbin.org": {
		"ip": "origin",
	},
}

// fieldKeywords defines keyword patterns for automatic field matching.
var fieldKeywords = map[string][]string{
	"ip":           {"ip", "query", "host", "origin"},
	"continent":    {"continent"},
	"country":      {"country"},
	"country_code": {"countryCode", "country_code", "cc"},
	"region":       {"region", "state", "province"},
	"region_name":  {"regionName", "region_name", "state_name"},
	"city":         {"city"},
	"district":     {"district"},
	"zip":          {"zip", "postal"},
	"latitude":     {"lat", "latitude"},
	"longitude":    {"lon", "lng", "longitude"},
	"timezone":     {"timezone"},
	"isp":          {"isp"},
	"org":          {"org", "organization"},
	"as":           {"as", "asn"},
}

// ExtractIPInfo extracts standardized IP info from raw API response.
func ExtractIPInfo(rawData map[string]interface{}, apiSource string) IPGeoStandard {
	var standard IPGeoStandard

	mapping, ok := apiFieldMappings[apiSource]
	if ok {
		standard = extractWithMapping(rawData, mapping)
	} else {
		standard = extractWithKeywords(rawData)
	}

	if apiSource == "ipinfo.io" {
		if loc, ok := rawData["loc"].(string); ok {
			standard.Latitude, standard.Longitude = parseLocation(loc)
		}
	}

	return standard
}

// extractWithMapping extracts fields using predefined mapping.
func extractWithMapping(rawData map[string]interface{}, mapping map[string]string) IPGeoStandard {
	var standard IPGeoStandard

	for stdField, rawField := range mapping {
		if val, ok := rawData[rawField]; ok {
			setStandardField(&standard, stdField, val)
		}
	}

	return standard
}

// extractWithKeywords extracts fields using keyword matching.
func extractWithKeywords(rawData map[string]interface{}) IPGeoStandard {
	var standard IPGeoStandard

	for stdField, keywords := range fieldKeywords {
		for _, keyword := range keywords {
			if val, ok := rawData[keyword]; ok {
				setStandardField(&standard, stdField, val)
				break
			}
		}
	}

	return standard
}

// setStandardField sets a standardized field value.
func setStandardField(standard *IPGeoStandard, field string, value interface{}) {
	switch field {
	case "ip":
		standard.IP = toString(value)
	case "continent":
		standard.Continent = toString(value)
	case "country":
		standard.Country = toString(value)
	case "country_code":
		standard.CountryCode = toString(value)
	case "region":
		standard.Region = toString(value)
	case "region_name":
		standard.RegionName = toString(value)
	case "city":
		standard.City = toString(value)
	case "district":
		standard.District = toString(value)
	case "zip":
		standard.Zip = toString(value)
	case "latitude":
		if _, ok := value.(string); !ok {
			standard.Latitude = toFloat64(value)
		}
	case "longitude":
		if _, ok := value.(string); !ok {
			standard.Longitude = toFloat64(value)
		}
	case "timezone":
		standard.Timezone = toString(value)
	case "isp":
		standard.ISP = toString(value)
	case "org":
		standard.Org = toString(value)
	case "as":
		standard.AS = toString(value)
	}
}

// parseLocation parses "lat,lon" format string.
func parseLocation(loc string) (float64, float64) {
	parts := strings.Split(loc, ",")
	if len(parts) == 2 {
		lat, errLat := strconv.ParseFloat(parts[0], 64)
		lon, errLon := strconv.ParseFloat(parts[1], 64)
		if errLat == nil && errLon == nil {
			return lat, lon
		}
	}
	return 0, 0
}

// toString converts interface{} to string.
func toString(v interface{}) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

// toFloat64 converts interface{} to float64.
func toFloat64(v interface{}) float64 {
	if f, ok := v.(float64); ok {
		return f
	}
	return 0
}
