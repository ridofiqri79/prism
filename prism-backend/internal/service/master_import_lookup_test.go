package service

import (
	"testing"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/ridofiqri79/prism-backend/internal/database/queries"
)

func TestMasterImportInstitutionLookupTracksScopedDuplicates(t *testing.T) {
	lookups := &masterImportLookups{
		institutionsByName:        map[string]queries.Institution{},
		ambiguousInstitutionNames: map[string]struct{}{},
		institutionScopeKeys:      map[string]struct{}{},
	}

	parentA := testUUID(1)
	parentB := testUUID(2)
	childA := queries.Institution{ID: testUUID(3), ParentID: parentA, Name: "Unit Pelaksana"}
	childB := queries.Institution{ID: testUUID(4), ParentID: parentB, Name: "Unit Pelaksana"}

	lookups.addInstitution(childA)
	if !lookups.hasInstitutionInScope("Unit Pelaksana", parentA) {
		t.Fatal("expected institution to exist in first parent scope")
	}
	if lookups.hasInstitutionInScope("Unit Pelaksana", parentB) {
		t.Fatal("did not expect institution to exist in second parent scope before it is added")
	}
	if _, exists, ambiguous := lookups.lookupInstitutionByName("Unit Pelaksana"); !exists || ambiguous {
		t.Fatalf("lookup after single scoped institution = exists %v ambiguous %v, want exists true ambiguous false", exists, ambiguous)
	}

	lookups.addInstitution(childB)
	if !lookups.hasInstitutionInScope("Unit Pelaksana", parentB) {
		t.Fatal("expected institution to exist in second parent scope")
	}
	if _, exists, ambiguous := lookups.lookupInstitutionByName("Unit Pelaksana"); exists || !ambiguous {
		t.Fatalf("lookup after duplicate child names = exists %v ambiguous %v, want exists false ambiguous true", exists, ambiguous)
	}
}

func testUUID(seed byte) pgtype.UUID {
	return pgtype.UUID{
		Bytes: [16]byte{seed},
		Valid: true,
	}
}
