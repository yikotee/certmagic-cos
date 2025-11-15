// Package: cos
// File: config.go
// Author: Kavi
// Created: 2025-11-14
// Updated: 2025-11-14
// Description: 存储cos基本配置信息
//
// -----------------------------------------------------------------------------
// Copyright (c) 2025 Kavi. All rights reserved.
// This file is licensed under the Hsiyue Software License v1.0.
// See the LICENSE file in the project root for the full license text.
// For commercial licensing, contact: kavi@hsiyue.com
// -----------------------------------------------------------------------------

package cos

type Config struct {
	Bucket    string // 存储桶名
	Region    string // 区域，例如 “ap-nanjing”
	SecretID  string // SecretID
	SecretKey string // SecretKey
	Prefix    string // 存放文件的子目录
}
