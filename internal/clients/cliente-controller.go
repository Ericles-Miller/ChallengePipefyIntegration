package clients

import (
	"errors"
	"net/http"
	"github.com/Ericles-Miller/ChallengePipefyIntegration/internal/clients/models"
	"github.com/Ericles-Miller/ChallengePipefyIntegration/internal/clients/services"
	"github.com/Ericles-Miller/ChallengePipefyIntegration/pkg"
	AppError "github.com/Ericles-Miller/ChallengePipefyIntegration/pkg/appError"
	"github.com/gin-gonic/gin"
)

type ClientController struct {
	service services.ClientService
}

func NewClientController(service services.ClientService) *ClientController {
	return &ClientController{service: service}
}

func (c *ClientController) RegisterRoutes(r *gin.Engine) {
	clientes := r.Group("/clientes")
	
	{
		clientes.POST("", c.CreateClient)
	}
}

// CreateClient godoc
// @Summary     Create a new client
// @Description Creates a client and maps it to a Pipefy card
// @Tags        clients
// @Accept      json
// @Produce     json
// @Param       request body     models.CreateClientRequest          true "Client data"
// @Success     201     {object} pkg.Response[models.ClientResponse] "Created"
// @Failure     400     {object} pkg.Response[models.ClientResponse] "Bad request"
// @Failure     404     {object} pkg.Response[models.ClientResponse] "Not found"
// @Failure     500     {object} pkg.Response[models.ClientResponse] "Internal server error"
// @Router      /clientes [post]
func (c *ClientController) CreateClient(ctx *gin.Context) {
	var req models.CreateClientRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, pkg.Fail[*models.ClientResponse](pkg.ParseValidationErrors(err)))
		return
	}

	client, err := c.service.CreateClient(ctx.Request.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, AppError.ErrBadRequest):
			ctx.JSON(http.StatusBadRequest, pkg.Fail[*models.ClientResponse](err.Error()))
		case errors.Is(err, AppError.ErrNotFound):
			ctx.JSON(http.StatusNotFound, pkg.Fail[*models.ClientResponse](err.Error()))
		default:
			ctx.JSON(http.StatusInternalServerError, pkg.Fail[*models.ClientResponse](err.Error()))
		}
		return
	}

	ctx.JSON(http.StatusCreated, pkg.Success(client))
}
