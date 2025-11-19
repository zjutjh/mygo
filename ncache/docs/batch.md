# ncache 批量查询（Batch）设计提案

本文提出在现有 ncache 上引入批量 API 的方案，优先保证与现有接口不冲突、实现简单、收益明显，并为后续高级聚合（跨请求合并）预留扩展点。

## 目标
- 提供批量读写与批量回源：MGet / MSet / MRemember。
- 继续支持多级缓存（L1 内存 + L2 Redis）、负缓存与 TTL 抖动。
- 批量回源时仅触发一次 loader 调用（对本次调用内缺失的 keys）。
- 不破坏现有 `Cache` 接口；以新增接口形式提供。

## 非目标（V1 不做）
- 跨请求的自动合并（coalescing）：不同 goroutine/请求同时请求同一批 keys 的聚合。
- 写合并与写回（write-behind）。
- 热点保护、布隆过滤器等高级能力。

## API 设计（新增）
```go
// BatchCache 在不破坏现有 Cache 的前提下引入批量能力
// multiCache 将实现该接口。
type BatchCache interface {
    Cache

    // MGet：批量读取缓存。
    // 返回：
    //  - hits: 命中的 key->值；
    //  - missing: 本次未命中的 keys；
    //  - err: 系统错误（层读失败等）。
    MGet(ctx context.Context, keys []string) (hits map[string][]byte, missing []string, err error)

    // MSet：批量写入缓存（所有层）。
    MSet(ctx context.Context, items map[string][]byte, ttl time.Duration) error

    // BatchLoaderFunc：批量回源函数。返回找到的 key->(值, TTL)。
    // 对于未找到的 key，请不要出现在返回的 map 中（即“缺席即未命中”）。
    type BatchLoaderFunc func(ctx context.Context, keys []string) (map[string]LoadedValue, error)

    type LoadedValue struct {
        Value []byte
        TTL   time.Duration // 0 表示使用默认 TTL
    }

    // MRemember：按需批量回源并回填。
    // - 读取路径：先多级 MGet；
    // - 对本次调用内 missing 由 loader 一次性回源；
    // - 回填：命中与负缓存（未找到）分别写入；
    // - 返回：最终所有 keys 的值（不包含未命中的 key）。
    MRemember(ctx context.Context, keys []string, loader BatchLoaderFunc, opts ...Option) (map[string][]byte, error)
}
```

说明：
- 兼容现有 `Option`（TTL、回源策略），但 V1 中批量仍默认使用「旁路 + 单飞」的语义，仅在单个调用内生效；跨请求合并留待 V2。
- 负缓存策略沿用：loader 未返回的 keys 视为未命中，使用 `NegativeTTL` 写入负缓存。

## 读取算法（MGet/MRemember）
1. 规范化 keys（去重、过滤空）。
2. 自上而下按层读取：
   - L1 内存：逐键读取，识别负缓存（N==true）记为 missing。
   - 记录已命中项；对未命中继续向下层查找。
   - 从更低层命中后，按需回填更高层（与单键逻辑一致）。
3. 汇总本次仍然 missing 的 keys：
   - 若为 MGet：直接返回 hits + missing。
   - 若为 MRemember：调用 batch loader 一次，返回 map[key]LoadedValue。
4. 回填：
   - 对 loader 返回的命中：逐层 Set（TTL 使用返回的 TTL 或默认 TTL，经 `applyJitter`）。
   - 对未返回的 keys：写入负缓存（`NegativeTTL` + 抖动）。
5. 返回合并后的结果（不包含未命中的键）。

## 写入算法（MSet）
- 对每层执行批量写入：
  - 内存层：循环 Store。
  - Redis 层：使用 Pipeline 批量 SET EX（go-redis v9 无 MSETEX，可逐键 SET EX）。
- TTL 统一应用 `applyJitter`，建议在逐键时应用，保证分散过期。

## Redis 细节
- 批量读：MGET，逐个解码 `record{V,N}`；N=true 视为未命中。
- 批量写：Pipeline (SET key payload EX ttl)。
- 命中回填：当从 Redis 命中时，同步回填 L1（同单键行为）。

## 单飞/合并策略
- V1：单调用内的「双重检查 + 批量回源」。对同一调用中的缺失 keys 只调用一次 loader；不同调用之间不合并。
- V2（可选）：引入微小聚合窗口（5~20ms）：
  - 以 (prefix+loaderName) 作为“批量组键”，收敛多个并发请求；
  - 使用 map[key]set[requestID] 收集缺失 keys；到时统一触发 loader；
  - 分发结果到各请求。实现复杂度较高，需严格控制窗口与内存占用。

## 边界与错误处理
- keys 为空：立即返回 empty。
- 部分系统错误：
  - 任一层读失败：中止并返回 error（保持与现有语义一致）。
  - 写回失败：记录第一个错误，继续尝试其余层；最终返回第一个错误。
- TTL：
  - loader 未指定 TTL（0）：使用默认 TTL；均应用抖动且下限 1s。
  - 负缓存 TTL：使用 `NegativeTTL`（抖动）。
- 命名空间前缀：统一通过现有 `namespaced` 处理。

## 兼容性与迁移
- 不改变现有 `Cache` 接口。
- 新增 `BatchCache` 接口由 `multiCache` 实现；调用方可通过类型断言使用：
  ```go
  if bc, ok := cache.(ncache.BatchCache); ok {
      hits, missing, _ := bc.MGet(ctx, keys)
      // ...
  }
  ```

## 开发计划（建议）
- V1（约 1~2 天）
  - 实现 MGet/MSet/MRemember（无跨请求合并）。
  - 内存层/Redis 层各自实现批量读写（Redis 使用 Pipeline）。
  - 单元测试：
    - MGet 命中/未命中/负缓存识别；
    - MRemember 仅一次 loader 调用、负缓存落盘；
    - TTL 抖动上下界；
    - 多层回填（从 Redis 命中回填内存）。
- V2（可选，视需求）
  - 聚合窗口 coalescing；
  - 熔断/限速（防止大批量 keys 导致 loader 压力峰值）。

## 选择建议（与第三方库对比）
- 若仅需基础批量能力（80% 场景）：推荐内部实现 V1，
  - 代码量小、与现有 DI（nedis/nlog）与约定完全一致；
  - 无外部依赖、可快速迭代；
  - 行为与现有单键路径一致（负缓存、抖动、写回）。
- 若需要高级批量合并/热点保护/写回等：可以调研第三方库的实现思路并择优吸收，避免直接替换带来迁移成本与行为差异。

## 测试验证（已完工）

- `ncache/batch_test.go` 使用内存层 + 自实现的 `fakeBatchLayer`（模拟 Redis 层）覆盖以下场景：
  - `MGet` 从下层命中后回填上层，并正确返回缺失键；
  - `MSet` 逐键写入两层，校验编码后的 payload；
  - `MRemember` 同一次调用只触发一次 loader，命中与负缓存写回正确；
  - 负缓存会在批量读取中被识别为 missing；
  - `dedupeKeys` 去空去重逻辑。 
- 测试指令：`go test ./ncache`（默认包含上述批量用例）。
- 如需对接真实 Redis，可额外编写集成测试（需事先 Boot 对应的 nedis scope）。

## 与 Redis 分布式锁的兼容性

- `ncache` 通过 `nedis.Pick(scope)` 获取 Redis UniversalClient；
- `lock` 包（基于 `redsync`）同样依赖 `nedis` 提供的客户端。两者仅共享底层 Redis 连接，不会直接冲突；
- 需确保配置文件中给缓存和锁分配不同的 key 前缀，避免命名空间互相污染（例如在 `ncache.Config.Prefix` 中设定专用前缀）；
- 如果在同一 Redis 实例下使用，两者连接数、命令超时等参数统一由 nedis 配置控制。 

