package aibalance

import (
	"fmt"
	"math"
	"math/rand"
	"sync"

	"github.com/yaklang/yaklang/common/log"
	"github.com/yaklang/yaklang/common/utils/omap"
)

type ModelEntry struct {
	ModelName string      `json:"model_name"`
	Providers []*Provider `json:"providers"`

	// 为每个 ModelEntry 添加互斥锁，避免对同一模型的提供者列表进行并发修改
	m sync.RWMutex
}

type Entrypoint struct {
	ModelEntries *omap.OrderedMap[string, *ModelEntry]

	m sync.RWMutex
}

func NewEntrypoint() *Entrypoint {
	return &Entrypoint{
		ModelEntries: omap.NewOrderedMap(make(map[string]*ModelEntry)),
	}
}

func (e *Entrypoint) CreateModelEntry(modelName string) *ModelEntry {
	return &ModelEntry{
		ModelName: modelName,
		Providers: []*Provider{},
	}
}

func (e *Entrypoint) AddProvider(modelName string, provider *Provider) {
	e.m.Lock()
	defer e.m.Unlock()

	if entry, ok := e.ModelEntries.Get(modelName); ok {
		entry.m.Lock()
		entry.Providers = append(entry.Providers, provider)
		entry.m.Unlock()
	} else {
		newEntry := &ModelEntry{
			ModelName: modelName,
			Providers: []*Provider{provider},
		}
		e.ModelEntries.Set(modelName, newEntry)
	}
}

func (e *Entrypoint) PeekProvider(modelName string) *Provider {
	e.m.RLock()
	entry, ok := e.ModelEntries.Get(modelName)
	e.m.RUnlock()

	if !ok {
		return nil
	}

	entry.m.RLock()
	defer entry.m.RUnlock()

	if len(entry.Providers) == 0 {
		return nil
	}

	// 如果只有一个 Provider，直接返回
	if len(entry.Providers) == 1 {
		return entry.Providers[0]
	}

	// 为每个 Provider 获取数据库状态信息
	type ProviderScore struct {
		provider      *Provider
		isHealthy     bool
		totalRequests int64
		latency       int64
		score         float64
	}

	// 收集所有可用的 Provider 信息
	var availableProviders []ProviderScore
	var totalHealthyRequests int64 = 0
	var minRequests int64 = -1
	var maxRequests int64 = 0

	// 提前复制 providers 列表以减少锁持有时间
	providers := make([]*Provider, len(entry.Providers))
	copy(providers, entry.Providers)

	// 第一轮：收集基本信息
	for _, provider := range providers {
		dbProvider, err := provider.GetDbProvider()
		if err != nil {
			log.Warnf("Failed to get db provider: %v", err)
			continue
		}

		// 必须有数据库记录且健康状态正常才考虑使用
		isHealthy := dbProvider.IsHealthy
		totalRequests := dbProvider.TotalRequests
		latency := dbProvider.LastLatency

		// 只有延迟大于 0 的提供者才会被考虑
		if latency <= 0 {
			continue
		}

		// 累计健康的 Provider 总请求数
		if isHealthy {
			totalHealthyRequests += totalRequests

			// 更新最小和最大请求数
			if minRequests == -1 || totalRequests < minRequests {
				minRequests = totalRequests
			}
			if totalRequests > maxRequests {
				maxRequests = totalRequests
			}
		}

		// 收集可用的 Provider 信息
		availableProviders = append(availableProviders, ProviderScore{
			provider:      provider,
			isHealthy:     isHealthy,
			totalRequests: totalRequests,
			latency:       latency,
		})
	}

	// 如果没有健康的 Provider，则尝试使用任何 Provider
	if len(availableProviders) == 0 || totalHealthyRequests == 0 {
		// 随机选择一个提供者
		if len(providers) == 0 {
			return nil
		}
		randomIndex := rand.Intn(len(providers))
		return providers[randomIndex]
	}

	// 最小与最大请求数差异
	requestDiff := maxRequests - minRequests

	// 第二轮：计算综合得分
	for i := range availableProviders {
		p := &availableProviders[i]

		if !p.isHealthy {
			// 非健康 Provider 的得分为 0，最低优先级
			p.score = 0
			continue
		}

		// 负载均衡得分：请求数越少得分越高，最高1.0
		// 当所有 Provider 请求数基本相同时，这个值接近于1
		loadBalanceScore := 1.0
		if requestDiff > 0 {
			// 将请求数归一化到 [0, 1] 范围
			normalizedRequests := float64(p.totalRequests-minRequests) / float64(requestDiff)
			// 反转，使得请求数少的得分高
			loadBalanceScore = 1.0 - normalizedRequests
		}

		// 延迟得分：延迟越低得分越高，范围 [0, 1]
		// 使用对数函数来平滑延迟差异，避免少量延迟差异导致分数差距过大
		latencyScore := 1.0
		if p.latency > 0 {
			// 根据经验，300ms 以下的延迟都很好，大于 1000ms 开始明显变差
			normalizedLatency := math.Min(math.Log10(float64(p.latency)/100), 1.0)
			latencyScore = 1.0 - normalizedLatency
		}

		// 综合得分：负载均衡占 60%，延迟占 40%
		p.score = loadBalanceScore*0.6 + latencyScore*0.4
	}

	// 第三轮：使用加权随机选择
	// 得分高的 Provider 被选中的概率更高，但不是确定性的
	// 这样可以避免所有请求都涌向同一个最优 Provider

	// 计算总得分
	var totalScore float64 = 0
	for _, p := range availableProviders {
		if p.isHealthy {
			totalScore += p.score
		}
	}

	if totalScore <= 0 {
		// 如果所有 Provider 得分都为 0，随机选择一个健康的
		healthyProviders := []ProviderScore{}
		for _, p := range availableProviders {
			if p.isHealthy {
				healthyProviders = append(healthyProviders, p)
			}
		}

		if len(healthyProviders) > 0 {
			randIndex := rand.Intn(len(healthyProviders))
			return healthyProviders[randIndex].provider
		}

		// 如果没有健康的，随机选择任意一个
		if len(availableProviders) > 0 {
			randIndex := rand.Intn(len(availableProviders))
			return availableProviders[randIndex].provider
		}

		return nil
	}

	// 加权随机选择
	r := rand.Float64() * totalScore
	var cumulativeScore float64 = 0
	for _, p := range availableProviders {
		if !p.isHealthy {
			continue
		}

		cumulativeScore += p.score
		if cumulativeScore >= r {
			return p.provider
		}
	}

	// 以防万一，返回第一个健康的 Provider
	for _, p := range availableProviders {
		if p.isHealthy {
			return p.provider
		}
	}

	return nil
}

// InitializeFromDB 从数据库加载Provider
func (e *Entrypoint) InitializeFromDB() error {
	// 从数据库加载所有提供者
	providers, err := GetAllAiProviders()
	if err != nil {
		return fmt.Errorf("failed to load providers from database: %v", err)
	}

	// 添加到入口点
	for _, dbProvider := range providers {
		provider := &Provider{
			TypeName:    dbProvider.TypeName,
			ModelName:   dbProvider.ModelName,
			DomainOrURL: dbProvider.DomainOrURL,
			APIKey:      dbProvider.APIKey,
			NoHTTPS:     dbProvider.NoHTTPS,
			DbProvider:  dbProvider, // 直接关联数据库对象
		}
		e.AddProvider(dbProvider.WrapperName, provider)
	}

	return nil
}
