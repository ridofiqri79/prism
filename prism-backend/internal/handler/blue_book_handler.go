package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"

	apperrors "github.com/ridofiqri79/prism-backend/internal/errors"
	"github.com/ridofiqri79/prism-backend/internal/model"
	"github.com/ridofiqri79/prism-backend/internal/service"
)

type BlueBookHandler struct {
	service *service.BlueBookService
}

func NewBlueBookHandler(service *service.BlueBookService) *BlueBookHandler {
	return &BlueBookHandler{service: service}
}

func (h *BlueBookHandler) ListBlueBooks(c echo.Context) error {
	res, err := h.service.ListBlueBooks(c.Request().Context(), blueBookListFilter(c), paginationParams(c))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, res)
}

func blueBookListFilter(c echo.Context) model.BlueBookListFilter {
	return model.BlueBookListFilter{
		PeriodIDs: queryStrings(c, "period_id"),
		Statuses:  queryStrings(c, "status"),
	}
}

func (h *BlueBookHandler) GetBlueBook(c echo.Context) error {
	id, err := parseIDParam(c, "id")
	if err != nil {
		return err
	}
	res, err := h.service.GetBlueBook(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, model.DataResponse[*model.BlueBookResponse]{Data: res})
}

func (h *BlueBookHandler) CreateBlueBook(c echo.Context) error {
	var req model.BlueBookRequest
	if err := bind(c, &req); err != nil {
		return err
	}
	res, err := h.service.CreateBlueBook(c.Request().Context(), req)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, model.DataResponse[*model.BlueBookResponse]{Data: res})
}

func (h *BlueBookHandler) UpdateBlueBook(c echo.Context) error {
	id, err := parseIDParam(c, "id")
	if err != nil {
		return err
	}
	var req model.BlueBookRequest
	if err := bind(c, &req); err != nil {
		return err
	}
	res, err := h.service.UpdateBlueBook(c.Request().Context(), id, req)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, model.DataResponse[*model.BlueBookResponse]{Data: res})
}

func (h *BlueBookHandler) DeleteBlueBook(c echo.Context) error {
	id, err := parseIDParam(c, "id")
	if err != nil {
		return err
	}
	if err := h.service.DeleteBlueBook(c.Request().Context(), id); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *BlueBookHandler) ListBBProjects(c echo.Context) error {
	bbID, err := parseIDParam(c, "bbId")
	if err != nil {
		return err
	}
	res, err := h.service.ListBBProjects(c.Request().Context(), bbID, bbProjectListFilter(c), paginationParams(c))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, res)
}

func bbProjectListFilter(c echo.Context) model.BBProjectListFilter {
	return model.BBProjectListFilter{
		ExecutingAgencyIDs: queryStrings(c, "executing_agency_ids"),
		LocationIDs:        queryStrings(c, "location_ids"),
	}
}

func (h *BlueBookHandler) GetBBProject(c echo.Context) error {
	bbID, err := parseIDParam(c, "bbId")
	if err != nil {
		return err
	}
	id, err := parseIDParam(c, "id")
	if err != nil {
		return err
	}
	res, err := h.service.GetBBProject(c.Request().Context(), bbID, id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, model.DataResponse[*model.BBProjectResponse]{Data: res})
}

func (h *BlueBookHandler) CreateBBProject(c echo.Context) error {
	bbID, err := parseIDParam(c, "bbId")
	if err != nil {
		return err
	}
	var req model.CreateBBProjectRequest
	if err := bind(c, &req); err != nil {
		return err
	}
	res, err := h.service.CreateBBProject(c.Request().Context(), bbID, req)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, model.DataResponse[*model.BBProjectResponse]{Data: res})
}

func (h *BlueBookHandler) UpdateBBProject(c echo.Context) error {
	bbID, err := parseIDParam(c, "bbId")
	if err != nil {
		return err
	}
	id, err := parseIDParam(c, "id")
	if err != nil {
		return err
	}
	var req model.UpdateBBProjectRequest
	if err := bind(c, &req); err != nil {
		return err
	}
	res, err := h.service.UpdateBBProject(c.Request().Context(), bbID, id, req)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, model.DataResponse[*model.BBProjectResponse]{Data: res})
}

func (h *BlueBookHandler) DeleteBBProject(c echo.Context) error {
	bbID, err := parseIDParam(c, "bbId")
	if err != nil {
		return err
	}
	id, err := parseIDParam(c, "id")
	if err != nil {
		return err
	}
	user, ok := c.Get("user").(*model.AuthUser)
	if !ok || user == nil {
		return apperrors.Unauthorized("User tidak ditemukan")
	}
	if err := h.service.DeleteBBProject(c.Request().Context(), bbID, id, user); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *BlueBookHandler) GetBBProjectHistory(c echo.Context) error {
	id, err := parseIDParam(c, "bbProjectId")
	if err != nil {
		return err
	}
	user, _ := c.Get("user").(*model.AuthUser)
	res, err := h.service.GetBBProjectHistoryWithAudit(c.Request().Context(), id, user != nil && user.Role == "ADMIN")
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, model.DataResponse[[]model.BBProjectHistoryItem]{Data: res})
}

func (h *BlueBookHandler) ImportBBProjects(c echo.Context) error {
	return h.handleImportBBProjects(c, false)
}

func (h *BlueBookHandler) PreviewImportBBProjects(c echo.Context) error {
	return h.handleImportBBProjects(c, true)
}

func (h *BlueBookHandler) DownloadBBProjectImportTemplate(c echo.Context) error {
	bbID, err := parseIDParam(c, "bbId")
	if err != nil {
		return err
	}

	template, err := h.service.BuildProjectImportTemplate(c.Request().Context(), bbID)
	if err != nil {
		return err
	}

	c.Response().Header().Set(echo.HeaderContentDisposition, `attachment; filename="`+template.FileName+`"`)
	return c.Blob(http.StatusOK, template.ContentType, template.Data)
}

func (h *BlueBookHandler) handleImportBBProjects(c echo.Context, preview bool) error {
	bbID, err := parseIDParam(c, "bbId")
	if err != nil {
		return err
	}

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
		res, err = h.service.PreviewBlueBookProjects(c.Request().Context(), bbID, file.Filename, src, file.Size)
	} else {
		res, err = h.service.ImportBlueBookProjects(c.Request().Context(), bbID, file.Filename, src, file.Size)
	}
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, model.DataResponse[*model.MasterImportResponse]{Data: res})
}

func (h *BlueBookHandler) ListLoI(c echo.Context) error {
	bbProjectID, err := parseIDParam(c, "bbProjectId")
	if err != nil {
		return err
	}
	res, err := h.service.ListLoI(c.Request().Context(), bbProjectID)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, model.DataResponse[[]model.LoIResponse]{Data: res})
}

func (h *BlueBookHandler) CreateLoI(c echo.Context) error {
	bbProjectID, err := parseIDParam(c, "bbProjectId")
	if err != nil {
		return err
	}
	var req model.LoIRequest
	if err := bind(c, &req); err != nil {
		return err
	}
	res, err := h.service.CreateLoI(c.Request().Context(), bbProjectID, req)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, model.DataResponse[*model.LoIResponse]{Data: res})
}

func (h *BlueBookHandler) DeleteLoI(c echo.Context) error {
	bbProjectID, err := parseIDParam(c, "bbProjectId")
	if err != nil {
		return err
	}
	id, err := parseIDParam(c, "id")
	if err != nil {
		return err
	}
	if err := h.service.DeleteLoI(c.Request().Context(), bbProjectID, id); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}
