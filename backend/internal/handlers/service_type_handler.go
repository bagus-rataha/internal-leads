package handlers

import (
	"fiber-api-boilerplate/internal/dto"
	"fiber-api-boilerplate/internal/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// serviceTypeService is the business-logic contract ServiceTypeHandler
// depends on. Defined consumer-side so it can be satisfied by the real
// service or a mock.
type serviceTypeService interface {
	Create(input dto.CreateServiceTypeInput) (*dto.ServiceTypeAdminResponse, error)
	List() ([]dto.ServiceTypeAdminResponse, error)
	Update(id uuid.UUID, input dto.UpdateServiceTypeInput) (*dto.ServiceTypeAdminResponse, error)
}

type ServiceTypeHandler struct {
	serviceTypeService serviceTypeService
}

func NewServiceTypeHandler(serviceTypeService serviceTypeService) *ServiceTypeHandler {
	return &ServiceTypeHandler{serviceTypeService: serviceTypeService}
}

// CreateServiceType godoc
// @Summary Create service type (admin only)
// @Tags service-types
// @Security BearerAuth
// @Param request body dto.CreateServiceTypeInput true "Create service type"
// @Success 201 {object} utils.Response{data=dto.ServiceTypeAdminResponse}
// @Router /service-types [post]
func (h *ServiceTypeHandler) CreateServiceType(c *fiber.Ctx) error {
	var input dto.CreateServiceTypeInput
	if err := utils.ParseAndValidate(c, &input); err != nil {
		return err
	}

	serviceType, err := h.serviceTypeService.Create(input)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusCreated, "Service type created successfully", serviceType)
}

// ListServiceTypesAdmin godoc
// @Summary List all service types, active and inactive (admin only)
// @Tags service-types
// @Security BearerAuth
// @Success 200 {object} utils.Response{data=[]dto.ServiceTypeAdminResponse}
// @Router /service-types [get]
func (h *ServiceTypeHandler) ListServiceTypesAdmin(c *fiber.Ctx) error {
	types, err := h.serviceTypeService.List()
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Service types retrieved successfully", types)
}

// UpdateServiceType godoc
// @Summary Update service type (admin only)
// @Tags service-types
// @Security BearerAuth
// @Param id path string true "Service type ID"
// @Param request body dto.UpdateServiceTypeInput true "Update service type"
// @Success 200 {object} utils.Response{data=dto.ServiceTypeAdminResponse}
// @Router /service-types/{id} [patch]
func (h *ServiceTypeHandler) UpdateServiceType(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid service type id")
	}

	var input dto.UpdateServiceTypeInput
	if err := utils.ParseAndValidate(c, &input); err != nil {
		return err
	}

	serviceType, err := h.serviceTypeService.Update(id, input)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Service type updated successfully", serviceType)
}
