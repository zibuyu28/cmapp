package api_c

import (
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/zibuyu28/cmapp/core/internal/service_c/machine"
)

type MachineAction string

const (
	Create MachineAction = "create"
	Stop   MachineAction = "stop"
	Start  MachineAction = "start"
	Remove MachineAction = "remove"
)

type MWReq struct {
	DriverID  int           `json:"driver_id"`
	MachineID int           `json:"machine_id"`
	Action    MachineAction `json:"action" validate:"required"`
	Param     interface{}   `json:"param" validate:"required"`
}

func mwExec(g *gin.Context) {
	var p = MWReq{}
	err := g.BindJSON(&p)
	if err != nil {
		fail(g, err)
		return
	}
	err = mwExecHandler(g, &p)
	if err != nil {
		fail(g, err)
		return
	}
}

func mwExecHandler(g *gin.Context, req *MWReq) error {
	switch req.Action {
	case Create:
		if req.DriverID == 0 {
			return errors.New("create action: driver id is nil")
		}
		err := machine.Create(g.Request.Context(), req.DriverID, req.Param)
		if err != nil {
			return errors.Wrap(err, "do create machine")
		}
		ok(g, "ok")
	default:
		return errors.Errorf("action [%s] not support now", req.Action)
	}
	return nil
}
