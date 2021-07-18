package worker

import (
	"context"
	"github.com/goinggo/mapstructure"
	"github.com/pkg/errors"
	"github.com/zibuyu28/cmapp/common/log"
	"github.com/zibuyu28/cmapp/plugin/proto/worker"
	"google.golang.org/grpc/metadata"
	"k8s.io/apimachinery/pkg/util/json"
	"reflect"
)

type broker struct {
	wsFont *wsFront
	plg    *plugin
}

func (b *broker) Execute(ctx context.Context) {
	b.wsFont.run(ctx)
	var sig = make(chan bool, 0)
	go func() {
		<-sig
		b.makeAction()
	}()
	b.plg.run(ctx, sig)
}

type Action string

const (
	GetWorkspace            Action = "GetWorkspace"
	DestroyWorkspace        Action = "DestroyWorkspace"
	DownloadToPath          Action = "DownloadToPath"
	Upload                  Action = "Upload"
	Compress                Action = "Compress"
	Decompress              Action = "Decompress"
	Copy                    Action = "Copy"
	UpdateFileContent       Action = "UpdateFileContent"
	DeleteFile              Action = "DeleteFile"
	CreateFile              Action = "CreateFile"
	CreateDir               Action = "CreateDir"
	RemoveDir               Action = "RemoveDir"
	FetchFileContent        Action = "FetchFileContent"
	CheckTargetPortUseful   Action = "CheckTargetPortUseful"
	SetupApp                Action = "SetupApp"
	Done                    Action = "Done"
	ShutdownApp             Action = "ShutdownApp"
	AppHealth               Action = "AppHealth"
	TargetPortIntranetRoute Action = "TargetPortIntranetRoute"
	TargetPortExternalRoute Action = "TargetPortExternalRoute"
)

var mmap = map[Action]reflect.Type{
	GetWorkspace:            nil,
	DestroyWorkspace:        reflect.TypeOf(worker.WorkspaceInfo{}),
	DownloadToPath:          reflect.TypeOf(worker.DownloadInfo{}),
	Upload:                  reflect.TypeOf(worker.UploadInfo{}),
	Compress:                reflect.TypeOf(worker.CompressInfo{}),
	Decompress:              reflect.TypeOf(worker.DeCompressInfo{}),
	Copy:                    reflect.TypeOf(worker.CopyInfo{}),
	UpdateFileContent:       reflect.TypeOf(worker.UpdateFileContentInfo{}),
	DeleteFile:              reflect.TypeOf(worker.DeleteFileInfo{}),
	CreateFile:              reflect.TypeOf(worker.CreateFileInfo{}),
	CreateDir:               reflect.TypeOf(worker.CreateDirInfo{}),
	RemoveDir:               reflect.TypeOf(worker.RemoveDirInfo{}),
	FetchFileContent:        reflect.TypeOf(worker.FetchFileContentInfo{}),
	CheckTargetPortUseful:   reflect.TypeOf(worker.CheckTargetPortInfo{}),
	SetupApp:                reflect.TypeOf(worker.SetupAppInfo{}),
	Done:                    nil,
	ShutdownApp:             reflect.TypeOf(worker.App{}),
	AppHealth:               reflect.TypeOf(worker.App{}),
	TargetPortIntranetRoute: reflect.TypeOf(worker.TargetPortIntranetInfo{}),
	TargetPortExternalRoute: reflect.TypeOf(worker.TargetPortExternalInfo{}),
}

type ActionREQ struct {
	UUID     string
	Action   Action
	Metadata map[string]string
	Args     map[string]interface{}
}

type ActionRESP struct {
	UUID   string
	Action Action
	ERR    string
	DATA   []byte
}

func (b *broker) makeAction() {
	for {
		req := &ActionREQ{}
		err := b.wsFont.ReceiveAction(req)
		if err != nil {
			log.Errorf(context.Background(), "Manage err when receive action. Now to continue. Err: [%v]", err)
			continue
		}
		if _, ok := mmap[req.Action]; !ok {
			log.Errorf(context.Background(), "Manage err when parse Action request [%+v]. Now to continue. Err: [%v]", req, err)
			continue
		} else {
			go b.trans(*req)
		}
	}
}

func (b *broker) trans(req ActionREQ) {
	md := metadata.New(req.Metadata)
	md.Set("ACTION_UUID", req.UUID)
	outgoingContext := metadata.NewOutgoingContext(context.Background(), md)

	var err error
	var respErr error
	var respData interface{}
	switch req.Action {
	case GetWorkspace:
		respData, err = b.plg.rpcClient.GetWorkspace(outgoingContext, &worker.Empty{})
		if err != nil {
			log.Errorf(outgoingContext, "Currently get workspace failed, req [%+v]. Now to continue. Err: [%v]", req, err)
			respErr = errors.Wrap(err, "call GetWorkspace")
		}
	case DestroyWorkspace:
		info := &worker.WorkspaceInfo{}
		err = mapstructure.Decode(req.Args, info)
		if err != nil {
			log.Errorf(outgoingContext, "Manage err when decode action request args [%+v]. Now to continue. Err: [%v]", req.Args, err)
			return
		}
		respData, err = b.plg.rpcClient.DestroyWorkspace(outgoingContext, info)
		if err != nil {
			log.Errorf(outgoingContext, "Currently destroy workspace failed, req [%+v]. Now to continue. Err: [%v]", req, err)
			respErr = errors.Wrap(err, "call DestroyWorkspace")
		}
	case DownloadToPath:
		info := &worker.DownloadInfo{}
		err = mapstructure.Decode(req.Args, info)
		if err != nil {
			log.Errorf(outgoingContext, "Manage err when decode action request args [%+v]. Now to continue. Err: [%v]", req.Args, err)
			return
		}
		respData, err = b.plg.rpcClient.DownloadToPath(outgoingContext, info)
		if err != nil {
			log.Errorf(outgoingContext, "Currently download to path failed, req [%+v]. Now to continue. Err: [%v]", req, err)
			respErr = errors.Wrap(err, "call DownloadToPath")
		}
	case Upload:
		info := &worker.UploadInfo{}
		err = mapstructure.Decode(req.Args, info)
		if err != nil {
			log.Errorf(outgoingContext, "Manage err when decode action request args [%+v]. Now to continue. Err: [%v]", req.Args, err)
			return
		}
		respData, err = b.plg.rpcClient.Upload(outgoingContext, info)
		if err != nil {
			log.Errorf(outgoingContext, "Currently Upload failed, req [%+v]. Now to continue. Err: [%v]", req, err)
			respErr = errors.Wrap(err, "call Upload")
		}
	case Compress:
		info := &worker.CompressInfo{}
		err = mapstructure.Decode(req.Args, info)
		if err != nil {
			log.Errorf(outgoingContext, "Manage err when decode action request args [%+v]. Now to continue. Err: [%v]", req.Args, err)
			return
		}
		respData, err = b.plg.rpcClient.Compress(outgoingContext, info)
		if err != nil {
			log.Errorf(outgoingContext, "Currently Compress failed, req [%+v]. Now to continue. Err: [%v]", req, err)
			respErr = errors.Wrap(err, "call Compress")
		}
	case Decompress:
		info := &worker.DeCompressInfo{}
		err = mapstructure.Decode(req.Args, info)
		if err != nil {
			log.Errorf(outgoingContext, "Manage err when decode action request args [%+v]. Now to continue. Err: [%v]", req.Args, err)
			return
		}
		respData, err = b.plg.rpcClient.Decompress(outgoingContext, info)
		if err != nil {
			log.Errorf(outgoingContext, "Currently Decompress failed, req [%+v]. Now to continue. Err: [%v]", req, err)
			respErr = errors.Wrap(err, "call Decompress")
		}
	case Copy:
		info := &worker.CopyInfo{}
		err = mapstructure.Decode(req.Args, info)
		if err != nil {
			log.Errorf(outgoingContext, "Manage err when decode action request args [%+v]. Now to continue. Err: [%v]", req.Args, err)
			return
		}
		respData, err = b.plg.rpcClient.Copy(outgoingContext, info)
		if err != nil {
			log.Errorf(outgoingContext, "Currently Copy failed, req [%+v]. Now to continue. Err: [%v]", req, err)
			respErr = errors.Wrap(err, "call Copy")
		}
	case UpdateFileContent:
		info := &worker.UpdateFileContentInfo{}
		err = mapstructure.Decode(req.Args, info)
		if err != nil {
			log.Errorf(outgoingContext, "Manage err when decode action request args [%+v]. Now to continue. Err: [%v]", req.Args, err)
			return
		}
		respData, err = b.plg.rpcClient.UpdateFileContent(outgoingContext, info)
		if err != nil {
			log.Errorf(outgoingContext, "Currently UpdateFileContent failed, req [%+v]. Now to continue. Err: [%v]", req, err)
			respErr = errors.Wrap(err, "call UpdateFileContent")
		}
	case DeleteFile:
		info := &worker.DeleteFileInfo{}
		err = mapstructure.Decode(req.Args, info)
		if err != nil {
			log.Errorf(outgoingContext, "Manage err when decode action request args [%+v]. Now to continue. Err: [%v]", req.Args, err)
			return
		}
		respData, err = b.plg.rpcClient.DeleteFile(outgoingContext, info)
		if err != nil {
			log.Errorf(outgoingContext, "Currently DeleteFile failed, req [%+v]. Now to continue. Err: [%v]", req, err)
			respErr = errors.Wrap(err, "call DeleteFile")
		}
	case CreateFile:
		info := &worker.CreateFileInfo{}
		err = mapstructure.Decode(req.Args, info)
		if err != nil {
			log.Errorf(outgoingContext, "Manage err when decode action request args [%+v]. Now to continue. Err: [%v]", req.Args, err)
			return
		}
		respData, err = b.plg.rpcClient.CreateFile(outgoingContext, info)
		if err != nil {
			log.Errorf(outgoingContext, "Currently CreateFile failed, req [%+v]. Now to continue. Err: [%v]", req, err)
			respErr = errors.Wrap(err, "call CreateFile")
		}
	case CreateDir:
		info := &worker.CreateDirInfo{}
		err = mapstructure.Decode(req.Args, info)
		if err != nil {
			log.Errorf(outgoingContext, "Manage err when decode action request args [%+v]. Now to continue. Err: [%v]", req.Args, err)
			return
		}
		respData, err = b.plg.rpcClient.CreateDir(outgoingContext, info)
		if err != nil {
			log.Errorf(outgoingContext, "Currently CreateDir failed, req [%+v]. Now to continue. Err: [%v]", req, err)
			respErr = errors.Wrap(err, "call CreateDir")
		}
	case RemoveDir:
		info := &worker.RemoveDirInfo{}
		err = mapstructure.Decode(req.Args, info)
		if err != nil {
			log.Errorf(outgoingContext, "Manage err when decode action request args [%+v]. Now to continue. Err: [%v]", req.Args, err)
			return
		}
		respData, err = b.plg.rpcClient.RemoveDir(outgoingContext, info)
		if err != nil {
			log.Errorf(outgoingContext, "Currently RemoveDir failed, req [%+v]. Now to continue. Err: [%v]", req, err)
			respErr = errors.Wrap(err, "call RemoveDir")
		}
	case FetchFileContent:
		// TODO implement
		log.Errorf(outgoingContext, "Currently implement me [%s]", FetchFileContent)
		return
	case CheckTargetPortUseful:
		info := &worker.CheckTargetPortInfo{}
		err = mapstructure.Decode(req.Args, info)
		if err != nil {
			log.Errorf(outgoingContext, "Manage err when decode action request args [%+v]. Now to continue. Err: [%v]", req.Args, err)
			return
		}
		respData, err = b.plg.rpcClient.CheckTargetPortUseful(outgoingContext, info)
		if err != nil {
			log.Errorf(outgoingContext, "Currently CheckTargetPortUseful failed, req [%+v]. Now to continue. Err: [%v]", req, err)
			respErr = errors.Wrap(err, "call CheckTargetPortUseful")
		}
	case SetupApp:
		info := &worker.SetupAppInfo{}
		err = mapstructure.Decode(req.Args, info)
		if err != nil {
			log.Errorf(outgoingContext, "Manage err when decode action request args [%+v]. Now to continue. Err: [%v]", req.Args, err)
			return
		}
		respData, err = b.plg.rpcClient.SetupApp(outgoingContext, info)
		if err != nil {
			log.Errorf(outgoingContext, "Currently SetupApp failed, req [%+v]. Now to continue. Err: [%v]", req, err)
			respErr = errors.Wrap(err, "call SetupApp")
		}
	case Done:
		respData, err = b.plg.rpcClient.Done(outgoingContext, &worker.Empty{})
		if err != nil {
			log.Errorf(outgoingContext, "Currently Done failed, req [%+v]. Now to continue. Err: [%v]", req, err)
			respErr = errors.Wrap(err, "call Done")
		}
	case ShutdownApp:
		info := &worker.App{}
		err = mapstructure.Decode(req.Args, info)
		if err != nil {
			log.Errorf(outgoingContext, "Manage err when decode action request args [%+v]. Now to continue. Err: [%v]", req.Args, err)
			return
		}
		respData, err = b.plg.rpcClient.ShutdownApp(outgoingContext, info)
		if err != nil {
			log.Errorf(outgoingContext, "Currently ShutdownApp failed, req [%+v]. Now to continue. Err: [%v]", req, err)
			respErr = errors.Wrap(err, "call ShutdownApp")
		}
	case AppHealth:
		info := &worker.App{}
		err = mapstructure.Decode(req.Args, info)
		if err != nil {
			log.Errorf(outgoingContext, "Manage err when decode action request args [%+v]. Now to continue. Err: [%v]", req.Args, err)
			return
		}
		respData, err = b.plg.rpcClient.AppHealth(outgoingContext, info)
		if err != nil {
			log.Errorf(outgoingContext, "Currently AppHealth failed, req [%+v]. Now to continue. Err: [%v]", req, err)
			respErr = errors.Wrap(err, "call AppHealth")
		}
	case TargetPortIntranetRoute:
		info := &worker.TargetPortIntranetInfo{}
		err = mapstructure.Decode(req.Args, info)
		if err != nil {
			log.Errorf(outgoingContext, "Manage err when decode action request args [%+v]. Now to continue. Err: [%v]", req.Args, err)
			return
		}
		respData, err = b.plg.rpcClient.TargetPortIntranetRoute(outgoingContext, info)
		if err != nil {
			log.Errorf(outgoingContext, "Currently TargetPortIntranetRoute failed, req [%+v]. Now to continue. Err: [%v]", req, err)
			respErr = errors.Wrap(err, "call TargetPortIntranetRoute")
		}
	case TargetPortExternalRoute:
		info := &worker.TargetPortExternalInfo{}
		err = mapstructure.Decode(req.Args, info)
		if err != nil {
			log.Errorf(outgoingContext, "Manage err when decode action request args [%+v]. Now to continue. Err: [%v]", req.Args, err)
			return
		}
		respData, err = b.plg.rpcClient.TargetPortExternalRoute(outgoingContext, info)
		if err != nil {
			log.Errorf(outgoingContext, "Currently TargetPortExternalRoute failed, req [%+v]. Now to continue. Err: [%v]", req, err)
			respErr = errors.Wrap(err, "call TargetPortExternalRoute")
		}
	}

	resp := ActionRESP{
		UUID:   req.UUID,
		Action: req.Action,
	}
	if respErr != nil {
		resp.ERR = respErr.Error()
	}
	if respData != nil {
		marshal, err := json.Marshal(respData)
		if err != nil {
			log.Errorf(outgoingContext, "Currently marshal resp data, resp data [%+v]. Now to continue. Err: [%v]", respData, err)
			return
		}
		resp.DATA = marshal
	}
	e := b.wsFont.ActionResp(resp)
	if e != nil {
		log.Errorf(outgoingContext, "Currently send resp, resp [%+v]. Now to continue. Err: [%v]", resp, e)
	}
}
