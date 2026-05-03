package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"

	apperrors "github.com/ridofiqri79/prism-backend/internal/errors"
	"github.com/ridofiqri79/prism-backend/internal/model"
	"github.com/ridofiqri79/prism-backend/internal/service"
)

type MonitoringHandler struct {
	service *service.MonitoringService
}

func NewMonitoringHandler(service *service.MonitoringService) *MonitoringHandler {
	return &MonitoringHandler{service: service}
}

func (h *MonitoringHandler) ListLoanAgreementReferences(c echo.Context) error {
	res, err := h.service.ListLoanAgreementReferences(c.Request().Context(), model.MonitoringLoanAgreementListFilter{
		IsEffective: queryStringPtr(c, "is_effective"),
	}, paginationParams(c))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, res)
}

func (h *MonitoringHandler) List(c echo.Context) error {
	laID, err := parseIDParam(c, "laId")
	if err != nil {
		return err
	}
	res, err := h.service.ListMonitoring(c.Request().Context(), laID, monitoringListFilter(c), paginationParams(c))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, res)
}

func monitoringListFilter(c echo.Context) model.MonitoringListFilter {
	return model.MonitoringListFilter{
		BudgetYear: queryStringPtr(c, "budget_year"),
		Quarter:    queryStringPtr(c, "quarter"),
	}
}

func (h *MonitoringHandler) Get(c echo.Context) error {
	laID, err := parseIDParam(c, "laId")
	if err != nil {
		return err
	}
	id, err := parseIDParam(c, "id")
	if err != nil {
		return err
	}
	res, err := h.service.GetMonitoring(c.Request().Context(), laID, id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, model.DataResponse[*model.MonitoringResponse]{Data: res})
}

func (h *MonitoringHandler) Create(c echo.Context) error {
	laID, err := parseIDParam(c, "laId")
	if err != nil {
		return err
	}
	var req model.MonitoringRequest
	if err := bind(c, &req); err != nil {
		return err
	}
	res, err := h.service.CreateMonitoring(c.Request().Context(), laID, req)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, model.DataResponse[*model.MonitoringResponse]{Data: res})
}

func (h *MonitoringHandler) DownloadImportTemplate(c echo.Context) error {
	template, err := h.service.BuildImportTemplate(c.Request().Context())
	if err != nil {
		return err
	}

	c.Response().Header().Set(echo.HeaderContentDisposition, `attachment; filename="`+template.FileName+`"`)
	return c.Blob(http.StatusOK, template.ContentType, template.Data)
}

func (h *MonitoringHandler) PreviewImport(c echo.Context) error {
	return h.handleImport(c, true)
}

func (h *MonitoringHandler) Import(c echo.Context) error {
	return h.handleImport(c, false)
}

func (h *MonitoringHandler) Update(c echo.Context) error {
	laID, err := parseIDParam(c, "laId")
	if err != nil {
		return err
	}
	id, err := parseIDParam(c, "id")
	if err != nil {
		return err
	}
	var req model.MonitoringRequest
	if err := bind(c, &req); err != nil {
		return err
	}
	res, err := h.service.UpdateMonitoring(c.Request().Context(), laID, id, req)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, model.DataResponse[*model.MonitoringResponse]{Data: res})
}

func (h *MonitoringHandler) Delete(c echo.Context) error {
	laID, err := parseIDParam(c, "laId")
	if err != nil {
		return err
	}
	id, err := parseIDParam(c, "id")
	if err != nil {
		return err
	}
	if err := h.service.DeleteMonitoring(c.Request().Context(), laID, id); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *MonitoringHandler) handleImport(c echo.Context, preview bool) error {
	file, err := c.FormFile("file")
	if err != nil {
		return apperrors.Validation(apperrors.FieldError{Field: "file", Message: "file wajib diunggah"})
	}

	src, err := file.Open()
	if err != nil {
		return apperrors.Validation(apperrors.FieldError{Field: "file", Message: "file tidak dapat dibaca"})
	}
	defer src.Close()

	var res *model.MasterImportResponse
	if preview {
		res, err = h.service.PreviewMonitoringImport(c.Request().Context(), file.Filename, src, file.Size)
	} else {
		res, err = h.service.ImportMonitoring(c.Request().Context(), file.Filename, src, file.Size)
	}
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, model.DataResponse[*model.MasterImportResponse]{Data: res})
}
