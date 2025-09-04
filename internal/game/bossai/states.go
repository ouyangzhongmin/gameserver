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

	// 控制Monster真实状态
	ctx.Boss.Idle()

	// 清除战斗目标
	ctx.Boss.SetCombatTarget(nil)
	ctx.Target = nil

	return nil
}

func (s *IdleState) OnUpdate(ctx *BossContext, deltaTime time.Duration) error {
	// 定期扫描敌人
	if ctx.CurrentTime.Sub(s.lastScanTime) >= s.scanInterval {
		s.lastScanTime = ctx.CurrentTime

		// 使用BossContext的方法从已缓存的NearbyEnemies中查找
		enemies := ctx.GetEnemiesInRange(s.patrolRadius)
		if len(enemies) > 0 {
			// 找到敌人，准备切换到追击状态
			nearest := ctx.GetNearestEnemyInRange(s.patrolRadius)
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

func (s *ChaseState) OnEnter(ctx *BossContext) error {
	if err := s.BaseBossState.OnEnter(ctx); err != nil {
		return err
	}

	// 控制Monster真实状态
	ctx.Boss.Chase()

	return nil
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

	// 控制Monster真实状态
	ctx.Boss.AttackAction()

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
	currentPos := ctx.Boss.GetPos()
	distance := calculateDistance(
		float64(currentPos.X), float64(currentPos.Y),
		s.spawnPoint.X, s.spawnPoint.Y,
	)

	// 检查是否到达出生点
	if distance < 10.0 {
		// 到达出生点，可以切换到巡逻状态
		logger.Debugf("Boss %d reached spawn point, switching to patrol", ctx.Boss.GetID())
		return nil
	}

	// 继续返回出生点
	return ctx.Boss.MoveTo(
		coord.Coord(s.spawnPoint.X),
		coord.Coord(s.spawnPoint.Y),
		coord.Coord(s.spawnPoint.Z),
	)
}

func (s *RetreatState) OnExit(ctx *BossContext) error {
	if err := s.BaseBossState.OnExit(ctx); err != nil {
		return err
	}

	// 恢复到巡逻状态
	ctx.Boss.Walk()

	return nil
}

// StunnedState 眩晕状态
type StunnedState struct {
	*BaseBossState
	stunDuration    time.Duration
	damageReduction float64
}

func NewStunnedState(duration time.Duration) *StunnedState {
	base := NewBaseBossState(StateStunned, "Stunned")
	return &StunnedState{
		BaseBossState:   base,
		stunDuration:    duration,
		damageReduction: 0.3, // 眩晕时减少30%伤害
	}
}

func (s *StunnedState) OnEnter(ctx *BossContext) error {
	if err := s.BaseBossState.OnEnter(ctx); err != nil {
		return err
	}

	// 控制Monster真实状态 - 眩晕时保持当前状态但停止移动
	ctx.Boss.Stop()

	logger.Debugf("Boss %d stunned for %v", ctx.Boss.GetID(), s.stunDuration)

	return nil
}

func (s *StunnedState) OnUpdate(ctx *BossContext, deltaTime time.Duration) error {
	// 检查眩晕时间是否结束
	if s.GetTimeInState(ctx.CurrentTime) >= s.stunDuration {
		// 眩晕结束，可以切换到其他状态
		logger.Debugf("Boss %d stun expired", ctx.Boss.GetID())
		return nil
	}

	// 眩晕期间无法行动
	return nil
}

func (s *StunnedState) OnExit(ctx *BossContext) error {
	if err := s.BaseBossState.OnExit(ctx); err != nil {
		return err
	}

	// 眩晕结束，恢复正常状态判断
	if ctx.Target != nil && ctx.Target.IsAlive() {
		ctx.Boss.Chase()
	} else {
		ctx.Boss.Idle()
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
	// 死亡状态不需要更新逻辑，等待系统处理
	return nil
}

func (s *DyingState) CanTransitionTo(stateID int32, ctx *BossContext) bool {
	// 死亡状态不能转换到其他状态
	return false
}

func (s *DyingState) GetNextState(ctx *BossContext) int32 {
	// 保持死亡状态
	return int32(StateDying)
}

// CastSkillState 释放技能状态
type CastSkillState struct {
	*BaseBossState
	currentSkillID int32
	castStartTime  time.Time
}

func NewCastSkillState() *CastSkillState {
	base := NewBaseBossState(StateCastSkill, "CastSkill")
	return &CastSkillState{
		BaseBossState: base,
	}
}

func (s *CastSkillState) OnEnter(ctx *BossContext) error {
	if err := s.BaseBossState.OnEnter(ctx); err != nil {
		return err
	}

	// 控制Monster真实状态
	ctx.Boss.AttackAction()

	s.castStartTime = ctx.CurrentTime
	logger.Debugf("Boss %d is casting skill %d", ctx.Boss.GetID(), s.currentSkillID)

	return nil
}

func (s *CastSkillState) OnUpdate(ctx *BossContext, deltaTime time.Duration) error {
	// 技能释放状态的更新逻辑
	// 检查技能释放是否完成
	return nil
}

func (s *CastSkillState) OnExit(ctx *BossContext) error {
	if err := s.BaseBossState.OnExit(ctx); err != nil {
		return err
	}

	s.currentSkillID = 0
	return nil
}

// EnragedState 狂暴状态
type EnragedState struct {
	*BaseBossState
	enrageStartTime  time.Time
	damageMultiplier float64
	attackSpeedBonus float64
}

func NewEnragedState() *EnragedState {
	base := NewBaseBossState(StateEnraged, "Enraged")
	return &EnragedState{
		BaseBossState:    base,
		damageMultiplier: 1.5, // 默认1.5倍伤害
		attackSpeedBonus: 2.0, // 默认2個攻击速度
	}
}

func (s *EnragedState) OnEnter(ctx *BossContext) error {
	if err := s.BaseBossState.OnEnter(ctx); err != nil {
		return err
	}

	// 控制Monster真实状态
	ctx.Boss.AttackAction()

	s.enrageStartTime = ctx.CurrentTime
	logger.Debugf("Boss %d is entering enraged state", ctx.Boss.GetID())

	return nil
}

func (s *EnragedState) OnUpdate(ctx *BossContext, deltaTime time.Duration) error {
	// 检查是否有有效目标
	if ctx.Target == nil || !ctx.Target.IsAlive() {
		// 寻找新目标
		enemies := ctx.GetEnemiesInRange(300.0)
		if len(enemies) > 0 {
			ctx.Target = enemies[0]
			ctx.Boss.SetCombatTarget(ctx.Target)
		}
	}

	// 狂暴状态下的攻击逻辑
	if ctx.Target != nil {
		// 计算目标距离
		bossPos := ctx.Boss.GetPos()
		targetPos := ctx.Target.GetPos()
		distance := float64(bossPos.DistanceTo(targetPos))

		// 如果在攻击范围内，执行攻击
		if distance <= 60.0 {
			if err := ctx.Boss.DoAttackTarget(ctx.Target); err != nil {
				logger.Errorf("Boss enraged attack failed: %v", err)
			}
		} else {
			// 追击目标
			ctx.Boss.Chase()
		}
	}

	return nil
}

func (s *EnragedState) OnExit(ctx *BossContext) error {
	if err := s.BaseBossState.OnExit(ctx); err != nil {
		return err
	}

	logger.Debugf("Boss %d exiting enraged state", ctx.Boss.GetID())
	return nil
}
