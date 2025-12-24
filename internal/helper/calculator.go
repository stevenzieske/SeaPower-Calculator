package helper

func CalculateFlightTime(velocityInKnots int64, rangeInNM int64) float64 {
	return float64(rangeInNM) / float64(velocityInKnots)
}

func CalculateRange(velocityInKnots int64, flightTime float64) float64 {
	return float64(velocityInKnots) * flightTime
}
