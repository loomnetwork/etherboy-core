package contracthelpers

import (
	"log"
	"reflect"
	"strings"

	proto "github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/loomnetwork/etherboy-core/txmsg"
	"github.com/loomnetwork/loom/plugin"
)

type SimpleContract struct {
	callbacks *serviceMap
}

func (s *SimpleContract) RegisterService(serviceName string, receiver interface{}) error {
	return s.callbacks.Register(receiver, serviceName, true)
}

func (s *SimpleContract) Init() {
	s.callbacks = new(serviceMap)
}

func (s *SimpleContract) StaticCall(ctx plugin.StaticContext, req *plugin.Request) (*plugin.Response, error) {
	return &plugin.Response{}, nil
}

func (s *SimpleContract) Call(ctx plugin.Context, req *plugin.Request) (*plugin.Response, error) {
	log.Println("Entering Etherboy contract")
	var tx txmsg.SimpleContractMethod
	if err := proto.Unmarshal(req.Body, &tx); err != nil {
		return nil, err
	}
	// TODO: owner shouldn't be in txmsg.SimpleContractMethod, should be in whatever is type is in tx.Data
	owner := strings.TrimSpace(tx.Owner)

	serviceSpec, methodSpec, err := s.callbacks.Get(tx.Method)
	if err != nil {
		return nil, err
	}

	txData := reflect.New(methodSpec.argsType)

	if err := ptypes.UnmarshalAny(tx.Data, txData.Interface().(proto.Message)); err != nil {
		return nil, err
	}

	//Lookup the method we need to call
	errValue := methodSpec.method.Func.Call([]reflect.Value{
		serviceSpec.rcvr,
		reflect.ValueOf(ctx),
		reflect.ValueOf(owner),
		txData,
	})

	// Cast the result to error if needed.
	var errResult error
	errInter := errValue[0].Interface()
	if errInter != nil {
		errResult = errInter.(error)
	}

	if errResult != nil {
		return nil, errResult
	}
	return &plugin.Response{}, nil
}
