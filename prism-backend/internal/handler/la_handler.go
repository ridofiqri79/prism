package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"

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
