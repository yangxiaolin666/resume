// Package cache 封装 Redis 的使用场景。
//
// 【这个文件在项目里的定位】
// 简历里写了「结合 Redis 缓存热点职位数据」「Redis 管理在线状态」，
// 本文件就是这两件事的落地实现，同时把「收藏集合」「消息消费幂等键」也统一收口。
//
// 【为什么要做内存降级】
// 学习/演示环境不一定装了 Redis。如果 Redis 连接失败就直接 panic，
// 整个项目就跑不起来，体验很差。这里的策略是：
//   - 启动时 Ping 一次，失败则标记为「内存模式」；
//   - 所有方法在内存模式下用 map 模拟同样的语义；
//   - 业务代码不需要关心当前用的是 Redis 还是内存。
//
// 【学习重点】
//  1. go-redis v9 的基本用法：Set / Get / SAdd / SMembers / SetNX / Incr；
//  2. 缓存三大问题的应对：
//     - 缓存穿透：空结果也缓存（本项目的职位详情不存在时直接返回错误，不缓存）；
//     - 缓存击穿：singleflight 让同一 key 的并发回源只有一个请求打到 MySQL；
//     - 缓存雪崩：TTL 加随机抖动，避免大量 key 同时过期。
//  3. 缓存一致性：更新职位后「删除缓存 + 版本号自增」，读请求使用新版本 key，
//     旧版本数据会在 TTL 到期后自然淘汰。
package cache

import (
	"context"
	"encoding/json"
	"log"
	"math/rand"
	"strconv"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/singleflight"
)

// Config 是 Redis 的连接参数（由 config.Config 传入，避免 cache 反向依赖 config）。
type Config struct {
	Addr     string
	Password string
	DB       int
	Enabled  bool
}

// Store 是缓存门面：内部要么走 Redis，要么走进程内 map。
type Store struct {
	rdb *redis.Client // 为 nil 表示当前处于内存降级模式
	sf  singleflight.Group

	mu      sync.RWMutex
	mem     map[string]memEntry      // 通用 KV 缓存（带过期时间）
	favs    map[int]map[int]struct{} // userID -> 收藏的 studentID 集合
	online  map[string]int           // "student:11" -> 连接数
	usedKey map[string]time.Time     // 幂等键（内存模式）
}

type memEntry struct {
	data    []byte
	expires time.Time
}

// New 创建缓存实例。
// 注意：这里不会因为 Redis 连不上而返回错误，而是打印日志后降级，
// 保证「没有中间件也能把项目跑起来」，方便学习和本地调试。
func New(cfg Config) *Store {
	s := &Store{
		mem:     make(map[string]memEntry),
		favs:    make(map[int]map[int]struct{}),
		online:  make(map[string]int),
		usedKey: make(map[string]time.Time),
	}
	if !cfg.Enabled {
		log.Printf("[cache] 已通过配置关闭 Redis，使用进程内内存缓存")
		return s
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:         cfg.Addr,
		Password:     cfg.Password,
		DB:           cfg.DB,
		DialTimeout:  2 * time.Second,
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 2 * time.Second,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Printf("[cache] Redis 连接失败(%v)，自动降级为内存缓存", err)
		_ = rdb.Close()
		return s
	}
	s.rdb = rdb
	log.Printf("[cache] Redis 连接成功: %s db=%d", cfg.Addr, cfg.DB)
	return s
}

// Enabled 表示当前是否真的在使用 Redis。
func (s *Store) Enabled() bool { return s != nil && s.rdb != nil }

// Close 释放 Redis 连接。
func (s *Store) Close() error {
	if s == nil || s.rdb == nil {
		return nil
	}
	return s.rdb.Close()
}

// Get 读取原始字节。第二个返回值表示是否命中。
func (s *Store) Get(ctx context.Context, key string) ([]byte, bool) {
	if s == nil {
		return nil, false
	}
	if s.rdb != nil {
		data, err := s.rdb.Get(ctx, key).Bytes()
		if err == nil {
			return data, true
		}
		if err != redis.Nil {
			log.Printf("[cache] GET %s 失败: %v", key, err)
		}
		return nil, false
	}

	s.mu.RLock()
	entry, ok := s.mem[key]
	s.mu.RUnlock()
	if !ok || time.Now().After(entry.expires) {
		return nil, false
	}
	return entry.data, true
}

// Set 写入并设置过期时间。
// ttl 会被加上最多 10% 的随机抖动，防止缓存雪崩。
func (s *Store) Set(ctx context.Context, key string, data []byte, ttl time.Duration) {
	if s == nil || len(data) == 0 {
		return
	}
	if ttl > 0 {
		ttl += time.Duration(rand.Int63n(int64(ttl) / 10))
	}
	if s.rdb != nil {
		if err := s.rdb.Set(ctx, key, data, ttl).Err(); err != nil {
			log.Printf("[cache] SET %s 失败: %v", key, err)
		}
		return
	}

	s.mu.Lock()
	s.mem[key] = memEntry{data: data, expires: time.Now().Add(ttl)}
	s.mu.Unlock()
}

// Del 删除一个或多个 key（缓存一致性用）。
func (s *Store) Del(ctx context.Context, keys ...string) {
	if s == nil || len(keys) == 0 {
		return
	}
	if s.rdb != nil {
		if err := s.rdb.Del(ctx, keys...).Err(); err != nil {
			log.Printf("[cache] DEL 失败: %v", err)
		}
		return
	}
	s.mu.Lock()
	for _, key := range keys {
		delete(s.mem, key)
	}
	s.mu.Unlock()
}

// GetOrLoad 是「读缓存 -> 未命中回源 -> 写缓存」的通用封装。
//
// 泛型 T 让职位列表、职位详情等不同结构体都能复用这段逻辑。
// singleflight 的作用：同一时刻 100 个请求查同一个 key，只放 1 个去查 MySQL，
// 其余 99 个等待并共享结果，避免缓存击穿把数据库打挂。
func GetOrLoad[T any](ctx context.Context, s *Store, key string, ttl time.Duration, load func() (T, error)) (T, error) {
	var zero T
	if data, ok := s.Get(ctx, key); ok {
		var value T
		if err := json.Unmarshal(data, &value); err == nil {
			return value, nil
		}
	}

	value, err, _ := s.sf.Do(key, func() (any, error) {
		item, err := load()
		if err != nil {
			return zero, err
		}
		if data, err := json.Marshal(item); err == nil {
			s.Set(ctx, key, data, ttl)
		}
		return item, nil
	})
	if err != nil {
		return zero, err
	}
	result, _ := value.(T)
	return result, nil
}

// ---------- 热点职位缓存版本号 ----------

// JobVersionKey 是「职位缓存版本」在 Redis 中的 key。
const JobVersionKey = "resume:job:version"

// JobVersion 返回当前职位缓存版本号。
// 读职位列表/详情时把它拼进缓存 key：
//
//	resume:job:v3:list?status=active
//
// 一旦职位发生增删改，就调用 BumpJobVersion 把版本 +1，
// 旧版本的缓存整体失效，不再需要按前缀扫描删除（SCAN 在大 key 空间下很慢）。
func (s *Store) JobVersion(ctx context.Context) int64 {
	if s == nil {
		return 0
	}
	if s.rdb != nil {
		value, err := s.rdb.Get(ctx, JobVersionKey).Int64()
		if err == redis.Nil {
			return 0
		}
		if err == nil {
			return value
		}
		log.Printf("[cache] 读取职位缓存版本失败: %v", err)
	}
	raw, ok := s.Get(ctx, JobVersionKey)
	if !ok {
		return 0
	}
	value, _ := strconv.ParseInt(string(raw), 10, 64)
	return value
}

// BumpJobVersion 让职位缓存版本自增，副作用是「旧缓存立即逻辑失效」。
func (s *Store) BumpJobVersion(ctx context.Context) {
	if s == nil {
		return
	}
	if s.rdb != nil {
		if err := s.rdb.Incr(ctx, JobVersionKey).Err(); err != nil {
			log.Printf("[cache] 职位缓存版本自增失败: %v", err)
		}
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	entry := s.mem[JobVersionKey]
	version, _ := strconv.ParseInt(string(entry.data), 10, 64)
	s.mem[JobVersionKey] = memEntry{
		data:    []byte(strconv.FormatInt(version+1, 10)),
		expires: time.Now().Add(24 * time.Hour),
	}
}

// ---------- 收藏集合（简历里的 LikeIfByRedis） ----------

func favoriteKey(userID int) string { return "resume:favorite:" + strconv.Itoa(userID) }

// AddFavorite 企业收藏学生。
func (s *Store) AddFavorite(ctx context.Context, userID, studentID int) {
	if s == nil {
		return
	}
	if s.rdb != nil {
		if err := s.rdb.SAdd(ctx, favoriteKey(userID), studentID).Err(); err != nil {
			log.Printf("[cache] SADD 失败: %v", err)
		}
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	set := s.favs[userID]
	if set == nil {
		set = make(map[int]struct{})
		s.favs[userID] = set
	}
	set[studentID] = struct{}{}
}

// RemoveFavorite 取消收藏。
func (s *Store) RemoveFavorite(ctx context.Context, userID, studentID int) {
	if s == nil {
		return
	}
	if s.rdb != nil {
		if err := s.rdb.SRem(ctx, favoriteKey(userID), studentID).Err(); err != nil {
			log.Printf("[cache] SREM 失败: %v", err)
		}
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if set := s.favs[userID]; set != nil {
		delete(set, studentID)
	}
}

// BatchIsFavorite 批量判断「这些学生是否被我收藏」。
// Redis 用 SMISMEMBER 一次网络往返拿到全部结果，比循环 SISMEMBER 快得多。
func (s *Store) BatchIsFavorite(ctx context.Context, userID int, studentIDs []int) map[int]bool {
	result := make(map[int]bool, len(studentIDs))
	if s == nil || len(studentIDs) == 0 {
		return result
	}
	if s.rdb != nil {
		args := make([]any, len(studentIDs))
		for i, id := range studentIDs {
			args[i] = id
		}
		values, err := s.rdb.SMIsMember(ctx, favoriteKey(userID), args...).Result()
		if err == nil {
			for i, id := range studentIDs {
				result[id] = values[i]
			}
			return result
		}
		log.Printf("[cache] SMISMEMBER 失败: %v", err)
		return result
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	set := s.favs[userID]
	for _, id := range studentIDs {
		_, result[id] = set[id]
	}
	return result
}

// ---------- 在线状态 ----------

func onlineKey(userID int, userType string) string {
	return "resume:online:" + userType + ":" + strconv.Itoa(userID)
}

// SetOnline 记录一个 WebSocket 连接上线。
//
// 用「计数」而不是简单的 1/0，是因为同一个用户可能开多个标签页。
// 设置 5 分钟 TTL 是兜底：即使进程被 kill 来不及执行下线逻辑，
// 状态也会自动过期，不会永远显示在线。
func (s *Store) SetOnline(ctx context.Context, userID int, userType string) {
	if s == nil {
		return
	}
	key := onlineKey(userID, userType)
	if s.rdb != nil {
		pipe := s.rdb.TxPipeline()
		pipe.Incr(ctx, key)
		pipe.Expire(ctx, key, 5*time.Minute)
		if _, err := pipe.Exec(ctx); err != nil {
			log.Printf("[cache] 更新在线状态失败: %v", err)
		}
		return
	}
	s.mu.Lock()
	s.online[key]++
	s.mu.Unlock()
}

// SetOffline 一个连接断开后计数减一，减到 0 就删除 key。
func (s *Store) SetOffline(ctx context.Context, userID int, userType string) {
	if s == nil {
		return
	}
	key := onlineKey(userID, userType)
	if s.rdb != nil {
		count, err := s.rdb.Decr(ctx, key).Result()
		if err != nil {
			log.Printf("[cache] 更新离线状态失败: %v", err)
			return
		}
		if count <= 0 {
			s.rdb.Del(ctx, key)
		}
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.online[key]--
	if s.online[key] <= 0 {
		delete(s.online, key)
	}
}

// IsOnline 查询在线状态（跨实例可见，因为真正的状态在 Redis 里）。
func (s *Store) IsOnline(ctx context.Context, userID int, userType string) bool {
	if s == nil {
		return false
	}
	if s.rdb != nil {
		count, err := s.rdb.Exists(ctx, onlineKey(userID, userType)).Result()
		return err == nil && count > 0
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.online[onlineKey(userID, userType)] > 0
}

// ---------- 消息消费幂等 ----------

// MarkOnce 尝试占位一个幂等键，返回 true 表示「第一次见到」。
// 典型用法：消费者收到 MQ 消息 -> MarkOnce(eventID) -> false 就直接 ACK 丢弃。
func (s *Store) MarkOnce(ctx context.Context, key string, ttl time.Duration) bool {
	if s == nil {
		return true
	}
	if s.rdb != nil {
		ok, err := s.rdb.SetNX(ctx, key, "1", ttl).Result()
		if err != nil {
			// Redis 故障时放行，让 MySQL 的唯一索引做最终兜底
			log.Printf("[cache] SetNX 失败: %v", err)
			return true
		}
		return ok
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	for k, expire := range s.usedKey {
		if now.After(expire) {
			delete(s.usedKey, k)
		}
	}
	if expire, ok := s.usedKey[key]; ok && now.Before(expire) {
		return false
	}
	s.usedKey[key] = now.Add(ttl)
	return true
}
