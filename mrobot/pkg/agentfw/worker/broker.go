package worker

//type broker struct {
//	wsFont *wsFront
//	plg    *plugin
//}
//
//func (b *broker) Execute(ctx context.Context) {
//	b.wsFont.run(ctx)
//	var sig = make(chan bool, 0)
//	go func() {
//		<-sig
//		b.makeAction()
//	}()
//	b.plg.run(ctx, sig)
//}
//
//type Action string
//
//// new worker0  action ======= new worker
//const (
//	NewApp        Action = "NewApp"
//	StartApp      Action = "StartApp"
//	StopApp       Action = "StopApp"
//	DestroyApp    Action = "DestroyApp"
//	TagEx         Action = "TagEx"
//	FileMountEx   Action = "FileMountEx"
//	EnvEx         Action = "EnvEx"
//	NetworkEx     Action = "NetworkEx"
//	FilePremiseEx Action = "FilePremiseEx"
//	LimitEx       Action = "LimitEx"
//	HealthEx      Action = "HealthEx"
//	LogEx         Action = "LogEx"
//)
//
//var mmap = map[Action]struct{}{
//	NewApp:        {},
//	StartApp:      {},
//	StopApp:       {},
//	DestroyApp:    {},
//	TagEx:         {},
//	FileMountEx:   {},
//	EnvEx:         {},
//	NetworkEx:     {},
//	FilePremiseEx: {},
//	LimitEx:       {},
//	HealthEx:      {},
//	LogEx:         {},
//}
//
//type ActionREQ struct {
//	UUID     string
//	Action   Action
//	Metadata map[string]string
//	Args     map[string]interface{}
//}
//
//type ActionRESP struct {
//	UUID   string
//	Action Action
//	ERR    string
//	DATA   []byte
//}
//
//func (b *broker) makeAction() {
//	for {
//		req := &ActionREQ{}
//		err := b.wsFont.ReceiveAction(req)
//		if err != nil {
//			log.Errorf(context.Background(), "Manage err when receive action. Now to continue. Err: [%v]", err)
//			continue
//		}
//		if _, ok := mmap[req.Action]; !ok {
//			log.Errorf(context.Background(), "Manage err when parse Action request [%+v]. Now to continue. Err: [%v]", req, err)
//			continue
//		} else {
//			go b.trans(*req)
//		}
//	}
//}
//
//func (b *broker) trans(req ActionREQ) {
//	//md := metadata.New(req.Metadata)
//	//md.Set("MA_UUID", req.UUID)
//	//outgoingContext := metadata.NewOutgoingContext(context.Background(), md)
//
//	//var err error
//	//var respErr error
//	//var respData interface{}
//	//switch req.Action {
//	//case NewApp:
//	//	info := &worker0.NewAppReq{}
//	//	err = mapstructure.Decode(req.Args, info)
//	//	if err != nil {
//	//		log.Errorf(outgoingContext, "Manage err when decode action request args [%+v]. Now to continue. Err: [%v]", req.Args, err)
//	//		return
//	//	}
//	//	respData, err = b.plg.rpcClient.NewApp(outgoingContext, info)
//	//	if err != nil {
//	//		log.Errorf(outgoingContext, "Currently new app failed, req [%+v]. Now to continue. Err: [%v]", req, err)
//	//		respErr = errors.Wrap(err, "call NewApp")
//	//	}
//	//case StartApp:
//	//	respData, err = b.plg.rpcClient.StartApp(outgoingContext, nil)
//	//	if err != nil {
//	//		log.Errorf(outgoingContext, "Currently start app failed, req [%+v]. Now to continue. Err: [%v]", req, err)
//	//		respErr = errors.Wrap(err, "call StartApp")
//	//	}
//	//case StopApp:
//	//	respData, err = b.plg.rpcClient.StopApp(outgoingContext, nil)
//	//	if err != nil {
//	//		log.Errorf(outgoingContext, "Currently stop app failed, req [%+v]. Now to continue. Err: [%v]", req, err)
//	//		respErr = errors.Wrap(err, "call StopApp")
//	//	}
//	//case DestroyApp:
//	//	respData, err = b.plg.rpcClient.DestroyApp(outgoingContext, nil)
//	//	if err != nil {
//	//		log.Errorf(outgoingContext, "Currently destroy app failed, req [%+v]. Now to continue. Err: [%v]", req, err)
//	//		respErr = errors.Wrap(err, "call DestroyApp")
//	//	}
//	//case TagEx:
//	//	info := &worker0.App_Tag{}
//	//	err = mapstructure.Decode(req.Args, info)
//	//	if err != nil {
//	//		log.Errorf(outgoingContext, "Manage err when decode action request args [%+v]. Now to continue. Err: [%v]", req.Args, err)
//	//		return
//	//	}
//	//	respData, err = b.plg.rpcClient.TagEx(outgoingContext, info)
//	//	if err != nil {
//	//		log.Errorf(outgoingContext, "Currently tag ex failed, req [%+v]. Now to continue. Err: [%v]", req, err)
//	//		respErr = errors.Wrap(err, "call TagEx")
//	//	}
//	//case FileMountEx:
//	//	info := &worker0.App_FileMount{}
//	//	err = mapstructure.Decode(req.Args, info)
//	//	if err != nil {
//	//		log.Errorf(outgoingContext, "Manage err when decode action request args [%+v]. Now to continue. Err: [%v]", req.Args, err)
//	//		return
//	//	}
//	//	respData, err = b.plg.rpcClient.FileMountEx(outgoingContext, info)
//	//	if err != nil {
//	//		log.Errorf(outgoingContext, "Currently file mount ex failed, req [%+v]. Now to continue. Err: [%v]", req, err)
//	//		respErr = errors.Wrap(err, "call FileMountEx")
//	//	}
//	//case EnvEx:
//	//	info := &worker0.App_EnvVar{}
//	//	err = mapstructure.Decode(req.Args, info)
//	//	if err != nil {
//	//		log.Errorf(outgoingContext, "Manage err when decode action request args [%+v]. Now to continue. Err: [%v]", req.Args, err)
//	//		return
//	//	}
//	//	respData, err = b.plg.rpcClient.EnvEx(outgoingContext, info)
//	//	if err != nil {
//	//		log.Errorf(outgoingContext, "Currently env ex failed, req [%+v]. Now to continue. Err: [%v]", req, err)
//	//		respErr = errors.Wrap(err, "call EnvEx")
//	//	}
//	//case NetworkEx:
//	//	info := &worker0.App_Network{}
//	//	err = mapstructure.Decode(req.Args, info)
//	//	if err != nil {
//	//		log.Errorf(outgoingContext, "Manage err when decode action request args [%+v]. Now to continue. Err: [%v]", req.Args, err)
//	//		return
//	//	}
//	//	respData, err = b.plg.rpcClient.NetworkEx(outgoingContext, info)
//	//	if err != nil {
//	//		log.Errorf(outgoingContext, "Currently network ex failed, req [%+v]. Now to continue. Err: [%v]", req, err)
//	//		respErr = errors.Wrap(err, "call NetworkEx")
//	//	}
//	//case FilePremiseEx:
//	//	info := &worker0.App_File{}
//	//	err = mapstructure.Decode(req.Args, info)
//	//	if err != nil {
//	//		log.Errorf(outgoingContext, "Manage err when decode action request args [%+v]. Now to continue. Err: [%v]", req.Args, err)
//	//		return
//	//	}
//	//	respData, err = b.plg.rpcClient.FilePremiseEx(outgoingContext, info)
//	//	if err != nil {
//	//		log.Errorf(outgoingContext, "Currently file premise ex failed, req [%+v]. Now to continue. Err: [%v]", req, err)
//	//		respErr = errors.Wrap(err, "call FilePremiseEx")
//	//	}
//	//case LimitEx:
//	//	info := &worker0.App_Limit{}
//	//	err = mapstructure.Decode(req.Args, info)
//	//	if err != nil {
//	//		log.Errorf(outgoingContext, "Manage err when decode action request args [%+v]. Now to continue. Err: [%v]", req.Args, err)
//	//		return
//	//	}
//	//	respData, err = b.plg.rpcClient.LimitEx(outgoingContext, info)
//	//	if err != nil {
//	//		log.Errorf(outgoingContext, "Currently limit ex failed, req [%+v]. Now to continue. Err: [%v]", req, err)
//	//		respErr = errors.Wrap(err, "call LimitEx")
//	//	}
//	//case HealthEx:
//	//	info := &worker0.App_Health{}
//	//	err = mapstructure.Decode(req.Args, info)
//	//	if err != nil {
//	//		log.Errorf(outgoingContext, "Manage err when decode action request args [%+v]. Now to continue. Err: [%v]", req.Args, err)
//	//		return
//	//	}
//	//	respData, err = b.plg.rpcClient.HealthEx(outgoingContext, info)
//	//	if err != nil {
//	//		log.Errorf(outgoingContext, "Currently health ex failed, req [%+v]. Now to continue. Err: [%v]", req, err)
//	//		respErr = errors.Wrap(err, "call HealthEx")
//	//	}
//	//case LogEx:
//	//	info := &worker0.App_Log{}
//	//	err = mapstructure.Decode(req.Args, info)
//	//	if err != nil {
//	//		log.Errorf(outgoingContext, "Manage err when decode action request args [%+v]. Now to continue. Err: [%v]", req.Args, err)
//	//		return
//	//	}
//	//	respData, err = b.plg.rpcClient.LogEx(outgoingContext, info)
//	//	if err != nil {
//	//		log.Errorf(outgoingContext, "Currently log ex failed, req [%+v]. Now to continue. Err: [%v]", req, err)
//	//		respErr = errors.Wrap(err, "call LogEx")
//	//	}
//	//}
//
//	//resp := ActionRESP{
//	//	UUID:   req.UUID,
//	//	Action: req.Action,
//	//}
//	//if respErr != nil {
//	//	resp.ERR = respErr.Error()
//	//}
//	//if respData != nil {
//	//	marshal, err := json.Marshal(respData)
//	//	if err != nil {
//	//		log.Errorf(outgoingContext, "Currently marshal resp data, resp data [%+v]. Now to continue. Err: [%v]", respData, err)
//	//		return
//	//	}
//	//	resp.DATA = marshal
//	//}
//	//e := b.wsFont.ActionResp(resp)
//	//if e != nil {
//	//	log.Errorf(outgoingContext, "Currently send resp, resp [%+v]. Now to continue. Err: [%v]", resp, e)
//	//}
//}
