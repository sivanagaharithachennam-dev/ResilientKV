package raft

import "testing"

func TestMajority(t *testing.T) {
	tests := []struct {
		clusterSize int
		want        int
	}{
		{1, 1},
		{2, 2},
		{3, 2},
		{4, 3},
		{5, 3},
		{7, 4},
	}

	for _, tt := range tests {
		got, err := Majority(tt.clusterSize)
		if err != nil {
			t.Fatalf("Majority(%d): %v", tt.clusterSize, err)
		}
		if got != tt.want {
			t.Errorf("Majority(%d) = %d, want %d", tt.clusterSize, got, tt.want)
		}
	}
}

func TestMajorityRejectsInvalidClusterSize(t *testing.T) {
	if _, err := Majority(0); err == nil {
		t.Fatal("expected zero-sized cluster to be rejected")
	}
	if _, err := Majority(-1); err == nil {
		t.Fatal("expected negative cluster size to be rejected")
	}
}

func TestHasMajority(t *testing.T) {
	tests := []struct {
		name        string
		acks        int
		clusterSize int
		want        bool
	}{
		{"single node", 1, 1, true},
		{"three nodes insufficient", 1, 3, false},
		{"three nodes majority", 2, 3, true},
		{"five nodes insufficient", 2, 5, false},
		{"five nodes majority", 3, 5, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := HasMajority(tt.acks, tt.clusterSize)
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Fatalf("HasMajority(%d, %d) = %v, want %v",
					tt.acks, tt.clusterSize, got, tt.want)
			}
		})
	}
}

func TestHasMajorityRejectsInvalidAcknowledgements(t *testing.T) {
	for _, acks := range []int{-1, 4} {
		if _, err := HasMajority(acks, 3); err == nil {
			t.Errorf("expected invalid acknowledgement count %d to be rejected", acks)
		}
	}
}
