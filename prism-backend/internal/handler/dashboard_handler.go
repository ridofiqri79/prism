package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/ridofiqri79/prism-backend/internal/model"
	"github.com/ridofiqri79/prism-backend/internal/service"
)

type DashboardHandler struct {
	service *service.DashboardService
}

func NewDashboardHandler(service *service.DashboardService) *DashboardHandler {
	return &DashboardHandler{service: service}
}

func (h *DashboardHandler) BlueBookDistribution(c echo.Context) error {
	res, err := h.service.GetBlueBookDistribution(c.Request().Context())
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, model.DataResponse[*model.DashboardBlueBookDistributionResponse]{Data: res})
}

func (h *DashboardHandler) GreenBookDistribution(c echo.Context) error {
	res, err := h.service.GetGreenBookDistribution(c.Request().Context())
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, model.DataResponse[*model.DashboardGreenBookDistributionResponse]{Data: res})
}

func (h *DashboardHandler) DaftarKegiatanDistribution(c echo.Context) error {
	res, err := h.service.GetDaftarKegiatanDistribution(c.Request().Context())
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, model.DataResponse[*model.DashboardDaftarKegiatanDistributionResponse]{Data: res})
}

func (h *DashboardHandler) LoanAgreementDistribution(c echo.Context) error {
	res, err := h.service.GetLoanAgreementDistribution(c.Request().Context())
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, model.DataResponse[*model.DashboardLoanAgreementDistributionResponse]{Data: res})
}
