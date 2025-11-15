// Package: cos
// File: cos.go
// Author: Kavi
// Created: 2025-11-14
// Updated: 2025-11-14
// Description: CertMagic实现COS存储
//
// -----------------------------------------------------------------------------
// Copyright (c) 2025 Kavi. All rights reserved.
// This file is licensed under the Hsiyue Software License v1.0.
// See the LICENSE file in the project root for the full license text.
// For commercial licensing, contact: kavi@hsiyue.com
// -----------------------------------------------------------------------------

package cos

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/caddyserver/certmagic"
	"github.com/tencentyun/cos-go-sdk-v5"
)

// certmagic.Storage接口
type Storage struct {
	client *cos.Client
	prefix string
}

func NewStorage(config Config) (*Storage, error) {
	bucketURL := fmt.Sprintf("https://%s.cos.%s.myqcloud.com",
		config.Bucket, config.Region)

	parsedURL, err := url.Parse(bucketURL)
	if err != nil {
		return nil, fmt.Errorf("解析 bucket 地址失败: %w", err)
	}

	cosClient := cos.NewClient(&cos.BaseURL{BucketURL: parsedURL}, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  config.SecretID,
			SecretKey: config.SecretKey,
		},
	})

	return &Storage{
		client: cosClient,
		prefix: strings.Trim(config.Prefix, "/"),
	}, nil
}

// 添加前缀到键名
func (s *Storage) buildKey(key string) string {
	if s.prefix == "" {
		return key
	}
	return path.Join(s.prefix, key)
}

// 从键名移除前缀
func (s *Storage) removePrefix(key string) string {
	if s.prefix == "" {
		return key
	}
	return strings.TrimPrefix(key, s.prefix+"/")
}

// 转换COS错误为标准错误
func (s *Storage) convertError(err error) error {
	if err == nil {
		return nil
	}
	if strings.Contains(err.Error(), "404") {
		return fs.ErrNotExist
	}
	return err
}

// 保存数据到COS
func (s *Storage) Store(ctx context.Context, key string, value []byte) error {
	fullKey := s.buildKey(key)
	_, err := s.client.Object.Put(ctx, fullKey, strings.NewReader(string(value)), nil)
	return err
}

// 从COS读取数据
func (s *Storage) Load(ctx context.Context, key string) ([]byte, error) {
	fullKey := s.buildKey(key)
	response, err := s.client.Object.Get(ctx, fullKey, nil)
	if err != nil {
		return nil, s.convertError(err)
	}
	defer response.Body.Close()

	return io.ReadAll(response.Body)
}

// 检查文件是否存在
func (s *Storage) Exists(ctx context.Context, key string) bool {
	fullKey := s.buildKey(key)
	_, err := s.client.Object.Head(ctx, fullKey, nil)
	return err == nil
}

// 删除文件
func (s *Storage) Delete(ctx context.Context, key string) error {
	fullKey := s.buildKey(key)
	_, err := s.client.Object.Delete(ctx, fullKey)
	return s.convertError(err)
}

// 列出指定前缀下的所有键
func (s *Storage) List(ctx context.Context, prefix string, recursive bool) ([]string, error) {
	var keys []string
	marker := ""
	for {
		opt := &cos.BucketGetOptions{
			Prefix:  s.buildKey(prefix),
			Marker:  marker,
			MaxKeys: 1000,
		}
		if !recursive {
			opt.Delimiter = "/"
		}

		result, _, err := s.client.Bucket.Get(ctx, opt)
		if err != nil {
			return nil, s.convertError(err)
		}

		for _, obj := range result.Contents {
			keys = append(keys, s.removePrefix(obj.Key))
		}

		if !result.IsTruncated {
			break
		}
		marker = result.NextMarker
	}
	return keys, nil
}

// 获取文件元信息
func (s *Storage) Stat(ctx context.Context, key string) (certmagic.KeyInfo, error) {
	fullKey := s.buildKey(key)
	resp, err := s.client.Object.Head(ctx, fullKey, nil)
	if err != nil {
		return certmagic.KeyInfo{}, s.convertError(err)
	}

	var mod time.Time
	if lm := resp.Header.Get("Last-Modified"); lm != "" {
		if t, perr := http.ParseTime(lm); perr == nil {
			mod = t
		}
	}

	return certmagic.KeyInfo{
		Key:        s.removePrefix(fullKey),
		Modified:   mod,
		Size:       resp.ContentLength,
		IsTerminal: true,
	}, nil
}

// 获取分布式锁
func (s *Storage) Lock(ctx context.Context, key string) error {
	lockKey := s.buildKey(key + ".lock")
	lockContent := time.Now().String()

	for {
		_, err := s.client.Object.Head(ctx, lockKey, nil)
		if err != nil {
			_, putErr := s.client.Object.Put(ctx, lockKey,
				strings.NewReader(lockContent), nil)
			if putErr == nil {
				return nil
			}
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Second):
		}
	}
}

// 释放分布式锁
func (s *Storage) Unlock(ctx context.Context, key string) error {
	lockKey := s.buildKey(key + ".lock")
	_, err := s.client.Object.Delete(ctx, lockKey)
	return err
}

var _ certmagic.Storage = (*Storage)(nil)
