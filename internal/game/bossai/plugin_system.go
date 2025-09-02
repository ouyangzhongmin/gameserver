package bossai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/ouyangzhongmin/gameserver/pkg/logger"
)

// BaseAIPlugin 基础AI插件实现
type BaseAIPlugin struct {
	name     string
	version  string
	enabled  bool
	priority int
	config   map[string]interface{}

	// 统计信息
	updateCount     int64
	errorCount      int64
	totalUpdateTime time.Duration
	lastUpdateTime  time.Time

	mutex sync.RWMutex
}

// NewBaseAIPlugin 创建基础AI插件
func NewBaseAIPlugin(name, version string) *BaseAIPlugin {
	return &BaseAIPlugin{
		name:     name,
		version:  version,
		enabled:  true,
		priority: 0,
		config:   make(map[string]interface{}),
	}
}

func (p *BaseAIPlugin) GetName() string {
	return p.name
}

func (p *BaseAIPlugin) GetVersion() string {
	return p.version
}

func (p *BaseAIPlugin) SetEnabled(enabled bool) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.enabled = enabled
}

func (p *BaseAIPlugin) SetPriority(priority int) {
	p.priority = priority
}

func (p *BaseAIPlugin) GetPriority() int {
	return p.priority
}

func (p *BaseAIPlugin) SetConfig(config map[string]interface{}) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.config = make(map[string]interface{})
	for k, v := range config {
		p.config[k] = v
	}
}

func (p *BaseAIPlugin) IsEnabled() bool {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	return p.enabled
}

func (p *BaseAIPlugin) GetConfig() map[string]interface{} {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	// 返回副本
	result := make(map[string]interface{})
	for k, v := range p.config {
		result[k] = v
	}

	return result
}

func (p *BaseAIPlugin) Initialize(ctx *BossContext) error {
	// 基础插件不需要特殊初始化
	return nil
}

func (p *BaseAIPlugin) Update(ctx *BossContext, deltaTime time.Duration) error {
	if !p.enabled {
		return nil
	}

	p.mutex.Lock()
	defer p.mutex.Unlock()

	start := time.Now()
	defer func() {
		p.totalUpdateTime += time.Since(start)
		p.lastUpdateTime = ctx.CurrentTime
		p.updateCount++
	}()

	// 基础插件不执行任何操作
	return nil
}

func (p *BaseAIPlugin) Cleanup() error {
	// 基础插件不需要特殊清理
	return nil
}

func (p *BaseAIPlugin) SuggestAction(ctx *BossContext) *AIAction {
	// 基础插件不提供建议
	return nil
}

func (p *BaseAIPlugin) ModifyBehavior(ctx *BossContext, originalAction *AIAction) *AIAction {
	// 基础插件不修改行为
	return originalAction
}

func (p *BaseAIPlugin) GetStatistics() PluginStatistics {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	stats := PluginStatistics{
		Name:            p.name,
		Version:         p.version,
		Enabled:         p.enabled,
		Priority:        p.priority,
		UpdateCount:     p.updateCount,
		ErrorCount:      p.errorCount,
		TotalUpdateTime: p.totalUpdateTime,
		LastUpdateTime:  p.lastUpdateTime,
	}

	if p.updateCount > 0 {
		stats.AverageUpdateTime = p.totalUpdateTime / time.Duration(p.updateCount)
	}

	return stats
}

type PluginStatistics struct {
	Name              string        `json:"name"`
	Version           string        `json:"version"`
	Enabled           bool          `json:"enabled"`
	Priority          int           `json:"priority"`
	UpdateCount       int64         `json:"update_count"`
	ErrorCount        int64         `json:"error_count"`
	TotalUpdateTime   time.Duration `json:"total_update_time"`
	AverageUpdateTime time.Duration `json:"average_update_time"`
	LastUpdateTime    time.Time     `json:"last_update_time"`
}

// CombatAnalysisPlugin 战斗分析插件
type CombatAnalysisPlugin struct {
	*BaseAIPlugin

	// 分析配置
	analyzeInterval time.Duration
	lastAnalyzeTime time.Time

	// 分析结果
	threatAssessment   map[int64]float64 // 实体ID -> 威胁值
	combatEfficiency   float64
	recommendedActions []AIAction
}

// NewCombatAnalysisPlugin 创建战斗分析插件
func NewCombatAnalysisPlugin() *CombatAnalysisPlugin {
	base := NewBaseAIPlugin("CombatAnalysis", "1.0.0")
	return &CombatAnalysisPlugin{
		BaseAIPlugin:       base,
		analyzeInterval:    time.Second * 2,
		threatAssessment:   make(map[int64]float64),
		recommendedActions: make([]AIAction, 0),
	}
}

func (p *CombatAnalysisPlugin) Initialize(ctx *BossContext) error {
	if err := p.BaseAIPlugin.Initialize(ctx); err != nil {
		return err
	}

	p.analyzeInterval = time.Duration(p.GetConfigInt("analyze_interval_ms", 2000)) * time.Millisecond

	logger.Debugf("CombatAnalysisPlugin initialized with interval %v", p.analyzeInterval)
	return nil
}

func (p *CombatAnalysisPlugin) Update(ctx *BossContext, deltaTime time.Duration) error {
	if err := p.BaseAIPlugin.Update(ctx, deltaTime); err != nil {
		return err
	}

	if !p.enabled {
		return nil
	}

	// 定期进行战斗分析
	if ctx.CurrentTime.Sub(p.lastAnalyzeTime) >= p.analyzeInterval {
		p.lastAnalyzeTime = ctx.CurrentTime
		p.performCombatAnalysis(ctx)
	}

	return nil
}

func (p *CombatAnalysisPlugin) SuggestAction(ctx *BossContext) *AIAction {
	if !p.enabled || len(p.recommendedActions) == 0 {
		return nil
	}

	// 返回优先级最高的推荐动作
	bestAction := &p.recommendedActions[0]
	for i := 1; i < len(p.recommendedActions); i++ {
		if p.recommendedActions[i].Priority > bestAction.Priority {
			bestAction = &p.recommendedActions[i]
		}
	}

	return bestAction
}

func (p *CombatAnalysisPlugin) performCombatAnalysis(ctx *BossContext) {
	// 清空旧的分析结果
	p.threatAssessment = make(map[int64]float64)
	p.recommendedActions = p.recommendedActions[:0]

	// 分析附近敌人的威胁值
	for _, enemy := range ctx.NearbyEnemies {
		threat := p.calculateThreatLevel(ctx, enemy)
		p.threatAssessment[enemy.GetID()] = threat
	}

	// 计算战斗效率
	p.combatEfficiency = p.calculateCombatEfficiency(ctx)

	// 生成推荐动作
	p.generateRecommendations(ctx)
}

func (p *CombatAnalysisPlugin) calculateThreatLevel(ctx *BossContext, enemy IEntity) float64 {
	// 基础威胁值计算
	threat := 1.0

	// 根据距离调整
	distance := calculateDistance(
		float64(ctx.Boss.GetPos().X), float64(ctx.Boss.GetPos().Y),
		float64(enemy.GetPos().X), float64(enemy.GetPos().Y),
	)

	if distance < 50 {
		threat *= 2.0 // 近距离威胁更高
	} else if distance > 200 {
		threat *= 0.5 // 远距离威胁较低
	}

	// 这里可以根据敌人类型、血量等因素进一步调整

	return threat
}

func (p *CombatAnalysisPlugin) calculateCombatEfficiency(ctx *BossContext) float64 {
	if ctx.CombatTime == 0 {
		return 1.0
	}

	// 简单的效率计算：伤害输出 / 战斗时间
	efficiency := float64(ctx.DamageDealt) / ctx.CombatTime.Seconds()

	// 受到伤害会降低效率
	if ctx.DamageReceived > 0 {
		efficiency *= (1.0 - float64(ctx.DamageReceived)/float64(ctx.Boss.GetMaxLife()))
	}

	return efficiency
}

func (p *CombatAnalysisPlugin) generateRecommendations(ctx *BossContext) {
	// 如果血量低，推荐防御性动作
	healthPercent := float64(ctx.Boss.GetCurrentLife()) / float64(ctx.Boss.GetMaxLife())
	if healthPercent < 0.3 {
		action := AIAction{
			Type:      ActionRetreat,
			Priority:  8,
			Timestamp: ctx.CurrentTime,
		}
		p.recommendedActions = append(p.recommendedActions, action)
	}

	// 如果有高威胁目标，推荐攻击
	for entityID, threat := range p.threatAssessment {
		if threat > 1.5 {
			action := AIAction{
				Type:      ActionAttack,
				TargetID:  entityID,
				Priority:  int(threat * 5),
				Timestamp: ctx.CurrentTime,
			}
			p.recommendedActions = append(p.recommendedActions, action)
		}
	}
}

func (p *CombatAnalysisPlugin) GetConfigInt(key string, defaultValue int) int {
	if value, exists := p.config[key]; exists {
		if intValue, ok := value.(int); ok {
			return intValue
		}
	}
	return defaultValue
}

// LLMProvider 大语言模型提供者实现
type OpenAIProvider struct {
	name     string
	apiKey   string
	endpoint string
	model    string
	client   *http.Client

	// 配置
	maxTokens   int
	temperature float64
	timeout     time.Duration

	// 统计
	requestCount int64
	errorCount   int64
	totalLatency time.Duration

	mutex sync.RWMutex
}

// NewOpenAIProvider 创建OpenAI提供者
func NewOpenAIProvider(apiKey string) *OpenAIProvider {
	return &OpenAIProvider{
		name:        "OpenAI",
		apiKey:      apiKey,
		endpoint:    "https://api.openai.com/v1/chat/completions",
		model:       "gpt-3.5-turbo",
		maxTokens:   1000,
		temperature: 0.7,
		timeout:     time.Second * 30,
		client: &http.Client{
			Timeout: time.Second * 30,
		},
	}
}

func (p *OpenAIProvider) GetName() string {
	return p.name
}

func (p *OpenAIProvider) Configure(config map[string]interface{}) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if endpoint, exists := config["endpoint"]; exists {
		if endpointStr, ok := endpoint.(string); ok {
			p.endpoint = endpointStr
		}
	}

	if model, exists := config["model"]; exists {
		if modelStr, ok := model.(string); ok {
			p.model = modelStr
		}
	}

	if maxTokens, exists := config["max_tokens"]; exists {
		if tokens, ok := maxTokens.(int); ok {
			p.maxTokens = tokens
		}
	}

	if temperature, exists := config["temperature"]; exists {
		if temp, ok := temperature.(float64); ok {
			p.temperature = temp
		}
	}

	if timeout, exists := config["timeout_ms"]; exists {
		if timeoutMs, ok := timeout.(int); ok {
			p.timeout = time.Duration(timeoutMs) * time.Millisecond
			p.client.Timeout = p.timeout
		}
	}

	return nil
}

func (p *OpenAIProvider) IsAvailable() bool {
	return p.apiKey != "" && p.endpoint != ""
}

func (p *OpenAIProvider) GenerateStrategy(ctx context.Context, bossState *BossSnapshot, gameState *GameSnapshot) (*AIStrategy, error) {
	if !p.IsAvailable() {
		return nil, fmt.Errorf("LLM provider not available")
	}

	p.mutex.Lock()
	p.requestCount++
	p.mutex.Unlock()

	start := time.Now()
	defer func() {
		p.mutex.Lock()
		p.totalLatency += time.Since(start)
		p.mutex.Unlock()
	}()

	// 构建提示词
	prompt := p.buildStrategyPrompt(bossState, gameState)

	// 调用LLM API
	response, err := p.callLLMAPI(ctx, prompt)
	if err != nil {
		p.mutex.Lock()
		p.errorCount++
		p.mutex.Unlock()
		return nil, err
	}

	// 解析响应
	strategy, err := p.parseStrategyResponse(response)
	if err != nil {
		p.mutex.Lock()
		p.errorCount++
		p.mutex.Unlock()
		return nil, err
	}

	return strategy, nil
}

func (p *OpenAIProvider) AdaptBehavior(ctx context.Context, feedback *ActionFeedback) (*BehaviorAdjustment, error) {
	if !p.IsAvailable() {
		return nil, fmt.Errorf("LLM provider not available")
	}

	p.mutex.Lock()
	p.requestCount++
	p.mutex.Unlock()

	start := time.Now()
	defer func() {
		p.mutex.Lock()
		p.totalLatency += time.Since(start)
		p.mutex.Unlock()
	}()

	// 构建适应性提示词
	prompt := p.buildAdaptationPrompt(feedback)

	// 调用LLM API
	response, err := p.callLLMAPI(ctx, prompt)
	if err != nil {
		p.mutex.Lock()
		p.errorCount++
		p.mutex.Unlock()
		return nil, err
	}

	// 解析响应
	adjustment, err := p.parseAdaptationResponse(response)
	if err != nil {
		p.mutex.Lock()
		p.errorCount++
		p.mutex.Unlock()
		return nil, err
	}

	return adjustment, nil
}

func (p *OpenAIProvider) buildStrategyPrompt(bossState *BossSnapshot, gameState *GameSnapshot) string {
	prompt := fmt.Sprintf(`You are an AI controller for a boss character in an MMORPG game.

Boss Status:
- Name: %s
- Health: %.1f%%
- Current State: %s
- Current Phase: %s
- Combat Duration: %.1f seconds
- Available Skills: %d

Game Environment:
- Nearby Enemies: %d
- Nearby Allies: %d

Please analyze the situation and suggest the best strategy. Consider:
1. Boss's current health and phase
2. Available skills and their effectiveness
3. Number and threat level of enemies
4. Optimal positioning and timing

Respond in JSON format with the following structure:
{
    "recommended_action": {
        "type": "attack|retreat|cast_skill|move",
        "skill_id": 123 (if casting skill),
        "target_id": 456 (if targeting specific entity),
        "priority": 1-10,
        "reasoning": "explanation"
    },
    "confidence": 0.0-1.0,
    "reasoning": "detailed explanation of the strategy"
}`,
		bossState.Name,
		bossState.HealthPercent*100,
		bossState.CurrentState,
		bossState.CurrentPhase,
		bossState.CombatTime,
		len(bossState.AvailableSkills),
		len(gameState.NearbyEnemies),
		len(gameState.NearbyAllies))

	return prompt
}

func (p *OpenAIProvider) buildAdaptationPrompt(feedback *ActionFeedback) string {
	prompt := fmt.Sprintf(`Analyze the effectiveness of a recent boss action and suggest behavioral adjustments.

Action Performed:
- Type: %s
- Success: %t
- Damage Dealt: %d
- Damage Received: %d
- Effectiveness: %.2f
- Target Reaction: %s

Please suggest adjustments to improve future performance. Consider:
1. Whether the action was effective
2. How to improve similar future actions
3. Priority adjustments for different skill types
4. Behavioral pattern modifications

Respond in JSON format:
{
    "skill_priorities": {"skill_id": priority_adjustment},
    "state_modifiers": {"modifier_name": adjustment_value},
    "confidence": 0.0-1.0,
    "reasoning": "explanation of adjustments"
}`,
		feedback.Action.Type,
		feedback.Success,
		feedback.DamageDealt,
		feedback.DamageReceived,
		feedback.Effectiveness,
		feedback.TargetReaction)

	return prompt
}

func (p *OpenAIProvider) callLLMAPI(ctx context.Context, prompt string) (string, error) {
	// 构建请求体
	requestBody := map[string]interface{}{
		"model": p.model,
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"max_tokens":  p.maxTokens,
		"temperature": p.temperature,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// 创建HTTP请求
	req, err := http.NewRequestWithContext(ctx, "POST", p.endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	// 发送请求
	resp, err := p.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// 解析响应
	var response struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no response from LLM")
	}

	return response.Choices[0].Message.Content, nil
}

func (p *OpenAIProvider) parseStrategyResponse(response string) (*AIStrategy, error) {
	var strategy AIStrategy

	if err := json.Unmarshal([]byte(response), &strategy); err != nil {
		return nil, fmt.Errorf("failed to parse strategy response: %w", err)
	}

	return &strategy, nil
}

func (p *OpenAIProvider) parseAdaptationResponse(response string) (*BehaviorAdjustment, error) {
	var adjustment BehaviorAdjustment

	if err := json.Unmarshal([]byte(response), &adjustment); err != nil {
		return nil, fmt.Errorf("failed to parse adaptation response: %w", err)
	}

	return &adjustment, nil
}

// LLMPlugin LLM集成插件
type LLMPlugin struct {
	*BaseAIPlugin

	provider       ILLMProvider
	updateInterval time.Duration
	lastUpdateTime time.Time

	// 策略缓存
	currentStrategy *AIStrategy
	strategyAge     time.Duration
	maxStrategyAge  time.Duration

	// 反馈收集
	actionFeedback  []ActionFeedback
	maxFeedbackSize int
}

// NewLLMPlugin 创建LLM插件
func NewLLMPlugin(provider ILLMProvider) *LLMPlugin {
	base := NewBaseAIPlugin("LLM", "1.0.0")
	return &LLMPlugin{
		BaseAIPlugin:    base,
		provider:        provider,
		updateInterval:  time.Second * 5,
		maxStrategyAge:  time.Second * 30,
		actionFeedback:  make([]ActionFeedback, 0),
		maxFeedbackSize: 100,
	}
}

func (p *LLMPlugin) Initialize(ctx *BossContext) error {
	if err := p.BaseAIPlugin.Initialize(ctx); err != nil {
		return err
	}

	// 配置LLM提供者
	if config, exists := p.config["llm_config"]; exists {
		if configMap, ok := config.(map[string]interface{}); ok {
			if err := p.provider.Configure(configMap); err != nil {
				return fmt.Errorf("failed to configure LLM provider: %w", err)
			}
		}
	}

	p.updateInterval = time.Duration(p.GetConfigInt("update_interval_ms", 5000)) * time.Millisecond
	p.maxStrategyAge = time.Duration(p.GetConfigInt("max_strategy_age_ms", 30000)) * time.Millisecond

	logger.Debugf("LLMPlugin initialized with provider %s", p.provider.GetName())
	return nil
}

func (p *LLMPlugin) Update(ctx *BossContext, deltaTime time.Duration) error {
	if err := p.BaseAIPlugin.Update(ctx, deltaTime); err != nil {
		return err
	}

	if !p.enabled || !p.provider.IsAvailable() {
		return nil
	}

	// 更新策略年龄
	if p.currentStrategy != nil {
		p.strategyAge += deltaTime
	}

	// 定期更新策略
	if ctx.CurrentTime.Sub(p.lastUpdateTime) >= p.updateInterval ||
		p.strategyAge > p.maxStrategyAge {

		p.lastUpdateTime = ctx.CurrentTime
		go p.updateStrategy(ctx)
	}

	return nil
}

func (p *LLMPlugin) SuggestAction(ctx *BossContext) *AIAction {
	if !p.enabled || p.currentStrategy == nil {
		return nil
	}

	return p.currentStrategy.RecommendedAction
}

func (p *LLMPlugin) ModifyBehavior(ctx *BossContext, originalAction *AIAction) *AIAction {
	if !p.enabled || p.currentStrategy == nil {
		return originalAction
	}

	// 可以根据LLM策略修改原始行为
	if p.currentStrategy.RecommendedAction != nil {
		// 如果LLM建议的行为优先级更高，使用LLM建议
		if p.currentStrategy.RecommendedAction.Priority > originalAction.Priority {
			return p.currentStrategy.RecommendedAction
		}
	}

	return originalAction
}

func (p *LLMPlugin) updateStrategy(ctx *BossContext) {
	// 创建Boss状态快照
	bossSnapshot := p.createBossSnapshot(ctx)

	// 创建游戏环境快照
	gameSnapshot := p.createGameSnapshot(ctx)

	// 调用LLM生成策略
	strategy, err := p.provider.GenerateStrategy(context.Background(), bossSnapshot, gameSnapshot)
	if err != nil {
		logger.Errorf("Failed to generate LLM strategy: %v", err)
		p.mutex.Lock()
		p.errorCount++
		p.mutex.Unlock()
		return
	}

	// 更新当前策略
	p.mutex.Lock()
	p.currentStrategy = strategy
	p.strategyAge = 0
	p.mutex.Unlock()

	logger.Debugf("LLM strategy updated: %s (confidence: %.2f)",
		strategy.Reasoning, strategy.Confidence)
}

func (p *LLMPlugin) createBossSnapshot(ctx *BossContext) *BossSnapshot {
	snapshot := &BossSnapshot{
		ID:              ctx.Boss.GetID(),
		HealthPercent:   float64(ctx.Boss.GetCurrentLife()) / float64(ctx.Boss.GetMaxLife()),
		Position:        ctx.Boss.GetPos(),
		CombatTime:      ctx.CombatTime.Seconds(),
		AvailableSkills: make([]SkillSnapshot, 0),
		StatusEffects:   make([]StatusEffect, 0),
	}

	if ctx.CurrentState != nil {
		snapshot.CurrentState = ctx.CurrentState.GetName()
	}

	if ctx.CurrentPhase != nil {
		snapshot.CurrentPhase = ctx.CurrentPhase.GetName()
	}

	return snapshot
}

func (p *LLMPlugin) createGameSnapshot(ctx *BossContext) *GameSnapshot {
	snapshot := &GameSnapshot{
		Timestamp:     ctx.CurrentTime,
		NearbyEnemies: make([]EntitySnapshot, 0),
		NearbyAllies:  make([]EntitySnapshot, 0),
	}

	// 添加附近敌人信息
	for _, enemy := range ctx.NearbyEnemies {
		enemySnapshot := EntitySnapshot{
			ID:       enemy.GetID(),
			Type:     "enemy",
			Position: enemy.GetPos(),
			Distance: calculateDistance(
				float64(ctx.Boss.GetPos().X), float64(ctx.Boss.GetPos().Y),
				float64(enemy.GetPos().X), float64(enemy.GetPos().Y),
			),
		}
		snapshot.NearbyEnemies = append(snapshot.NearbyEnemies, enemySnapshot)
	}

	// 添加附近盟友信息
	for _, ally := range ctx.NearbyAllies {
		allySnapshot := EntitySnapshot{
			ID:       ally.GetID(),
			Type:     "ally",
			Position: ally.GetPos(),
			Distance: calculateDistance(
				float64(ctx.Boss.GetPos().X), float64(ctx.Boss.GetPos().Y),
				float64(ally.GetPos().X), float64(ally.GetPos().Y),
			),
		}
		snapshot.NearbyAllies = append(snapshot.NearbyAllies, allySnapshot)
	}

	return snapshot
}

func (p *LLMPlugin) GetConfigInt(key string, defaultValue int) int {
	if value, exists := p.config[key]; exists {
		if intValue, ok := value.(int); ok {
			return intValue
		}
	}
	return defaultValue
}
