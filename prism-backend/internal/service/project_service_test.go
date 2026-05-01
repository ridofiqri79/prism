package service

import (
	"testing"

	"github.com/ridofiqri79/prism-backend/internal/model"
)

func TestProjectMasterSearchByBBCodeCountsAndSummarizes(t *testing.T) {
	env := setupBlueBookVersioningTest(t)
	projectService := NewProjectService(env.queries)

	blueBook := env.createBlueBook(t, 0, nil)
	project := env.createBBProject(t, blueBook.ID, "BB-PM-SEARCH-001", "Searchable Code Project")
	search := "BB-PM-SEARCH"

	result, err := projectService.ListProjectMaster(env.ctx, model.ProjectMasterFilter{
		Search: &search,
	}, model.PaginationParams{Page: 1, Limit: 10})
	if err != nil {
		t.Fatalf("ListProjectMaster(search by bb_code) error = %v", err)
	}
	if result.Meta.Total != 1 {
		t.Fatalf("total = %d, want 1", result.Meta.Total)
	}
	if len(result.Data) != 1 || result.Data[0].ID != project.ID {
		t.Fatalf("data = %+v, want project %s", result.Data, project.ID)
	}
	if result.Summary.TotalLoanUSD != 1000000 {
		t.Fatalf("summary total loan = %v, want 1000000", result.Summary.TotalLoanUSD)
	}
}
