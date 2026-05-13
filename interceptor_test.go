/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-05-11 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-05-11 22:15:57
 * @FilePath: \go-scope-provider\interceptor_test.go
 * @Description: gRPC 拦截器测试
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */
package provider

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// 场景：UnaryPayloadInterceptor 从 metadata 中解析 Payload
// 预期：Payload 被注入到上下文
func TestUnaryPayloadInterceptor(t *testing.T) {
	p := &Payload{
		Domain:   1,
		TenantID: "T001",
		RoleCode: "owner",
		ScopeBindings: []*ScopeEntry{
			{ScopeType: 2, RegionCodes: []string{"MM"}},
		},
	}

	jsonData, _ := json.Marshal(p)
	encoded := base64.StdEncoding.EncodeToString(jsonData)

	md := metadata.Pairs(string(MetadataAuthPayload), encoded)
	ctx := metadata.NewIncomingContext(context.Background(), md)

	var capturedCtx context.Context
	interceptor := UnaryPayloadInterceptor()
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		capturedCtx = ctx
		return "ok", nil
	}

	_, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{}, handler)
	assert.NoError(t, err)

	got := GetPayloadFromCtx(capturedCtx)
	assert.NotNil(t, got)
	assert.Equal(t, int32(1), got.Domain)
	assert.Equal(t, "T001", got.TenantID)
	assert.Equal(t, "owner", got.RoleCode)
	assert.Len(t, got.ScopeBindings, 1)
}

// 场景：metadata 中无 x-auth-payload
// 预期：上下文中无 Payload
func TestUnaryPayloadInterceptor_NoMetadata(t *testing.T) {
	ctx := metadata.NewIncomingContext(context.Background(), metadata.New(nil))

	var capturedCtx context.Context
	interceptor := UnaryPayloadInterceptor()
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		capturedCtx = ctx
		return "ok", nil
	}

	_, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{}, handler)
	assert.NoError(t, err)
	assert.Nil(t, GetPayloadFromCtx(capturedCtx))
}

// 场景：metadata 中 base64 解码失败
// 预期：上下文中无 Payload，不 panic
func TestUnaryPayloadInterceptor_InvalidBase64(t *testing.T) {
	md := metadata.Pairs(string(MetadataAuthPayload), "not-valid-base64!!!")
	ctx := metadata.NewIncomingContext(context.Background(), md)

	var capturedCtx context.Context
	interceptor := UnaryPayloadInterceptor()
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		capturedCtx = ctx
		return "ok", nil
	}

	_, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{}, handler)
	assert.NoError(t, err)
	assert.Nil(t, GetPayloadFromCtx(capturedCtx))
}

// 场景：metadata 中 JSON 格式无效
// 预期：上下文中无 Payload，不 panic
func TestUnaryPayloadInterceptor_InvalidJSON(t *testing.T) {
	encoded := base64.StdEncoding.EncodeToString([]byte("{invalid}"))
	md := metadata.Pairs(string(MetadataAuthPayload), encoded)
	ctx := metadata.NewIncomingContext(context.Background(), md)

	var capturedCtx context.Context
	interceptor := UnaryPayloadInterceptor()
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		capturedCtx = ctx
		return "ok", nil
	}

	_, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{}, handler)
	assert.NoError(t, err)
	assert.Nil(t, GetPayloadFromCtx(capturedCtx))
}

// 场景：无 metadata 的上下文
// 预期：不 panic，上下文中无 Payload
func TestUnaryPayloadInterceptor_NoIncomingMetadata(t *testing.T) {
	var capturedCtx context.Context
	interceptor := UnaryPayloadInterceptor()
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		capturedCtx = ctx
		return "ok", nil
	}

	_, err := interceptor(context.Background(), nil, &grpc.UnaryServerInfo{}, handler)
	assert.NoError(t, err)
	assert.Nil(t, GetPayloadFromCtx(capturedCtx))
}

// mockServerStream 用于测试 StreamPayloadInterceptor
type mockServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (m *mockServerStream) Context() context.Context {
	return m.ctx
}

// 场景：StreamPayloadInterceptor 从 metadata 中解析 Payload
// 预期：Payload 被注入到流上下文
func TestStreamPayloadInterceptor(t *testing.T) {
	p := &Payload{
		Domain:   2,
		TenantID: "T002",
		RoleCode: "admin",
	}

	jsonData, _ := json.Marshal(p)
	encoded := base64.StdEncoding.EncodeToString(jsonData)

	md := metadata.Pairs(string(MetadataAuthPayload), encoded)
	ctx := metadata.NewIncomingContext(context.Background(), md)

	stream := &mockServerStream{ctx: ctx}

	var capturedCtx context.Context
	interceptor := StreamPayloadInterceptor()
	handler := func(srv interface{}, ss grpc.ServerStream) error {
		capturedCtx = ss.Context()
		return nil
	}

	err := interceptor(nil, stream, &grpc.StreamServerInfo{}, handler)
	assert.NoError(t, err)

	got := GetPayloadFromCtx(capturedCtx)
	assert.NotNil(t, got)
	assert.Equal(t, int32(2), got.Domain)
	assert.Equal(t, "T002", got.TenantID)
}
