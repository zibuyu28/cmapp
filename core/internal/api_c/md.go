package api_c

import (
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/zibuyu28/cmapp/core/internal/service_c"
	"github.com/zibuyu28/cmapp/core/pkg/ag"
)

type function string

func parse(s string) function { return function(s) }

const (
	newApp        function = "NewApp"
	startApp      function = "StartApp"
	stopApp       function = "StopApp"
	destroyApp    function = "DestroyApp"
	tagEx         function = "TagEx"
	fileMountEx   function = "FileMountEx"
	envEx         function = "EnvEx"
	networkEx     function = "NetworkEx"
	filePremiseEx function = "FilePremiseEx"
	limitEx       function = "LimitEx"
	healthEx      function = "HealthEx"
	logEx         function = "LogEx"
)

// mdExecHandler machine driver exec handler
func mdExecHandler(req *MDReq) (interface{}, error) {
	pb, err := json.Marshal(req.Param)
	if err != nil {
		return nil, errors.Wrap(err, "marshal request's param")
	}
	switch parse(req.Fnc) {
	case newApp:
		var nar ag.NewAppReq
		err = json.Unmarshal(pb, &nar)
		if err != nil {
			return nil, err
		}
		app, err := service_c.RMDIns.NewApp(&nar)
		if err != nil {
			return nil, errors.Wrap(err, "new app")
		}
		return *app, nil
	case startApp:
		var nar ag.App
		err = json.Unmarshal(pb, &nar)
		if err != nil {
			return nil, err
		}
		err = service_c.RMDIns.StartApp(&nar)
		if err != nil {
			return nil, errors.Wrap(err, "start app")
		}
		return nar, nil
	case stopApp:
		var nar ag.App
		err = json.Unmarshal(pb, &nar)
		if err != nil {
			return nil, err
		}
		err = service_c.RMDIns.StopApp(&nar)
		if err != nil {
			return nil, errors.Wrap(err, "stop app")
		}
		return nar, nil
	case destroyApp:
		var nar ag.App
		err = json.Unmarshal(pb, &nar)
		if err != nil {
			return nil, err
		}
		err = service_c.RMDIns.DestroyApp(&nar)
		if err != nil {
			return nil, errors.Wrap(err, "destroy app")
		}
		return nar, nil
	case tagEx:
		var nar ag.Tag
		err = json.Unmarshal(pb, &nar)
		if err != nil {
			return nil, err
		}
		err = service_c.RMDIns.TagEx(req.AppUUID, &nar)
		if err != nil {
			return nil, errors.Wrap(err, "add tag")
		}
		return nar, nil
	case fileMountEx:
		var nar ag.FileMount
		err = json.Unmarshal(pb, &nar)
		if err != nil {
			return nil, err
		}
		err = service_c.RMDIns.FileMountEx(req.AppUUID, &nar)
		if err != nil {
			return nil, errors.Wrap(err, "file mount")
		}
		return nar, nil

	case envEx:
		var nar ag.EnvVar
		err = json.Unmarshal(pb, &nar)
		if err != nil {
			return nil, err
		}
		err = service_c.RMDIns.EnvEx(req.AppUUID, &nar)
		if err != nil {
			return nil, errors.Wrap(err, "set env")
		}
		return nar, nil
	case networkEx:
		var nar ag.Network
		err = json.Unmarshal(pb, &nar)
		if err != nil {
			return nil, err
		}
		err = service_c.RMDIns.NetworkEx(req.AppUUID, &nar)
		if err != nil {
			return nil, errors.Wrap(err, "set network")
		}
		return nar, nil
	case filePremiseEx:
		var nar ag.File
		err = json.Unmarshal(pb, &nar)
		if err != nil {
			return nil, err
		}
		err = service_c.RMDIns.FilePremiseEx(req.AppUUID, &nar)
		if err != nil {
			return nil, errors.Wrap(err, "set file premise")
		}
		return nar, nil
	case limitEx:
		var nar ag.Limit
		err = json.Unmarshal(pb, &nar)
		if err != nil {
			return nil, err
		}
		err = service_c.RMDIns.LimitEx(req.AppUUID, &nar)
		if err != nil {
			return nil, errors.Wrap(err, "set app limit")
		}
		return nar, nil
	case healthEx:
		var nar ag.Health
		err = json.Unmarshal(pb, &nar)
		if err != nil {
			return nil, err
		}
		err = service_c.RMDIns.HealthEx(req.AppUUID, &nar)
		if err != nil {
			return nil, errors.Wrap(err, "set app health")
		}
		return nar, nil
	case logEx:
		var nar ag.Log
		err = json.Unmarshal(pb, &nar)
		if err != nil {
			return nil, err
		}
		err = service_c.RMDIns.LogEx(req.AppUUID, &nar)
		if err != nil {
			return nil, errors.Wrap(err, "set app log")
		}
		return nar, nil
	default:
		return nil, errors.Errorf("function [%s] not correct", req.Fnc)
	}
}
