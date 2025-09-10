package bossai

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/ouyangzhongmin/gameserver/pkg/coord"
	"github.com/ouyangzhongmin/gameserver/pkg/logger"
	"github.com/ouyangzhongmin/gameserver/pkg/shape"
)

var rnd = rand.New(rand.NewSource(time.Now().UnixNano()))

// ===================== 新增的行为节点实现 =====================

// RandomMoveActionNode 随机移动节点
type RandomMoveActionNode struct {
	*ActionNode
	moveRadius   int
	lastMoveTime time.Time
	moveInterval time.Duration
	// 移动目标位置
	targetX   int
	targetY   int
	targetZ   int
	hasTarget bool
}

func NewRandomMoveActionNode(name string) *RandomMoveActionNode {
	return &RandomMoveActionNode{
		ActionNode:   NewActionNode(name),
		moveRadius:   100.0,
		moveInterval: time.Second * 30, // 默认5秒移动一次
	}
}

func (n *RandomMoveActionNode) Execute(ctx *BossContext) BehaviorResult {
	// 使用基类的通用状态检查
	if stateResult := n.ActionNode.CheckStateAndBeginExecution(ctx); stateResult != ResultRunning {
		return stateResult
	}

	// 检查移动时间间隔
	if ctx.CurrentTime.Sub(n.lastMoveTime) < n.moveInterval {
		n.ActionNode.SetComplete(ResultFailure)
		return ResultFailure
	}

	if ctx.Boss.IsChasing() || ctx.Boss.IsEscaping() || ctx.Boss.IsAttacking() || ctx.Boss.IsDied() {
		n.ActionNode.SetComplete(ResultFailure)
		return ResultFailure
	}

	// 获取移动半径参数
	if radius := n.GetParamAsInt("radius", 0); radius > 0 {
		n.moveRadius = radius
	}
	if interval := n.GetParam("interval"); interval != nil {
		if intervalStr, ok := interval.(string); ok {
			if parsed, err := time.ParseDuration(intervalStr); err == nil {
				n.moveInterval = parsed
			}
		}
	}

	bossPos := ctx.Boss.GetPos()

	// 如果没有目标位置，生成随机目标
	if !n.hasTarget {
		rpos, err := ctx.Boss.GetRandomPos(n.moveRadius)
		if err != nil {
			n.ActionNode.SetFailed()
			return ResultFailure
		}
		n.targetX = int(rpos.X)
		n.targetY = int(rpos.Y)
		n.targetZ = int(bossPos.Z)
		n.hasTarget = true
		logger.Debugf("RandomMove: Generated target (%d, %d, %d)", n.targetX, n.targetY, n.targetZ)
	}

	// 检查是否已到达目标位置
	dist := shape.CalculateDistance(
		float64(bossPos.X), float64(bossPos.Y),
		float64(n.targetX), float64(n.targetY),
	)

	if dist < 1 { // 到达目标附近
		logger.Debugf("RandomMove: Reached target, distance: %.2f", dist)
		n.lastMoveTime = ctx.CurrentTime
		n.hasTarget = false // 清除目标，下次会重新生成
		n.ActionNode.SetComplete(ResultSuccess)
		return ResultSuccess
	}

	// 检查Boss是否空闲（可以移动）
	if !ctx.Boss.IsIdle() {
		logger.Debugf("RandomMove: Boss not idle, distance: %.2f", dist)
		return ResultRunning // Boss在忩，等待
	}

	// 执行移动
	err := ctx.Boss.MoveTo(coord.Coord(n.targetX), coord.Coord(n.targetY), coord.Coord(n.targetZ))
	if err != nil {
		logger.Debugf("RandomMove: MoveTo failed: %v", err)
		n.hasTarget = false // 移动失败，清除目标
		n.ActionNode.SetFailed()
		return ResultFailure
	}

	logger.Debugf("RandomMove: Moving to target (%d, %.d), distance: %.2f", n.targetX, n.targetY, dist)
	return ResultRunning // 继续移动
}

func (n *RandomMoveActionNode) Reset() {
	n.ActionNode.Reset() // 调用基类的Reset
	n.hasTarget = false  // 清除目标位置
}

// PatrolMoveActionNode 巡逻移动节点
type PatrolMoveActionNode struct {
	*ActionNode
	patrolPoints    []coord.Vector3
	currentPointIdx int
	// 性能优化：限制MoveTo调用频率
	lastMoveTime time.Time     // 上次执行MoveTo的时间
	moveInterval time.Duration // MoveTo执行间隔
}

func NewPatrolMoveActionNode(name string) *PatrolMoveActionNode {
	return &PatrolMoveActionNode{
		ActionNode:      NewActionNode(name),
		patrolPoints:    make([]coord.Vector3, 0),
		currentPointIdx: 0,
		moveInterval:    time.Millisecond * 3, // 默认3s间隔（巡逻可以稍慢一些）
	}
}

func (n *PatrolMoveActionNode) Execute(ctx *BossContext) BehaviorResult {
	n.ActionNode.Execute(ctx) // 调用基类统计

	// 获取巡逻点配置
	if points := n.GetParam("patrol_points"); points != nil {
		if pointsSlice, ok := points.([]interface{}); ok {
			n.patrolPoints = make([]coord.Vector3, 0, len(pointsSlice))
			for _, point := range pointsSlice {
				if pointMap, ok := point.(map[string]interface{}); ok {
					if x, okX := pointMap["x"].(coord.Coord); okX {
						if y, okY := pointMap["y"].(coord.Coord); okY {
							var z coord.Coord = 0
							if zVal, okZ := pointMap["z"].(coord.Coord); okZ {
								z = coord.Coord(zVal)
							}
							n.patrolPoints = append(n.patrolPoints, coord.Vector3{X: x, Y: y, Z: z})
						}
					}
				}
			}
		}
	}
	// 获取移动间隔参数
	if interval := n.GetParam("move_interval"); interval != nil {
		if intervalStr, ok := interval.(string); ok {
			if parsed, err := time.ParseDuration(intervalStr); err == nil {
				n.moveInterval = parsed
			}
		}
	}

	if len(n.patrolPoints) == 0 {
		return ResultFailure
	}

	// 移动到目标点
	target := n.patrolPoints[n.currentPointIdx]
	bossPos := ctx.Boss.GetPos()
	distance := shape.CalculateDistance(
		float64(bossPos.X), float64(bossPos.Y),
		float64(target.X), float64(target.Y),
	)

	if distance < 1.0 { // 到达目标点
		n.currentPointIdx = (n.currentPointIdx + 1) % len(n.patrolPoints)
		return ResultSuccess
	}

	// 性能优化：限制MoveTo调用频率
	if ctx.CurrentTime.Sub(n.lastMoveTime) >= n.moveInterval {
		err := ctx.Boss.MoveTo(coord.Coord(target.X), coord.Coord(target.Y), coord.Coord(target.Z))
		if err != nil {
			logger.Debugf("PatrolMove: MoveTo failed: %v", err)
			return ResultFailure
		}

		n.lastMoveTime = ctx.CurrentTime
		logger.Debugf("PatrolMove: Moving to patrol point %d (%.2f, %.2f), distance: %.2f",
			n.currentPointIdx, target.X, target.Y, distance)
	} else {
		// 跨过MoveTo调用，节省性能
		logger.Debugf("PatrolMove: Skipping MoveTo - interval not reached")
	}

	return ResultRunning
}

// RandomSpeechActionNode 随机讲话节点
type RandomSpeechActionNode struct {
	*ActionNode
	speechTexts    []string
	lastSpeechTime time.Time
	speechInterval time.Duration
	speechChance   float64
}

func NewRandomSpeechActionNode(name string) *RandomSpeechActionNode {
	speechTexts := []string{"Hello!", "Hi there!", "How are you?"}
	return &RandomSpeechActionNode{
		ActionNode:     NewActionNode(name),
		speechTexts:    speechTexts,
		speechInterval: time.Second * 10, // 默认10秒检查一次
		speechChance:   0.1,              // 默认30%概率
	}
}

func (n *RandomSpeechActionNode) Execute(ctx *BossContext) BehaviorResult {
	n.ActionNode.Execute(ctx) // 调用基类统计

	// 检查时间间隔
	if ctx.CurrentTime.Sub(n.lastSpeechTime) < n.speechInterval {
		return ResultSuccess
	}

	// 获取配置参数
	if texts := n.GetParam("texts"); texts != nil {
		if textsSlice, ok := texts.([]interface{}); ok {
			n.speechTexts = make([]string, 0, len(textsSlice))
			for _, text := range textsSlice {
				if textStr, ok := text.(string); ok {
					n.speechTexts = append(n.speechTexts, textStr)
				}
			}
		}
	}
	if chance := n.GetParamAsFloat64("chance", 0); chance > 0 {
		n.speechChance = chance
	}

	// 检查概率
	if rnd.Float64() > n.speechChance {
		n.lastSpeechTime = ctx.CurrentTime
		return ResultSuccess
	}

	// 随机选择讲话内容
	if len(n.speechTexts) > 0 {
		text := n.speechTexts[rnd.Intn(len(n.speechTexts))]
		logger.Infof("Boss %d says: %s", ctx.Boss.GetID(), text)
		// 这里可以添加实际的讲话发送逻辑
	}

	n.lastSpeechTime = ctx.CurrentTime
	return ResultSuccess
}

// ScanEnemiesActionNode 扫描敌人节点
type ScanEnemiesActionNode struct {
	*ActionNode
	scanRadius   float64
	lastScanTime time.Time
	scanInterval time.Duration
}

func NewScanEnemiesActionNode(name string) *ScanEnemiesActionNode {
	return &ScanEnemiesActionNode{
		ActionNode:   NewActionNode(name),
		scanRadius:   10,
		scanInterval: time.Second * 2, // 默认2秒扫描一次
	}
}

func (n *ScanEnemiesActionNode) Execute(ctx *BossContext) BehaviorResult {
	// 使用基类的通用状态检查
	if stateResult := n.ActionNode.CheckStateAndBeginExecution(ctx); stateResult != ResultRunning {
		return stateResult
	}

	// 检查扫描时间间隔
	if ctx.CurrentTime.Sub(n.lastScanTime) < n.scanInterval {
		n.ActionNode.SetComplete(ResultFailure)
		return ResultFailure
	}

	// 获取扫描半径参数
	if radius := n.GetParamAsFloat64("radius", 0); radius > 0 {
		n.scanRadius = radius
	}

	// 从已缓存的NearbyEnemies中查找指定范围内的敌人
	// 找到敌人，设置目标
	nearest := ctx.GetNearestEnemyInRange(n.scanRadius)
	if nearest != nil {
		ctx.Target = nearest
		ctx.Boss.SetCombatTarget(nearest)
		logger.Debugf("Boss %d found enemy %d in range", ctx.Boss.GetID(), nearest.GetID())
	} else {
		logger.Debugf("Boss %d not found enemy", ctx.Boss.GetID())
	}

	n.lastScanTime = ctx.CurrentTime
	n.ActionNode.SetComplete(ResultSuccess)
	return ResultSuccess
}

func (n *ScanEnemiesActionNode) Reset() {
	n.ActionNode.Reset() // 调用基类的Reset
}

// BasicAttackActionNode 基础攻击节点
type BasicAttackActionNode struct {
	*ActionNode
	lastAttackTime time.Time
	attackCooldown time.Duration
}

func NewBasicAttackActionNode(name string) *BasicAttackActionNode {
	return &BasicAttackActionNode{
		ActionNode:     NewActionNode(name),
		attackCooldown: time.Millisecond * 1500, // 默认1.5秒攻击间隔
	}
}

func (n *BasicAttackActionNode) Execute(ctx *BossContext) BehaviorResult {
	// 使用基类的通用状态检查
	if stateResult := n.ActionNode.CheckStateAndBeginExecution(ctx); stateResult != ResultRunning {
		return stateResult
	}

	if ctx.Target == nil || !ctx.Target.IsAlive() {
		n.ActionNode.SetFailed()
		return ResultFailure
	}

	// 检查攻击冷却
	if ctx.CurrentTime.Sub(n.lastAttackTime) < n.attackCooldown {
		return ResultRunning
	}

	// 获取攻击冷却参数
	if cooldown := n.GetParam("cooldown"); cooldown != nil {
		if cooldownStr, ok := cooldown.(string); ok {
			if parsed, err := time.ParseDuration(cooldownStr); err == nil {
				n.attackCooldown = parsed
			}
		}
	}

	// 执行攻击
	err := ctx.Boss.DoAttackTarget(ctx.Target)
	if err != nil {
		logger.Errorf("Basic attack failed: %v", err)
		n.ActionNode.SetFailed()
		return ResultFailure
	}

	n.lastAttackTime = ctx.CurrentTime

	// 记录攻击动作
	action := &AIAction{
		Type:      ActionAttack,
		TargetID:  ctx.Target.GetID(),
		Timestamp: ctx.CurrentTime,
		Priority:  5,
		Executed:  true,
	}
	ctx.LastAction = action
	if ctx.ActionHistory == nil {
		ctx.ActionHistory = make([]*AIAction, 0)
	}
	ctx.ActionHistory = append(ctx.ActionHistory, action)

	n.ActionNode.SetComplete(ResultSuccess)
	return ResultSuccess
}

func (n *BasicAttackActionNode) Reset() {
	n.ActionNode.Reset() // 调用基类的Reset
}

// SkillUsageConditionNode 技能使用条件节点
type SkillUsageNode struct {
	*ActionNode
	skillId        int32 // 技能id， 如果是配置未指定，则通过GetAvailableSkill获取一个可用的技能
	lastAttackTime time.Time
	attackCooldown time.Duration
}

func NewSkillUsageNode(name string) *SkillUsageNode {
	return &SkillUsageNode{
		ActionNode:     NewActionNode(name),
		skillId:        0,
		attackCooldown: time.Millisecond * 1500, // 默认1.5秒攻击间隔
	}
}

func (n *SkillUsageNode) Execute(ctx *BossContext) BehaviorResult {
	n.ActionNode.Execute(ctx)
	if stateResult := n.ActionNode.CheckStateAndBeginExecution(ctx); stateResult != ResultRunning {
		return stateResult
	}

	if ctx.Target == nil || !ctx.Target.IsAlive() {
		n.ActionNode.SetFailed()
		return ResultFailure
	}

	// 检查攻击冷却
	if ctx.CurrentTime.Sub(n.lastAttackTime) < n.attackCooldown {
		return ResultRunning
	}
	n.lastAttackTime = ctx.CurrentTime

	// 技能id， 如果是配置未指定，则通过GetAvailableSkill获取一个可用的技能
	if n.skillId == 0 {
		skillId := int32(n.GetParamAsInt("skill_id", 0))
		if skillId == 0 {
			rules := n.GetParamAsString("rules", "")
			skillId = ctx.Boss.GetAvailableSkill(rules)
		} else {
			if !ctx.Boss.CanUseSkill(skillId) {
				// 技能不可用
				n.ActionNode.SetFailed()
				return ResultFailure
			}
		}
		n.skillId = skillId
	}

	// 检查当前阶段是否限制了可用技能
	if !ctx.IsSkillAllowedInCurrentPhase(n.skillId) {
		// 技能不在当前阶段允许列表中，尝试从阶段可用技能中选择一个
		availableSkills := ctx.GetAvailableSkillsFromCurrentPhase()
		if len(availableSkills) > 0 {
			skillFound := false
			for _, skillID := range availableSkills {
				if ctx.Boss.CanUseSkill(skillID) {
					n.skillId = skillID
					skillFound = true
					break
				}
			}

			// 如果阶段中的技能都不可用，则技能使用失败
			if !skillFound {
				logger.Errorf("SkillUsageConditionNode: No available skill found in current phase")
				n.ActionNode.SetFailed()
				return ResultFailure
			}
		}
	}

	if n.skillId == 0 {
		logger.Errorf("SkillUsageConditionNode: No available skill found")
		n.ActionNode.SetFailed()
		return ResultFailure
	}

	err := ctx.Boss.UseSkill(n.skillId, ctx.Target)
	if err != nil {
		logger.Errorf("SkillUsageConditionNode: UseSkill failed: %v", err)
		n.ActionNode.SetFailed()
		return ResultFailure
	}

	// 记录攻击动作
	action := &AIAction{
		Type:      ActionCastSkill,
		TargetID:  ctx.Target.GetID(),
		Timestamp: ctx.CurrentTime,
		Priority:  5,
		Executed:  true,
	}
	ctx.LastAction = action
	if ctx.ActionHistory == nil {
		ctx.ActionHistory = make([]*AIAction, 0)
	}
	ctx.ActionHistory = append(ctx.ActionHistory, action)

	n.ActionNode.SetComplete(ResultSuccess)
	return ResultSuccess
}

// EscapeCheckConditionNode 逃跑检查条件节点
type EscapeCheckConditionNode struct {
	*ConditionNode
	escapeThreshold float64
}

func NewEscapeCheckConditionNode(name string) *EscapeCheckConditionNode {
	return &EscapeCheckConditionNode{
		ConditionNode:   NewConditionNode(name),
		escapeThreshold: 0.2, // 20%血量以下逃跑
	}
}

func (n *EscapeCheckConditionNode) Execute(ctx *BossContext) BehaviorResult {
	n.ConditionNode.Execute(ctx)

	currentHP := float64(ctx.Boss.GetCurrentLife()) / float64(ctx.Boss.GetMaxLife())
	if currentHP > n.escapeThreshold {
		return ResultFailure
	}

	// 条件达成，请求转换到逃跑状态
	reason := fmt.Sprintf("Health below threshold: %.2f < %.2f", currentHP, n.escapeThreshold)
	ctx.RequestStateTransition(StateRetreat, reason, 100) // 高优先级

	n.ConditionNode.SetComplete(ResultSuccess)
	return ResultSuccess
}

// AutoRecoverActionNode 自动恢复节点
type AutoRecoverActionNode struct {
	*ActionNode
	healRate     float64
	lastHealTime time.Time
	heatInterval time.Duration
}

func NewAutoRecoverActionNode(name string) *AutoRecoverActionNode {
	return &AutoRecoverActionNode{
		ActionNode:   NewActionNode(name),
		healRate:     0.05, // 每秒恢复5%生命值
		heatInterval: time.Second,
	}
}

func (n *AutoRecoverActionNode) Execute(ctx *BossContext) BehaviorResult {
	n.ActionNode.Execute(ctx)

	// 检查治疗时间间隔
	if ctx.CurrentTime.Sub(n.lastHealTime) < n.heatInterval {
		return ResultSuccess
	}

	// 获取治疗参数
	if healRate := n.GetParamAsFloat64("heal_rate", 0); healRate > 0 {
		n.healRate = healRate
	}

	// 执行治疗（这里需要Boss实体提供Heal方法）
	healAmount := int32(float64(ctx.Boss.GetMaxLife()) * n.healRate)
	logger.Debugf("Boss %d auto healing %d", ctx.Boss.GetID(), healAmount)
	// ctx.Boss.Heal(healAmount) // 需要在IBossEntity接口中添加Heal方法

	n.lastHealTime = ctx.CurrentTime
	return ResultSuccess
}

// TriggerRewardActionNode 触发奖励节点
type TriggerRewardActionNode struct {
	*ActionNode
	rewardTriggered bool
}

func NewTriggerRewardActionNode(name string) *TriggerRewardActionNode {
	return &TriggerRewardActionNode{
		ActionNode: NewActionNode(name),
	}
}

func (n *TriggerRewardActionNode) Execute(ctx *BossContext) BehaviorResult {
	n.ActionNode.Execute(ctx)

	if n.rewardTriggered {
		return ResultSuccess
	}

	// 触发奖励逻辑
	logger.Infof("Boss %d death rewards triggered", ctx.Boss.GetID())
	// 这里可以添加具体的奖励触发逻辑
	// 例如：经验值、装备掉落、称号等

	n.rewardTriggered = true
	return ResultSuccess
}

// PatrolMovementActionNode 巡逻移动节点
type PatrolMovementActionNode struct {
	*ActionNode
	moveSpeed float64
}

func NewPatrolMovementActionNode(name string) *PatrolMovementActionNode {
	return &PatrolMovementActionNode{
		ActionNode: NewActionNode(name),
		moveSpeed:  50.0,
	}
}

func (n *PatrolMovementActionNode) Execute(ctx *BossContext) BehaviorResult {
	n.ActionNode.Execute(ctx)

	// 获取移动速度参数
	if speed := n.GetParamAsFloat64("speed", 0); speed > 0 {
		n.moveSpeed = speed
	}

	// 实际的巡逻移动逻辑将由PatrolMoveActionNode处理
	return ResultSuccess
}
