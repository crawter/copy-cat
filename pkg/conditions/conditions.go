// Package conditions holds the two win conditions of the runner's levels:
// surviving the set time and collecting the set quota of jar lids. It is
// pure Go and imports no raylib, so it can be tested headless and reused
// unchanged on Android.
package conditions

// SurvivalTime is how long the hero must survive, in seconds, to complete
// the survival level.
const SurvivalTime = 60.0

// CollectionQuota is how many jar lids the hero must collect to complete the
// collection level.
const CollectionQuota = 8

// SurvivalMet reports whether the hero survived at least SurvivalTime seconds.
func SurvivalMet(elapsed float64) bool {
	return elapsed >= SurvivalTime
}

// CollectionMet reports whether the hero collected at least CollectionQuota
// jar lids.
func CollectionMet(collected int) bool {
	return collected >= CollectionQuota
}
