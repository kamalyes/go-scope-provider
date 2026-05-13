/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-05-11 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-05-13 21:26:19
 * @FilePath: \go-scope-provider\provider.go
 * @Description: 作用域提供者 - 从上下文中解析作用域数据并应用到 SQL 查询
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */
package provider

import (
	"context"
	"strconv"

	"github.com/kamalyes/go-sqlbuilder/repository"
	"github.com/kamalyes/go-sqlbuilder/scope"
)

// payloadCtxKeyType 用于 context 存储认证载荷的私有键类型
type payloadCtxKeyType struct{}

// payloadKey payload 在 context 中的键
var payloadKey = payloadCtxKeyType{}

// ContextKey 上下文键名常量
type ContextKey string

const (
	ContextKeyDomain   ContextKey = "domain"
	ContextKeyTenantID ContextKey = "tenant_id"
	ContextKeyUserID   ContextKey = "user_id"
	ContextKeyRoleCode ContextKey = "role_code"
	ContextKeyIsOwner  ContextKey = "is_owner"
)

// ResolveScopeData 从上下文中解析作用域数据
// 优先从注入的 Payload 中提取，否则从上下文值中逐个读取
func ResolveScopeData(ctx context.Context, opts ...scope.Option) scope.ScopeData {
	payload := GetPayloadFromCtx(ctx)
	if payload != nil {
		return payload.ToScopeData(opts...)
	}

	data := scope.NewScopeData(opts...)
	if v := ctx.Value(ContextKeyDomain); v != nil {
		switch d := v.(type) {
		case int32:
			data.Domain = d
		case int:
			data.Domain = int32(d)
		case string:
			if n, err := strconv.ParseInt(d, 10, 32); err == nil {
				data.Domain = int32(n)
			}
		}
	}
	if v := ctx.Value(ContextKeyTenantID); v != nil {
		if tid, ok := v.(string); ok {
			data.TenantID = tid
		}
	}
	if v := ctx.Value(ContextKeyRoleCode); v != nil {
		if rc, ok := v.(string); ok {
			data.IsOwner = IsOwnerRoleCode(rc)
		}
	}
	if v := ctx.Value(ContextKeyIsOwner); v != nil {
		if isOwner, ok := v.(bool); ok {
			data.IsOwner = isOwner
		}
	}

	return data
}

// ApplyScopeQuery 从上下文中解析作用域数据并应用到 SQL 查询
// 等价于 ResolveScopeData + ApplySQLScope 的组合
func ApplyScopeQuery(ctx context.Context, query *repository.Query, opts ...scope.Option) *repository.Query {
	data := ResolveScopeData(ctx, opts...)
	return scope.ApplySQLScope(query, data)
}

// GetPayloadFromCtx 从上下文中获取认证载荷（Payload）
// 需要先通过 WithPayload 或拦截器注入载荷
func GetPayloadFromCtx(ctx context.Context) *Payload {
	raw := ctx.Value(payloadKey)
	if raw == nil {
		return nil
	}
	if p, ok := raw.(*Payload); ok {
		return p
	}
	return nil
}

// WithPayload 将 Payload 注入到上下文
func WithPayload(ctx context.Context, payload *Payload) context.Context {
	return context.WithValue(ctx, payloadKey, payload)
}

// WithScopeContext 便捷函数，将作用域信息注入到上下文
func WithScopeContext(ctx context.Context, domain int32, tenantID, roleCode string) context.Context {
	ctx = context.WithValue(ctx, ContextKeyDomain, domain)
	ctx = context.WithValue(ctx, ContextKeyTenantID, tenantID)
	ctx = context.WithValue(ctx, ContextKeyRoleCode, roleCode)
	return ctx
}

// WithOwner 将显式 Owner 标记注入到上下文。
func WithOwner(ctx context.Context, isOwner bool) context.Context {
	return context.WithValue(ctx, ContextKeyIsOwner, isOwner)
}
