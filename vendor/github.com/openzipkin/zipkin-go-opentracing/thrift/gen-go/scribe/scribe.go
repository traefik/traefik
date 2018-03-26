// Autogenerated by Thrift Compiler (0.9.3)
// DO NOT EDIT UNLESS YOU ARE SURE THAT YOU KNOW WHAT YOU ARE DOING

package scribe

import (
	"bytes"
	"fmt"
	"github.com/apache/thrift/lib/go/thrift"
)

// (needed to ensure safety because of naive import list construction.)
var _ = thrift.ZERO
var _ = fmt.Printf
var _ = bytes.Equal

type Scribe interface {
	// Parameters:
	//  - Messages
	Log(messages []*LogEntry) (r ResultCode, err error)
}

type ScribeClient struct {
	Transport       thrift.TTransport
	ProtocolFactory thrift.TProtocolFactory
	InputProtocol   thrift.TProtocol
	OutputProtocol  thrift.TProtocol
	SeqId           int32
}

func NewScribeClientFactory(t thrift.TTransport, f thrift.TProtocolFactory) *ScribeClient {
	return &ScribeClient{Transport: t,
		ProtocolFactory: f,
		InputProtocol:   f.GetProtocol(t),
		OutputProtocol:  f.GetProtocol(t),
		SeqId:           0,
	}
}

func NewScribeClientProtocol(t thrift.TTransport, iprot thrift.TProtocol, oprot thrift.TProtocol) *ScribeClient {
	return &ScribeClient{Transport: t,
		ProtocolFactory: nil,
		InputProtocol:   iprot,
		OutputProtocol:  oprot,
		SeqId:           0,
	}
}

// Parameters:
//  - Messages
func (p *ScribeClient) Log(messages []*LogEntry) (r ResultCode, err error) {
	if err = p.sendLog(messages); err != nil {
		return
	}
	return p.recvLog()
}

func (p *ScribeClient) sendLog(messages []*LogEntry) (err error) {
	oprot := p.OutputProtocol
	if oprot == nil {
		oprot = p.ProtocolFactory.GetProtocol(p.Transport)
		p.OutputProtocol = oprot
	}
	p.SeqId++
	if err = oprot.WriteMessageBegin("Log", thrift.CALL, p.SeqId); err != nil {
		return
	}
	args := ScribeLogArgs{
		Messages: messages,
	}
	if err = args.Write(oprot); err != nil {
		return
	}
	if err = oprot.WriteMessageEnd(); err != nil {
		return
	}
	return oprot.Flush()
}

func (p *ScribeClient) recvLog() (value ResultCode, err error) {
	iprot := p.InputProtocol
	if iprot == nil {
		iprot = p.ProtocolFactory.GetProtocol(p.Transport)
		p.InputProtocol = iprot
	}
	method, mTypeId, seqId, err := iprot.ReadMessageBegin()
	if err != nil {
		return
	}
	if method != "Log" {
		err = thrift.NewTApplicationException(thrift.WRONG_METHOD_NAME, "Log failed: wrong method name")
		return
	}
	if p.SeqId != seqId {
		err = thrift.NewTApplicationException(thrift.BAD_SEQUENCE_ID, "Log failed: out of sequence response")
		return
	}
	if mTypeId == thrift.EXCEPTION {
		error0 := thrift.NewTApplicationException(thrift.UNKNOWN_APPLICATION_EXCEPTION, "Unknown Exception")
		var error1 error
		error1, err = error0.Read(iprot)
		if err != nil {
			return
		}
		if err = iprot.ReadMessageEnd(); err != nil {
			return
		}
		err = error1
		return
	}
	if mTypeId != thrift.REPLY {
		err = thrift.NewTApplicationException(thrift.INVALID_MESSAGE_TYPE_EXCEPTION, "Log failed: invalid message type")
		return
	}
	result := ScribeLogResult{}
	if err = result.Read(iprot); err != nil {
		return
	}
	if err = iprot.ReadMessageEnd(); err != nil {
		return
	}
	value = result.GetSuccess()
	return
}

type ScribeProcessor struct {
	processorMap map[string]thrift.TProcessorFunction
	handler      Scribe
}

func (p *ScribeProcessor) AddToProcessorMap(key string, processor thrift.TProcessorFunction) {
	p.processorMap[key] = processor
}

func (p *ScribeProcessor) GetProcessorFunction(key string) (processor thrift.TProcessorFunction, ok bool) {
	processor, ok = p.processorMap[key]
	return processor, ok
}

func (p *ScribeProcessor) ProcessorMap() map[string]thrift.TProcessorFunction {
	return p.processorMap
}

func NewScribeProcessor(handler Scribe) *ScribeProcessor {

	self2 := &ScribeProcessor{handler: handler, processorMap: make(map[string]thrift.TProcessorFunction)}
	self2.processorMap["Log"] = &scribeProcessorLog{handler: handler}
	return self2
}

func (p *ScribeProcessor) Process(iprot, oprot thrift.TProtocol) (success bool, err thrift.TException) {
	name, _, seqId, err := iprot.ReadMessageBegin()
	if err != nil {
		return false, err
	}
	if processor, ok := p.GetProcessorFunction(name); ok {
		return processor.Process(seqId, iprot, oprot)
	}
	iprot.Skip(thrift.STRUCT)
	iprot.ReadMessageEnd()
	x3 := thrift.NewTApplicationException(thrift.UNKNOWN_METHOD, "Unknown function "+name)
	oprot.WriteMessageBegin(name, thrift.EXCEPTION, seqId)
	x3.Write(oprot)
	oprot.WriteMessageEnd()
	oprot.Flush()
	return false, x3

}

type scribeProcessorLog struct {
	handler Scribe
}

func (p *scribeProcessorLog) Process(seqId int32, iprot, oprot thrift.TProtocol) (success bool, err thrift.TException) {
	args := ScribeLogArgs{}
	if err = args.Read(iprot); err != nil {
		iprot.ReadMessageEnd()
		x := thrift.NewTApplicationException(thrift.PROTOCOL_ERROR, err.Error())
		oprot.WriteMessageBegin("Log", thrift.EXCEPTION, seqId)
		x.Write(oprot)
		oprot.WriteMessageEnd()
		oprot.Flush()
		return false, err
	}

	iprot.ReadMessageEnd()
	result := ScribeLogResult{}
	var retval ResultCode
	var err2 error
	if retval, err2 = p.handler.Log(args.Messages); err2 != nil {
		x := thrift.NewTApplicationException(thrift.INTERNAL_ERROR, "Internal error processing Log: "+err2.Error())
		oprot.WriteMessageBegin("Log", thrift.EXCEPTION, seqId)
		x.Write(oprot)
		oprot.WriteMessageEnd()
		oprot.Flush()
		return true, err2
	} else {
		result.Success = &retval
	}
	if err2 = oprot.WriteMessageBegin("Log", thrift.REPLY, seqId); err2 != nil {
		err = err2
	}
	if err2 = result.Write(oprot); err == nil && err2 != nil {
		err = err2
	}
	if err2 = oprot.WriteMessageEnd(); err == nil && err2 != nil {
		err = err2
	}
	if err2 = oprot.Flush(); err == nil && err2 != nil {
		err = err2
	}
	if err != nil {
		return
	}
	return true, err
}

// HELPER FUNCTIONS AND STRUCTURES

// Attributes:
//  - Messages
type ScribeLogArgs struct {
	Messages []*LogEntry `thrift:"messages,1" json:"messages"`
}

func NewScribeLogArgs() *ScribeLogArgs {
	return &ScribeLogArgs{}
}

func (p *ScribeLogArgs) GetMessages() []*LogEntry {
	return p.Messages
}
func (p *ScribeLogArgs) Read(iprot thrift.TProtocol) error {
	if _, err := iprot.ReadStructBegin(); err != nil {
		return thrift.PrependError(fmt.Sprintf("%T read error: ", p), err)
	}

	for {
		_, fieldTypeId, fieldId, err := iprot.ReadFieldBegin()
		if err != nil {
			return thrift.PrependError(fmt.Sprintf("%T field %d read error: ", p, fieldId), err)
		}
		if fieldTypeId == thrift.STOP {
			break
		}
		switch fieldId {
		case 1:
			if err := p.readField1(iprot); err != nil {
				return err
			}
		default:
			if err := iprot.Skip(fieldTypeId); err != nil {
				return err
			}
		}
		if err := iprot.ReadFieldEnd(); err != nil {
			return err
		}
	}
	if err := iprot.ReadStructEnd(); err != nil {
		return thrift.PrependError(fmt.Sprintf("%T read struct end error: ", p), err)
	}
	return nil
}

func (p *ScribeLogArgs) readField1(iprot thrift.TProtocol) error {
	_, size, err := iprot.ReadListBegin()
	if err != nil {
		return thrift.PrependError("error reading list begin: ", err)
	}
	tSlice := make([]*LogEntry, 0, size)
	p.Messages = tSlice
	for i := 0; i < size; i++ {
		_elem4 := &LogEntry{}
		if err := _elem4.Read(iprot); err != nil {
			return thrift.PrependError(fmt.Sprintf("%T error reading struct: ", _elem4), err)
		}
		p.Messages = append(p.Messages, _elem4)
	}
	if err := iprot.ReadListEnd(); err != nil {
		return thrift.PrependError("error reading list end: ", err)
	}
	return nil
}

func (p *ScribeLogArgs) Write(oprot thrift.TProtocol) error {
	if err := oprot.WriteStructBegin("Log_args"); err != nil {
		return thrift.PrependError(fmt.Sprintf("%T write struct begin error: ", p), err)
	}
	if err := p.writeField1(oprot); err != nil {
		return err
	}
	if err := oprot.WriteFieldStop(); err != nil {
		return thrift.PrependError("write field stop error: ", err)
	}
	if err := oprot.WriteStructEnd(); err != nil {
		return thrift.PrependError("write struct stop error: ", err)
	}
	return nil
}

func (p *ScribeLogArgs) writeField1(oprot thrift.TProtocol) (err error) {
	if err := oprot.WriteFieldBegin("messages", thrift.LIST, 1); err != nil {
		return thrift.PrependError(fmt.Sprintf("%T write field begin error 1:messages: ", p), err)
	}
	if err := oprot.WriteListBegin(thrift.STRUCT, len(p.Messages)); err != nil {
		return thrift.PrependError("error writing list begin: ", err)
	}
	for _, v := range p.Messages {
		if err := v.Write(oprot); err != nil {
			return thrift.PrependError(fmt.Sprintf("%T error writing struct: ", v), err)
		}
	}
	if err := oprot.WriteListEnd(); err != nil {
		return thrift.PrependError("error writing list end: ", err)
	}
	if err := oprot.WriteFieldEnd(); err != nil {
		return thrift.PrependError(fmt.Sprintf("%T write field end error 1:messages: ", p), err)
	}
	return err
}

func (p *ScribeLogArgs) String() string {
	if p == nil {
		return "<nil>"
	}
	return fmt.Sprintf("ScribeLogArgs(%+v)", *p)
}

// Attributes:
//  - Success
type ScribeLogResult struct {
	Success *ResultCode `thrift:"success,0" json:"success,omitempty"`
}

func NewScribeLogResult() *ScribeLogResult {
	return &ScribeLogResult{}
}

var ScribeLogResult_Success_DEFAULT ResultCode

func (p *ScribeLogResult) GetSuccess() ResultCode {
	if !p.IsSetSuccess() {
		return ScribeLogResult_Success_DEFAULT
	}
	return *p.Success
}
func (p *ScribeLogResult) IsSetSuccess() bool {
	return p.Success != nil
}

func (p *ScribeLogResult) Read(iprot thrift.TProtocol) error {
	if _, err := iprot.ReadStructBegin(); err != nil {
		return thrift.PrependError(fmt.Sprintf("%T read error: ", p), err)
	}

	for {
		_, fieldTypeId, fieldId, err := iprot.ReadFieldBegin()
		if err != nil {
			return thrift.PrependError(fmt.Sprintf("%T field %d read error: ", p, fieldId), err)
		}
		if fieldTypeId == thrift.STOP {
			break
		}
		switch fieldId {
		case 0:
			if err := p.readField0(iprot); err != nil {
				return err
			}
		default:
			if err := iprot.Skip(fieldTypeId); err != nil {
				return err
			}
		}
		if err := iprot.ReadFieldEnd(); err != nil {
			return err
		}
	}
	if err := iprot.ReadStructEnd(); err != nil {
		return thrift.PrependError(fmt.Sprintf("%T read struct end error: ", p), err)
	}
	return nil
}

func (p *ScribeLogResult) readField0(iprot thrift.TProtocol) error {
	if v, err := iprot.ReadI32(); err != nil {
		return thrift.PrependError("error reading field 0: ", err)
	} else {
		temp := ResultCode(v)
		p.Success = &temp
	}
	return nil
}

func (p *ScribeLogResult) Write(oprot thrift.TProtocol) error {
	if err := oprot.WriteStructBegin("Log_result"); err != nil {
		return thrift.PrependError(fmt.Sprintf("%T write struct begin error: ", p), err)
	}
	if err := p.writeField0(oprot); err != nil {
		return err
	}
	if err := oprot.WriteFieldStop(); err != nil {
		return thrift.PrependError("write field stop error: ", err)
	}
	if err := oprot.WriteStructEnd(); err != nil {
		return thrift.PrependError("write struct stop error: ", err)
	}
	return nil
}

func (p *ScribeLogResult) writeField0(oprot thrift.TProtocol) (err error) {
	if p.IsSetSuccess() {
		if err := oprot.WriteFieldBegin("success", thrift.I32, 0); err != nil {
			return thrift.PrependError(fmt.Sprintf("%T write field begin error 0:success: ", p), err)
		}
		if err := oprot.WriteI32(int32(*p.Success)); err != nil {
			return thrift.PrependError(fmt.Sprintf("%T.success (0) field write error: ", p), err)
		}
		if err := oprot.WriteFieldEnd(); err != nil {
			return thrift.PrependError(fmt.Sprintf("%T write field end error 0:success: ", p), err)
		}
	}
	return err
}

func (p *ScribeLogResult) String() string {
	if p == nil {
		return "<nil>"
	}
	return fmt.Sprintf("ScribeLogResult(%+v)", *p)
}
