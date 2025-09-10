package bossai

import (
	"math"
	"time"

	"github.com/ouyangzhongmin/gameserver/pkg/coord"
)

// BehaviorNodeType 行为树节点类型
type BehaviorNodeType int

const (
	NodeTypeAction    BehaviorNodeType = iota // 行为节点
	NodeTypeCondition                         // 条件节点
	NodeTypeSequence                          // 顺序节点
	NodeTypeSelector                          // 选择节点
	NodeTypeParallel                          // 并行节点
	NodeTypeDecorator                         // 装饰节点
)

// BehaviorResult 行为树执行结果
type BehaviorResult int

const (
	ResultSuccess BehaviorResult = iota
	ResultFailure
	ResultRunning
)

// BossStateID Boss状态ID
type BossStateID string

const (
	StateIdle    BossStateID = "idle"
	StateChase   BossStateID = "chase"
	StateAttack  BossStateID = "attack"
	StateRetreat BossStateID = "retreat"
	StateDying   BossStateID = "dying"
)

// PhaseCondition 阶段触发条件
type PhaseCondition struct {
	Type     ConditionType          `json:"type"`
	Params   map[string]interface{} `json:"params"`
	Operator LogicalOperator        `json:"operator,omitempty"`
	SubConds []PhaseCondition       `json:"sub_conditions,omitempty"`
}

type ConditionType int

const (
	CondHealthPercent ConditionType = iota
	CondTimeElapsed
	CondSkillUsed
	CondTargetCount
	CondCustomScript
)

type LogicalOperator int

const (
	OpAnd LogicalOperator = iota
	OpOr
	OpNot
)

// BossContext Boss AI上下文
type BossContext struct {
	// 核心实体
	Boss   IBossEntity
	Target IEntity

	// 时间信息
	CurrentTime time.Time
	DeltaTime   time.Duration
	CombatTime  time.Duration

	// 状态信息
	CurrentState   IBossState
	PreviousState  IBossState
	CurrentPhase   IBossPhase
	PreviousPhase  IBossPhase
	IsStateChanged bool
	IsPhaseChanged bool

	// 战斗信息
	DamageDealt    int32
	DamageReceived int32
	SkillsUsed     map[int32]int // skillID -> usage count

	// 环境信息（由BossAIManager在updateContext中更新）
	NearbyEnemies  []IEntity     // 附近敌人
	NearbyAllies   []IEntity     // 附近盟友
	OriginPosition coord.Vector3 // 追击前的原始位置

	// AI决策缓存
	LastAction    *AIAction
	ActionHistory []*AIAction

	// 自定义数据
	CustomData map[string]interface{}

	// 状态转换请求
	stateTransitionRequest *StateTransitionRequest
}

// StateTransitionRequest 状态转换请求
type StateTransitionRequest struct {
	TargetState BossStateID
	Reason      string
	Priority    int
}

// RequestStateTransition 请求状态转换
func (ctx *BossContext) RequestStateTransition(stateID BossStateID, reason string, priority int) {
	ctx.stateTransitionRequest = &StateTransitionRequest{
		TargetState: stateID,
		Reason:      reason,
		Priority:    priority,
	}
}

// GetStateTransitionRequest 获取状态转换请求
func (ctx *BossContext) GetStateTransitionRequest() *StateTransitionRequest {
	return ctx.stateTransitionRequest
}

// ClearStateTransitionRequest 清除状态转换请求
func (ctx *BossContext) ClearStateTransitionRequest() {
	ctx.stateTransitionRequest = nil
}

// GetEnemiesInRange 从已缓存的NearbyEnemies中筛选指定范围内的敌人
// 避免重复调用AOI查询，提高性能
func (ctx *BossContext) GetEnemiesInRange(radius float64) []IEntity {
	if len(ctx.NearbyEnemies) == 0 {
		return []IEntity{}
	}

	bossPos := ctx.Boss.GetPos()
	filteredEnemies := make([]IEntity, 0, len(ctx.NearbyEnemies))

	for _, enemy := range ctx.NearbyEnemies {
		enemyPos := enemy.GetPos()
		// 计算距离
		dx := float64(bossPos.X - enemyPos.X)
		dy := float64(bossPos.Y - enemyPos.Y)
		distance := math.Sqrt(dx*dx + dy*dy)

		if distance <= radius {
			filteredEnemies = append(filteredEnemies, enemy)
		}
	}

	return filteredEnemies
}

// GetNearestEnemyInRange 从已缓存的NearbyEnemies中获取指定范围内最近的敌人
func (ctx *BossContext) GetNearestEnemyInRange(maxRadius float64) IEntity {
	enemiesInRange := ctx.GetEnemiesInRange(maxRadius)
	if len(enemiesInRange) == 0 {
		return nil
	}

	bossPos := ctx.Boss.GetPos()
	var nearestEnemy IEntity
	var minDistance float64 = math.Inf(1)

	for _, enemy := range enemiesInRange {
		enemyPos := enemy.GetPos()
		dx := float64(bossPos.X - enemyPos.X)
		dy := float64(bossPos.Y - enemyPos.Y)
		distance := math.Sqrt(dx*dx + dy*dy)

		if distance < minDistance {
			minDistance = distance
			nearestEnemy = enemy
		}
	}

	return nearestEnemy
}

// HasEnemiesInRange 检查指定范围内是否有敌人
func (ctx *BossContext) HasEnemiesInRange(radius float64) bool {
	return len(ctx.GetEnemiesInRange(radius)) > 0
}

// AIAction AI决策动作
type AIAction struct {
	Type      ActionType             `json:"type"`
	SkillID   int32                  `json:"skill_id,omitempty"`
	TargetID  int64                  `json:"target_id,omitempty"`
	Target    IEntity                `json:"-"` // 目标实体引用（不序列化）
	Position  *coord.Vector3         `json:"position,omitempty"`
	Duration  time.Duration          `json:"duration,omitempty"`
	Damage    int32                  `json:"damage,omitempty"` // 伤害值
	Params    map[string]interface{} `json:"params,omitempty"`
	Priority  int                    `json:"priority"`
	Executed  bool                   `json:"executed"` // 是否已执行
	Timestamp time.Time              `json:"timestamp"`
}

type ActionType int

const (
	ActionIdle ActionType = iota
	ActionMove
	ActionAttack
	ActionCastSkill
	ActionRetreat
	ActionSummon
	ActionBuff
	ActionDebuff
)

// BossSnapshot Boss状态快照（用于LLM）
type BossSnapshot struct {
	ID              int64           `json:"id"`
	Name            string          `json:"name"`
	HealthPercent   float64         `json:"health_percent"`
	Position        coord.Vector3   `json:"position"`
	CurrentState    string          `json:"current_state"`
	CurrentPhase    string          `json:"current_phase"`
	CombatTime      float64         `json:"combat_time"`
	AvailableSkills []SkillSnapshot `json:"available_skills"`
	StatusEffects   []StatusEffect  `json:"status_effects"`
}

// GameSnapshot 游戏环境快照（用于LLM）
type GameSnapshot struct {
	Timestamp     time.Time        `json:"timestamp"`
	NearbyEnemies []EntitySnapshot `json:"nearby_enemies"`
	NearbyAllies  []EntitySnapshot `json:"nearby_allies"`
	TerrainInfo   *TerrainInfo     `json:"terrain_info,omitempty"`
	WeatherInfo   *WeatherInfo     `json:"weather_info,omitempty"`
}

type EntitySnapshot struct {
	ID            int64         `json:"id"`
	Type          string        `json:"type"`
	Position      coord.Vector3 `json:"position"`
	HealthPercent float64       `json:"health_percent"`
	Distance      float64       `json:"distance"`
	ThreatLevel   int           `json:"threat_level"`
}

type SkillSnapshot struct {
	ID             int32   `json:"id"`
	Name           string  `json:"name"`
	CooldownRemain float64 `json:"cooldown_remain"`
	Range          float64 `json:"range"`
	IsAvailable    bool    `json:"is_available"`
}

type StatusEffect struct {
	Type      string  `json:"type"`
	Duration  float64 `json:"duration"`
	Intensity float64 `json:"intensity"`
}

type TerrainInfo struct {
	Type      string          `json:"type"`
	Obstacles []coord.Vector3 `json:"obstacles"`
	SafeZones []coord.Vector3 `json:"safe_zones"`
}

type WeatherInfo struct {
	Type       string  `json:"type"`
	Intensity  float64 `json:"intensity"`
	Visibility float64 `json:"visibility"`
}

// AIStrategy LLM生成的AI策略
type AIStrategy struct {
	RecommendedAction *AIAction              `json:"recommended_action"`
	Reasoning         string                 `json:"reasoning"`
	Confidence        float64                `json:"confidence"`
	Alternatives      []*AIAction            `json:"alternatives,omitempty"`
	Adaptations       map[string]interface{} `json:"adaptations,omitempty"`
}

// ActionFeedback 行为反馈
type ActionFeedback struct {
	Action         *AIAction     `json:"action"`
	Success        bool          `json:"success"`
	DamageDealt    int32         `json:"damage_dealt"`
	DamageReceived int32         `json:"damage_received"`
	Duration       time.Duration `json:"duration"`
	TargetReaction string        `json:"target_reaction"`
	Effectiveness  float64       `json:"effectiveness"`
	Timestamp      time.Time     `json:"timestamp"`
}

// BehaviorAdjustment 行为调整
type BehaviorAdjustment struct {
	SkillPriorities map[int32]int      `json:"skill_priorities"`
	StateModifiers  map[string]float64 `json:"state_modifiers"`
	BehaviorTree    string             `json:"behavior_tree,omitempty"`
	CustomLogic     string             `json:"custom_logic,omitempty"`
	Confidence      float64            `json:"confidence"`
}

// BossConfig Boss配置
type BossConfig struct {
	ID          int32  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`

	// 状态配置
	States       []StateConfig `json:"states"`
	InitialState BossStateID   `json:"initial_state"`

	// 阶段配置
	Phases []PhaseConfig `json:"phases"`

	// 技能配置
	Skills []SkillConfig `json:"skills"`

	// 行为树配置
	BehaviorTree BehaviorTreeConfig `json:"behavior_tree"`

	// AI插件配置
	Plugins []PluginConfig `json:"plugins"`

	// 高级配置
	LLMConfig         *LLMConfig         `json:"llm_config,omitempty"`
	PerformanceConfig *PerformanceConfig `json:"performance_config,omitempty"`
}

type StateConfig struct {
	ID          BossStateID            `json:"id"`
	Transitions []StateTransition      `json:"transitions"`
	Behaviors   interface{}            `json:"behaviors"` // 支持字符串数组或复杂行为树配置
	Modifiers   map[string]interface{} `json:"modifiers,omitempty"`
}

// BehaviorConfig 行为配置结构
type BehaviorConfig struct {
	Type     string                 `json:"type"`               // random, sequence, selector, repeat, node
	Node     string                 `json:"node,omitempty"`     // 对于node类型，指定具体的行为节点名称
	Params   map[string]interface{} `json:"params,omitempty"`   // 自定义参数
	Rand     int                    `json:"rand,omitempty"`     // 随机概率权重（仅用于random类型）
	Count    int                    `json:"count,omitempty"`    // 重复次数（仅用于repeat类型）
	Children []BehaviorConfig       `json:"children,omitempty"` // 子节点
}

type StateTransition struct {
	ToState   BossStateID    `json:"to_state"`
	Condition PhaseCondition `json:"condition"`
	Priority  int            `json:"priority"`
}

type PhaseConfig struct {
	ID               int32                  `json:"id"`
	Name             string                 `json:"name"`
	TriggerCondition PhaseCondition         `json:"trigger_condition"`
	AvailableSkills  []int32                `json:"available_skills"`
	BehaviorTree     string                 `json:"behavior_tree"`
	Modifiers        map[string]interface{} `json:"modifiers,omitempty"`
}

type SkillConfig struct {
	ID         int32            `json:"id"`
	Name       string           `json:"name"`
	Cooldown   time.Duration    `json:"cooldown"`
	Range      float64          `json:"range"`
	CastTime   time.Duration    `json:"cast_time"`
	Conditions []PhaseCondition `json:"conditions,omitempty"`
	Effects    []SkillEffect    `json:"effects"`
}

type SkillEffect struct {
	Type     string        `json:"type"`
	Value    interface{}   `json:"value"`
	Target   string        `json:"target"`
	Duration time.Duration `json:"duration,omitempty"`
}

type BehaviorTreeConfig struct {
	RootNode       BehaviorNodeConfig `json:"root_node"`
	UpdateInterval time.Duration      `json:"update_interval"`
}

type BehaviorNodeConfig struct {
	Type     BehaviorNodeType       `json:"type"`
	Name     string                 `json:"name"`
	Params   map[string]interface{} `json:"params,omitempty"`
	Children []BehaviorNodeConfig   `json:"children,omitempty"`
}

type PluginConfig struct {
	Name     string                 `json:"name"`
	Type     string                 `json:"type"`
	Enabled  bool                   `json:"enabled"`
	Priority int                    `json:"priority"`
	Config   map[string]interface{} `json:"config"`
}

type LLMConfig struct {
	Provider       string        `json:"provider"`
	Model          string        `json:"model"`
	APIKey         string        `json:"api_key,omitempty"`
	Endpoint       string        `json:"endpoint,omitempty"`
	UpdateInterval time.Duration `json:"update_interval"`
	MaxTokens      int           `json:"max_tokens"`
	Temperature    float64       `json:"temperature"`
	Enabled        bool          `json:"enabled"`
}

type PerformanceConfig struct {
	MaxBehaviorDepth   int     `json:"max_behavior_depth"`
	MaxActionsPerFrame int     `json:"max_actions_per_frame"`
	EnableProfiling    bool    `json:"enable_profiling"`
	MemoryLimit        int64   `json:"memory_limit"`
	CPUThreshold       float64 `json:"cpu_threshold"`
}

// AIDebugInfo AI调试信息
type AIDebugInfo struct {
	CurrentState       string              `json:"current_state"`
	CurrentPhase       string              `json:"current_phase"`
	LastAction         *AIAction           `json:"last_action"`
	BehaviorTreeStatus string              `json:"behavior_tree_status"`
	ActivePlugins      []string            `json:"active_plugins"`
	PerformanceMetrics *PerformanceMetrics `json:"performance_metrics"`
	LLMStatus          *LLMStatus          `json:"llm_status,omitempty"`
}

type PerformanceMetrics struct {
	UpdateTime          time.Duration `json:"update_time"`
	BehaviorExecuteTime time.Duration `json:"behavior_execute_time"`
	MemoryUsage         int64         `json:"memory_usage"`
	CPUUsage            float64       `json:"cpu_usage"`
	FrameRate           float64       `json:"frame_rate"`
}

type LLMStatus struct {
	Provider       string        `json:"provider"`
	IsConnected    bool          `json:"is_connected"`
	LastUpdate     time.Time     `json:"last_update"`
	RequestCount   int64         `json:"request_count"`
	ErrorCount     int64         `json:"error_count"`
	AverageLatency time.Duration `json:"average_latency"`
}

// BossContext 方法实现
// GetPhaseManager 获取阶段管理器
func (ctx *BossContext) GetPhaseManager() IPhaseManager {
	if phaseMgr, ok := ctx.CustomData["phase_manager"].(IPhaseManager); ok {
		return phaseMgr
	}
	return nil
}

// SetPhaseManager 设置阶段管理器
func (ctx *BossContext) SetPhaseManager(phaseMgr IPhaseManager) {
	if ctx.CustomData == nil {
		ctx.CustomData = make(map[string]interface{})
	}
	ctx.CustomData["phase_manager"] = phaseMgr
}

// GetStateMachine 获取状态机
func (ctx *BossContext) GetStateMachine() IStateMachine {
	if stateMachine, ok := ctx.CustomData["state_machine"].(IStateMachine); ok {
		return stateMachine
	}
	return nil
}

// SetStateMachine 设置状态机
func (ctx *BossContext) SetStateMachine(stateMachine IStateMachine) {
	if ctx.CustomData == nil {
		ctx.CustomData = make(map[string]interface{})
	}
	ctx.CustomData["state_machine"] = stateMachine
}

// IsTargetValid 检查目标是否有效
func (ctx *BossContext) IsTargetValid() bool {
	return ctx.Target != nil && ctx.Target.IsAlive()
}

// AddActionToHistory 添加动作到历史记录
func (ctx *BossContext) AddActionToHistory(action *AIAction) {
	if ctx.ActionHistory == nil {
		ctx.ActionHistory = make([]*AIAction, 0)
	}

	action.Timestamp = ctx.CurrentTime
	ctx.ActionHistory = append(ctx.ActionHistory, action)

	// 限制历史记录长度
	if len(ctx.ActionHistory) > 100 {
		ctx.ActionHistory = ctx.ActionHistory[1:]
	}
}

// GetLastActionOfType 获取指定类型的最后一个动作
func (ctx *BossContext) GetLastActionOfType(actionType ActionType) *AIAction {
	for i := len(ctx.ActionHistory) - 1; i >= 0; i-- {
		if ctx.ActionHistory[i].Type == actionType {
			return ctx.ActionHistory[i]
		}
	}
	return nil
}

// GetHealthPercent 获取当前血量百分比
func (ctx *BossContext) GetHealthPercent() float64 {
	if ctx.Boss == nil {
		return 0.0
	}
	return float64(ctx.Boss.GetCurrentLife()) / float64(ctx.Boss.GetMaxLife())
}

// IsInCombat 检查是否在战斗中
func (ctx *BossContext) IsInCombat() bool {
	return ctx.Boss != nil && ctx.Boss.IsInCombat()
}

// GetNearbyEnemyCount 获取附近敌人数量
func (ctx *BossContext) GetNearbyEnemyCount() int {
	return len(ctx.NearbyEnemies)
}

// GetCombatDuration 获取战斗持续时间
func (ctx *BossContext) GetCombatDuration() time.Duration {
	return ctx.CombatTime
}

// GetAvailableSkillsFromCurrentPhase 获取当前阶段的可用技能列表
func (ctx *BossContext) GetAvailableSkillsFromCurrentPhase() []int32 {
	if ctx.CurrentPhase != nil {
		return ctx.CurrentPhase.GetAvailableSkills()
	}
	return nil
}

// IsSkillAllowedInCurrentPhase 检查技能在当前阶段是否被允许使用
func (ctx *BossContext) IsSkillAllowedInCurrentPhase(skillID int32) bool {
	availableSkills := ctx.GetAvailableSkillsFromCurrentPhase()
	if availableSkills == nil || len(availableSkills) == 0 {
		// 如果当前阶段没有限制可用技能，则所有技能都被允许
		return true
	}

	// 检查技能是否在允许列表中
	for _, allowedSkillID := range availableSkills {
		if allowedSkillID == skillID {
			return true
		}
	}
	return false
}
