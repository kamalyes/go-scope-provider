/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-05-11 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-05-13 00:05:21
 * @FilePath: \go-scope-provider\provider_test.go
 * @Description: 作用域提供者测试
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */
package provider

import (
	"context"
	"testing"

	"github.com/kamalyes/go-sqlbuilder/repository"
	"github.com/kamalyes/go-sqlbuilder/scope"
	"github.com/stretchr/testify/assert"
)

// 场景：Payload 完整字段序列化和反序列化
// 预期：ToJSON 和 FromJSON 往返一致
func TestPayload_ToJSON_FromJSON(t *testing.T) {
	original := &Payload{
		Domain:   1,
		TenantID: "T001",
		UserID:   "U001",
		RoleCode: "owner",
		IsOwner:  true,
		ScopeBindings: []*ScopeEntry{
			{
				ScopeType:   2,
				RegionCodes: []string{"MM", "TH"},
			},
			{
				ScopeType: 3,
				RegionPlatforms: []*RegionPlatform{
					{RegionCode: "SG", PlatformIds: []string{"P1", "P2"}},
				},
			},
		},
	}

	jsonStr, err := original.ToJSON()
	assert.NoError(t, err)
	assert.NotEmpty(t, jsonStr)

	restored, err := FromJSON(jsonStr)
	assert.NoError(t, err)
	assert.Equal(t, original.Domain, restored.Domain)
	assert.Equal(t, original.TenantID, restored.TenantID)
	assert.Equal(t, original.UserID, restored.UserID)
	assert.Equal(t, original.RoleCode, restored.RoleCode)
	assert.Equal(t, original.IsOwner, restored.IsOwner)
	assert.Len(t, restored.ScopeBindings, 2)
	assert.Equal(t, int32(2), restored.ScopeBindings[0].ScopeType)
	assert.Equal(t, []string{"MM", "TH"}, restored.ScopeBindings[0].RegionCodes)
	assert.Equal(t, int32(3), restored.ScopeBindings[1].ScopeType)
	assert.Len(t, restored.ScopeBindings[1].RegionPlatforms, 1)
	assert.Equal(t, "SG", restored.ScopeBindings[1].RegionPlatforms[0].RegionCode)
	assert.Equal(t, []string{"P1", "P2"}, restored.ScopeBindings[1].RegionPlatforms[0].PlatformIds)
}

// 场景：Payload 使用显式 IsOwner 标记
// 预期：即使角色编码不是默认 owner，IsOwner 仍为 true
func TestPayload_ToScopeData_ExplicitOwner(t *testing.T) {
	p := &Payload{
		Domain:   1,
		TenantID: "T001",
		RoleCode: "tenant-admin",
		IsOwner:  true,
	}

	data := p.ToScopeData()
	assert.True(t, data.IsOwner)
}

// 场景：自定义默认 Owner 角色编码
// 预期：自定义编码可识别为 Owner
func TestPayload_ToScopeData_CustomOwnerRoleCode(t *testing.T) {
	SetDefaultOwnerRoleCodes("tenant-owner")
	defer SetDefaultOwnerRoleCodes(DefaultOwnerRoleCode)

	p := &Payload{
		Domain:   1,
		TenantID: "T001",
		RoleCode: "tenant-owner",
	}

	data := p.ToScopeData()
	assert.True(t, data.IsOwner)
}

// 场景：无效 JSON 反序列化
// 预期：返回错误
func TestFromJSON_Invalid(t *testing.T) {
	_, err := FromJSON("{invalid}")
	assert.Error(t, err)
}

// 场景：Payload 转换为 ScopeData（owner 角色）
// 预期：IsOwner=true，ScopeEntries 正确转换
func TestPayload_ToScopeData_Owner(t *testing.T) {
	p := &Payload{
		Domain:   1,
		TenantID: "T001",
		RoleCode: "owner",
		ScopeBindings: []*ScopeEntry{
			{ScopeType: 1},
		},
	}

	data := p.ToScopeData()
	assert.Equal(t, int32(1), data.Domain)
	assert.Equal(t, "T001", data.TenantID)
	assert.True(t, data.IsOwner)
	assert.Len(t, data.ScopeEntries, 1)
	assert.Equal(t, int32(1), data.ScopeEntries[0].ScopeType)
}

// 场景：Payload 转换为 ScopeData（非 owner 角色）
// 预期：IsOwner=false
func TestPayload_ToScopeData_NonOwner(t *testing.T) {
	p := &Payload{
		Domain:   1,
		TenantID: "T001",
		RoleCode: "admin",
	}
	data := p.ToScopeData()
	assert.False(t, data.IsOwner)
}

// 场景：Payload 转换时 RegionPlatforms 正确映射
// 预期：RegionPlatformEntry 正确转换
func TestPayload_ToScopeData_RegionPlatforms(t *testing.T) {
	p := &Payload{
		Domain:   1,
		TenantID: "T001",
		RoleCode: "user",
		ScopeBindings: []*ScopeEntry{
			{
				ScopeType: 3,
				RegionPlatforms: []*RegionPlatform{
					{RegionCode: "MM", PlatformIds: []string{"P1"}},
					{RegionCode: "SG", PlatformIds: []string{"P2", "P3"}},
				},
			},
		},
	}

	data := p.ToScopeData()
	assert.Len(t, data.ScopeEntries, 1)
	assert.Len(t, data.ScopeEntries[0].RegionPlatforms, 2)
	assert.Equal(t, "MM", data.ScopeEntries[0].RegionPlatforms[0].RegionCode)
	assert.Equal(t, []string{"P1"}, data.ScopeEntries[0].RegionPlatforms[0].PlatformIds)
	assert.Equal(t, "SG", data.ScopeEntries[0].RegionPlatforms[1].RegionCode)
	assert.Equal(t, []string{"P2", "P3"}, data.ScopeEntries[0].RegionPlatforms[1].PlatformIds)
}

// 场景：Payload 空的 ScopeBindings
// 预期：ScopeEntries 为 nil
func TestPayload_ToScopeData_EmptyBindings(t *testing.T) {
	p := &Payload{
		Domain:   2,
		TenantID: "T001",
		RoleCode: "admin",
	}
	data := p.ToScopeData()
	assert.Nil(t, data.ScopeEntries)
}

// 场景：Payload 中 nil ScopeEntry 被跳过
// 预期：结果不包含 nil 条目
func TestPayload_ToScopeData_NilEntry(t *testing.T) {
	p := &Payload{
		Domain:   1,
		TenantID: "T001",
		RoleCode: "user",
		ScopeBindings: []*ScopeEntry{
			nil,
			{ScopeType: 2, RegionCodes: []string{"MM"}},
			nil,
		},
	}
	data := p.ToScopeData()
	assert.Len(t, data.ScopeEntries, 1)
}

// 场景：WithPayload 和 GetPayloadFromCtx 注入与提取
// 预期：注入后可正确提取
func TestWithPayload_GetPayloadFromCtx(t *testing.T) {
	ctx := context.Background()
	assert.Nil(t, GetPayloadFromCtx(ctx))

	p := &Payload{Domain: 1, TenantID: "T001", RoleCode: "owner"}
	ctx = WithPayload(ctx, p)

	got := GetPayloadFromCtx(ctx)
	assert.Equal(t, p, got)
}

// 场景：GetPayloadFromCtx 上下文中存了非 Payload 类型
// 预期：返回 nil
func TestGetPayloadFromCtx_WrongType(t *testing.T) {
	ctx := context.WithValue(context.Background(), payloadKey, "not-a-payload")
	assert.Nil(t, GetPayloadFromCtx(ctx))
}

// 场景：WithScopeContext 注入作用域信息
// 预期：ResolveScopeData 正确读取
func TestWithScopeContext_ResolveScopeData(t *testing.T) {
	ctx := context.Background()
	ctx = WithScopeContext(ctx, 1, "T001", "owner")

	data := ResolveScopeData(ctx)
	assert.Equal(t, int32(1), data.Domain)
	assert.Equal(t, "T001", data.TenantID)
	assert.True(t, data.IsOwner)
}

// 场景：ResolveScopeData 优先使用 Payload
// 预期：Payload 优先级高于 WithScopeContext
func TestResolveScopeData_PayloadPriority(t *testing.T) {
	ctx := context.Background()
	ctx = WithScopeContext(ctx, 2, "T-CTX", "admin")

	p := &Payload{
		Domain:   1,
		TenantID: "T-PAYLOAD",
		RoleCode: "owner",
		ScopeBindings: []*ScopeEntry{
			{ScopeType: 2, RegionCodes: []string{"MM"}},
		},
	}
	ctx = WithPayload(ctx, p)

	data := ResolveScopeData(ctx)
	assert.Equal(t, int32(1), data.Domain)
	assert.Equal(t, "T-PAYLOAD", data.TenantID)
	assert.True(t, data.IsOwner)
	assert.Len(t, data.ScopeEntries, 1)
}

// 场景：ResolveScopeData 上下文中无作用域信息
// 预期：返回默认 ScopeData
func TestResolveScopeData_EmptyContext(t *testing.T) {
	data := ResolveScopeData(context.Background())
	assert.Equal(t, int32(0), data.Domain)
	assert.Empty(t, data.TenantID)
	assert.False(t, data.IsOwner)
}

// 场景：ApplyScopeQuery 完整流程
// 预期：返回带 FilterGroup 的 Query
func TestApplyScopeQuery(t *testing.T) {
	ctx := context.Background()
	p := &Payload{
		Domain:   1,
		TenantID: "T001",
		RoleCode: "user",
		ScopeBindings: []*ScopeEntry{
			{ScopeType: 2, RegionCodes: []string{"MM"}},
		},
	}
	ctx = WithPayload(ctx, p)

	query := repository.NewQuery()
	result := ApplyScopeQuery(ctx, query)

	assert.NotNil(t, result.FilterGroup)
	assert.Equal(t, "tenant_id", result.FilterGroup.Filters[0].Field)
	assert.Equal(t, "T001", result.FilterGroup.Filters[0].Value)
}

// 场景：ResolveScopeData 使用自定义 Option
// 预期：自定义字段映射生效
func TestResolveScopeData_WithOption(t *testing.T) {
	ctx := context.Background()
	p := &Payload{
		Domain:   1,
		TenantID: "T001",
		RoleCode: "user",
		ScopeBindings: []*ScopeEntry{
			{ScopeType: 2, RegionCodes: []string{"MM"}},
		},
	}
	ctx = WithPayload(ctx, p)

	data := ResolveScopeData(ctx, scope.WithTenantIDField("org_id"))
	assert.Equal(t, "org_id", data.Config.Mapping.TenantIDField)
}

// 场景：ResolveScopeData 从上下文读取 string 类型的 Domain
// 预期：字符串 "2" 被解析为 int32(2)
func TestResolveScopeData_DomainAsString(t *testing.T) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, ContextKeyDomain, "2")
	ctx = context.WithValue(ctx, ContextKeyTenantID, "T001")
	ctx = context.WithValue(ctx, ContextKeyRoleCode, "admin")

	data := ResolveScopeData(ctx)
	assert.Equal(t, int32(2), data.Domain)
	assert.Equal(t, "T001", data.TenantID)
	assert.False(t, data.IsOwner)
}

// 场景：ResolveScopeData 从上下文读取 int 类型的 Domain
// 预期：int(1) 被转换为 int32(1)
func TestResolveScopeData_DomainAsInt(t *testing.T) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, ContextKeyDomain, 1)
	ctx = context.WithValue(ctx, ContextKeyTenantID, "T001")
	ctx = context.WithValue(ctx, ContextKeyRoleCode, "owner")

	data := ResolveScopeData(ctx)
	assert.Equal(t, int32(1), data.Domain)
	assert.True(t, data.IsOwner)
}

// 场景：ResolveScopeData 从上下文读取无效 string 类型的 Domain
// 预期：Domain 保持默认值 0
func TestResolveScopeData_DomainAsInvalidString(t *testing.T) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, ContextKeyDomain, "not-a-number")

	data := ResolveScopeData(ctx)
	assert.Equal(t, int32(0), data.Domain)
}
