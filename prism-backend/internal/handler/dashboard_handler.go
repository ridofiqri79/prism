package handler

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"

	apperrors "github.com/ridofiqri79/prism-backend/internal/errors"
	"github.com/ridofiqri79/prism-backend/internal/model"
	"github.com/ridofiqri79/prism-backend/internal/service"
)

type DashboardHandler struct {
	service *service.DashboardService
}

func NewDashboardHandler(service *service.DashboardService) *DashboardHandler {
	return &DashboardHandler{service: service}
}

func (h *DashboardHandler) StageOverview(c echo.Context) error {
	periodIDs, err := dashboardPeriodIDs(c)
	if err != nil {
		return err
	}

	res, err := h.service.GetStageOverview(c.Request().Context(), periodIDs)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, model.DataResponse[*model.DashboardStageOverviewResponse]{Data: res})
}

func (h *DashboardHandler) BlueBookDistribution(c echo.Context) error {
	periodIDs, err := dashboardPeriodIDs(c)
	if err != nil {
		return err
	}

	res, err := h.service.GetBlueBookDistribution(c.Request().Context(), periodIDs)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, model.DataResponse[*model.DashboardBlueBookDistributionResponse]{Data: res})
}

func (h *DashboardHandler) GreenBookDistribution(c echo.Context) error {
	periodIDs, err := dashboardPeriodIDs(c)
	if err != nil {
		return err
	}

	res, err := h.service.GetGreenBookDistribution(c.Request().Context(), periodIDs)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, model.DataResponse[*model.DashboardGreenBookDistributionResponse]{Data: res})
}

func (h *DashboardHandler) DaftarKegiatanDistribution(c echo.Context) error {
	periodIDs, err := dashboardPeriodIDs(c)
	if err != nil {
		return err
	}

	res, err := h.service.GetDaftarKegiatanDistribution(c.Request().Context(), periodIDs)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, model.DataResponse[*model.DashboardDaftarKegiatanDistributionResponse]{Data: res})
}

func (h *DashboardHandler) LoanAgreementDistribution(c echo.Context) error {
	periodIDs, err := dashboardPeriodIDs(c)
	if err != nil {
		return err
	}

	res, err := h.service.GetLoanAgreementDistribution(c.Request().Context(), periodIDs)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, model.DataResponse[*model.DashboardLoanAgreementDistributionResponse]{Data: res})
}

func dashboardPeriodIDs(c echo.Context) ([]pgtype.UUID, error) {
	rawValues := queryValues(c, "period_ids", "period_ids[]", "period_id")
	if len(rawValues) == 0 {
		return nil, nil
	}

	periodIDs := make([]pgtype.UUID, 0, len(rawValues))
	for _, raw := range rawValues {
		periodID, err := model.ParseUUID(raw)
		if err != nil {
			return nil, apperrors.Validation(
				apperrors.FieldError{Field: "period_ids", Message: "UUID tidak valid"},
			)
		}

		periodIDs = append(periodIDs, periodID)
	}

	return periodIDs, nil
}
