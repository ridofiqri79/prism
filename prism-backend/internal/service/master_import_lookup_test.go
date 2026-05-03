package service

import (
	"testing"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/ridofiqri79/prism-backend/internal/database/queries"
	"github.com/ridofiqri79/prism-backend/internal/model"
)

func TestMasterImportInstitutionLookupTracksScopedDuplicates(t *testing.T) {
	lookups := &masterImportLookups{
		institutionsByID:          map[string]queries.Institution{},
		institutionsByName:        map[string]queries.Institution{},
		institutionsByPath:        map[string]queries.Institution{},
		ambiguousInstitutionNames: map[string]struct{}{},
		institutionScopeKeys:      map[string]struct{}{},
	}

	parentA := testUUID(1)
	parentB := testUUID(2)
	rootA := queries.Institution{ID: parentA, Name: "Kementerian A"}
	rootB := queries.Institution{ID: parentB, Name: "Kementerian B"}
	childA := queries.Institution{ID: testUUID(3), ParentID: parentA, Name: "Unit Pelaksana"}
	childB := queries.Institution{ID: testUUID(4), ParentID: parentB, Name: "Unit Pelaksana"}

	lookups.addInstitution(rootA)
	lookups.addInstitution(rootB)
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
	if got, exists, ambiguous := lookups.lookupInstitutionReference("Unit Pelaksana; Kementerian B;"); !exists || ambiguous || got.ID != childB.ID {
		t.Fatalf("lookup scoped path = id %v exists %v ambiguous %v, want childB exists true ambiguous false", got.ID, exists, ambiguous)
	}
	if got, exists, ambiguous := lookups.lookupInstitutionReference("Unit Pelaksana"); exists || !ambiguous || got.ID.Valid {
		t.Fatalf("lookup ambiguous name = id %v exists %v ambiguous %v, want no id exists false ambiguous true", got.ID, exists, ambiguous)
	}
	if got, exists, ambiguous := lookups.lookupInstitutionReference(modelUUIDString(childA.ID)); !exists || ambiguous || got.ID != childA.ID {
		t.Fatalf("lookup id = id %v exists %v ambiguous %v, want childA exists true ambiguous false", got.ID, exists, ambiguous)
	}
}

func TestMasterImportLenderLookupFallsBackToUniqueShortName(t *testing.T) {
	lookups := &masterImportLookups{}
	lender := masterImportTestLender(testUUID(10), "Asian Development Bank", "ADB")
	lookups.addLender(lender.ID, lender.Name, lender.Type, lender.ShortName)

	if got, exists, ambiguous := lookups.lookupLenderReference("Asian Development Bank"); !exists || ambiguous || got.ID != lender.ID {
		t.Fatalf("lookup by name = id %v exists %v ambiguous %v, want lender exists true ambiguous false", got.ID, exists, ambiguous)
	}
	if got, exists, ambiguous := lookups.lookupLenderReference("adb"); !exists || ambiguous || got.ID != lender.ID {
		t.Fatalf("lookup by short_name = id %v exists %v ambiguous %v, want lender exists true ambiguous false", got.ID, exists, ambiguous)
	}
}

func TestMasterImportLenderLookupRejectsAmbiguousShortName(t *testing.T) {
	lookups := &masterImportLookups{}
	first := masterImportTestLender(testUUID(11), "First Development Bank", "FDB")
	second := masterImportTestLender(testUUID(12), "Second Development Bank", "FDB")
	lookups.addLender(first.ID, first.Name, first.Type, first.ShortName)
	lookups.addLender(second.ID, second.Name, second.Type, second.ShortName)

	if got, exists, ambiguous := lookups.lookupLenderReference("FDB"); exists || !ambiguous || got.ID.Valid {
		t.Fatalf("lookup ambiguous short_name = id %v exists %v ambiguous %v, want no id exists false ambiguous true", got.ID, exists, ambiguous)
	}
	if got, exists, ambiguous := lookups.lookupLenderReference("First Development Bank"); !exists || ambiguous || got.ID != first.ID {
		t.Fatalf("lookup full name after ambiguous short_name = id %v exists %v ambiguous %v, want first exists true ambiguous false", got.ID, exists, ambiguous)
	}
}

func testUUID(seed byte) pgtype.UUID {
	return pgtype.UUID{
		Bytes: [16]byte{seed},
		Valid: true,
	}
}

func masterImportTestLender(id pgtype.UUID, name, shortName string) masterImportLenderRef {
	return masterImportLenderRef{
		ID:        id,
		Name:      name,
		Type:      "Multilateral",
		ShortName: pgtype.Text{String: shortName, Valid: shortName != ""},
	}
}

func modelUUIDString(value pgtype.UUID) string {
	return model.UUIDToString(value)
}
