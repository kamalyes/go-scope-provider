/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-05-11 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-05-13 23:31:34
 * @FilePath: \go-scope-provider\payload.go
 * @Description: 作用域载荷定义 - 通用的作用域数据结构，不依赖 protobuf
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */
package provider

import (
	"encoding/json"
	"strings"
	"sync"

	"github.com/kamalyes/go-sqlbuilder/scope"
)

const DefaultOwnerRoleCode = "owner"

var ownerRoleCodes = struct {
	sync.RWMutex
	values []string
}{values: []string{DefaultOwnerRoleCode}}

// Payload 认证载荷，包含用户身份和作用域信息
// 可通过 JSON 序列化传输（如 gRPC metadata、HTTP Header）
type Payload struct {
	Domain        int32         `json:"domain"`
	TenantID      string        `json:"tenant_id"`
	UserID        string        `json:"user_id"`
	RoleCode      string        `json:"role_code"`
	IsOwner       bool          `json:"is_owner"`
	ScopeBindings []*ScopeEntry `json:"scope_bindings,omitempty"`
}

// SetDefaultOwnerRoleCodes 设置默认 Owner 角色编码集合。
func SetDefaultOwnerRoleCodes(codes ...string) {
	ownerRoleCodes.Lock()
	defer ownerRoleCodes.Unlock()
	ownerRoleCodes.values = normalizeOwnerRoleCodes(codes)
}

// IsOwnerRoleCode 判断角色编码是否为 Owner 角色编码。
func IsOwnerRoleCode(roleCode string, codes ...string) bool {
	roleCode = strings.TrimSpace(roleCode)
	if roleCode == "" {
		return false
	}
	if len(codes) == 0 {
		ownerRoleCodes.RLock()
		codes = append([]string(nil), ownerRoleCodes.values...)
		ownerRoleCodes.RUnlock()
	}
	for _, code := range codes {
		if roleCode == strings.TrimSpace(code) {
			return true
		}
	}
	return false
}

func normalizeOwnerRoleCodes(codes []string) []string {
	seen := make(map[string]struct{}, len(codes))
	result := make([]string, 0, len(codes))
	for _, code := range codes {
		code = strings.TrimSpace(code)
		if code == "" {
			continue
		}
		if _, ok := seen[code]; ok {
			continue
		}
		seen[code] = struct{}{}
		result = append(result, code)
	}
	if len(result) == 0 {
		return []string{DefaultOwnerRoleCode}
	}
	return result
}

// ScopeEntry 作用域条目，定义用户的访问范围
type ScopeEntry struct {
	ScopeType       int32             `json:"scope_type"`
	RegionCodes     []string          `json:"region_codes,omitempty"`
	RegionPlatforms []*RegionPlatform `json:"region_platforms,omitempty"`
	TenantIds       []string          `json:"tenant_ids,omitempty"`
}

// RegionPlatform 地区-平台绑定关系
type RegionPlatform struct {
	RegionCode  string   `json:"region_code"`
	PlatformIds []string `json:"platform_ids,omitempty"`
}

// ToJSON 将 Payload 序列化为 JSON 字符串
func (p *Payload) ToJSON() (string, error) {
	data, err := json.Marshal(p)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// FromJSON 从 JSON 字符串反序列化 Payload
func FromJSON(data string) (*Payload, error) {
	var p Payload
	if err := json.Unmarshal([]byte(data), &p); err != nil {
		return nil, err
	}
	return &p, nil
}

// ToScopeData 将 Payload 转换为 scope.ScopeData
func (p *Payload) ToScopeData(opts ...scope.Option) scope.ScopeData {
	data := scope.NewScopeData(opts...)
	data.Domain = p.Domain
	data.TenantID = p.TenantID
	data.IsOwner = p.IsOwner || IsOwnerRoleCode(p.RoleCode)
	data.ScopeEntries = convertScopeEntries(p.ScopeBindings)
	return data
}

// convertScopeEntries 将 provider.ScopeEntry 转换为 scope.ScopeEntry
func convertScopeEntries(entries []*ScopeEntry) []*scope.ScopeEntry {
	if len(entries) == 0 {
		return nil
	}
	result := make([]*scope.ScopeEntry, 0, len(entries))
	for _, e := range entries {
		if e == nil {
			continue
		}
		se := &scope.ScopeEntry{
			ScopeType:   e.ScopeType,
			RegionCodes: e.RegionCodes,
			TenantIds:   e.TenantIds,
		}
		for _, rp := range e.RegionPlatforms {
			if rp != nil {
				se.RegionPlatforms = append(se.RegionPlatforms, &scope.RegionPlatformEntry{
					RegionCode:  rp.RegionCode,
					PlatformIds: rp.PlatformIds,
				})
			}
		}
		result = append(result, se)
	}
	return result
}
