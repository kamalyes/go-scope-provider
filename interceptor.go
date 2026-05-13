/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-05-11 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-05-11 23:15:57
 * @FilePath: \go-scope-provider\interceptor.go
 * @Description: gRPC 拦截器 - 从 metadata 中解析认证载荷并注入到上下文
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */
package provider

import (
	"context"
	"encoding/base64"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// MetadataKey gRPC metadata 键名常量
type MetadataKey string

const (
	MetadataAuthPayload MetadataKey = "x-auth-payload"
)

// UnaryPayloadInterceptor gRPC 一元拦截器，从 metadata 中解析认证载荷并注入到上下文
// 载荷以 base64 编码的 JSON 存放在 x-auth-payload metadata 中
func UnaryPayloadInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		ctx = injectPayloadFromMetadata(ctx)
		return handler(ctx, req)
	}
}

// StreamPayloadInterceptor gRPC 流拦截器，从 metadata 中解析认证载荷并注入到上下文
func StreamPayloadInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx := injectPayloadFromMetadata(ss.Context())
		wrapped := &wrappedServerStream{ServerStream: ss, ctx: ctx}
		return handler(srv, wrapped)
	}
}

// injectPayloadFromMetadata 从 gRPC metadata 中提取并解码认证载荷，注入到上下文
func injectPayloadFromMetadata(ctx context.Context) context.Context {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ctx
	}

	values := md.Get(string(MetadataAuthPayload))
	if len(values) == 0 {
		return ctx
	}

	decoded, err := base64.StdEncoding.DecodeString(values[0])
	if err != nil {
		return ctx
	}

	payload, err := FromJSON(string(decoded))
	if err != nil {
		return ctx
	}

	return WithPayload(ctx, payload)
}

// wrappedServerStream 包装 ServerStream 以支持自定义上下文
type wrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedServerStream) Context() context.Context {
	return w.ctx
}
