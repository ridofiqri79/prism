package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"

	apperrors "github.com/ridofiqri79/prism-backend/internal/errors"
	"github.com/ridofiqri79/prism-backend/internal/model"
	"github.com/ridofiqri79/prism-backend/internal/service"
)

type LAHandler struct {
	service *service.LAService
}

func NewLAHandler(service *service.LAService) *LAHandler {
	return &LAHandler{service: service}
}

func (h *LAHandler) ListLA(c echo.Context) error {
	res, err := h.service.ListLoanAgreements(c.Request().Context(), loanAgreementListFilter(c), paginationParams(c))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, res)
}

func loanAgreementListFilter(c echo.Context) model.LoanAgreementListFilter {
	return model.LoanAgreementListFilter{
		LenderID:          queryStringPtr(c, "lender_id"),
		IsExtended:        queryStringPtr(c, "is_extended"),
		ClosingDateBefore: queryStringPtr(c, "closing_date_before"),
		RiskCodes:         queryValues(c, "risk_codes", "risk_codes[]"),
	}
}

func (h *LAHandler) GetLA(c echo.Context) error {
	id, err := parseIDParam(c, "id")
	if err != nil {
		return err
	}
	res, err := h.service.GetLoanAgreement(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, model.DataResponse[*model.LoanAgreementResponse]{Data: res})
}

func (h *LAHandler) CreateLA(c echo.Context) error {
	var req model.LoanAgreementRequest
	if err := bind(c, &req); err != nil {
		return err
	}
	res, err := h.service.CreateLoanAgreement(c.Request().Context(), req)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, model.DataResponse[*model.LoanAgreementResponse]{Data: res})
}

func (h *LAHandler) DownloadLAImportTemplate(c echo.Context) error {
	template, err := h.service.BuildImportTemplate(c.Request().Context())
	if err != nil {
		return err
	}

	c.Response().Header().Set(echo.HeaderContentDisposition, `attachment; filename="`+template.FileName+`"`)
	return c.Blob(http.StatusOK, template.ContentType, template.Data)
}

func (h *LAHandler) PreviewImportLA(c echo.Context) error {
	return h.handleImportLA(c, true)
}

func (h *LAHandler) ImportLA(c echo.Context) error {
	return h.handleImportLA(c, false)
}

func (h *LAHandler) UpdateLA(c echo.Context) error {
	id, err := parseIDParam(c, "id")
	if err != nil {
		return err
	}
	var req model.LoanAgreementRequest
	if err := bind(c, &req); err != nil {
		return err
	}
	res, err := h.service.UpdateLoanAgreement(c.Request().Context(), id, req)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, model.DataResponse[*model.LoanAgreementResponse]{Data: res})
}

func (h *LAHandler) DeleteLA(c echo.Context) error {
	id, err := parseIDParam(c, "id")
	if err != nil {
		return err
	}
	if err := h.service.DeleteLoanAgreement(c.Request().Context(), id); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *LAHandler) handleImportLA(c echo.Context, preview bool) error {
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
		res, err = h.service.PreviewLoanAgreementImport(c.Request().Context(), file.Filename, src, file.Size)
	} else {
		res, err = h.service.ImportLoanAgreement(c.Request().Context(), file.Filename, src, file.Size)
	}
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, model.DataResponse[*model.MasterImportResponse]{Data: res})
}
