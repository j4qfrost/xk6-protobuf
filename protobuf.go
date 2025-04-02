package protobuf

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/bufbuild/protocompile"
	"google.golang.org/protobuf/encoding/protojson"

	"go.k6.io/k6/js/modules"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
)

func init() {
	modules.Register("k6/x/protobuf", new(Protobuf))
}

type Protobuf struct{}

type ProtoFile struct {
	messageDesc protoreflect.MessageDescriptor
}

// Load a .proto file and find the requested message type
func (p *Protobuf) Load(protoFilePath, lookupType string) (ProtoFile, error) {
	compiler := protocompile.Compiler{
		Resolver: &protocompile.SourceResolver{},
	}

	files, err := compiler.Compile(context.Background(), protoFilePath)
	if err != nil {
		return ProtoFile{}, fmt.Errorf("failed to compile proto file: %w", err)
	}
	if len(files) == 0 {
		return ProtoFile{}, fmt.Errorf("no protobuf files were compiled")
	}

	messageDesc := files[0].Messages().ByName(protoreflect.Name(lookupType))
	if messageDesc == nil {
		return ProtoFile{}, fmt.Errorf("message type '%s' not found", lookupType)
	}

	return ProtoFile{messageDesc}, nil
}

// Encode JSON to protobuf binary format
func (p *ProtoFile) Encode(data string) ([]byte, error) {
	dynamicMessage := dynamicpb.NewMessage(p.messageDesc)

	if err := protojson.Unmarshal([]byte(data), dynamicMessage); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	encodedBytes, err := proto.Marshal(dynamicMessage)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal protobuf: %w", err)
	}

	return encodedBytes, nil
}

// Decode protobuf binary format back to JSON
func (p *ProtoFile) Decode(encodedBytes []byte) (string, error) {
	dynamicMessage := dynamicpb.NewMessage(p.messageDesc)

	if err := proto.Unmarshal(encodedBytes, dynamicMessage); err != nil {
		return "", fmt.Errorf("failed to unmarshal protobuf: %w", err)
	}

	marshalOptions := protojson.MarshalOptions{UseProtoNames: true}

	jsonString, err := marshalOptions.Marshal(dynamicMessage)
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return string(jsonString), nil
}

// Encode a protobuf message with a length prefix (Delimited Encoding)
func (p *ProtoFile) EncodeDelimited(data string) ([]byte, error) {
	encodedBytes, err := p.Encode(data)
	if err != nil {
		return nil, err
	}

	// Create a buffer to store the length-prefixed message
	var buf bytes.Buffer

	// Write the length as a varint
	if err := binary.Write(&buf, binary.LittleEndian, uint64(len(encodedBytes))); err != nil {
		return nil, fmt.Errorf("failed to write length prefix: %w", err)
	}

	// Write the actual encoded message
	if _, err := buf.Write(encodedBytes); err != nil {
		return nil, fmt.Errorf("failed to write message: %w", err)
	}

	return buf.Bytes(), nil
}

// Decode a length-prefixed (delimited) protobuf message
func (p *ProtoFile) DecodeDelimited(delimitedBytes []byte) (string, error) {
	buf := bytes.NewReader(delimitedBytes)

	// Read the length prefix
	var length uint64
	if err := binary.Read(buf, binary.LittleEndian, &length); err != nil {
		return "", fmt.Errorf("failed to read length prefix: %w", err)
	}

	// Read the actual protobuf message
	messageBytes := make([]byte, length)
	if _, err := io.ReadFull(buf, messageBytes); err != nil {
		return "", fmt.Errorf("failed to read full message: %w", err)
	}

	// Decode the message
	return p.Decode(messageBytes)
}