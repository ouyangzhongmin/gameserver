package bossai

import (
	"context"
	"time"

	"github.com/ouyangzhongmin/gameserver/pkg/coord"
)

// 注意：Boss AI系统复用原有游戏系统的核心接口以避免重复定义
// 这些接口在 internal/game/igame.go 中定义：
// - IEntity: 游戏实体基础接口
// - IMovableEntity: 可移动实体接口
//
// Boss AI系统通过适配器模式与原有系统集成，确保类型兼容性

// IEntityForAI Boss AI系统使用的实体接口，扩展了基础功能
type IEntityForAI interface {
	// 基础实体功能（来自原有IEntity接口）
	GetID() int64
	GetPos() coord.Vector3
	GetEntityType() int
	IsDestroyed() bool

	// AI专用扩展功能
	IsAlive() bool // 生命状态检查
}

// IMovableEntityForAI Boss AI系统使用的可移动实体接口
type IMovableEntityForAI interface {
	IEntityForAI

	// 移动功能
	MoveTo(x, y, z coord.Coord) error
	Stop() error

	// 战斗功能
	CanAttackTarget(target IEntityForAI) bool
	IsInAttackRange(x, y coord.Coord) bool
}

// 为了保持向后兼容，提供类型别名
type IEntity = IEntityForAI
type IMovableEntity = IMovableEntityForAI

// IBossEntity Boss实体接口
type IBossEntity interface {
	IMovableEntity

	// 基础属性
	GetMaxLife() int32
	GetCurrentLife() int32
	GetAttackPower() int32

	// 技能相关
	CanUseSkill(skillID int32) bool
	UseSkill(skillID int32, target IEntity) error
	GetSkillCooldown(skillID int32) time.Duration

	// 状态相关
	IsInCombat() bool
	GetCombatTarget() IEntity
	SetCombatTarget(target IEntity)

	// 视野相关
	GetEntitiesInRange(radius float64) []IEntity
	GetNearestEnemy() IEntity
}

// IBossState Boss状态接口
type IBossState interface {
	GetName() string
	GetID() int32

	// 状态生命周期
	OnEnter(ctx *BossContext) error
	OnUpdate(ctx *BossContext, deltaTime time.Duration) error
	OnExit(ctx *BossContext) error

	// 状态转换条件
	CanTransitionTo(stateID int32, ctx *BossContext) bool
	GetNextState(ctx *BossContext) int32
}

// IBehaviorNode 行为树节点接口
type IBehaviorNode interface {
	GetName() string
	GetType() BehaviorNodeType

	// 节点执行
	Execute(ctx *BossContext) BehaviorResult
	Reset()

	// 子节点管理 (仅组合节点需要)
	AddChild(child IBehaviorNode) error
	GetChildren() []IBehaviorNode
}

// IBossSkill Boss技能接口
type IBossSkill interface {
	GetID() int32
	GetName() string
	GetCooldown() time.Duration
	GetRange() float64
	GetCastTime() time.Duration

	// 技能条件检查
	CanUse(ctx *BossContext) bool
	GetTargets(ctx *BossContext) []IEntity

	// 技能执行
	Cast(ctx *BossContext, targets []IEntity) error

	// 冷却管理
	StartCooldown()
	IsOnCooldown() bool
	GetRemainingCooldown() time.Duration
}

// IBossPhase Boss阶段接口
type IBossPhase interface {
	GetID() int32
	GetName() string
	GetTriggerCondition() PhaseCondition

	// 阶段生命周期
	OnEnter(ctx *BossContext) error
	OnExit(ctx *BossContext) error

	// 阶段行为
	GetBehaviorTree() IBehaviorNode
	GetBehaviorTreeName() string
	GetAvailableSkills() []int32
	GetStateModifiers() map[string]interface{}
}

// IAIPlugin AI插件接口
type IAIPlugin interface {
	GetName() string
	GetVersion() string
	GetPriority() int
	IsEnabled() bool
	SetConfig(config map[string]interface{})
	SetPriority(priority int)

	// 插件生命周期
	Initialize(ctx *BossContext) error
	Update(ctx *BossContext, deltaTime time.Duration) error
	Cleanup() error

	// 决策支持
	SuggestAction(ctx *BossContext) *AIAction
	ModifyBehavior(ctx *BossContext, originalAction *AIAction) *AIAction
}

// ILLMProvider 大语言模型接口
type ILLMProvider interface {
	GetName() string

	// 决策生成
	GenerateStrategy(ctx context.Context, bossState *BossSnapshot, gameState *GameSnapshot) (*AIStrategy, error)
	AdaptBehavior(ctx context.Context, feedback *ActionFeedback) (*BehaviorAdjustment, error)

	// 配置管理
	Configure(config map[string]interface{}) error
	IsAvailable() bool
}

// IBossAI Boss AI主接口
type IBossAI interface {
	// 基础控制
	Initialize(boss IBossEntity) error
	Update(deltaTime time.Duration) error
	Destroy() error

	// 状态管理
	GetCurrentState() IBossState
	TransitionTo(stateID int32) error

	// 阶段管理
	GetCurrentPhase() IBossPhase
	CheckPhaseTransition() error

	// 行为控制
	ExecuteBehavior() error
	InterruptBehavior() error

	// 插件管理
	AddPlugin(plugin IAIPlugin) error
	RemovePlugin(pluginName string) error

	// 事件处理
	OnDamageReceived(attacker IEntity, damage int32) error
	OnSkillUsed(skillID int32, success bool) error
	OnTargetChanged(newTarget IEntity) error

	// 配置和调试
	LoadConfig(config *BossConfig) error
	GetDebugInfo() *AIDebugInfo
}
