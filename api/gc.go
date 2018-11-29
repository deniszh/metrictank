package api

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/grafana/metrictank/api/middleware"
	"github.com/grafana/metrictank/api/models"
	"github.com/grafana/metrictank/api/response"
	"github.com/grafana/metrictank/kafkamdm"
)

func (s *Server) getStartupGCPercent(ctx *middleware.Context) {
	response.Write(ctx, response.NewJson(200, models.StartupGCPercent{Value: kafkamdm.StartupGCPercent}, ""))
}

func (s *Server) getNormalGCPercent(ctx *middleware.Context) {
	response.Write(ctx, response.NewJson(200, models.NormalGCPercent{Value: kafkamdm.NormalGCPercent}, ""))
}

func (s *Server) setStartupGCPercent(ctx *middleware.Context, percent models.StartupGCPercent) {
	value, err := strconv.ParseInt(percent.value, 10, 32)
	if err != nil {
		response.Write(ctx, response.NewError(http.StatusBadRequest, fmt.Sprintf(
			"could not parse status to Int. %s",
			err.Error())),
		)
		return
	}
	kafkamdm.StartupGCPercent = value
	ctx.PlainText(200, []byte("OK"))
}

func (s *Server) setNormalGCPercent(ctx *middleware.Context, percent models.NormalGCPercent) {
	value, err := strconv.ParseInt(percent.value, 10, 32)
	if err != nil {
		response.Write(ctx, response.NewError(http.StatusBadRequest, fmt.Sprintf(
			"could not parse status to Int. %s",
			err.Error())),
		)
		return
	}
	kafkamdm.SNormalGCPercent = value
	ctx.PlainText(200, []byte("OK"))
}
