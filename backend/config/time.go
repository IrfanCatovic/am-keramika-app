package config

import (
	"time"

	_ "time/tzdata"
)

const BusinessTimeZone = "Europe/Belgrade"

var businessLocation = loadBusinessLocation()

func loadBusinessLocation() *time.Location {
	location, err := time.LoadLocation(BusinessTimeZone)
	if err != nil {
		// The official Go distributions include the IANA tz database. Keep a
		// deterministic fallback for minimal container images.
		return time.FixedZone(BusinessTimeZone, 1*60*60)
	}
	return location
}

func FormatBusinessDateTime(value time.Time) string {
	return value.In(businessLocation).Format("2006-01-02 15:04")
}

func BusinessLocation() *time.Location {
	return businessLocation
}
