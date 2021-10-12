package machine

import (
	"github.com/zibuyu28/cmapp/core/pkg/ag"
	"github.com/zibuyu28/cmapp/plugin/proto/worker0"
)

func wappstruct(a *ag.App) *worker0.App {
	ap := worker0.App{
		UUID:            a.UUID,
		MainP:           &worker0.App_MainProcess{},
		FileMounts:      []*worker0.App_FileMount{},
		EnvironmentVars: []*worker0.App_EnvVar{},
		Networks:        []*worker0.App_Network{},
		Workspace: &worker0.App_WorkspaceInfo{
			Workspace: a.Workspace.Workspace,
		},
		FilePremise: []*worker0.App_File{},
		LimitInfo: &worker0.App_Limit{
			CPU:    int32(a.LimitInfo.CPU),
			Memory: int32(a.LimitInfo.Memory),
		},
		HealthInfo: &worker0.App_Health{
			Liveness: &worker0.App_Health_Basic{
				MethodType: worker0.App_Health_Basic_Method(int(a.HealthInfo.Liveness.MethodType)),
				Path:       a.HealthInfo.Liveness.Path,
				Port:       int32(a.HealthInfo.Liveness.Port),
			},
			Readness: &worker0.App_Health_Basic{
				MethodType: worker0.App_Health_Basic_Method(int(a.HealthInfo.Readness.MethodType)),
				Path:       a.HealthInfo.Readness.Path,
				Port:       int32(a.HealthInfo.Readness.Port),
			},
		},
		LogInfo: &worker0.App_Log{
			RealTimeFile: a.LogInfo.RealTimeFile,
			FilePath:     a.LogInfo.FilePath,
		},
		Tags: []*worker0.App_Tag{},
	}
	if a.Tags != nil {
		for _, tag := range a.Tags {
			ap.Tags = append(ap.Tags, &worker0.App_Tag{
				Key:   tag.Key,
				Value: tag.Value,
			})
		}
	}
	if a.FilePremise != nil {
		for _, file := range ap.FilePremise {
			ap.FilePremise = append(ap.FilePremise, &worker0.App_File{
				Name:        file.Name,
				AcquireAddr: file.AcquireAddr,
				Shell:       file.Shell,
			})
		}
	}

	if a.FileMounts != nil {
		for _, mount := range a.FileMounts {
			ap.FileMounts = append(ap.FileMounts, &worker0.App_FileMount{
				File:    mount.File,
				MountTo: mount.MountTo,
				Volume:  mount.Volume,
			})
		}
	}
	if a.EnvironmentVars != nil {
		for _, environmentVar := range a.EnvironmentVars {
			ap.EnvironmentVars = append(ap.EnvironmentVars, &worker0.App_EnvVar{
				Key:   environmentVar.Key,
				Value: environmentVar.Value,
			})
		}
	}
	if a.Networks != nil {
		for _, net := range a.Networks {
			if net.PortInfo.Port == 0 {
				continue
			}
			network := worker0.App_Network{
				PortInfo: &worker0.App_Network_PortInf{
					Port:         int32(net.PortInfo.Port),
					Name:         net.PortInfo.Name,
					ProtocolType: worker0.App_Network_PortInf_Protocol(net.PortInfo.ProtocolType),
				},
				RouteInfo: []*worker0.App_Network_RouteInf{},
			}

			if net.RouteInfo != nil {
				for _, inf := range net.RouteInfo {
					if len(inf.Router) == 0 {
						continue
					}
					network.RouteInfo = append(network.RouteInfo, &worker0.App_Network_RouteInf{
						RouteType: worker0.App_Network_RouteInf_Route(inf.RouteType),
						Router:    inf.Router,
					})
				}
			}
			ap.Networks = append(ap.Networks)
		}
	}
	return &ap
}

func tagset(ap *ag.App, ts []*worker0.App_Tag) {
	if ts != nil {
		if ap.Tags == nil {
			ap.Tags = []ag.Tag{}
		}
		for _, tag := range ts {
			ap.Tags = append(ap.Tags, ag.Tag{
				Key:   tag.Key,
				Value: tag.Value,
			})
		}
	}
}

func logset(a *ag.App, t *worker0.App_Log) {
	if t != nil {
		a.LogInfo.FilePath = t.FilePath
		a.LogInfo.RealTimeFile = t.RealTimeFile
	}
}

func healthset(a *ag.App, h *worker0.App_Health) {
	if h != nil {
		if h.Liveness != nil {
			a.HealthInfo.Liveness.Port = int(h.Liveness.Port)
			a.HealthInfo.Liveness.Path = h.Liveness.Path
			a.HealthInfo.Liveness.MethodType = ag.Method(int(h.Liveness.MethodType))
		}
		if h.Readness != nil {
			a.HealthInfo.Readness.Port = int(h.Readness.Port)
			a.HealthInfo.Readness.Path = h.Readness.Path
			a.HealthInfo.Readness.MethodType = ag.Method(int(h.Readness.MethodType))
		}
	}
}

func limitset(a *ag.App, l *worker0.App_Limit) {
	if l != nil {
		a.LimitInfo.CPU = int(l.CPU)
		a.LimitInfo.Memory = int(l.Memory)
	}
}

func filepremiseset(ap *ag.App, fs []*worker0.App_File) {
	if fs != nil {
		if ap.FilePremise == nil {
			ap.FilePremise = []ag.File{}
		}
		for _, file := range fs {
			ap.FilePremise = append(ap.FilePremise, ag.File{
				Name:        file.Name,
				AcquireAddr: file.AcquireAddr,
				Shell:       file.Shell,
			})
		}
	}
}

func workspaceset(a *ag.App, w *worker0.App_WorkspaceInfo) {
	if w != nil {
		a.Workspace.Workspace = w.Workspace
	}
}

func filemountset(ap *ag.App, fs []*worker0.App_FileMount) {
	if fs != nil {
		if ap.FileMounts == nil {
			ap.FileMounts = []ag.FileMount{}
		}
		for _, mount := range fs {
			ap.FileMounts = append(ap.FileMounts, ag.FileMount{
				File:    mount.File,
				MountTo: mount.MountTo,
				Volume:  mount.Volume,
			})
		}
	}
}

func envset(ap *ag.App, es []*worker0.App_EnvVar) {
	if es != nil {
		if ap.EnvironmentVars == nil {
			ap.EnvironmentVars = []ag.EnvVar{}
		}
		for _, environmentVar := range es {
			ap.EnvironmentVars = append(ap.EnvironmentVars, ag.EnvVar{
				Key:   environmentVar.Key,
				Value: environmentVar.Value,
			})
		}
	}
}

func networkset(ap *ag.App, ns []*worker0.App_Network) {
	if ns != nil {
		if ap.Networks == nil {
			ap.Networks = []ag.Network{}
		}
		for _, net := range ns {
			if net.PortInfo == nil {
				continue
			}
			network := ag.Network{
				PortInfo: struct {
					Port         int         `json:"port"`
					Name         string      `json:"name"`
					ProtocolType ag.Protocol `json:"protocol_type"`
				}{
					Port:         int(net.PortInfo.Port),
					Name:         net.PortInfo.Name,
					ProtocolType: ag.Protocol(int(net.PortInfo.ProtocolType)),
				},
				RouteInfo: []struct {
					RouteType ag.Route `json:"route_type"`
					Router    string   `json:"router"`
				}{},
			}
			if net.RouteInfo != nil {
				for _, inf := range net.RouteInfo {
					network.RouteInfo = append(network.RouteInfo, struct {
						RouteType ag.Route `json:"route_type"`
						Router    string   `json:"router"`
					}{RouteType: ag.Route(int(inf.RouteType)), Router: inf.Router})
				}
			}
			ap.Networks = append(ap.Networks)
		}
	}
}

func appstruct(a *worker0.App) *ag.App {
	ap := ag.App{
		UUID: a.UUID,
		MainP: struct {
			CheckSum string   `json:"check_sum"`
			Name     string   `json:"name"`
			Version  string   `json:"version"`
			Type     ag.PType `json:"type"`
			WorkDir  string   `json:"work_dir"`
			StartCMD []string `json:"start_cmd"`
		}{},
		FileMounts:      []ag.FileMount{},
		EnvironmentVars: []ag.EnvVar{},
		Networks:        []ag.Network{},
		Workspace:       ag.WorkspaceInfo{},
		FilePremise:     []ag.File{},
		LimitInfo:       ag.Limit{},
		HealthInfo: ag.Health{
			Liveness: ag.Basic{},
			Readness: ag.Basic{},
		},
		LogInfo: ag.Log{},
		Tags:    []ag.Tag{},
	}
	tagset(&ap, a.Tags)
	logset(&ap, a.LogInfo)
	healthset(&ap, a.HealthInfo)
	limitset(&ap, a.LimitInfo)
	filepremiseset(&ap, a.FilePremise)
	workspaceset(&ap, a.Workspace)
	filemountset(&ap, a.FileMounts)
	envset(&ap, a.EnvironmentVars)
	networkset(&ap, a.Networks)
	return &ap
}
