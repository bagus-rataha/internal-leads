package handlers

import (
	"errors"
	"fmt"
	"time"

	"fiber-api-boilerplate/internal/repository"
	"fiber-api-boilerplate/internal/services"
	"fiber-api-boilerplate/internal/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
)

// exportService is the business-logic contract ExportHandler depends on.
// Defined consumer-side so it can be satisfied by the real service or a mock.
type exportService interface {
	Export(callerID uuid.UUID, role, password string, filter repository.LeadFilter) (*excelize.File, error)
}

type ExportHandler struct {
	exportService exportService
}

func NewExportHandler(exportService exportService) *ExportHandler {
	return &ExportHandler{exportService: exportService}
}

// ExportLeads godoc
// @Summary Export leads to Excel
// @Tags leads
// @Security BearerAuth
// @Param X-Export-Password header string true "Caller's own account password"
// @Success 200 {file} binary
// @Router /leads/export [get]
func (h *ExportHandler) ExportLeads(c *fiber.Ctx) error {
	userID, ok := utils.GetUserID(c)
	if !ok {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, "Invalid session")
	}
	role, _ := c.Locals("role").(string)

	password := c.Get("X-Export-Password")
	if password == "" {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, "invalid credentials")
	}

	query, err := parseLeadListQuery(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	filter := repository.LeadFilter{
		Q:            query.Q,
		Status:       query.Status,
		SourceID:     query.SourceID,
		TeamID:       query.TeamID,
		OwnerID:      query.OwnerID,
		ProvinceID:   query.ProvinceID,
		CityID:       query.CityID,
		DateFrom:     query.DateFrom,
		DateTo:       query.DateTo,
		FollowUpFrom: query.FollowUpFrom,
		FollowUpTo:   query.FollowUpTo,
		Stale:        query.Stale,
	}

	workbook, err := h.exportService.Export(userID, role, password, filter)
	if err != nil {
		var tooMany *services.ErrExportTooManyRows
		if errors.As(err, &tooMany) {
			return utils.ErrorResponseWithData(c, fiber.StatusUnprocessableEntity,
				"Hasil melebihi batas export, persempit filter", fiber.Map{"row_count": tooMany.Count})
		}
		if errors.Is(err, services.ErrExportInvalidPassword) {
			return utils.ErrorResponse(c, fiber.StatusUnauthorized, "invalid credentials")
		}
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	filename := fmt.Sprintf("leads-export-%s.xlsx", time.Now().Format("20060102-150405"))
	c.Set(fiber.HeaderContentType, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Set(fiber.HeaderContentDisposition, fmt.Sprintf(`attachment; filename="%s"`, filename))
	return workbook.Write(c.Response().BodyWriter())
}
