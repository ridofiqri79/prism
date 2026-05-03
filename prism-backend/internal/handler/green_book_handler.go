package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"

	apperrors "github.com/ridofiqri79/prism-backend/internal/errors"
	"github.com/ridofiqri79/prism-backend/internal/model"
	"github.com/ridofiqri79/prism-backend/internal/service"
)

type GreenBookHandler struct {
	service *service.GreenBookService
}

func NewGreenBookHandler(service *service.GreenBookService) *GreenBookHandler {
	return &GreenBookHandler{service: service}
}

func (h *GreenBookHandler) ListGreenBooks(c echo.Context) error {
	res, err := h.service.ListGreenBooks(c.Request().Context(), greenBookListFilter(c), paginationParams(c))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, res)
}

func greenBookListFilter(c echo.Context) model.GreenBookListFilter {
	return model.GreenBookListFilter{
		PublishYears: queryStrings(c, "publish_year"),
		Statuses:     queryStrings(c, "status"),
	}
}

func (h *GreenBookHandler) GetGreenBook(c echo.Context) error {
	id, err := parseIDParam(c, "id")
	if err != nil {
		return err
	}
	res, err := h.service.GetGreenBook(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, model.DataResponse[*model.GreenBookResponse]{Data: res})
}

func (h *GreenBookHandler) CreateGreenBook(c echo.Context) error {
	var req model.GreenBookRequest
	if err := bind(c, &req); err != nil {
		return err
	}
	res, err := h.service.CreateGreenBook(c.Request().Context(), req)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, model.DataResponse[*model.GreenBookResponse]{Data: res})
}

func (h *GreenBookHandler) UpdateGreenBook(c echo.Context) error {
	id, err := parseIDParam(c, "id")
	if err != nil {
		return err
	}
	var req model.GreenBookRequest
	if err := bind(c, &req); err != nil {
		return err
	}
	res, err := h.service.UpdateGreenBook(c.Request().Context(), id, req)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, model.DataResponse[*model.GreenBookResponse]{Data: res})
}

func (h *GreenBookHandler) DeleteGreenBook(c echo.Context) error {
	id, err := parseIDParam(c, "id")
	if err != nil {
		return err
	}
	if err := h.service.DeleteGreenBook(c.Request().Context(), id); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *GreenBookHandler) ListGBProjects(c echo.Context) error {
	gbID, err := parseIDParam(c, "gbId")
	if err != nil {
		return err
	}
	res, err := h.service.ListGBProjects(c.Request().Context(), gbID, gbProjectListFilter(c), paginationParams(c))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, res)
}

func gbProjectListFilter(c echo.Context) model.GBProjectListFilter {
	return model.GBProjectListFilter{
		BBProjectIDs:       queryStrings(c, "bb_project_ids"),
		ExecutingAgencyIDs: queryStrings(c, "executing_agency_ids"),
		LocationIDs:        queryStrings(c, "location_ids"),
		Statuses:           queryStrings(c, "status"),
	}
}

func (h *GreenBookHandler) GetGBProject(c echo.Context) error {
	gbID, err := parseIDParam(c, "gbId")
	if err != nil {
		return err
	}
	id, err := parseIDParam(c, "id")
	if err != nil {
		return err
	}
	res, err := h.service.GetGBProject(c.Request().Context(), gbID, id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, model.DataResponse[*model.GBProjectResponse]{Data: res})
}

func (h *GreenBookHandler) CreateGBProject(c echo.Context) error {
	gbID, err := parseIDParam(c, "gbId")
	if err != nil {
		return err
	}
	var req model.CreateGBProjectRequest
	if err := bind(c, &req); err != nil {
		return err
	}
	res, err := h.service.CreateGBProject(c.Request().Context(), gbID, req)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, model.DataResponse[*model.GBProjectResponse]{Data: res})
}

func (h *GreenBookHandler) UpdateGBProject(c echo.Context) error {
	gbID, err := parseIDParam(c, "gbId")
	if err != nil {
		return err
	}
	id, err := parseIDParam(c, "id")
	if err != nil {
		return err
	}
	var req model.UpdateGBProjectRequest
	if err := bind(c, &req); err != nil {
		return err
	}
	res, err := h.service.UpdateGBProject(c.Request().Context(), gbID, id, req)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, model.DataResponse[*model.GBProjectResponse]{Data: res})
}

func (h *GreenBookHandler) DeleteGBProject(c echo.Context) error {
	gbID, err := parseIDParam(c, "gbId")
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
	if err := h.service.DeleteGBProject(c.Request().Context(), gbID, id, user); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *GreenBookHandler) GetGBProjectHistory(c echo.Context) error {
	id, err := parseIDParam(c, "gbProjectId")
	if err != nil {
		return err
	}
	user, _ := c.Get("user").(*model.AuthUser)
	res, err := h.service.GetGBProjectHistoryWithAudit(c.Request().Context(), id, user != nil && user.Role == "ADMIN")
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, model.DataResponse[[]model.GBProjectHistoryItem]{Data: res})
}

func (h *GreenBookHandler) ImportGBProjects(c echo.Context) error {
	return h.handleImportGBProjects(c, false)
}

func (h *GreenBookHandler) PreviewImportGBProjects(c echo.Context) error {
	return h.handleImportGBProjects(c, true)
}

func (h *GreenBookHandler) DownloadGBProjectImportTemplate(c echo.Context) error {
	gbID, err := parseIDParam(c, "gbId")
	if err != nil {
		return err
	}

	template, err := h.service.BuildProjectImportTemplate(c.Request().Context(), gbID)
	if err != nil {
		return err
	}

	c.Response().Header().Set(echo.HeaderContentDisposition, `attachment; filename="`+template.FileName+`"`)
	return c.Blob(http.StatusOK, template.ContentType, template.Data)
}

func (h *GreenBookHandler) handleImportGBProjects(c echo.Context, preview bool) error {
	gbID, err := parseIDParam(c, "gbId")
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
		res, err = h.service.PreviewGreenBookProjects(c.Request().Context(), gbID, file.Filename, src, file.Size)
	} else {
		res, err = h.service.ImportGreenBookProjects(c.Request().Context(), gbID, file.Filename, src, file.Size)
	}
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, model.DataResponse[*model.MasterImportResponse]{Data: res})
}
