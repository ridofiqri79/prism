package service

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"

	"github.com/ridofiqri79/prism-backend/internal/model"
)

func TestDeleteDaftarKegiatanHardDeletesWhenNoProjects(t *testing.T) {
	env := setupBlueBookVersioningTest(t)
	dkService := NewDKService(env.pool, env.queries, nil)
	dk := env.createDaftarKegiatan(t, dkService, "DK-DELETE-EMPTY")
	dkID := mustParseUUID(t, dk.ID)

	if err := dkService.DeleteDaftarKegiatan(env.ctx, dkID); err != nil {
		t.Fatalf("DeleteDaftarKegiatan(no projects) error = %v", err)
	}
	if _, err := env.queries.GetDaftarKegiatan(env.ctx, dkID); !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("GetDaftarKegiatan after hard delete error = %v, want pgx.ErrNoRows", err)
	}
	assertAuditDeleteExists(t, env, "daftar_kegiatan", dk.ID)
}

func TestDeleteDaftarKegiatanRejectsWhenProjectsExist(t *testing.T) {
	env := setupBlueBookVersioningTest(t)
	gbService := NewGreenBookService(env.pool, env.queries, nil)
	dkService := NewDKService(env.pool, env.queries, nil)
	lender := env.createTestLender(t, "DK Delete Lender", "DKL")

	blueBook := env.createBlueBook(t, 0, nil)
	bbProject := env.createBBProject(t, blueBook.ID, "BB-DK-DEL-001", "BB Used By DK Delete")
	greenBook := env.createGreenBook(t, gbService, 0, nil)
	gbProject := env.createGBProjectWithFundingLender(t, gbService, greenBook.ID, bbProject.ID, "GB-DK-DEL-001", "GB Used By DK Delete", lender)
	dk := env.createDaftarKegiatan(t, dkService, "DK-DELETE-USED")
	_ = env.createDKProject(t, dkService, dk.ID, gbProject.ID, lender)

	err := dkService.DeleteDaftarKegiatan(env.ctx, mustParseUUID(t, dk.ID))
	assertAppErrorCode(t, err, "CONFLICT")
	if current, getErr := dkService.GetDaftarKegiatan(env.ctx, mustParseUUID(t, dk.ID)); getErr != nil {
		t.Fatalf("GetDaftarKegiatan after blocked delete error = %v", getErr)
	} else if current.ProjectCount != 1 {
		t.Fatalf("ProjectCount after blocked delete = %d, want 1", current.ProjectCount)
	}

	list, err := dkService.ListDaftarKegiatan(env.ctx, model.DaftarKegiatanListFilter{}, model.PaginationParams{Page: 1, Limit: 10})
	if err != nil {
		t.Fatalf("ListDaftarKegiatan after blocked delete error = %v", err)
	}
	found := false
	for _, row := range list.Data {
		if row.ID == dk.ID {
			found = true
			if row.ProjectCount != 1 {
				t.Fatalf("list ProjectCount = %d, want 1", row.ProjectCount)
			}
		}
	}
	if !found {
		t.Fatalf("Daftar Kegiatan %s not found in list after blocked delete", dk.ID)
	}
}
