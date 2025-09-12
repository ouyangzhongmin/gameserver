package bossai

import (
	"context"
	"time"

	"github.com/ouyangzhongmin/gameserver/pkg/shape"

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
	WalkTo(x, y, z coord.Coord) error
	EscapeTo(x, y, z coord.Coord) error
	ChaseTo(x, y, z coord.Coord) error
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
	GetAttackDuration() int    // 攻击间隔
	GetStepTime() int          // 复用Monster的getStepTime方法
	GetBornPos() coord.Vector3 // 获取出生点
	GetMaxChaseDist() int      // 获取最大的追击距离

	// 技能相关
	CanUseSkill(skillID int32) bool
	IsSkillInCD(skillID int32) bool // 获取技能是否还在cd中
	// 获取一个可以释放的技能，这个技能的获取可以按配置的规则
	// 比如:血量低于50%优先获取回血的技能释放，蓝量低于50%优先获取回蓝技能，有霸体技能cd好了则优先霸体技能
	GetAvailableSkill(rules string) int32
	IsInSkillAttackRange(skillID int32, x, y coord.Coord) bool
	UseSkill(skillID int32, target IEntity) error

	// 战斗功能（复用Monster基础功能）
	DoAttackTarget(target IEntity) error                               // 复用Monster的doAttackTarget方法
	GetCanAttackPos(target IEntity, offset int) (coord.Vector3, error) // 复用Monster的GetCanAttackPos方法
	GetRandomPos(radius int) (coord.Vector3, error)
	GetMovableRect() shape.Rect

	// 状态相关
	IsInCombat() bool
	SetInCombat(val bool)
	GetCombatTarget() IEntity
	SetCombatTarget(target IEntity)

	// 视野相关
	GetEntitesInRange(radius float64) []IEntity
	IsEnemy(entity IEntity) bool
	// 是否盟友
	IsAlly(entity IEntity) bool

	// Monster状态控制方法（与原有系统的ActionState同步）
	Idle()         // 空闲状态
	Walk()         // 行走状态
	Run()          // 跑步状态
	Chase()        // 追击状态
	Escape()       // 逃跑状态
	AttackAction() // 攻击状态
	Die()          // 死亡状态

	IsIdle() bool
	IsWalking() bool
	IsRunning() bool
	IsChasing() bool
	IsEscaping() bool
	IsAttacking() bool
	IsDied() bool

	OnPhaseEnter(phase IBossPhase) // 当进入新阶段时调用
}

// IBossState Boss状态接口
type IBossState interface {
	GetID() BossStateID
	GetName() string
	// 状态生命周期
	OnEnter(ctx *BossContext) error
	OnUpdate(ctx *BossContext, deltaTime time.Duration) error
	OnExit(ctx *BossContext) error

	IsActive() bool

	// 状态转换条件
	CanTransitionTo(stateID BossStateID, ctx *BossContext) bool
	GetNextState(ctx *BossContext) BossStateID

	SetModifier(key string, value interface{})
	GetModifier(key string) interface{}

	AddTransition(toState BossStateID, transition StateTransition)
	SetBehaviorTree(tree *BehaviorTree)
	GetBehaviorTree() *BehaviorTree

	GetTimeInState(currentTime time.Time) time.Duration
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
	setParent(parent IBehaviorNode)
	GetParent() IBehaviorNode

	SetParams(val map[string]interface{})
}

// IBossPhase Boss阶段接口
type IBossPhase interface {
	GetID() int32
	GetName() string
	GetTriggerCondition() TriggerCondition

	// 阶段生命周期
	OnEnter(ctx *BossContext) error
	OnExit(ctx *BossContext) error

	// 阶段行为
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

// IPhaseManager 阶段管理器接口
type IPhaseManager interface {
	// 阶段管理
	AddPhase(phase IBossPhase) error
	RemovePhase(phaseID int32) error
	GetPhase(phaseID int32) (IBossPhase, error)
	GetCurrentPhase() IBossPhase

	// 阶段转换
	Update(ctx *BossContext, deltaTime time.Duration) error
	SetPhaseOrder(order []int32) error
	ForceTransition(phaseID int32, ctx *BossContext) error

	// 生命周期
	Destroy()
}

// IStateMachine 状态机接口
type IStateMachine interface {
	// 状态管理
	AddState(state IBossState) error
	RemoveState(stateID int32) error
	GetState(stateID int32) (IBossState, error)
	GetCurrentState() IBossState

	// 状态转换
	Update(ctx *BossContext, deltaTime time.Duration) error
	SetInitialState(stateID int32, ctx *BossContext) error
	ForceTransition(stateID BossStateID, ctx *BossContext) error

	// 生命周期
	Destroy()
}
