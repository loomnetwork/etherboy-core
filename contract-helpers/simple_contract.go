package contracthelpers

import (
	"encoding/json"
	"errors"
	"reflect"

	proto "github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
	"github.com/loomnetwork/etherboy-core/txmsg"
	"github.com/loomnetwork/loom/plugin"
)

type SimpleContract struct {
	callbacks *serviceMap
}

func (s *SimpleContract) Init(contract plugin.Contract) error {
	s.callbacks = new(serviceMap)
	return s.callbacks.Register(contract, contract.Meta().Name)
}

func (s *SimpleContract) StaticCall(ctx plugin.StaticContext, req *plugin.Request) (*plugin.Response, error) {
	var result []reflect.Value
	if req.ContentType == plugin.ContentType_JSON {
		var query txmsg.SimpleContractMethodJSON
		if err := json.Unmarshal(req.Body, &query); err != nil {
			return nil, err
		}
		serviceSpec, methodSpec, err := s.callbacks.Get(query.Method, true)
		if err != nil {
			return nil, err
		}
		queryParams := reflect.New(methodSpec.argsType)
		if err := json.Unmarshal(query.Data, queryParams.Interface()); err != nil {
			return nil, err
		}
		result = methodSpec.method.Func.Call([]reflect.Value{
			serviceSpec.rcvr,
			reflect.ValueOf(ctx),
			queryParams,
		})
	} else if req.ContentType == plugin.ContentType_PROTOBUF3 {
		var query txmsg.SimpleContractMethod
		if err := proto.Unmarshal(req.Body, &query); err != nil {
			return nil, err
		}
		serviceSpec, methodSpec, err := s.callbacks.Get(query.Method, true)
		if err != nil {
			return nil, err
		}
		queryParams := reflect.New(methodSpec.argsType)
		if err := types.UnmarshalAny(query.Data, queryParams.Interface().(proto.Message)); err != nil {
			return nil, err
		}
		result = methodSpec.method.Func.Call([]reflect.Value{
			serviceSpec.rcvr,
			reflect.ValueOf(ctx),
			queryParams,
		})
	} else {
		return nil, errors.New("unsupported content type")
	}

	// If the method returned an error, extract & return it
	var err error
	errInter := result[1].Interface()
	if errInter != nil {
		err = errInter.(error)
	}
	if err != nil {
		return nil, err
	}

	resp := &plugin.Response{ContentType: req.Accept}
	if req.Accept == plugin.ContentType_JSON {
		resp.Body, err = json.Marshal(result[0].Interface())
	} else if req.Accept == plugin.ContentType_PROTOBUF3 {
		resp.Body, err = proto.Marshal(result[0].Interface().(proto.Message))
	} else {
		return nil, errors.New("unsupported accept type")
	}
	return resp, err
}

func (s *SimpleContract) Call(ctx plugin.Context, req *plugin.Request) (*plugin.Response, error) {
	var tx txmsg.SimpleContractMethod
	if err := proto.Unmarshal(req.Body, &tx); err != nil {
		return nil, err
	}

	serviceSpec, methodSpec, err := s.callbacks.Get(tx.Method, false)
	if err != nil {
		return nil, err
	}

	txData := reflect.New(methodSpec.argsType)

	if err := types.UnmarshalAny(tx.Data, txData.Interface().(proto.Message)); err != nil {
		return nil, err
	}

	//Lookup the method we need to call
	errValue := methodSpec.method.Func.Call([]reflect.Value{
		serviceSpec.rcvr,
		reflect.ValueOf(ctx),
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
