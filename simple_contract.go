package main

import (
	"fmt"
	"log"
	"reflect"
	"strings"

	proto "github.com/golang/protobuf/proto"
	"github.com/loomnetwork/etherboy-core/txmsg"
	"github.com/loomnetwork/loom/plugin"
)

type SimpleContract struct {
	callbacks *serviceMap
}

func (s *SimpleContract) RegisterService(receiver interface{}, name interface{}) error {
	return d.callbacks.Register(receiver, reflect.TypeOf(name).String(), true)
}

func (s *SimpleContract) SInit(ctx plugin.Context, req *plugin.Request) error {
	d.callbacks = new(serviceMap)
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
	var tx txmsg.EtherboyAppTx
	proto.Unmarshal(req.Body, &tx)
	owner := strings.TrimSpace(tx.Owner)

	typeName := reflect.TypeOf(tx.Data).String()
	fmt.Printf("typename -%s\n", typeName)
	serviceSpec, methodSpec, err := d.callbacks.Get(typeName)
	if err != nil {
		return d.jsonResponse(), err
	}

	//Lookup the method we need to call
	errValue := methodSpec.method.Func.Call([]reflect.Value{
		serviceSpec.rcvr,
		reflect.ValueOf(ctx),
		reflect.ValueOf(owner),
		reflect.ValueOf(tx.Data),
	})

	// Cast the result to error if needed.
	var errResult error
	errInter := errValue[0].Interface()
	if errInter != nil {
		errResult = errInter.(error)
	}

	if errResult != nil {
		return d.jsonResponse(), errResult
	}
	return &plugin.Response{}, nil
}

func (s *SimpleContract) ownerKey(owner string) []byte {
	return []byte("owner:" + owner)
}
