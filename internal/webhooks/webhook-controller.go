package webhooks

import (
	"errors"
	"net/http"

	webhookModels "github.com/Ericles-Miller/ChallengePipefyIntegration/internal/webhooks/models"
	"github.com/Ericles-Miller/ChallengePipefyIntegration/internal/webhooks/services"
	"github.com/Ericles-Miller/ChallengePipefyIntegration/pkg"
	AppError "github.com/Ericles-Miller/ChallengePipefyIntegration/pkg/appError"
	"github.com/gin-gonic/gin"
)

type WebhookController struct {
	service services.WebhookService
}

func NewWebhookController(service services.WebhookService) *WebhookController {
	return &WebhookController{service: service}
}

func (c *WebhookController) RegisterRoutes(r *gin.Engine) {
	webhooks := r.Group("/webhooks/pipefy")
	{
		webhooks.POST("/card-updated", c.CardUpdated)
	}
}

// CardUpdated godoc
// @Summary     Process Pipefy card update event
// @Description Receives a webhook event from Pipefy, applies priority rules and updates the client status
// @Tags        webhooks
// @Accept      json
// @Produce     json
// @Param       request body     webhookModels.WebhookEventRequest          true "Webhook event data"
// @Success     200     {object} pkg.Response[webhookModels.WebhookEventResponse] "Processed"
// @Failure     400     {object} pkg.Response[webhookModels.WebhookEventResponse] "Bad request or duplicate event"
// @Failure     404     {object} pkg.Response[webhookModels.WebhookEventResponse] "Client not found"
// @Failure     500     {object} pkg.Response[webhookModels.WebhookEventResponse] "Internal server error"
// @Router      /webhooks/pipefy/card-updated [post]
func (c *WebhookController) CardUpdated(ctx *gin.Context) {
	var req webhookModels.WebhookEventRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, pkg.Fail[*webhookModels.WebhookEventResponse](pkg.ParseValidationErrors(err)))
		return
	}

	response, err := c.service.ProcessEvent(ctx.Request.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, AppError.ErrBadRequest):
			ctx.JSON(http.StatusBadRequest, pkg.Fail[*webhookModels.WebhookEventResponse](err.Error()))
		case errors.Is(err, AppError.ErrNotFound):
			ctx.JSON(http.StatusNotFound, pkg.Fail[*webhookModels.WebhookEventResponse](err.Error()))
		default:
			ctx.JSON(http.StatusInternalServerError, pkg.Fail[*webhookModels.WebhookEventResponse](err.Error()))
		}
		return
	}

	ctx.JSON(http.StatusOK, pkg.Success(response))
}
