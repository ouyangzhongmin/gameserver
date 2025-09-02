package bossai

import (
	"math"
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
}

// NewBaseBossState 创建基础状态
func NewBaseBossState(id BossStateID, name string) *BaseBossState {
	return &BaseBossState{
		id:          id,
		name:        name,
		transitions: make(map[BossStateID]StateTransition),
		modifiers:   make(map[string]interface{}),
		isActive:    false,
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
	// 基础状态不执行任何操作
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
	// 暂时返回false，后续可以扩展
	logger.Debugf("Custom script evaluation not implemented: %s", scriptName)
	return false
}

// IdleState 空闲状态
type IdleState struct {
	*BaseBossState
	patrolRadius float64
	scanInterval time.Duration
	lastScanTime time.Time
}

func NewIdleState() *IdleState {
	base := NewBaseBossState(StateIdle, "Idle")
	return &IdleState{
		BaseBossState: base,
		patrolRadius:  100.0,
		scanInterval:  time.Second * 2,
	}
}

func (s *IdleState) OnEnter(ctx *BossContext) error {
	if err := s.BaseBossState.OnEnter(ctx); err != nil {
		return err
	}

	// 清除战斗目标
	ctx.Boss.SetCombatTarget(nil)
	ctx.Target = nil

	return nil
}

func (s *IdleState) OnUpdate(ctx *BossContext, deltaTime time.Duration) error {
	// 定期扫描敌人
	if ctx.CurrentTime.Sub(s.lastScanTime) >= s.scanInterval {
		s.lastScanTime = ctx.CurrentTime

		// 搜索附近敌人
		enemies := ctx.Boss.GetEntitiesInRange(s.patrolRadius)
		if len(enemies) > 0 {
			// 找到敌人，准备切换到追击状态
			nearest := ctx.Boss.GetNearestEnemy()
			if nearest != nil {
				ctx.Target = nearest
				ctx.Boss.SetCombatTarget(nearest)
			}
		}
	}

	return nil
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

func (s *PatrolState) OnUpdate(ctx *BossContext, deltaTime time.Duration) error {
	// 检查是否有敌人进入警戒范围
	enemies := ctx.Boss.GetEntitiesInRange(s.alertRadius)
	if len(enemies) > 0 {
		nearest := ctx.Boss.GetNearestEnemy()
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

// ChaseState 追击状态
type ChaseState struct {
	*BaseBossState
	chaseRadius  float64
	attackRange  float64
	maxChaseTime time.Duration
}

func NewChaseState() *ChaseState {
	base := NewBaseBossState(StateChase, "Chase")
	return &ChaseState{
		BaseBossState: base,
		chaseRadius:   300.0,
		attackRange:   50.0,
		maxChaseTime:  time.Second * 30,
	}
}

func (s *ChaseState) OnUpdate(ctx *BossContext, deltaTime time.Duration) error {
	if ctx.Target == nil || !ctx.Target.IsAlive() {
		// 目标丢失或死亡，返回idle
		ctx.Target = nil
		ctx.Boss.SetCombatTarget(nil)
		return nil
	}

	// 检查是否追击时间过长
	if s.GetTimeInState(ctx.CurrentTime) > s.maxChaseTime {
		logger.Debugf("Boss %d chase timeout, returning to idle", ctx.Boss.GetID())
		ctx.Target = nil
		ctx.Boss.SetCombatTarget(nil)
		return nil
	}

	// 计算与目标的距离
	targetPos := ctx.Target.GetPos()
	currentPos := ctx.Boss.GetPos()
	distance := calculateDistance(
		float64(currentPos.X), float64(currentPos.Y),
		float64(targetPos.X), float64(targetPos.Y),
	)

	// 检查是否超出追击范围
	if distance > s.chaseRadius {
		logger.Debugf("Boss %d target out of chase range", ctx.Boss.GetID())
		ctx.Target = nil
		ctx.Boss.SetCombatTarget(nil)
		return nil
	}

	// 检查是否进入攻击范围
	if distance <= s.attackRange {
		// 可以攻击了
		return nil
	}

	// 继续追击
	return ctx.Boss.MoveTo(targetPos.X, targetPos.Y, targetPos.Z)
}

// AttackState 攻击状态
type AttackState struct {
	*BaseBossState
	attackCooldown time.Duration
	lastAttackTime time.Time
	comboCount     int
	maxComboCount  int
}

func NewAttackState() *AttackState {
	base := NewBaseBossState(StateAttack, "Attack")
	return &AttackState{
		BaseBossState:  base,
		attackCooldown: time.Millisecond * 1500,
		comboCount:     0,
		maxComboCount:  3,
	}
}

func (s *AttackState) OnUpdate(ctx *BossContext, deltaTime time.Duration) error {
	if ctx.Target == nil || !ctx.Target.IsAlive() {
		ctx.Target = nil
		ctx.Boss.SetCombatTarget(nil)
		return nil
	}

	// 检查攻击冷却
	if ctx.CurrentTime.Sub(s.lastAttackTime) < s.attackCooldown {
		return nil
	}

	// 检查攻击范围
	if !ctx.Boss.IsInAttackRange(ctx.Target.GetPos().X, ctx.Target.GetPos().Y) {
		// 目标超出攻击范围，切换回追击状态
		return nil
	}

	// 执行攻击
	s.performAttack(ctx)
	s.lastAttackTime = ctx.CurrentTime
	s.comboCount++

	// 检查连击数
	if s.comboCount >= s.maxComboCount {
		s.comboCount = 0
		// 可以考虑切换到其他状态或使用技能
	}

	return nil
}

func (s *AttackState) OnEnter(ctx *BossContext) error {
	if err := s.BaseBossState.OnEnter(ctx); err != nil {
		return err
	}

	// 停止移动
	ctx.Boss.Stop()
	s.comboCount = 0

	return nil
}

func (s *AttackState) performAttack(ctx *BossContext) {
	// 这里可以实现具体的攻击逻辑
	logger.Debugf("Boss %d attacking target %d", ctx.Boss.GetID(), ctx.Target.GetID())

	// 记录攻击动作
	action := &AIAction{
		Type:      ActionAttack,
		TargetID:  ctx.Target.GetID(),
		Timestamp: ctx.CurrentTime,
		Priority:  5,
	}

	ctx.LastAction = action
	if ctx.ActionHistory == nil {
		ctx.ActionHistory = make([]*AIAction, 0)
	}
	ctx.ActionHistory = append(ctx.ActionHistory, action)
}

// 辅助函数
func calculateDistance(x1, y1, x2, y2 float64) float64 {
	dx := x2 - x1
	dy := y2 - y1
	return math.Sqrt(dx*dx + dy*dy)
}
