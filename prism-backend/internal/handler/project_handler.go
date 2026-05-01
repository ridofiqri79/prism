package handler

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/ridofiqri79/prism-backend/internal/model"
	"github.com/ridofiqri79/prism-backend/internal/service"
)

type ProjectHandler struct {
	service *service.ProjectService
}

func NewProjectHandler(service *service.ProjectService) *ProjectHandler {
	return &ProjectHandler{service: service}
}

func (h *ProjectHandler) ListMaster(c echo.Context) error {
	res, err := h.service.ListProjectMaster(c.Request().Context(), projectMasterFilter(c), paginationParams(c))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, res)
}

func (h *ProjectHandler) ExportMaster(c echo.Context) error {
	file, err := h.service.ExportProjectMaster(c.Request().Context(), projectMasterFilter(c), paginationParams(c))
	if err != nil {
		return err
	}

	c.Response().Header().Set(echo.HeaderContentDisposition, `attachment; filename="`+file.FileName+`"`)
	return c.Blob(http.StatusOK, file.ContentType, file.Data)
}

func projectMasterFilter(c echo.Context) model.ProjectMasterFilter {
	return model.ProjectMasterFilter{
		LoanTypes:           queryValues(c, "loan_types", "loan_types[]"),
		IndicationLenderIDs: queryValues(c, "indication_lender_ids", "indication_lender_ids[]"),
		ExecutingAgencyIDs:  queryValues(c, "executing_agency_ids", "executing_agency_ids[]"),
		FixedLenderIDs:      queryValues(c, "fixed_lender_ids", "fixed_lender_ids[]"),
		ProjectStatuses:     queryValues(c, "project_statuses", "project_statuses[]"),
		PipelineStatuses:    queryValues(c, "pipeline_statuses", "pipeline_statuses[]"),
		ProgramTitleIDs:     queryValues(c, "program_title_ids", "program_title_ids[]"),
		RegionIDs:           queryValues(c, "region_ids", "region_ids[]"),
		ForeignLoanMin:      queryStringPtr(c, "foreign_loan_min"),
		ForeignLoanMax:      queryStringPtr(c, "foreign_loan_max"),
		DKDateFrom:          queryStringPtr(c, "dk_date_from"),
		DKDateTo:            queryStringPtr(c, "dk_date_to"),
		Search:              queryStringPtr(c, "search"),
		IncludeHistory:      strings.EqualFold(c.QueryParam("include_history"), "true"),
	}
}

func queryValues(c echo.Context, names ...string) []string {
	raw := c.QueryParams()
	values := []string{}
	seen := map[string]struct{}{}
	for _, name := range names {
		for _, value := range raw[name] {
			for _, item := range strings.Split(value, ",") {
				item = strings.TrimSpace(item)
				if item == "" {
					continue
				}
				if _, ok := seen[item]; ok {
					continue
				}
				seen[item] = struct{}{}
				values = append(values, item)
			}
		}
	}
	return values
}
