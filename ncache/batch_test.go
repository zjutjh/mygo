package ncache

import (
	"context"
	"sort"
	"testing"
	"time"
)

// fakeBatchLayer implements Layer and batchLayer for testing without real Redis.
type fakeBatchLayer struct {
	items map[string]fakeBatchItem
}

type fakeBatchItem struct {
	value  []byte
	expire time.Time
}

func newFakeBatchLayer() *fakeBatchLayer {
	return &fakeBatchLayer{items: make(map[string]fakeBatchItem)}
}

func (f *fakeBatchLayer) Get(_ context.Context, key string) ([]byte, time.Duration, bool, error) {
	item, ok := f.items[key]
	if !ok {
		return nil, 0, false, nil
	}
	if !item.expire.IsZero() && time.Now().After(item.expire) {
		delete(f.items, key)
		return nil, 0, false, nil
	}
	var ttlLeft time.Duration
	if !item.expire.IsZero() {
		ttlLeft = time.Until(item.expire)
	}
	return item.value, ttlLeft, true, nil
}

func (f *fakeBatchLayer) Set(_ context.Context, key string, val []byte, ttl time.Duration) error {
	var expire time.Time
	if ttl > 0 {
		expire = time.Now().Add(ttl)
	}
	f.items[key] = fakeBatchItem{value: append([]byte(nil), val...), expire: expire}
	return nil
}

func (f *fakeBatchLayer) Delete(_ context.Context, key string) error {
	delete(f.items, key)
	return nil
}

func (f *fakeBatchLayer) MGet(ctx context.Context, keys []string) (map[string][]byte, []string, error) {
	hits := make(map[string][]byte)
	missing := make([]string, 0, len(keys))
	for _, key := range keys {
		val, _, ok, err := f.Get(ctx, key)
		if err != nil {
			return nil, nil, err
		}
		if !ok {
			missing = append(missing, key)
			continue
		}
		hits[key] = append([]byte(nil), val...)
	}
	return hits, missing, nil
}

func (f *fakeBatchLayer) MSet(ctx context.Context, items map[string][]byte, ttl time.Duration) error {
	for key, val := range items {
		if err := f.Set(ctx, key, val, ttl); err != nil {
			return err
		}
	}
	return nil
}

func TestMultiCache_BatchOps_MemoryAndFakeLayer(t *testing.T) {
	t.Helper()

	mem := newMemoryLayer(MemoryConfig{MaxEntries: 128})
	t.Cleanup(func() { _ = mem.Close() })
	fake := newFakeBatchLayer()

	mc := &multiCache{
		conf: Config{
			Enable:      true,
			DefaultTTL:  time.Minute,
			NegativeTTL: 30 * time.Second,
		},
		layers: []Layer{mem, fake},
	}

	ctx := context.Background()

	// Seed L2 only and ensure MGet backfills memory layer.
	seedPayload := encodeRecord([]byte("seed"), false)
	_ = fake.Set(ctx, mc.namespaced("from_redis"), seedPayload, time.Minute)

	hits, missing, err := mc.MGet(ctx, []string{"from_redis", "absent"})
	if err != nil {
		t.Fatalf("MGet returned error: %v", err)
	}
	if len(hits) != 1 || string(hits["from_redis"]) != "seed" {
		t.Fatalf("unexpected hits: %+v", hits)
	}
	if len(missing) != 1 || missing[0] != "absent" {
		t.Fatalf("unexpected missing: %+v", missing)
	}

	// Verify backfill to memory layer.
	if _, _, ok, _ := mem.Get(ctx, mc.namespaced("from_redis")); !ok {
		t.Fatalf("memory layer not backfilled")
	}

	// Test MSet writes to both layers with per-key jitter applied (ttl > 0).
	items := map[string][]byte{
		"k1": []byte("v1"),
		"k2": []byte("v2"),
	}
	if err := mc.MSet(ctx, items, time.Minute); err != nil {
		t.Fatalf("MSet returned error: %v", err)
	}
	for key, want := range items {
		if got, err := mc.Get(ctx, key); err != nil || string(got) != string(want) {
			t.Fatalf("expected %s=%s, got %s (err=%v)", key, want, string(got), err)
		}
		entry, ok := fake.items[mc.namespaced(key)]
		if !ok {
			t.Fatalf("fake layer missing %s", key)
		}
		rec, err := decodeRecord(entry.value)
		if err != nil {
			t.Fatalf("decode fake payload error: %v", err)
		}
		if string(rec.V) != string(want) {
			t.Fatalf("fake layer stored unexpected value for %s", key)
		}
	}

	// Remove k2 so loader must repopulate it during MRemember.
	if err := mc.Delete(ctx, "k2"); err != nil {
		t.Fatalf("Delete error: %v", err)
	}

	// Prepare for MRemember: ensure k1 already cached, loader should not see it.
	loaderCalls := 0
	var loaderKeys []string
	loader := func(ctx context.Context, keys []string) (map[string]LoadedValue, error) {
		loaderCalls++
		loaderKeys = append(loaderKeys, keys...)
		values := make(map[string]LoadedValue)
		for _, k := range keys {
			if k == "k2" {
				values[k] = LoadedValue{Value: []byte("v2_loaded"), TTL: 2 * time.Minute}
			}
			if k == "k_missing" {
				// simulate not found by omitting from map
			}
		}
		return values, nil
	}

	keys := []string{"k1", "k2", "k_missing"}
	vals, err := mc.MRemember(ctx, keys, loader)
	if err != nil {
		t.Fatalf("MRemember error: %v", err)
	}
	if loaderCalls != 1 {
		t.Fatalf("loader called %d times", loaderCalls)
	}
	if len(loaderKeys) != 2 || loaderKeys[0] != "k2" || loaderKeys[1] != "k_missing" {
		t.Fatalf("unexpected loader keys: %v", loaderKeys)
	}
	if string(vals["k1"]) != "v1" || string(vals["k2"]) != "v2_loaded" {
		t.Fatalf("unexpected remember results: %+v", vals)
	}
	if _, ok := vals["k_missing"]; ok {
		t.Fatalf("k_missing should not appear in results")
	}
	if _, err := mc.Get(ctx, "k_missing"); err == nil {
		t.Fatalf("expected k_missing to be negative cached")
	}
}

func TestMultiCache_MGet_NegativeCache(t *testing.T) {
	mem := newMemoryLayer(MemoryConfig{})
	t.Cleanup(func() { _ = mem.Close() })

	mc := &multiCache{
		conf:   Config{DefaultTTL: time.Minute, NegativeTTL: 20 * time.Second},
		layers: []Layer{mem},
	}
	ctx := context.Background()

	// Populate negative cache via MRemember loader returning nothing.
	loader := func(ctx context.Context, keys []string) (map[string]LoadedValue, error) {
		return nil, nil
	}
	if _, err := mc.MRemember(ctx, []string{"neg"}, loader); err != nil {
		t.Fatalf("MRemember error: %v", err)
	}

	hits, missing, err := mc.MGet(ctx, []string{"neg"})
	if err != nil {
		t.Fatalf("MGet error: %v", err)
	}
	if len(hits) != 0 {
		t.Fatalf("expected no hits, got %+v", hits)
	}
	if len(missing) != 1 || missing[0] != "neg" {
		t.Fatalf("negative key should be reported missing, got %+v", missing)
	}
}

func TestDedupeKeys(t *testing.T) {
	out := dedupeKeys([]string{"a", "", "b", "a", "c"})
	sort.Strings(out)
	expected := []string{"a", "b", "c"}
	for i, v := range expected {
		if out[i] != v {
			t.Fatalf("expected %v, got %v", expected, out)
		}
	}
}
