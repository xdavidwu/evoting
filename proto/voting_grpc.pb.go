// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.21.12
// source: proto/voting.proto

package proto

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// RegistrationClient is the client API for Registration service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type RegistrationClient interface {
	RegisterVoter(ctx context.Context, in *Voter, opts ...grpc.CallOption) (*Status, error)
	UnregisterVoter(ctx context.Context, in *VoterName, opts ...grpc.CallOption) (*Status, error)
}

type registrationClient struct {
	cc grpc.ClientConnInterface
}

func NewRegistrationClient(cc grpc.ClientConnInterface) RegistrationClient {
	return &registrationClient{cc}
}

func (c *registrationClient) RegisterVoter(ctx context.Context, in *Voter, opts ...grpc.CallOption) (*Status, error) {
	out := new(Status)
	err := c.cc.Invoke(ctx, "/voting.Registration/RegisterVoter", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *registrationClient) UnregisterVoter(ctx context.Context, in *VoterName, opts ...grpc.CallOption) (*Status, error) {
	out := new(Status)
	err := c.cc.Invoke(ctx, "/voting.Registration/UnregisterVoter", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// RegistrationServer is the server API for Registration service.
// All implementations must embed UnimplementedRegistrationServer
// for forward compatibility
type RegistrationServer interface {
	RegisterVoter(context.Context, *Voter) (*Status, error)
	UnregisterVoter(context.Context, *VoterName) (*Status, error)
	mustEmbedUnimplementedRegistrationServer()
}

// UnimplementedRegistrationServer must be embedded to have forward compatible implementations.
type UnimplementedRegistrationServer struct {
}

func (UnimplementedRegistrationServer) RegisterVoter(context.Context, *Voter) (*Status, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RegisterVoter not implemented")
}
func (UnimplementedRegistrationServer) UnregisterVoter(context.Context, *VoterName) (*Status, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UnregisterVoter not implemented")
}
func (UnimplementedRegistrationServer) mustEmbedUnimplementedRegistrationServer() {}

// UnsafeRegistrationServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to RegistrationServer will
// result in compilation errors.
type UnsafeRegistrationServer interface {
	mustEmbedUnimplementedRegistrationServer()
}

func RegisterRegistrationServer(s grpc.ServiceRegistrar, srv RegistrationServer) {
	s.RegisterService(&Registration_ServiceDesc, srv)
}

func _Registration_RegisterVoter_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Voter)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RegistrationServer).RegisterVoter(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/voting.Registration/RegisterVoter",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RegistrationServer).RegisterVoter(ctx, req.(*Voter))
	}
	return interceptor(ctx, in, info, handler)
}

func _Registration_UnregisterVoter_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(VoterName)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RegistrationServer).UnregisterVoter(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/voting.Registration/UnregisterVoter",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RegistrationServer).UnregisterVoter(ctx, req.(*VoterName))
	}
	return interceptor(ctx, in, info, handler)
}

// Registration_ServiceDesc is the grpc.ServiceDesc for Registration service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Registration_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "voting.Registration",
	HandlerType: (*RegistrationServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "RegisterVoter",
			Handler:    _Registration_RegisterVoter_Handler,
		},
		{
			MethodName: "UnregisterVoter",
			Handler:    _Registration_UnregisterVoter_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "proto/voting.proto",
}

// EVotingClient is the client API for EVoting service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type EVotingClient interface {
	PreAuth(ctx context.Context, in *VoterName, opts ...grpc.CallOption) (*Challenge, error)
	Auth(ctx context.Context, in *AuthRequest, opts ...grpc.CallOption) (*AuthToken, error)
	CreateElection(ctx context.Context, in *Election, opts ...grpc.CallOption) (*Status, error)
	CastVote(ctx context.Context, in *Vote, opts ...grpc.CallOption) (*Status, error)
	GetResult(ctx context.Context, in *ElectionName, opts ...grpc.CallOption) (*ElectionResult, error)
}

type eVotingClient struct {
	cc grpc.ClientConnInterface
}

func NewEVotingClient(cc grpc.ClientConnInterface) EVotingClient {
	return &eVotingClient{cc}
}

func (c *eVotingClient) PreAuth(ctx context.Context, in *VoterName, opts ...grpc.CallOption) (*Challenge, error) {
	out := new(Challenge)
	err := c.cc.Invoke(ctx, "/voting.eVoting/PreAuth", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *eVotingClient) Auth(ctx context.Context, in *AuthRequest, opts ...grpc.CallOption) (*AuthToken, error) {
	out := new(AuthToken)
	err := c.cc.Invoke(ctx, "/voting.eVoting/Auth", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *eVotingClient) CreateElection(ctx context.Context, in *Election, opts ...grpc.CallOption) (*Status, error) {
	out := new(Status)
	err := c.cc.Invoke(ctx, "/voting.eVoting/CreateElection", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *eVotingClient) CastVote(ctx context.Context, in *Vote, opts ...grpc.CallOption) (*Status, error) {
	out := new(Status)
	err := c.cc.Invoke(ctx, "/voting.eVoting/CastVote", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *eVotingClient) GetResult(ctx context.Context, in *ElectionName, opts ...grpc.CallOption) (*ElectionResult, error) {
	out := new(ElectionResult)
	err := c.cc.Invoke(ctx, "/voting.eVoting/GetResult", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// EVotingServer is the server API for EVoting service.
// All implementations must embed UnimplementedEVotingServer
// for forward compatibility
type EVotingServer interface {
	PreAuth(context.Context, *VoterName) (*Challenge, error)
	Auth(context.Context, *AuthRequest) (*AuthToken, error)
	CreateElection(context.Context, *Election) (*Status, error)
	CastVote(context.Context, *Vote) (*Status, error)
	GetResult(context.Context, *ElectionName) (*ElectionResult, error)
	mustEmbedUnimplementedEVotingServer()
}

// UnimplementedEVotingServer must be embedded to have forward compatible implementations.
type UnimplementedEVotingServer struct {
}

func (UnimplementedEVotingServer) PreAuth(context.Context, *VoterName) (*Challenge, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PreAuth not implemented")
}
func (UnimplementedEVotingServer) Auth(context.Context, *AuthRequest) (*AuthToken, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Auth not implemented")
}
func (UnimplementedEVotingServer) CreateElection(context.Context, *Election) (*Status, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateElection not implemented")
}
func (UnimplementedEVotingServer) CastVote(context.Context, *Vote) (*Status, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CastVote not implemented")
}
func (UnimplementedEVotingServer) GetResult(context.Context, *ElectionName) (*ElectionResult, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetResult not implemented")
}
func (UnimplementedEVotingServer) mustEmbedUnimplementedEVotingServer() {}

// UnsafeEVotingServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to EVotingServer will
// result in compilation errors.
type UnsafeEVotingServer interface {
	mustEmbedUnimplementedEVotingServer()
}

func RegisterEVotingServer(s grpc.ServiceRegistrar, srv EVotingServer) {
	s.RegisterService(&EVoting_ServiceDesc, srv)
}

func _EVoting_PreAuth_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(VoterName)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(EVotingServer).PreAuth(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/voting.eVoting/PreAuth",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(EVotingServer).PreAuth(ctx, req.(*VoterName))
	}
	return interceptor(ctx, in, info, handler)
}

func _EVoting_Auth_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AuthRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(EVotingServer).Auth(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/voting.eVoting/Auth",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(EVotingServer).Auth(ctx, req.(*AuthRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _EVoting_CreateElection_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Election)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(EVotingServer).CreateElection(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/voting.eVoting/CreateElection",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(EVotingServer).CreateElection(ctx, req.(*Election))
	}
	return interceptor(ctx, in, info, handler)
}

func _EVoting_CastVote_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Vote)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(EVotingServer).CastVote(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/voting.eVoting/CastVote",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(EVotingServer).CastVote(ctx, req.(*Vote))
	}
	return interceptor(ctx, in, info, handler)
}

func _EVoting_GetResult_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ElectionName)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(EVotingServer).GetResult(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/voting.eVoting/GetResult",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(EVotingServer).GetResult(ctx, req.(*ElectionName))
	}
	return interceptor(ctx, in, info, handler)
}

// EVoting_ServiceDesc is the grpc.ServiceDesc for EVoting service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var EVoting_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "voting.eVoting",
	HandlerType: (*EVotingServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "PreAuth",
			Handler:    _EVoting_PreAuth_Handler,
		},
		{
			MethodName: "Auth",
			Handler:    _EVoting_Auth_Handler,
		},
		{
			MethodName: "CreateElection",
			Handler:    _EVoting_CreateElection_Handler,
		},
		{
			MethodName: "CastVote",
			Handler:    _EVoting_CastVote_Handler,
		},
		{
			MethodName: "GetResult",
			Handler:    _EVoting_GetResult_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "proto/voting.proto",
}

// SyncClient is the client API for Sync service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type SyncClient interface {
	Join(ctx context.Context, in *NodeIdentifier, opts ...grpc.CallOption) (*Dump, error)
	NodesChanged(ctx context.Context, in *NodesList, opts ...grpc.CallOption) (*Empty, error)
	Sql(ctx context.Context, in *SqlRequest, opts ...grpc.CallOption) (*Empty, error)
	NewKey(ctx context.Context, in *Key, opts ...grpc.CallOption) (*Empty, error)
	Ping(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*Empty, error)
}

type syncClient struct {
	cc grpc.ClientConnInterface
}

func NewSyncClient(cc grpc.ClientConnInterface) SyncClient {
	return &syncClient{cc}
}

func (c *syncClient) Join(ctx context.Context, in *NodeIdentifier, opts ...grpc.CallOption) (*Dump, error) {
	out := new(Dump)
	err := c.cc.Invoke(ctx, "/voting.Sync/Join", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *syncClient) NodesChanged(ctx context.Context, in *NodesList, opts ...grpc.CallOption) (*Empty, error) {
	out := new(Empty)
	err := c.cc.Invoke(ctx, "/voting.Sync/NodesChanged", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *syncClient) Sql(ctx context.Context, in *SqlRequest, opts ...grpc.CallOption) (*Empty, error) {
	out := new(Empty)
	err := c.cc.Invoke(ctx, "/voting.Sync/Sql", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *syncClient) NewKey(ctx context.Context, in *Key, opts ...grpc.CallOption) (*Empty, error) {
	out := new(Empty)
	err := c.cc.Invoke(ctx, "/voting.Sync/NewKey", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *syncClient) Ping(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*Empty, error) {
	out := new(Empty)
	err := c.cc.Invoke(ctx, "/voting.Sync/Ping", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// SyncServer is the server API for Sync service.
// All implementations must embed UnimplementedSyncServer
// for forward compatibility
type SyncServer interface {
	Join(context.Context, *NodeIdentifier) (*Dump, error)
	NodesChanged(context.Context, *NodesList) (*Empty, error)
	Sql(context.Context, *SqlRequest) (*Empty, error)
	NewKey(context.Context, *Key) (*Empty, error)
	Ping(context.Context, *Empty) (*Empty, error)
	mustEmbedUnimplementedSyncServer()
}

// UnimplementedSyncServer must be embedded to have forward compatible implementations.
type UnimplementedSyncServer struct {
}

func (UnimplementedSyncServer) Join(context.Context, *NodeIdentifier) (*Dump, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Join not implemented")
}
func (UnimplementedSyncServer) NodesChanged(context.Context, *NodesList) (*Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method NodesChanged not implemented")
}
func (UnimplementedSyncServer) Sql(context.Context, *SqlRequest) (*Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Sql not implemented")
}
func (UnimplementedSyncServer) NewKey(context.Context, *Key) (*Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method NewKey not implemented")
}
func (UnimplementedSyncServer) Ping(context.Context, *Empty) (*Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Ping not implemented")
}
func (UnimplementedSyncServer) mustEmbedUnimplementedSyncServer() {}

// UnsafeSyncServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to SyncServer will
// result in compilation errors.
type UnsafeSyncServer interface {
	mustEmbedUnimplementedSyncServer()
}

func RegisterSyncServer(s grpc.ServiceRegistrar, srv SyncServer) {
	s.RegisterService(&Sync_ServiceDesc, srv)
}

func _Sync_Join_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(NodeIdentifier)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SyncServer).Join(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/voting.Sync/Join",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SyncServer).Join(ctx, req.(*NodeIdentifier))
	}
	return interceptor(ctx, in, info, handler)
}

func _Sync_NodesChanged_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(NodesList)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SyncServer).NodesChanged(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/voting.Sync/NodesChanged",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SyncServer).NodesChanged(ctx, req.(*NodesList))
	}
	return interceptor(ctx, in, info, handler)
}

func _Sync_Sql_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SqlRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SyncServer).Sql(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/voting.Sync/Sql",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SyncServer).Sql(ctx, req.(*SqlRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Sync_NewKey_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Key)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SyncServer).NewKey(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/voting.Sync/NewKey",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SyncServer).NewKey(ctx, req.(*Key))
	}
	return interceptor(ctx, in, info, handler)
}

func _Sync_Ping_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SyncServer).Ping(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/voting.Sync/Ping",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SyncServer).Ping(ctx, req.(*Empty))
	}
	return interceptor(ctx, in, info, handler)
}

// Sync_ServiceDesc is the grpc.ServiceDesc for Sync service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Sync_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "voting.Sync",
	HandlerType: (*SyncServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Join",
			Handler:    _Sync_Join_Handler,
		},
		{
			MethodName: "NodesChanged",
			Handler:    _Sync_NodesChanged_Handler,
		},
		{
			MethodName: "Sql",
			Handler:    _Sync_Sql_Handler,
		},
		{
			MethodName: "NewKey",
			Handler:    _Sync_NewKey_Handler,
		},
		{
			MethodName: "Ping",
			Handler:    _Sync_Ping_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "proto/voting.proto",
}
