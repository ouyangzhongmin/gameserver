package bossai

import (
	"time"

	"github.com/ouyangzhongmin/gameserver/pkg/coord"
	"github.com/ouyangzhongmin/gameserver/pkg/logger"
)

// BaseBossState 基础Boss状态实现
type BaseBossState struct {
	id          BossStateID
	name        string
	transitions map[BossStateID]StateTransition
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
		transitions:      make(map[BossStateID]StateTransition),
		modifiers:        make(map[string]interface{}),
		isActive:         false,
		behaviorConfigs:  make([]string, 0),
		behaviorInterval: time.Millisecond * 500, // 默认500ms执行一次行为树
	}
}

func (s *BaseBossState) GetName() string {
	return s.name
}

func (s *BaseBossState) GetID() int32 {
	return int32(s.id)
}

func (s *BaseBossState) OnEnter(ctx *BossContext) error {
	s.enterTime = ctx.CurrentTime
	s.isActive = true
	logger.Debugf("Boss %d entering state: %s", ctx.Boss.GetID(), s.name)
	return nil
}

func (s *BaseBossState) OnUpdate(ctx *BossContext, deltaTime time.Duration) error {
	// 执行行为树
	if s.behaviorTree != nil && ctx.CurrentTime.Sub(s.lastBehaviorTime) >= s.behaviorInterval {
		s.lastBehaviorTime = ctx.CurrentTime
		result := s.behaviorTree.Execute(ctx)

		logger.Debugf("State %s executed behavior tree, result: %v", s.name, result)
	}

	return nil
}

func (s *BaseBossState) OnExit(ctx *BossContext) error {
	s.isActive = false
	logger.Debugf("Boss %d exiting state: %s (duration: %v)",
		ctx.Boss.GetID(), s.name, ctx.CurrentTime.Sub(s.enterTime))
	return nil
}

func (s *BaseBossState) CanTransitionTo(stateID int32, ctx *BossContext) bool {
	transition, exists := s.transitions[BossStateID(stateID)]
	if !exists {
		return false
	}

	return s.evaluateCondition(transition.Condition, ctx)
}

func (s *BaseBossState) GetNextState(ctx *BossContext) int32 {
	// 按优先级检查可能的状态转换
	var bestTransition *StateTransition
	var bestStateID BossStateID = -1
	maxPriority := -1

	for stateID, transition := range s.transitions {
		if transition.Priority > maxPriority && s.evaluateCondition(transition.Condition, ctx) {
			maxPriority = transition.Priority
			bestTransition = &transition
			bestStateID = stateID
		}
	}

	if bestTransition != nil {
		return int32(bestStateID)
	}

	return int32(s.id) // 保持当前状态
}

func (s *BaseBossState) AddTransition(toState BossStateID, transition StateTransition) {
	s.transitions[toState] = transition
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

// evaluateCondition 评估条件
func (s *BaseBossState) evaluateCondition(condition PhaseCondition, ctx *BossContext) bool {
	switch condition.Type {
	case CondHealthPercent:
		threshold, ok := condition.Params["threshold"].(float64)
		if !ok {
			return false
		}
		currentPercent := float64(ctx.Boss.GetCurrentLife()) / float64(ctx.Boss.GetMaxLife())
		operator, ok := condition.Params["operator"].(string)
		if !ok {
			operator = "less_than"
		}

		switch operator {
		case "less_than":
			return currentPercent < threshold
		case "greater_than":
			return currentPercent > threshold
		case "equal":
			return currentPercent == threshold
		default:
			return false
		}

	case CondTimeElapsed:
		duration, ok := condition.Params["duration"].(time.Duration)
		if !ok {
			return false
		}
		return s.GetTimeInState(ctx.CurrentTime) >= duration

	case CondSkillUsed:
		skillID, ok := condition.Params["skill_id"].(int32)
		if !ok {
			return false
		}
		count, exists := ctx.SkillsUsed[skillID]
		if !exists {
			return false
		}
		minCount, ok := condition.Params["min_count"].(int)
		if !ok {
			minCount = 1
		}
		return count >= minCount

	case CondTargetCount:
		minCount, ok := condition.Params["min_count"].(int)
		if !ok {
			return false
		}
		return len(ctx.NearbyEnemies) >= minCount

	case CondCustomScript:
		// 这里可以扩展自定义脚本逻辑
		scriptName, ok := condition.Params["script"].(string)
		if !ok {
			return false
		}
		return s.evaluateCustomScript(scriptName, ctx)
	}

	// 处理复合条件
	if len(condition.SubConds) > 0 {
		results := make([]bool, len(condition.SubConds))
		for i, subCond := range condition.SubConds {
			results[i] = s.evaluateCondition(subCond, ctx)
		}

		switch condition.Operator {
		case OpAnd:
			for _, result := range results {
				if !result {
					return false
				}
			}
			return true
		case OpOr:
			for _, result := range results {
				if result {
					return true
				}
			}
			return false
		case OpNot:
			if len(results) > 0 {
				return !results[0]
			}
		}
	}

	return false
}

// evaluateCustomScript 评估自定义脚本
func (s *BaseBossState) evaluateCustomScript(scriptName string, ctx *BossContext) bool {
	// 这里可以实现脚本引擎或者预定义的逻辑
	switch scriptName {
	case "in_attack_range":
		if ctx.Target == nil {
			return false
		}
		return ctx.Boss.IsInAttackRange(ctx.Target.GetPos().X, ctx.Target.GetPos().Y)

	case "chase_timeout_or_out_of_range":
		// 检查追击超时或超出范围
		if ctx.Target == nil {
			return true
		}
		maxChaseTime := time.Second * 30
		if s.GetTimeInState(ctx.CurrentTime) > maxChaseTime {
			return true
		}
		// 检查距离
		bossPos := ctx.Boss.GetPos()
		targetPos := ctx.Target.GetPos()
		distance := calculateDistance(
			float64(bossPos.X), float64(bossPos.Y),
			float64(targetPos.X), float64(targetPos.Y),
		)
		return distance > 300.0 // 超出追击范围

	case "target_out_of_attack_range":
		if ctx.Target == nil {
			return true
		}
		return !ctx.Boss.IsInAttackRange(ctx.Target.GetPos().X, ctx.Target.GetPos().Y)

	case "skill_cast_complete":
		// 这里需要检查技能释放是否完成
		// 暂时简单的时间检查
		return s.GetTimeInState(ctx.CurrentTime) > time.Second*2

	case "stun_expired":
		// 检查眩晕是否结束
		stunDuration := time.Second * 3 // 默认3秒
		return s.GetTimeInState(ctx.CurrentTime) >= stunDuration

	default:
		logger.Debugf("Custom script evaluation not implemented: %s", scriptName)
		return false
	}
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

// PatrolState 巡逻状态
type PatrolState struct {
	*BaseBossState
	patrolPoints    []Position
	currentPointIdx int
	moveSpeed       float64
	patrolRadius    float64
	alertRadius     float64
}

type Position struct {
	X, Y, Z float64
}

func NewPatrolState(points []Position) *PatrolState {
	base := NewBaseBossState(StatePatrol, "Patrol")
	return &PatrolState{
		BaseBossState:   base,
		patrolPoints:    points,
		currentPointIdx: 0,
		moveSpeed:       50.0,
		patrolRadius:    200.0,
		alertRadius:     150.0,
	}
}

func (s *PatrolState) OnEnter(ctx *BossContext) error {
	if err := s.BaseBossState.OnEnter(ctx); err != nil {
		return err
	}

	// 控制Monster真实状态
	ctx.Boss.Walk()

	return nil
}

func (s *PatrolState) OnUpdate(ctx *BossContext, deltaTime time.Duration) error {
	// 检查是否有敌人进入警戒范围
	enemies := ctx.GetEnemiesInRange(s.alertRadius)
	if len(enemies) > 0 {
		nearest := ctx.GetNearestEnemyInRange(s.alertRadius)
		if nearest != nil {
			ctx.Target = nearest
			ctx.Boss.SetCombatTarget(nearest)
			return nil
		}
	}

	// 继续巡逻逻辑
	if len(s.patrolPoints) > 0 {
		target := s.patrolPoints[s.currentPointIdx]
		currentPos := ctx.Boss.GetPos()

		// 检查是否到达目标点
		distance := calculateDistance(
			float64(currentPos.X), float64(currentPos.Y),
			target.X, target.Y,
		)

		if distance < 5.0 { // 接近目标点
			s.currentPointIdx = (s.currentPointIdx + 1) % len(s.patrolPoints)
		} else {
			// 移动向目标点
			return ctx.Boss.MoveTo(coord.Coord(target.X), coord.Coord(target.Y), coord.Coord(target.Z))
		}
	}

	return nil
}

// ChaseState 追击状态 - 现在依赖行为树执行具体行为
type ChaseState struct {
	*BaseBossState
}

func NewChaseState() *ChaseState {
	base := NewBaseBossState(StateChase, "Chase")
	return &ChaseState{
		BaseBossState: base,
	}
}

func (s *ChaseState) OnEnter(ctx *BossContext) error {
	if err := s.BaseBossState.OnEnter(ctx); err != nil {
		return err
	}

	// 控制Monster真实状态
	ctx.Boss.Chase()

	return nil
}

func (s *ChaseState) OnUpdate(ctx *BossContext, deltaTime time.Duration) error {
	// 调用基类的OnUpdate，它将执行行为树
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

	return nil
}

func (s *AttackState) OnUpdate(ctx *BossContext, deltaTime time.Duration) error {
	// 调用基类的OnUpdate，它将执行行为树
	return s.BaseBossState.OnUpdate(ctx, deltaTime)
}

// RetreatState 返回/撤退状态
type RetreatState struct {
	*BaseBossState
	spawnPoint   Position
	moveSpeed    float64
	healRate     float64
	damageReduce float64
}

func NewRetreatState(spawnPoint Position) *RetreatState {
	base := NewBaseBossState(StateRetreat, "Retreat")
	return &RetreatState{
		BaseBossState: base,
		spawnPoint:    spawnPoint,
		moveSpeed:     100.0, // 2倍速度
		healRate:      0.05,  // 每秒恢复5%生命值
		damageReduce:  0.5,   // 减少50%伤害
	}
}

func (s *RetreatState) OnEnter(ctx *BossContext) error {
	if err := s.BaseBossState.OnEnter(ctx); err != nil {
		return err
	}

	// 控制Monster真实状态
	ctx.Boss.Escape()

	// 清除战斗目标
	ctx.Target = nil
	ctx.Boss.SetCombatTarget(nil)

	logger.Debugf("Boss %d retreating to spawn point (%f, %f)",
		ctx.Boss.GetID(), s.spawnPoint.X, s.spawnPoint.Y)

	return nil
}

func (s *RetreatState) OnUpdate(ctx *BossContext, deltaTime time.Duration) error {
	// 调用基类的OnUpdate，它将执行行为树
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

func (s *DyingState) CanTransitionTo(stateID int32, ctx *BossContext) bool {
	// 死亡状态不能转换到其他状态
	return false
}

func (s *DyingState) GetNextState(ctx *BossContext) int32 {
	// 保持死亡状态
	return int32(StateDying)
}
