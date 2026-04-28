package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"

	apperrors "github.com/ridofiqri79/prism-backend/internal/errors"
	"github.com/ridofiqri79/prism-backend/internal/model"
	"github.com/ridofiqri79/prism-backend/internal/service"
)

type DKHandler struct {
	service *service.DKService
}

func NewDKHandler(service *service.DKService) *DKHandler {
	return &DKHandler{service: service}
}

func (h *DKHandler) ListDK(c echo.Context) error {
	res, err := h.service.ListDaftarKegiatan(c.Request().Context(), paginationParams(c))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, res)
}

func (h *DKHandler) GetDK(c echo.Context) error {
	id, err := parseIDParam(c, "id")
	if err != nil {
		return err
	}
	res, err := h.service.GetDaftarKegiatan(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, model.DataResponse[*model.DaftarKegiatanResponse]{Data: res})
}

func (h *DKHandler) CreateDK(c echo.Context) error {
	var req model.DaftarKegiatanRequest
	if err := bind(c, &req); err != nil {
		return err
	}
	res, err := h.service.CreateDaftarKegiatan(c.Request().Context(), req)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, model.DataResponse[*model.DaftarKegiatanResponse]{Data: res})
}

func (h *DKHandler) UpdateDK(c echo.Context) error {
	id, err := parseIDParam(c, "id")
	if err != nil {
		return err
	}
	var req model.DaftarKegiatanRequest
	if err := bind(c, &req); err != nil {
		return err
	}
	res, err := h.service.UpdateDaftarKegiatan(c.Request().Context(), id, req)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, model.DataResponse[*model.DaftarKegiatanResponse]{Data: res})
}

func (h *DKHandler) DeleteDK(c echo.Context) error {
	id, err := parseIDParam(c, "id")
	if err != nil {
		return err
	}
	if err := h.service.DeleteDaftarKegiatan(c.Request().Context(), id); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *DKHandler) DownloadDKImportTemplate(c echo.Context) error {
	template, err := h.service.BuildImportTemplate(c.Request().Context())
	if err != nil {
		return err
	}

	c.Response().Header().Set(echo.HeaderContentDisposition, `attachment; filename="`+template.FileName+`"`)
	return c.Blob(http.StatusOK, template.ContentType, template.Data)
}

func (h *DKHandler) PreviewImportDK(c echo.Context) error {
	return h.handleImportDK(c, true)
}

func (h *DKHandler) ImportDK(c echo.Context) error {
	return h.handleImportDK(c, false)
}

func (h *DKHandler) ListDKProjects(c echo.Context) error {
	dkID, err := parseIDParam(c, "dkId")
	if err != nil {
		return err
	}
	res, err := h.service.ListDKProjects(c.Request().Context(), dkID, paginationParams(c))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, res)
}

func (h *DKHandler) GetDKProject(c echo.Context) error {
	dkID, err := parseIDParam(c, "dkId")
	if err != nil {
		return err
	}
	id, err := parseIDParam(c, "id")
	if err != nil {
		return err
	}
	res, err := h.service.GetDKProject(c.Request().Context(), dkID, id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, model.DataResponse[*model.DKProjectResponse]{Data: res})
}

func (h *DKHandler) CreateDKProject(c echo.Context) error {
	dkID, err := parseIDParam(c, "dkId")
	if err != nil {
		return err
	}
	var req model.CreateDKProjectRequest
	if err := bind(c, &req); err != nil {
		return err
	}
	res, err := h.service.CreateDKProject(c.Request().Context(), dkID, req)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, model.DataResponse[*model.DKProjectResponse]{Data: res})
}

func (h *DKHandler) UpdateDKProject(c echo.Context) error {
	dkID, err := parseIDParam(c, "dkId")
	if err != nil {
		return err
	}
	id, err := parseIDParam(c, "id")
	if err != nil {
		return err
	}
	var req model.UpdateDKProjectRequest
	if err := bind(c, &req); err != nil {
		return err
	}
	res, err := h.service.UpdateDKProject(c.Request().Context(), dkID, id, req)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, model.DataResponse[*model.DKProjectResponse]{Data: res})
}

func (h *DKHandler) DeleteDKProject(c echo.Context) error {
	dkID, err := parseIDParam(c, "dkId")
	if err != nil {
		return err
	}
	id, err := parseIDParam(c, "id")
	if err != nil {
		return err
	}
	if err := h.service.DeleteDKProject(c.Request().Context(), dkID, id); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *DKHandler) handleImportDK(c echo.Context, preview bool) error {
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
		res, err = h.service.PreviewDaftarKegiatanImport(c.Request().Context(), file.Filename, src, file.Size)
	} else {
		res, err = h.service.ImportDaftarKegiatan(c.Request().Context(), file.Filename, src, file.Size)
	}
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, model.DataResponse[*model.MasterImportResponse]{Data: res})
}
