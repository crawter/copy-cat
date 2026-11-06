package conditions

import "testing"

func TestSurvivalConditionReadsSetTime(t *testing.T) {
	if SurvivalMet(SurvivalTime - 1.0) {
		t.Fatalf("SurvivalMet(%v) = true, want false one second early", SurvivalTime-1.0)
	}
	if !SurvivalMet(SurvivalTime) {
		t.Fatalf("SurvivalMet(%v) = false, want true at the set time", SurvivalTime)
	}
}

func TestCollectionConditionReadsSetQuota(t *testing.T) {
	if CollectionMet(CollectionQuota - 1) {
		t.Fatalf("CollectionMet(%v) = true, want false one short of the quota", CollectionQuota-1)
	}
	if !CollectionMet(CollectionQuota) {
		t.Fatalf("CollectionMet(%v) = false, want true at the quota", CollectionQuota)
	}
}
