package ncache

import (
	"context"
	"time"
)

// 批量扩展：可选的层批量接口
type batchLayer interface {
	MGet(ctx context.Context, keys []string) (map[string][]byte, []string, error)
	MSet(ctx context.Context, items map[string][]byte, ttl time.Duration) error
}

// MGet 批量读取：自上而下按层查找，识别负缓存；从低层命中会向上回填。
func (c *multiCache) MGet(ctx context.Context, keys []string) (map[string][]byte, []string, error) {
	// 预处理 keys：去空 + 去重
	origKeys := dedupeKeys(keys)
	if len(origKeys) == 0 {
		return map[string][]byte{}, nil, nil
	}
	// 原始key -> 命名空间key
	ns := make(map[string]string, len(origKeys))
	pending := make(map[string]struct{}, len(origKeys))
	for _, k := range origKeys {
		nk := c.namespaced(k)
		ns[k] = nk
		pending[k] = struct{}{}
	}

	results := make(map[string][]byte, len(origKeys))

	// 逐层查找
	for i, l := range c.layers {
		if len(pending) == 0 {
			break
		}
		// 准备当前层需要查询的 ns keys
		curOrig := make([]string, 0, len(pending))
		curNS := make([]string, 0, len(pending))
		for k := range pending {
			curOrig = append(curOrig, k)
			curNS = append(curNS, ns[k])
		}

		// 批量或逐键读取，得到原始 payload（record 编码）
		hitsPayload := make(map[string][]byte)
		if bl, ok := l.(batchLayer); ok {
			h, _, err := bl.MGet(ctx, curNS)
			if err != nil {
				return nil, nil, err
			}
			// 反向映射：ns->orig
			idx := make(map[string]string, len(curOrig))
			for _, k := range curOrig {
				idx[ns[k]] = k
			}
			for nk, b := range h {
				if orig, ok := idx[nk]; ok {
					hitsPayload[orig] = b
				}
			}
		} else {
			for _, k := range curOrig {
				b, _, ok, err := l.Get(ctx, ns[k])
				if err != nil {
					return nil, nil, err
				}
				if ok {
					hitsPayload[k] = b
				}
			}
		}

		if len(hitsPayload) == 0 {
			continue
		}

		// 解码与分类；准备向上回填的 payload
		upfill := make(map[string][]byte) // nsKey -> payload
		for k, payload := range hitsPayload {
			rec, err := decodeRecord(payload)
			if err != nil {
				// 数据损坏，视为未命中
				continue
			}
			if rec.N {
				// 负缓存：不计入命中；保持在 pending（对 MGet 表现为 missing）
				delete(hitsPayload, k)
				continue
			}
			// 命中：加入结果，并从 pending 移除
			results[k] = rec.V
			delete(pending, k)
			upfill[ns[k]] = payload
		}

		if len(upfill) == 0 || i == 0 {
			continue
		}
		// 从低层命中，回填到更高层（0..i-1），使用默认TTL（与单键回填保持一致，不应用抖动）
		for j := 0; j < i; j++ {
			if bl, ok := c.layers[j].(batchLayer); ok {
				_ = bl.MSet(ctx, upfill, c.conf.DefaultTTL)
			} else {
				for nk, payload := range upfill {
					_ = c.layers[j].Set(ctx, nk, payload, c.conf.DefaultTTL)
				}
			}
		}
	}

	// 未命中的（包含负缓存键）
	missing := make([]string, 0, len(pending))
	for k := range pending {
		missing = append(missing, k)
	}
	return results, missing, nil
}

// MSet 批量写入：遍历所有层，逐键写入以保留每键抖动。
func (c *multiCache) MSet(ctx context.Context, items map[string][]byte, ttl time.Duration) error {
	if len(items) == 0 {
		return nil
	}
	// 构建 payload 与每键 TTL（应用抖动）
	type kv struct {
		nk      string
		payload []byte
		ttl     time.Duration
	}
	rows := make([]kv, 0, len(items))
	for k, v := range items {
		nk := c.namespaced(k)
		t := ttl
		if t <= 0 {
			t = c.conf.DefaultTTL
		}
		t = c.applyJitter(t)
		rows = append(rows, kv{nk: nk, payload: encodeRecord(v, false), ttl: t})
	}
	var firstErr error
	for _, l := range c.layers {
		for _, r := range rows {
			if err := l.Set(ctx, r.nk, r.payload, r.ttl); err != nil && firstErr == nil {
				firstErr = err
			}
		}
	}
	return firstErr
}

// MRemember 批量记忆：一次请求内对缺失 keys 调用一次 loader；命中回填、未找到写负缓存。
func (c *multiCache) MRemember(ctx context.Context, keys []string, loader BatchLoaderFunc, opts ...Option) (map[string][]byte, error) {
	// 选项
	ro := rememberOpts{ttl: c.conf.DefaultTTL, policy: SourcePolicy(c.conf.Policy)}
	for _, opt := range opts {
		opt(&ro)
	}
	if ro.ttl <= 0 {
		ro.ttl = c.conf.DefaultTTL
	}

	// 先批量读取
	hits, missing, err := c.MGet(ctx, keys)
	if err != nil {
		return nil, err
	}
	if len(missing) == 0 {
		return hits, nil
	}

	// V1：单次调用内的“合并回源”，不做跨请求单飞
	loaded, err := loader(ctx, missing)
	if err != nil {
		return nil, err
	}

	// 写回命中
	for k, lv := range loaded {
		v := lv.Value
		ttl := lv.TTL
		if ttl <= 0 {
			ttl = ro.ttl
		}
		_ = c.Set(ctx, k, v, ttl) // 复用单键 Set，以便每键抖动
		hits[k] = v
	}

	// 对 loader 未返回的 keys 写负缓存
	loadedSet := make(map[string]struct{}, len(loaded))
	for k := range loaded {
		loadedSet[k] = struct{}{}
	}
	for _, k := range missing {
		if _, ok := loadedSet[k]; ok {
			continue
		}
		_ = c.setNegative(ctx, k)
	}

	return hits, nil
}

// 工具：去空并去重
func dedupeKeys(keys []string) []string {
	if len(keys) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(keys))
	out := make([]string, 0, len(keys))
	for _, k := range keys {
		if k == "" {
			continue
		}
		if _, ok := seen[k]; ok {
			continue
		}
		seen[k] = struct{}{}
		out = append(out, k)
	}
	return out
}
