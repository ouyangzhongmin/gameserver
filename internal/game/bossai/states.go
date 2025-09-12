package bossai

import (
	"time"

	"github.com/ouyangzhongmin/gameserver/pkg/coord"
	"github.com/ouyangzhongmin/gameserver/pkg/logger"
	"github.com/ouyangzhongmin/gameserver/pkg/shape"
)

// BaseBossState 基础Boss状态实现
type BaseBossState struct {
	id          BossStateID
	name        string
	transitions []StateTransition
	modifiers   map[string]interface{}
	enterTime   time.Time
	isActive    bool

	// 行为树相关
	behaviorTree     *BehaviorTree
	behaviorConfigs  []string      // behaviors配置列表
	lastBehaviorTime time.Time     // 上次执行行为树的时间
	behaviorInterval time.Duration // 行为树执行间隔
}

// NewBaseBossState 创建基础状态
func NewBaseBossState(id BossStateID, name string) *BaseBossState {
	return &BaseBossState{
		id:               id,
		name:             name,
		transitions:      make([]StateTransition, 0),
		modifiers:        make(map[string]interface{}),
		isActive:         false,
		behaviorConfigs:  make([]string, 0),
		behaviorInterval: time.Millisecond * 500, // 默认500ms执行一次行为树
	}
}

func (s *BaseBossState) GetID() BossStateID {
	return s.id
}

func (s *BaseBossState) GetName() string {
	return s.name
}

func (s *BaseBossState) OnEnter(ctx *BossContext) error {
	s.enterTime = ctx.CurrentTime
	s.isActive = true
	if s.behaviorTree != nil {
		// 每次进入时需要重置所有的行为树
		s.behaviorTree.Reset()
	}
	logger.Debugf("Boss %d entering state: %s", ctx.Boss.GetID(), s.name)
	return nil
}

func (s *BaseBossState) OnUpdate(ctx *BossContext, deltaTime time.Duration) error {
	// 执行行为树
	if s.behaviorTree != nil && ctx.CurrentTime.Sub(s.lastBehaviorTime) >= s.behaviorInterval {
		s.lastBehaviorTime = ctx.CurrentTime
		s.behaviorTree.Execute(ctx)
		//logger.Debugf("State %s executed behavior tree, result: %v", s.name,  result)
	}

	if ctx.Boss.IsDied() {
		// 死亡时强制进入
		ctx.RequestStateTransition(StateDying, "boss_died", 1000)
	}

	return nil
}

func (s *BaseBossState) OnExit(ctx *BossContext) error {
	s.isActive = false
	logger.Debugf("Boss %d exiting state: %s (duration: %v)",
		ctx.Boss.GetID(), s.name, ctx.CurrentTime.Sub(s.enterTime))
	return nil
}

func (s *BaseBossState) CanTransitionTo(stateID BossStateID, ctx *BossContext) bool {
	for _, transition := range s.transitions {
		if transition.ToState == stateID {
			if ok := evaluateCondition(transition.Condition, ctx); ok {
				return true
			}
		}
	}
	return false
}

func (s *BaseBossState) GetNextState(ctx *BossContext) BossStateID {
	// 按优先级检查可能的状态转换
	var bestTransition *StateTransition
	var bestStateID BossStateID = ""
	maxPriority := -1

	for _, transition := range s.transitions {
		if transition.Priority > maxPriority && evaluateCondition(transition.Condition, ctx) {
			maxPriority = transition.Priority
			bestTransition = &transition
			bestStateID = transition.ToState
		}
	}

	if bestTransition != nil {
		return BossStateID(bestStateID)
	}

	return BossStateID(s.id) // 保持当前状态
}

func (s *BaseBossState) AddTransition(toState BossStateID, transition StateTransition) {
	s.transitions = append(s.transitions, transition)
}

func (s *BaseBossState) SetModifier(key string, value interface{}) {
	s.modifiers[key] = value

	// 特殊处理行为树配置
	if key == "behaviors" {
		if behaviors, ok := value.([]string); ok {
			s.behaviorConfigs = behaviors
		} else if behaviorInterface, ok := value.([]interface{}); ok {
			// 处理从JSON解析来的interface{}切片
			behaviors := make([]string, len(behaviorInterface))
			for i, behavior := range behaviorInterface {
				if behaviorStr, ok := behavior.(string); ok {
					behaviors[i] = behaviorStr
				}
			}
			s.behaviorConfigs = behaviors
		}
	}

	// 特殊处理行为执行间隔
	if key == "behavior_interval" {
		if interval, ok := value.(time.Duration); ok {
			s.behaviorInterval = interval
		} else if intervalStr, ok := value.(string); ok {
			if parsed, err := time.ParseDuration(intervalStr); err == nil {
				s.behaviorInterval = parsed
			}
		}
	}
}

func (s *BaseBossState) GetModifier(key string) interface{} {
	return s.modifiers[key]
}

func (s *BaseBossState) IsActive() bool {
	return s.isActive
}

func (s *BaseBossState) GetTimeInState(currentTime time.Time) time.Duration {
	if !s.isActive {
		return 0
	}
	return currentTime.Sub(s.enterTime)
}

// SetBehaviorTree 设置行为树
func (s *BaseBossState) SetBehaviorTree(tree *BehaviorTree) {
	s.behaviorTree = tree
}

// GetBehaviorTree 获取行为树
func (s *BaseBossState) GetBehaviorTree() *BehaviorTree {
	return s.behaviorTree
}

// GetBehaviorConfigs 获取行为配置列表
func (s *BaseBossState) GetBehaviorConfigs() []string {
	return s.behaviorConfigs
}

// SetBehaviorConfigs 设置行为配置列表
func (s *BaseBossState) SetBehaviorConfigs(configs []string) {
	s.behaviorConfigs = configs
}

// ExecuteBehaviorTree 手动执行行为树
func (s *BaseBossState) ExecuteBehaviorTree(ctx *BossContext) BehaviorResult {
	if s.behaviorTree == nil {
		return ResultFailure
	}
	return s.behaviorTree.Execute(ctx)
}

// IdleState 空闲状态 - 现在依赖行为树执行具体行为
type IdleState struct {
	*BaseBossState
}

func NewIdleState() *IdleState {
	base := NewBaseBossState(StateIdle, "Idle")
	return &IdleState{
		BaseBossState: base,
	}
}

func (s *IdleState) OnEnter(ctx *BossContext) error {
	if err := s.BaseBossState.OnEnter(ctx); err != nil {
		return err
	}

	// 控制Monster真实状态
	ctx.Boss.Idle()

	// 清除战斗目标
	ctx.Boss.SetCombatTarget(nil)
	ctx.Target = nil

	return nil
}

func (s *IdleState) OnUpdate(ctx *BossContext, deltaTime time.Duration) error {
	// 调用基类的OnUpdate，它将执行行为树
	return s.BaseBossState.OnUpdate(ctx, deltaTime)
}

// ChaseState 追击状态 - 现在依赖行为树执行具体行为
type ChaseState struct {
	*BaseBossState
	maxChaseDistance float64
	// 位置变化检测和执行间隔限制
	lastTargetPos coord.Vector3 // 记录目标的上次位置
	lastMoveTime  time.Time     // 上次执行MoveTo的时间
	moveInterval  time.Duration // MoveTo执行间隔
}

func NewChaseState() *ChaseState {
	base := NewBaseBossState(StateChase, "Chase")
	return &ChaseState{
		BaseBossState:    base,
		maxChaseDistance: 50.0,
	}
}

func (s *ChaseState) OnEnter(ctx *BossContext) error {
	if err := s.BaseBossState.OnEnter(ctx); err != nil {
		return err
	}
	// 这里需要记录开始追击的原始位置，怪物在返回时需要回到这个位置
	ctx.ChaseStartPos = ctx.Boss.GetPos()
	return nil
}

func (s *ChaseState) OnUpdate(ctx *BossContext, deltaTime time.Duration) error {
	if ctx.Target == nil || !ctx.Target.IsAlive() {
		return nil
	}
	s.moveInterval = time.Duration(ctx.Boss.GetStepTime()) * time.Millisecond
	// 检查距离
	bossPos := ctx.Boss.GetPos()
	targetPos := ctx.Target.GetPos()
	distance := shape.CalculateDistance(
		float64(bossPos.X), float64(bossPos.Y),
		float64(targetPos.X), float64(targetPos.Y),
	)

	if distance > s.maxChaseDistance {
		// 超出追击范围，放弃目标
		ctx.Target = nil
		ctx.Boss.SetCombatTarget(nil)
		return nil
	}

	// 到达攻击范围
	if ctx.Boss.IsInAttackRange(targetPos.X, targetPos.Y) {
		return nil
	}

	// 检查是否需要执行MoveTo（优化性能）
	needMove := false

	// 1. 检查时间间隔限制
	if ctx.CurrentTime.Sub(s.lastMoveTime) >= s.moveInterval {
		// 2. 检查目标位置是否发生显著变化
		targetMoveDistance := shape.CalculateDistance(
			float64(s.lastTargetPos.X), float64(s.lastTargetPos.Y),
			float64(targetPos.X), float64(targetPos.Y),
		)

		// 目标位置变化超过阈值，或者是第一次执行
		if targetMoveDistance >= 2.0 || s.lastTargetPos.X == 0 {
			needMove = true
		}
	}

	if needMove {
		// 执行移动
		err := ctx.Boss.ChaseTo(targetPos.X, targetPos.Y, targetPos.Z)
		if err != nil {
			logger.Debugf("ChaseTarget: MoveTo failed: %v", err)
			return err
		}

		// 更新记录
		s.lastMoveTime = ctx.CurrentTime
		s.lastTargetPos = targetPos

		logger.Debugf("ChaseTarget: Moving to target (%.2f, %.2f), distance: %.2f",
			float64(targetPos.X), float64(targetPos.Y), distance)
	} else {
		// 不需要移动，节省性能
		// logger.Debugf("ChaseTarget: Skipping MoveTo - target position unchanged or interval not reached")
	}

	return s.BaseBossState.OnUpdate(ctx, deltaTime)
}

// AttackState 攻击状态 - 现在依赖行为树执行具体行为
type AttackState struct {
	*BaseBossState
}

func NewAttackState() *AttackState {
	base := NewBaseBossState(StateAttack, "Attack")
	return &AttackState{
		BaseBossState: base,
	}
}

func (s *AttackState) OnEnter(ctx *BossContext) error {
	if err := s.BaseBossState.OnEnter(ctx); err != nil {
		return err
	}

	// 停止移动
	ctx.Boss.Stop()
	ctx.Boss.SetInCombat(true)

	return nil
}

func (s *AttackState) OnExit(ctx *BossContext) error {
	if err := s.BaseBossState.OnExit(ctx); err != nil {
		return err
	}
	ctx.Boss.SetInCombat(false)
	return nil
}

func (s *AttackState) OnUpdate(ctx *BossContext, deltaTime time.Duration) error {
	// 调用基类的OnUpdate，它将执行行为树
	return s.BaseBossState.OnUpdate(ctx, deltaTime)
}

// RetreatState 返回/撤退状态
type RetreatState struct {
	*BaseBossState
	lastMoveTime time.Time     // 上次执行MoveTo的时间
	moveInterval time.Duration // MoveTo执行间隔
}

func NewRetreatState() *RetreatState {
	base := NewBaseBossState(StateRetreat, "Retreat")
	return &RetreatState{
		BaseBossState: base,
	}
}

func (s *RetreatState) OnEnter(ctx *BossContext) error {
	if err := s.BaseBossState.OnEnter(ctx); err != nil {
		return err
	}

	bornPos := ctx.Boss.GetBornPos()
	// 返回出生点
	ctx.Boss.EscapeTo(bornPos.X, bornPos.Y, bornPos.Z)
	// 清除战斗目标
	ctx.Target = nil
	ctx.Boss.SetCombatTarget(nil)

	logger.Debugf("Boss %d retreating to born point (%f, %f)",
		ctx.Boss.GetID(), bornPos.X, bornPos.Y)

	return nil
}

func (s *RetreatState) OnUpdate(ctx *BossContext, deltaTime time.Duration) error {
	// 获取出生点参数
	bornPos := ctx.Boss.GetBornPos()
	if !ctx.Boss.IsEscaping() && !ctx.Boss.IsWalking() && !ctx.Boss.IsRunning() {
		ctx.Boss.EscapeTo(bornPos.X, bornPos.Y, bornPos.Z)
	}

	// 获取移动间隔参数
	moveInterval := time.Duration(ctx.Boss.GetStepTime()) * time.Millisecond

	if ctx.CurrentTime.Sub(s.lastMoveTime) >= moveInterval {
		bossPos := ctx.Boss.GetPos()
		distance := shape.CalculateDistance(
			float64(bossPos.X), float64(bossPos.Y),
			float64(bornPos.X), float64(bornPos.Y),
		)

		if distance <= 1 {
			// 到达目的地,请求转换到idle状态
			reason := "ReturnToBorn"
			ctx.RequestStateTransition(StateIdle, reason, 100) // 高优先级
			logger.Debugf("ReturnToBorn: Moved to born point (%.2f, %.2f), distance: %.2f",
				bornPos.X, bornPos.Y, distance)
		}
		s.lastMoveTime = ctx.CurrentTime
	}
	return s.BaseBossState.OnUpdate(ctx, deltaTime)
}

func (s *RetreatState) OnExit(ctx *BossContext) error {
	if err := s.BaseBossState.OnExit(ctx); err != nil {
		return err
	}

	return nil
}

// DyingState 死亡状态
type DyingState struct {
	*BaseBossState
	deathDuration time.Duration
}

func NewDyingState() *DyingState {
	base := NewBaseBossState(StateDying, "Dying")
	return &DyingState{
		BaseBossState: base,
		deathDuration: time.Second * 5, // 5秒死亡动画
	}
}

func (s *DyingState) OnEnter(ctx *BossContext) error {
	if err := s.BaseBossState.OnEnter(ctx); err != nil {
		return err
	}

	// 控制Monster真实状态
	ctx.Boss.Die()

	logger.Debugf("Boss %d is dying", ctx.Boss.GetID())

	return nil
}

func (s *DyingState) OnUpdate(ctx *BossContext, deltaTime time.Duration) error {
	// 调用基类的OnUpdate，它将执行行为树（如触发奖励）
	return s.BaseBossState.OnUpdate(ctx, deltaTime)
}

func (s *DyingState) CanTransitionTo(stateID BossStateID, ctx *BossContext) bool {
	// 死亡状态不能转换到其他状态
	return false
}

func (s *DyingState) GetNextState(ctx *BossContext) BossStateID {
	// 保持死亡状态
	return StateDying
}
