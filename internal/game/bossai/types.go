package bossai

import (
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
type BossStateID int32

const (
	StateIdle BossStateID = iota
	StatePatrol
	StateChase
	StateAttack
	StateCastSkill
	StateStunned
	StateEnraged
	StateRetreat
	StateDying
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

	// 环境信息
	NearbyEnemies []IEntity
	NearbyAllies  []IEntity
	Position      coord.Vector3

	// AI决策缓存
	LastAction    *AIAction
	ActionHistory []*AIAction

	// 自定义数据
	CustomData map[string]interface{}
}

// AIAction AI决策动作
type AIAction struct {
	Type      ActionType             `json:"type"`
	SkillID   int32                  `json:"skill_id,omitempty"`
	TargetID  int64                  `json:"target_id,omitempty"`
	Position  *coord.Vector3         `json:"position,omitempty"`
	Duration  time.Duration          `json:"duration,omitempty"`
	Params    map[string]interface{} `json:"params,omitempty"`
	Priority  int                    `json:"priority"`
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
	Name        string                 `json:"name"`
	Transitions []StateTransition      `json:"transitions"`
	Behaviors   []string               `json:"behaviors"`
	Modifiers   map[string]interface{} `json:"modifiers,omitempty"`
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
