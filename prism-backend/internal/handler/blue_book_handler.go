package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"

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
	res, err := h.service.ListBlueBooks(c.Request().Context(), paginationParams(c))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, res)
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
	res, err := h.service.ListBBProjects(c.Request().Context(), bbID, paginationParams(c))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, res)
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
	if err := h.service.DeleteBBProject(c.Request().Context(), bbID, id); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
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
