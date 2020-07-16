package provider

type Provider interface {
	GetName()string
	Start(args interface{})
	Stop()
}

//type ProviderManager struct {
//	 items map[string]Provider
//}
//
//func NewProviderManager()*ProviderManager{
//	return &ProviderManager{
//		items: map[string]Provider{},
//	}
//}
//
//func (p *ProviderManager)Register(provider Provider){
//	p.items[provider.GetName()] = provider
//}
//
//func (p *ProviderManager)Run(name string) (Provider,error){
//	var err error
//	provider,ok:=p.items[name]
//	if ok{
//		go func() {
//			 defer func() {
//			 	if err:=recover();err!=nil{
//			 		log.Fatalf("[%s] provider crashed,ERR: %s\ntrace%s",name,err,string(debug.Stack()))
//				}
//			 }()
//			 if err:=provider.Start();err!=nil{
//			 	log.Fatalf("%s provider fail,ERR:%s",err.Error())
//			 }
//		}()
//	}else{
//		err = fmt.Errorf("provider [%s] not found",name)
//	}
//	return provider,err
//}
