// Package: cos
// File: caddyfile.go
// Author: Kavi
// Created: 2025-11-15
// Updated: 2025-11-15
// Description: caddy插件集成
//
// -----------------------------------------------------------------------------
// Copyright (c) 2025 Kavi. All rights reserved.
// This file is licensed under the Hsiyue Software License v1.0.
// See the LICENSE file in the project root for the full license text.
// For commercial licensing, contact: kavi@hsiyue.com
// -----------------------------------------------------------------------------

package cos

import (
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/certmagic"
)

func init() {
	caddy.RegisterModule(CaddyStorage{})
}

type CaddyStorage struct {
	Bucket    string `json:"bucket,omitempty"`
	Region    string `json:"region,omitempty"`
	SecretID  string `json:"secret_id,omitempty"`
	SecretKey string `json:"secret_key,omitempty"`
	Prefix    string `json:"prefix,omitempty"`

	storage *Storage
}

func (CaddyStorage) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "caddy.storage.cos",
		New: func() caddy.Module { return new(CaddyStorage) },
	}
}

func (cs *CaddyStorage) Provision(ctx caddy.Context) error {
	storage, err := NewStorage(Config{
		Bucket:    cs.Bucket,
		Region:    cs.Region,
		SecretID:  cs.SecretID,
		SecretKey: cs.SecretKey,
		Prefix:    cs.Prefix,
	})
	if err != nil {
		return err
	}
	cs.storage = storage
	return nil
}

// 返回CertMagic存储接口
func (cs *CaddyStorage) CertMagicStorage() (certmagic.Storage, error) {
	return cs.storage, nil
}

// 解析Caddyfile配置
func (cs *CaddyStorage) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	for d.Next() {
		for d.NextBlock(0) {
			switch d.Val() {
			case "bucket":
				if !d.Args(&cs.Bucket) {
					return d.ArgErr()
				}
			case "region":
				if !d.Args(&cs.Region) {
					return d.ArgErr()
				}
			case "secret_id":
				if !d.Args(&cs.SecretID) {
					return d.ArgErr()
				}
			case "secret_key":
				if !d.Args(&cs.SecretKey) {
					return d.ArgErr()
				}
			case "prefix":
				if !d.Args(&cs.Prefix) {
					return d.ArgErr()
				}
			}
		}
	}
	return nil
}

// 接口检查
var (
	_ caddy.Provisioner     = (*CaddyStorage)(nil)
	_ caddyfile.Unmarshaler = (*CaddyStorage)(nil)
	_ certmagic.Storage     = (*Storage)(nil)
)
