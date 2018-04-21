package main

import (
	"fmt"
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

func (s *SimpleContract) RegisterService(receiver interface{}) error {
	return s.callbacks.Register(receiver, "simple", true)
}

func (s *SimpleContract) SInit(ctx plugin.Context, req *plugin.Request) error {
	s.callbacks = new(serviceMap)
	return nil
}

func (s *SimpleContract) jsonResponse() *plugin.Response {
	return &plugin.Response{
		ContentType: plugin.ContentType_JSON,
		Body:        []byte("{}"),
	}
}

func (s *SimpleContract) StaticCall(ctx plugin.StaticContext, req *plugin.Request) (*plugin.Response, error) {
	return &plugin.Response{}, nil
}

func (s *SimpleContract) Call(ctx plugin.Context, req *plugin.Request) (*plugin.Response, error) {
	log.Println("Entering Etherboy contract")
	var tx txmsg.SimpleContractMethod
	proto.Unmarshal(req.Body, &tx)
	owner := strings.TrimSpace(tx.Owner)
	fmt.Printf("wtf -%s\n", &tx)

	typeName := reflect.TypeOf(tx.Data).String()
	fmt.Printf("typename -%s\n", typeName)
	serviceSpec, methodSpec, err := s.callbacks.Get(tx.Method)
	if err != nil {
		return s.jsonResponse(), err
	}

	txData := reflect.New(methodSpec.argsType)

	if err := ptypes.UnmarshalAny(tx.Data, txData.Interface().(proto.Message)); err != nil {
		return s.jsonResponse(), err
	}

	fmt.Printf("typename-txdata -%s\n", txData)
	fmt.Printf("typename-txdata2 -%T\n", txData)

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
		return s.jsonResponse(), errResult
	}
	return &plugin.Response{}, nil
}

func (s *SimpleContract) ownerKey(owner string) []byte {
	return []byte("owner:" + owner)
}
