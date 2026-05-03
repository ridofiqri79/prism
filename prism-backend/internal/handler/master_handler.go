package handler

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	apperrors "github.com/ridofiqri79/prism-backend/internal/errors"
	"github.com/ridofiqri79/prism-backend/internal/model"
	"github.com/ridofiqri79/prism-backend/internal/service"
)

type MasterHandler struct {
	service *service.MasterService
}

func NewMasterHandler(service *service.MasterService) *MasterHandler {
	return &MasterHandler{service: service}
}

func (h *MasterHandler) ListCountries(c echo.Context) error {
	res, err := h.service.ListCountries(c.Request().Context(), paginationParams(c))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, res)
}

func (h *MasterHandler) GetCountry(c echo.Context) error {
	id, err := parseIDParam(c, "id")
	if err != nil {
		return err
	}
	res, err := h.service.GetCountry(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, model.DataResponse[*model.CountryResponse]{Data: res})
}

func (h *MasterHandler) CreateCountry(c echo.Context) error {
	var req model.CountryRequest
	if err := bind(c, &req); err != nil {
		return err
	}
	res, err := h.service.CreateCountry(c.Request().Context(), req)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, model.DataResponse[*model.CountryResponse]{Data: res})
}

func (h *MasterHandler) UpdateCountry(c echo.Context) error {
	id, err := parseIDParam(c, "id")
	if err != nil {
		return err
	}
	var req model.CountryRequest
	if err := bind(c, &req); err != nil {
		return err
	}
	res, err := h.service.UpdateCountry(c.Request().Context(), id, req)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, model.DataResponse[*model.CountryResponse]{Data: res})
}

func (h *MasterHandler) DeleteCountry(c echo.Context) error {
	id, err := parseIDParam(c, "id")
	if err != nil {
		return err
	}
	if err := h.service.DeleteCountry(c.Request().Context(), id); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *MasterHandler) ListCurrencies(c echo.Context) error {
	res, err := h.service.ListCurrencies(c.Request().Context(), paginationParams(c), c.QueryParam("active"))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, res)
}

func (h *MasterHandler) GetCurrency(c echo.Context) error {
	id, err := parseIDParam(c, "id")
	if err != nil {
		return err
	}
	res, err := h.service.GetCurrency(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, model.DataResponse[*model.CurrencyResponse]{Data: res})
}

func (h *MasterHandler) CreateCurrency(c echo.Context) error {
	var req model.CurrencyRequest
	if err := bind(c, &req); err != nil {
		return err
	}
	res, err := h.service.CreateCurrency(c.Request().Context(), req)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, model.DataResponse[*model.CurrencyResponse]{Data: res})
}

func (h *MasterHandler) UpdateCurrency(c echo.Context) error {
	id, err := parseIDParam(c, "id")
	if err != nil {
		return err
	}
	var req model.CurrencyRequest
	if err := bind(c, &req); err != nil {
		return err
	}
	res, err := h.service.UpdateCurrency(c.Request().Context(), id, req)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, model.DataResponse[*model.CurrencyResponse]{Data: res})
}

func (h *MasterHandler) DeleteCurrency(c echo.Context) error {
	id, err := parseIDParam(c, "id")
	if err != nil {
		return err
	}
	if err := h.service.DeleteCurrency(c.Request().Context(), id); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *MasterHandler) ListLenders(c echo.Context) error {
	res, err := h.service.ListLenders(c.Request().Context(), paginationParams(c), queryStrings(c, "type"))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, res)
}

func (h *MasterHandler) GetLender(c echo.Context) error {
	id, err := parseIDParam(c, "id")
	if err != nil {
		return err
	}
	res, err := h.service.GetLender(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, model.DataResponse[*model.LenderResponse]{Data: res})
}

func (h *MasterHandler) CreateLender(c echo.Context) error {
	var req model.CreateLenderRequest
	if err := bind(c, &req); err != nil {
		return err
	}
	res, err := h.service.CreateLender(c.Request().Context(), req)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, model.DataResponse[*model.LenderResponse]{Data: res})
}

func (h *MasterHandler) UpdateLender(c echo.Context) error {
	id, err := parseIDParam(c, "id")
	if err != nil {
		return err
	}
	var req model.UpdateLenderRequest
	if err := bind(c, &req); err != nil {
		return err
	}
	res, err := h.service.UpdateLender(c.Request().Context(), id, req)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, model.DataResponse[*model.LenderResponse]{Data: res})
}

func (h *MasterHandler) DeleteLender(c echo.Context) error {
	id, err := parseIDParam(c, "id")
	if err != nil {
		return err
	}
	if err := h.service.DeleteLender(c.Request().Context(), id); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *MasterHandler) ListInstitutions(c echo.Context) error {
	res, err := h.service.ListInstitutions(c.Request().Context(), paginationParams(c), queryStrings(c, "level"), queryStringPtr(c, "parent_id"))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, res)
}

func (h *MasterHandler) LookupInstitutions(c echo.Context) error {
	res, err := h.service.LookupInstitutions(c.Request().Context(), paginationParams(c), queryStrings(c, "level"), queryStringPtr(c, "parent_id"))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, res)
}

func (h *MasterHandler) GetInstitution(c echo.Context) error {
	id, err := parseIDParam(c, "id")
	if err != nil {
		return err
	}
	res, err := h.service.GetInstitution(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, model.DataResponse[*model.InstitutionResponse]{Data: res})
}

func (h *MasterHandler) CreateInstitution(c echo.Context) error {
	var req model.InstitutionRequest
	if err := bind(c, &req); err != nil {
		return err
	}
	res, err := h.service.CreateInstitution(c.Request().Context(), req)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, model.DataResponse[*model.InstitutionResponse]{Data: res})
}

func (h *MasterHandler) UpdateInstitution(c echo.Context) error {
	id, err := parseIDParam(c, "id")
	if err != nil {
		return err
	}
	var req model.InstitutionRequest
	if err := bind(c, &req); err != nil {
		return err
	}
	res, err := h.service.UpdateInstitution(c.Request().Context(), id, req)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, model.DataResponse[*model.InstitutionResponse]{Data: res})
}

func (h *MasterHandler) DeleteInstitution(c echo.Context) error {
	id, err := parseIDParam(c, "id")
	if err != nil {
		return err
	}
	if err := h.service.DeleteInstitution(c.Request().Context(), id); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *MasterHandler) ListRegions(c echo.Context) error {
	res, err := h.service.ListRegions(c.Request().Context(), paginationParams(c), queryStrings(c, "type"), c.QueryParam("parent_code"))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, res)
}

func (h *MasterHandler) LookupRegions(c echo.Context) error {
	res, err := h.service.LookupRegions(c.Request().Context(), paginationParams(c), queryStrings(c, "type"), c.QueryParam("parent_code"))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, res)
}

func (h *MasterHandler) GetRegion(c echo.Context) error {
	id, err := parseIDParam(c, "id")
	if err != nil {
		return err
	}
	res, err := h.service.GetRegion(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, model.DataResponse[*model.RegionResponse]{Data: res})
}

func (h *MasterHandler) CreateRegion(c echo.Context) error {
	var req model.RegionRequest
	if err := bind(c, &req); err != nil {
		return err
	}
	res, err := h.service.CreateRegion(c.Request().Context(), req)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, model.DataResponse[*model.RegionResponse]{Data: res})
}

func (h *MasterHandler) UpdateRegion(c echo.Context) error {
	id, err := parseIDParam(c, "id")
	if err != nil {
		return err
	}
	var req model.RegionRequest
	if err := bind(c, &req); err != nil {
		return err
	}
	res, err := h.service.UpdateRegion(c.Request().Context(), id, req)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, model.DataResponse[*model.RegionResponse]{Data: res})
}

func (h *MasterHandler) DeleteRegion(c echo.Context) error {
	id, err := parseIDParam(c, "id")
	if err != nil {
		return err
	}
	if err := h.service.DeleteRegion(c.Request().Context(), id); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *MasterHandler) ListProgramTitles(c echo.Context) error {
	res, err := h.service.ListProgramTitles(c.Request().Context(), paginationParams(c), queryStringPtr(c, "parent_id"))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, res)
}

func (h *MasterHandler) LookupProgramTitles(c echo.Context) error {
	res, err := h.service.LookupProgramTitles(c.Request().Context(), paginationParams(c))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, res)
}

func (h *MasterHandler) GetProgramTitle(c echo.Context) error {
	id, err := parseIDParam(c, "id")
	if err != nil {
		return err
	}
	res, err := h.service.GetProgramTitle(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, model.DataResponse[*model.ProgramTitleResponse]{Data: res})
}

func (h *MasterHandler) CreateProgramTitle(c echo.Context) error {
	var req model.ProgramTitleRequest
	if err := bind(c, &req); err != nil {
		return err
	}
	res, err := h.service.CreateProgramTitle(c.Request().Context(), req)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, model.DataResponse[*model.ProgramTitleResponse]{Data: res})
}

func (h *MasterHandler) UpdateProgramTitle(c echo.Context) error {
	id, err := parseIDParam(c, "id")
	if err != nil {
		return err
	}
	var req model.ProgramTitleRequest
	if err := bind(c, &req); err != nil {
		return err
	}
	res, err := h.service.UpdateProgramTitle(c.Request().Context(), id, req)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, model.DataResponse[*model.ProgramTitleResponse]{Data: res})
}

func (h *MasterHandler) DeleteProgramTitle(c echo.Context) error {
	id, err := parseIDParam(c, "id")
	if err != nil {
		return err
	}
	if err := h.service.DeleteProgramTitle(c.Request().Context(), id); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *MasterHandler) ListBappenasPartners(c echo.Context) error {
	res, err := h.service.ListBappenasPartners(c.Request().Context(), paginationParams(c), queryStrings(c, "level"), queryStringPtr(c, "parent_id"))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, res)
}

func (h *MasterHandler) LookupBappenasPartners(c echo.Context) error {
	res, err := h.service.LookupBappenasPartners(c.Request().Context(), paginationParams(c), queryStrings(c, "level"))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, res)
}

func (h *MasterHandler) GetBappenasPartner(c echo.Context) error {
	id, err := parseIDParam(c, "id")
	if err != nil {
		return err
	}
	res, err := h.service.GetBappenasPartner(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, model.DataResponse[*model.BappenasPartnerResponse]{Data: res})
}

func (h *MasterHandler) CreateBappenasPartner(c echo.Context) error {
	var req model.BappenasPartnerRequest
	if err := bind(c, &req); err != nil {
		return err
	}
	res, err := h.service.CreateBappenasPartner(c.Request().Context(), req)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, model.DataResponse[*model.BappenasPartnerResponse]{Data: res})
}

func (h *MasterHandler) UpdateBappenasPartner(c echo.Context) error {
	id, err := parseIDParam(c, "id")
	if err != nil {
		return err
	}
	var req model.BappenasPartnerRequest
	if err := bind(c, &req); err != nil {
		return err
	}
	res, err := h.service.UpdateBappenasPartner(c.Request().Context(), id, req)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, model.DataResponse[*model.BappenasPartnerResponse]{Data: res})
}

func (h *MasterHandler) DeleteBappenasPartner(c echo.Context) error {
	id, err := parseIDParam(c, "id")
	if err != nil {
		return err
	}
	if err := h.service.DeleteBappenasPartner(c.Request().Context(), id); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *MasterHandler) ListPeriods(c echo.Context) error {
	res, err := h.service.ListPeriods(c.Request().Context(), paginationParams(c))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, res)
}

func (h *MasterHandler) GetPeriod(c echo.Context) error {
	id, err := parseIDParam(c, "id")
	if err != nil {
		return err
	}
	res, err := h.service.GetPeriod(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, model.DataResponse[*model.PeriodResponse]{Data: res})
}

func (h *MasterHandler) CreatePeriod(c echo.Context) error {
	var req model.PeriodRequest
	if err := bind(c, &req); err != nil {
		return err
	}
	res, err := h.service.CreatePeriod(c.Request().Context(), req)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, model.DataResponse[*model.PeriodResponse]{Data: res})
}

func (h *MasterHandler) UpdatePeriod(c echo.Context) error {
	id, err := parseIDParam(c, "id")
	if err != nil {
		return err
	}
	var req model.PeriodRequest
	if err := bind(c, &req); err != nil {
		return err
	}
	res, err := h.service.UpdatePeriod(c.Request().Context(), id, req)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, model.DataResponse[*model.PeriodResponse]{Data: res})
}

func (h *MasterHandler) DeletePeriod(c echo.Context) error {
	id, err := parseIDParam(c, "id")
	if err != nil {
		return err
	}
	if err := h.service.DeletePeriod(c.Request().Context(), id); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *MasterHandler) ListNationalPriorities(c echo.Context) error {
	res, err := h.service.ListNationalPriorities(c.Request().Context(), paginationParams(c), queryStrings(c, "period_id"))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, res)
}

func (h *MasterHandler) GetNationalPriority(c echo.Context) error {
	id, err := parseIDParam(c, "id")
	if err != nil {
		return err
	}
	res, err := h.service.GetNationalPriority(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, model.DataResponse[*model.NationalPriorityResponse]{Data: res})
}

func (h *MasterHandler) CreateNationalPriority(c echo.Context) error {
	var req model.NationalPriorityRequest
	if err := bind(c, &req); err != nil {
		return err
	}
	res, err := h.service.CreateNationalPriority(c.Request().Context(), req)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, model.DataResponse[*model.NationalPriorityResponse]{Data: res})
}

func (h *MasterHandler) UpdateNationalPriority(c echo.Context) error {
	id, err := parseIDParam(c, "id")
	if err != nil {
		return err
	}
	var req model.NationalPriorityRequest
	if err := bind(c, &req); err != nil {
		return err
	}
	res, err := h.service.UpdateNationalPriority(c.Request().Context(), id, req)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, model.DataResponse[*model.NationalPriorityResponse]{Data: res})
}

func (h *MasterHandler) DeleteNationalPriority(c echo.Context) error {
	id, err := parseIDParam(c, "id")
	if err != nil {
		return err
	}
	if err := h.service.DeleteNationalPriority(c.Request().Context(), id); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *MasterHandler) ImportData(c echo.Context) error {
	return h.handleImportData(c, false)
}

func (h *MasterHandler) PreviewImportData(c echo.Context) error {
	return h.handleImportData(c, true)
}

func (h *MasterHandler) DownloadImportTemplate(c echo.Context) error {
	template, err := h.service.BuildMasterImportTemplate(c.Request().Context())
	if err != nil {
		return err
	}

	c.Response().Header().Set(echo.HeaderContentDisposition, `attachment; filename="`+template.FileName+`"`)
	return c.Blob(http.StatusOK, template.ContentType, template.Data)
}

func (h *MasterHandler) handleImportData(c echo.Context, preview bool) error {
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
		res, err = h.service.PreviewMasterData(c.Request().Context(), file.Filename, src, file.Size)
	} else {
		res, err = h.service.ImportMasterData(c.Request().Context(), file.Filename, src, file.Size)
	}
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, model.DataResponse[*model.MasterImportResponse]{Data: res})
}

func paginationParams(c echo.Context) model.PaginationParams {
	return model.PaginationParams{
		Page:   parseIntQuery(c, "page", 1),
		Limit:  parseIntQuery(c, "limit", 20),
		Sort:   c.QueryParam("sort"),
		Order:  c.QueryParam("order"),
		Search: c.QueryParam("search"),
	}
}

func queryStringPtr(c echo.Context, name string) *string {
	value := c.QueryParam(name)
	if value == "" {
		return nil
	}
	return &value
}

func queryStrings(c echo.Context, name string) []string {
	values := append([]string{}, c.QueryParams()[name]...)
	values = append(values, c.QueryParams()[name+"[]"]...)
	filters := make([]string, 0, len(values))

	for _, rawValue := range values {
		for _, part := range strings.Split(rawValue, ",") {
			value := strings.TrimSpace(part)
			if value != "" {
				filters = append(filters, value)
			}
		}
	}

	return filters
}

func bind(c echo.Context, dst any) error {
	if err := c.Bind(dst); err != nil {
		return apperrors.Validation(apperrors.FieldError{Field: "body", Message: "JSON tidak valid"})
	}
	return nil
}
