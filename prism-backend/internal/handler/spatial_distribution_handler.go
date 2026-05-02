package handler

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/ridofiqri79/prism-backend/internal/model"
	"github.com/ridofiqri79/prism-backend/internal/service"
)

type SpatialDistributionHandler struct {
	service *service.SpatialDistributionService
}

func NewSpatialDistributionHandler(service *service.SpatialDistributionService) *SpatialDistributionHandler {
	return &SpatialDistributionHandler{service: service}
}

func (h *SpatialDistributionHandler) Choropleth(c echo.Context) error {
	res, err := h.service.Choropleth(c.Request().Context(), spatialDistributionFilter(c))
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, model.DataResponse[*model.SpatialDistributionChoroplethResponse]{Data: res})
}

func (h *SpatialDistributionHandler) RegionProjects(c echo.Context) error {
	filter := model.SpatialDistributionProjectFilter{
		SpatialDistributionFilter: spatialDistributionFilter(c),
		RegionCode:                strings.TrimSpace(c.QueryParam("region_code")),
	}

	res, err := h.service.RegionProjects(c.Request().Context(), filter, paginationParams(c))
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, res)
}

func spatialDistributionFilter(c echo.Context) model.SpatialDistributionFilter {
	return model.SpatialDistributionFilter{
		Level:            c.QueryParam("level"),
		ProvinceCode:     queryStringPtr(c, "province_code"),
		LoanTypes:        queryValues(c, "loan_types", "loan_types[]"),
		ProjectStatuses:  queryValues(c, "project_statuses", "project_statuses[]"),
		PipelineStatuses: queryValues(c, "pipeline_statuses", "pipeline_statuses[]"),
		Search:           queryStringPtr(c, "search"),
		IncludeHistory:   strings.EqualFold(c.QueryParam("include_history"), "true"),
	}
}
