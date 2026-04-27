package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"

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
