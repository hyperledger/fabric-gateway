// Copyright IBM Corp. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"context"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hyperledger/fabric-protos-go-apiv2/common"
	"github.com/hyperledger/fabric-protos-go-apiv2/gateway"
	"github.com/hyperledger/fabric-protos-go-apiv2/msp"
	"github.com/hyperledger/fabric-protos-go-apiv2/peer"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/runtime/protoiface"
	"google.golang.org/protobuf/testing/protocmp"
)

// AssertProtoEqual ensures an expected protobuf message matches an actual message
func AssertProtoEqual(t *testing.T, expected protoreflect.ProtoMessage, actual protoreflect.ProtoMessage) {
	if diff := cmp.Diff(expected, actual, protocmp.Transform()); diff != "" {
		require.FailNow(t, fmt.Sprintf(
			"Not equal:\nexpected: %s\nactual  : %s\n\nDiff:\n- Expected\n+ Actual\n\n%s",
			formatProto(expected),
			formatProto(actual),
			diff,
		))
	}
}

func formatProto(message proto.Message) string {
	if message == nil {
		return fmt.Sprintf("%T", message)
	}

	marshal := prototext.MarshalOptions{
		Multiline:    true,
		Indent:       "\t",
		AllowPartial: true,
	}
	formatted := strings.TrimSpace(marshal.Format(message))
	return fmt.Sprintf("%s{\n%s\n}", protoMessageType(message), indent(formatted))
}

func protoMessageType(message proto.Message) string {
	return string(message.ProtoReflect().Descriptor().Name())
}

func indent(text string) string {
	return "\t" + strings.ReplaceAll(text, "\n", "\n\t")
}

// AssertUnmarshal ensures that a protobuf is umarshaled without error
func AssertUnmarshal(t *testing.T, b []byte, m protoreflect.ProtoMessage) {
	err := proto.Unmarshal(b, m)
	require.NoError(t, err)
}

// AssertUnmarshalProposalPayload ensures that a ChaincodeProposalPayload protobuf is umarshalled without error
func AssertUnmarshalProposalPayload(t *testing.T, proposedTransaction *peer.SignedProposal) *peer.ChaincodeProposalPayload {
	proposal := &peer.Proposal{}
	AssertUnmarshal(t, proposedTransaction.ProposalBytes, proposal)

	payload := &peer.ChaincodeProposalPayload{}
	AssertUnmarshal(t, proposal.Payload, payload)

	return payload
}

// AssertUnmarshalInvocationSpec ensures that a ChaincodeInvocationSpec protobuf is umarshalled without error
func AssertUnmarshalInvocationSpec(t *testing.T, proposedTransaction *peer.SignedProposal) *peer.ChaincodeInvocationSpec {
	proposal := &peer.Proposal{}
	AssertUnmarshal(t, proposedTransaction.ProposalBytes, proposal)

	payload := &peer.ChaincodeProposalPayload{}
	AssertUnmarshal(t, proposal.Payload, payload)

	input := &peer.ChaincodeInvocationSpec{}
	AssertUnmarshal(t, payload.Input, input)

	return input
}

// AssertUnmarshalChannelheader ensures that a ChannelHeader protobuf is umarshalled without error
func AssertUnmarshalChannelheader(t *testing.T, proposedTransaction *peer.SignedProposal) *common.ChannelHeader {
	header := AssertUnmarshalHeader(t, proposedTransaction)

	channelHeader := &common.ChannelHeader{}
	AssertUnmarshal(t, header.ChannelHeader, channelHeader)

	return channelHeader
}

// AssertUnmarshalHeader ensures that a Header protobuf is umarshalled without error
func AssertUnmarshalHeader(t *testing.T, proposedTransaction *peer.SignedProposal) *common.Header {
	proposal := &peer.Proposal{}
	AssertUnmarshal(t, proposedTransaction.ProposalBytes, proposal)

	header := &common.Header{}
	AssertUnmarshal(t, proposal.Header, header)

	return header
}

// AssertUnmarshalSignatureHeader ensures that a SignatureHeader protobuf is umarshalled without error
func AssertUnmarshalSignatureHeader(t *testing.T, proposedTransaction *peer.SignedProposal) *common.SignatureHeader {
	header := AssertUnmarshalHeader(t, proposedTransaction)

	signatureHeader := &common.SignatureHeader{}
	AssertUnmarshal(t, header.SignatureHeader, signatureHeader)

	return signatureHeader
}

type invokeFunction func(ctx context.Context, method string, args any, reply any, opts ...grpc.CallOption) error

func ExpectEvaluate(mockConnection *MockClientConnInterface, options ...invokeFunction) *MockClientConnInterface_Invoke_Call {
	invokeCall := mockConnection.EXPECT().
		Invoke(mock.Anything, "/gateway.Gateway/Evaluate", mock.Anything, mock.Anything, mock.Anything)
	fakeInvoke(invokeCall, options...)
	return invokeCall
}

func fakeInvoke(mock *MockClientConnInterface_Invoke_Call, options ...invokeFunction) {
	mock.RunAndReturn(func(ctx context.Context, method string, args any, reply any, opts ...grpc.CallOption) error {
		for _, option := range options {
			if err := option(ctx, method, args, reply, opts...); err != nil {
				return err
			}
		}

		return nil
	})
}

func WithEvaluateResponse(value []byte) invokeFunction {
	return func(ctx context.Context, method string, args any, reply any, opts ...grpc.CallOption) error {
		proto.Merge(reply.(proto.Message), &gateway.EvaluateResponse{
			Result: &peer.Response{
				Payload: value,
			},
		})
		return nil
	}
}

// WithInvokeContextErr causes the invoke to return any error associated with the invocation context; otherwise it has
// no effect.
func WithInvokeContextErr() invokeFunction {
	return func(ctx context.Context, method string, args any, reply any, opts ...grpc.CallOption) error {
		return ctx.Err()
	}
}

func WithInvokeError(err error) invokeFunction {
	return func(ctx context.Context, method string, args any, reply any, opts ...grpc.CallOption) error {
		return err
	}
}

func CaptureInvokeRequest[T proto.Message](requests chan<- T) invokeFunction {
	return func(ctx context.Context, method string, args any, reply any, opts ...grpc.CallOption) error {
		TrySend(requests, args.(T))
		return nil
	}
}

func CaptureInvokeOptions(callOptions chan<- []grpc.CallOption) invokeFunction {
	return func(ctx context.Context, method string, args any, reply any, opts ...grpc.CallOption) error {
		TrySend(callOptions, opts)
		return nil
	}
}

func CaptureInvokeContext(contexts chan<- context.Context) invokeFunction {
	return func(ctx context.Context, method string, args any, reply any, opts ...grpc.CallOption) error {
		TrySend(contexts, ctx)
		return nil
	}
}

func TrySend[T any](channel chan<- T, value T) bool {
	select {
	case channel <- value:
		return true
	default:
		return false
	}
}

func NewStatusError(t *testing.T, code codes.Code, message string, details ...protoiface.MessageV1) error {
	s, err := status.New(code, message).WithDetails(details...)
	require.NoError(t, err)

	return s.Err()
}

func ExpectEndorse(mockConnection *MockClientConnInterface, options ...invokeFunction) *MockClientConnInterface_Invoke_Call {
	invokeCall := mockConnection.EXPECT().
		Invoke(mock.Anything, "/gateway.Gateway/Endorse", mock.Anything, mock.Anything, mock.Anything)
	fakeInvoke(invokeCall, options...)
	return invokeCall
}

func ExpectSubmit(mockConnection *MockClientConnInterface, options ...invokeFunction) *MockClientConnInterface_Invoke_Call {
	invokeCall := mockConnection.EXPECT().
		Invoke(mock.Anything, "/gateway.Gateway/Submit", mock.Anything, mock.Anything, mock.Anything)
	fakeInvoke(invokeCall, options...)
	return invokeCall
}

func ExpectCommitStatus(mockConnection *MockClientConnInterface, options ...invokeFunction) *MockClientConnInterface_Invoke_Call {
	invokeCall := mockConnection.EXPECT().
		Invoke(mock.Anything, "/gateway.Gateway/CommitStatus", mock.Anything, mock.Anything, mock.Anything)
	fakeInvoke(invokeCall, options...)
	return invokeCall
}

func WithEndorseResponse(response *gateway.EndorseResponse) invokeFunction {
	return func(ctx context.Context, method string, args any, reply any, opts ...grpc.CallOption) error {
		proto.Merge(reply.(proto.Message), response)
		return nil
	}
}

func WithCommitStatusResponse(status peer.TxValidationCode, blockNumber uint64) invokeFunction {
	return func(ctx context.Context, method string, args any, reply any, opts ...grpc.CallOption) error {
		proto.Merge(reply.(proto.Message), &gateway.CommitStatusResponse{
			Result:      status,
			BlockNumber: blockNumber,
		})
		return nil
	}
}

type newStreamFunction func(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error)

func ExpectChaincodeEvents(mockConnection *MockClientConnInterface, options ...newStreamFunction) *MockClientConnInterface_NewStream_Call {
	newStreamCall := mockConnection.EXPECT().
		NewStream(mock.Anything, mock.Anything, "/gateway.Gateway/ChaincodeEvents", mock.Anything)
	fakeNewStream(newStreamCall, options...)
	return newStreamCall
}

func ExpectDeliver(mockConnection *MockClientConnInterface, options ...newStreamFunction) *MockClientConnInterface_NewStream_Call {
	newStreamCall := mockConnection.EXPECT().
		NewStream(mock.Anything, mock.Anything, "/protos.Deliver/Deliver", mock.Anything)
	fakeNewStream(newStreamCall, options...)
	return newStreamCall
}

func ExpectDeliverFiltered(mockConnection *MockClientConnInterface, options ...newStreamFunction) *MockClientConnInterface_NewStream_Call {
	newStreamCall := mockConnection.EXPECT().
		NewStream(mock.Anything, mock.Anything, "/protos.Deliver/DeliverFiltered", mock.Anything)
	fakeNewStream(newStreamCall, options...)
	return newStreamCall
}

func ExpectDeliverWithPrivateData(mockConnection *MockClientConnInterface, options ...newStreamFunction) *MockClientConnInterface_NewStream_Call {
	newStreamCall := mockConnection.EXPECT().
		NewStream(mock.Anything, mock.Anything, "/protos.Deliver/DeliverWithPrivateData", mock.Anything)
	fakeNewStream(newStreamCall, options...)
	return newStreamCall
}

func fakeNewStream(mock *MockClientConnInterface_NewStream_Call, options ...newStreamFunction) {
	mock.RunAndReturn(func(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		for _, option := range options {
			if stream, err := option(ctx, desc, method, opts...); stream != nil || err != nil {
				return stream, err
			}
		}

		return nil, nil
	})
}

func WithNewStreamResult(stream grpc.ClientStream) newStreamFunction {
	return func(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		return stream, nil
	}
}

func WithNewStreamError(err error) newStreamFunction {
	return func(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		return nil, err
	}
}

func CaptureNewStreamOptions(callOptions chan<- []grpc.CallOption) newStreamFunction {
	return func(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		TrySend(callOptions, opts)
		return nil, nil
	}
}

type sendMsgFunction func(message any) error

func ExpectSendMsg(mockStream *MockClientStream, options ...sendMsgFunction) *MockClientStream_SendMsg_Call {
	result := mockStream.EXPECT().SendMsg(mock.Anything)
	result.RunAndReturn(func(message any) error {
		for _, option := range options {
			if err := option(message); err != nil {
				return err
			}
		}

		return nil
	})
	return result
}

func CaptureSendMsg[T proto.Message](messages chan<- T) sendMsgFunction {
	return func(message any) error {
		TrySend(messages, message.(T))
		return nil
	}
}

type recvMsgFunction func(message any) error

func ExpectRecvMsg(mockStream *MockClientStream, options ...recvMsgFunction) *MockClientStream_RecvMsg_Call {
	result := mockStream.EXPECT().RecvMsg(mock.Anything)

	if len(options) > 0 {
		result.RunAndReturn(func(message any) error {
			for _, option := range options {
				if err := option(message); err != nil {
					return err
				}
			}

			return nil
		})
	}

	return result
}

func WithRecvMsgs[T proto.Message](responses ...T) recvMsgFunction {
	responseChannel := make(chan proto.Message, len(responses))
	for _, response := range responses {
		responseChannel <- response
	}
	close(responseChannel)

	return func(message any) error {
		response, ok := <-responseChannel
		if !ok {
			return io.EOF
		}

		proto.Merge(message.(proto.Message), response)
		return nil
	}
}

func AssertMarshal(t *testing.T, message protoreflect.ProtoMessage, msgAndArgs ...any) []byte {
	bytes, err := proto.Marshal(message)
	require.NoError(t, err, msgAndArgs...)
	return bytes
}

func AssertNewEndorseResponse(t *testing.T, result string, channelName string) *gateway.EndorseResponse {
	return &gateway.EndorseResponse{
		PreparedTransaction: &common.Envelope{
			Payload: AssertMarshal(t, &common.Payload{
				Header: &common.Header{
					ChannelHeader: AssertMarshal(t, &common.ChannelHeader{
						ChannelId: channelName,
					}),
				},
				Data: AssertMarshal(t, &peer.Transaction{
					Actions: []*peer.TransactionAction{
						{
							Payload: AssertMarshal(t, &peer.ChaincodeActionPayload{
								Action: &peer.ChaincodeEndorsedAction{
									ProposalResponsePayload: AssertMarshal(t, &peer.ProposalResponsePayload{
										Extension: AssertMarshal(t, &peer.ChaincodeAction{
											Response: &peer.Response{
												Payload: []byte(result),
											},
										}),
									}),
								},
							}),
						},
					},
				}),
			}),
		},
	}
}

func AssertValidBlockEventRequestHeader(t *testing.T, payload *common.Payload, expectedChannel string) {
	channelHeader := &common.ChannelHeader{}
	AssertUnmarshal(t, payload.GetHeader().GetChannelHeader(), channelHeader)

	require.Equal(t, expectedChannel, channelHeader.GetChannelId(), "channel name")

	signatureHeader := &common.SignatureHeader{}
	AssertUnmarshal(t, payload.GetHeader().GetSignatureHeader(), signatureHeader)

	expectedCreator := &msp.SerializedIdentity{
		Mspid:   TestCredentials.Identity().MspID(),
		IdBytes: TestCredentials.Identity().Credentials(),
	}
	actualCreator := &msp.SerializedIdentity{}
	AssertUnmarshal(t, signatureHeader.GetCreator(), actualCreator)
	AssertProtoEqual(t, expectedCreator, actualCreator)
}
