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
		moveInterval: time.Second * 20, // 默认10秒移动一次
	}
}

func (n *RandomMoveActionNode) Execute(ctx *BossContext) BehaviorResult {
	// 检查移动时间间隔
	if ctx.CurrentTime.Sub(n.lastMoveTime) < n.moveInterval && n.executeState != NodeStateRunning {
		return ResultFailure
	}

	// 使用基类的通用状态检查
	if n.executeState == NodeStateIdle {
		// 初始化数据
		n.executeState = NodeStateRunning
		n.hasTarget = false

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
	}

	if ctx.Boss.IsChasing() || ctx.Boss.IsEscaping() || ctx.Boss.IsAttacking() || ctx.Boss.IsDied() {
		n.SetComplete(ResultFailure)
		return ResultFailure
	}

	bossPos := ctx.Boss.GetPos()

	// 如果没有目标位置，生成随机目标
	if !n.hasTarget {
		rpos, err := ctx.Boss.GetRandomPos(n.moveRadius)
		if err != nil {
			n.SetFailed()
			return ResultFailure
		}
		n.targetX = int(rpos.X)
		n.targetY = int(rpos.Y)
		n.targetZ = int(bossPos.Z)
		n.hasTarget = true
		logger.Debugf("RandomMove: Generated target (%d, %d, %d)", n.targetX, n.targetY, n.targetZ)
	}
	n.lastMoveTime = ctx.CurrentTime
	// 检查是否已到达目标位置
	dist := shape.CalculateDistance(
		float64(bossPos.X), float64(bossPos.Y),
		float64(n.targetX), float64(n.targetY),
	)

	if dist < 1 { // 到达目标附近
		logger.Debugf("RandomMove: Reached target, distance: %.2f", dist)
		n.Reset() // 清除目标，下次会重新生成
		return ResultSuccess
	}

	// 检查Boss是否空闲（可以移动）
	if ctx.Boss.IsIdle() {
		// 执行移动
		err := ctx.Boss.WalkTo(coord.Coord(n.targetX), coord.Coord(n.targetY), coord.Coord(n.targetZ))
		if err != nil {
			logger.Debugf("RandomMove: WalkTo failed: %v", err)
			n.Reset()
			return ResultFailure
		}
		logger.Debugf("RandomMove: Walking to target (%d, %.d), distance: %.2f", n.targetX, n.targetY, dist)
		return ResultRunning // Boss在移动中，等待
	}

	logger.Debugf("RandomMove: Boss is Walking, distance: %.2f", dist)
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
	if len(n.patrolPoints) == 0 {
		n.SetFailed()
		return ResultFailure
	}

	// 检查移动时间间隔
	if ctx.CurrentTime.Sub(n.lastMoveTime) < n.moveInterval && n.executeState != NodeStateRunning {
		n.SetComplete(ResultFailure)
		return ResultFailure
	}

	if n.executeState == NodeStateIdle {
		// 初始化数据
		n.executeState = NodeStateRunning
		n.currentPointIdx = 0

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
	}

	if ctx.Boss.IsChasing() || ctx.Boss.IsEscaping() || ctx.Boss.IsAttacking() || ctx.Boss.IsDied() {
		n.SetComplete(ResultFailure)
		return ResultFailure
	}

	// 移动到目标点
	target := n.patrolPoints[n.currentPointIdx]
	bossPos := ctx.Boss.GetPos()
	distance := shape.CalculateDistance(
		float64(bossPos.X), float64(bossPos.Y),
		float64(target.X), float64(target.Y),
	)
	n.lastMoveTime = ctx.CurrentTime

	if distance < 1.0 { // 到达目标点
		n.currentPointIdx++
		n.currentPointIdx = n.currentPointIdx % len(n.patrolPoints) // 循环
		n.SetComplete(ResultSuccess)
		return ResultSuccess
	}

	// 检查Boss是否空闲（可以移动）
	if ctx.Boss.IsIdle() {
		err := ctx.Boss.WalkTo(coord.Coord(target.X), coord.Coord(target.Y), coord.Coord(target.Z))
		if err != nil {
			logger.Debugf("PatrolMove: MoveTo failed: %v", err)
			return ResultFailure
		}

		logger.Debugf("PatrolMove: Moving to patrol point %d (%.2f, %.2f), distance: %.2f",
			n.currentPointIdx, target.X, target.Y, distance)
		return ResultRunning
	}
	logger.Debugf("PatrolMove: Boss is moving, distance: %.2f", distance)
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
		speechInterval: time.Second * 20, // 默认10秒检查一次
		speechChance:   0.1,              // 默认30%概率
	}
}

func (n *RandomSpeechActionNode) Execute(ctx *BossContext) BehaviorResult {
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
	// 检查扫描时间间隔
	if ctx.CurrentTime.Sub(n.lastScanTime) < n.scanInterval {
		n.SetComplete(ResultFailure)
		return ResultFailure
	}

	if ctx.Target != nil && ctx.Target.IsAlive() {
		n.SetComplete(ResultFailure)
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
		//logger.Debugf("Boss %d not found enemy", ctx.Boss.GetID())
	}

	n.lastScanTime = ctx.CurrentTime
	return ResultSuccess
}

// BasicAttackActionNode 基础攻击节点
type BasicAttackActionNode struct {
	*ActionNode
	lastAttackTime time.Time
}

func NewBasicAttackActionNode(name string) *BasicAttackActionNode {
	return &BasicAttackActionNode{
		ActionNode: NewActionNode(name),
	}
}

func (n *BasicAttackActionNode) Execute(ctx *BossContext) BehaviorResult {
	if ctx.Target == nil || !ctx.Target.IsAlive() {
		n.SetComplete(ResultFailure)
		return ResultFailure
	}

	// 检查攻击冷却
	attackCooldown := time.Duration(ctx.Boss.GetAttackDuration()) * time.Millisecond
	delta := ctx.CurrentTime.Sub(n.lastAttackTime)
	if delta < attackCooldown {
		n.SetComplete(ResultFailure)
		return ResultFailure
	}

	if ctx.Boss.IsAttacking() {
		return ResultRunning
	}

	// 执行攻击
	err := ctx.Boss.DoAttackTarget(ctx.Target)
	if err != nil {
		logger.Errorf("Basic attack failed: %v", err)
		n.SetFailed()
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

	n.SetComplete(ResultSuccess)
	return ResultSuccess
}

// CombatAttackActionNode 组合基础连击节点
type CombatAttackActionNode struct {
	*ActionNode
	lastAttackTime time.Time
	attackCount    int           // 连击次数
	attackIndex    int           //
	combatInterval time.Duration // 连击间隔
	attackInterval time.Duration // 攻击间隔
}

func NewCombatAttackActionNode(name string) *CombatAttackActionNode {
	return &CombatAttackActionNode{
		ActionNode:     NewActionNode(name),
		attackCount:    3,
		attackIndex:    0,
		combatInterval: 300 * time.Millisecond,  // 连招间隔
		attackInterval: 5000 * time.Millisecond, // 攻击间隔
	}
}

func (n *CombatAttackActionNode) Execute(ctx *BossContext) BehaviorResult {
	// 检查攻击冷却
	if ctx.CurrentTime.Sub(n.lastAttackTime) < n.attackInterval && n.executeState != NodeStateRunning {
		n.SetFailed()
		return ResultFailure
	}

	if n.executeState == NodeStateIdle {
		// 初始化数据
		n.executeState = NodeStateRunning
		n.attackIndex = 0
		if count := n.GetParamAsInt("count", 0); count > 0 {
			n.attackCount = count
		}
		if interval := n.GetParamAsInt("interval", 0); interval > 0 {
			n.combatInterval = time.Duration(interval) * time.Millisecond
		}
	}

	if ctx.Target == nil || !ctx.Target.IsAlive() {
		n.SetFailed()
		return ResultFailure
	}

	// 连招的间隔
	delta2 := ctx.CurrentTime.Sub(n.lastAttackTime)
	if delta2 < n.combatInterval {
		return ResultRunning
	}

	// 执行攻击
	err := ctx.Boss.DoAttackTarget(ctx.Target)
	if err != nil {
		logger.Errorf("Basic attack failed: %v", err)
		n.SetFailed()
		return ResultFailure
	}

	n.lastAttackTime = ctx.CurrentTime

	// 记录攻击动作
	action := &AIAction{
		Type:      ActionCombatAttack,
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
	n.attackIndex++
	if n.attackIndex >= n.attackCount {
		// 连击完成
		n.Reset()
		return ResultSuccess
	}
	return ResultRunning
}

func (n *CombatAttackActionNode) Reset() {
	n.BaseBehaviorNode.Reset()
	n.attackIndex = 0
}

// SkillUsageConditionNode 技能使用条件节点
type SkillUsageNode struct {
	*ActionNode
	skillId        int32 // 技能id， 如果是配置未指定，则通过GetAvailableSkill获取一个可用的技能
	lastAttackTime time.Time
	attackInterval time.Duration
}

func NewSkillUsageNode(name string) *SkillUsageNode {
	return &SkillUsageNode{
		ActionNode:     NewActionNode(name),
		skillId:        0,
		attackInterval: time.Millisecond * 5000, // 默认5秒攻击间隔
	}
}

func (n *SkillUsageNode) Execute(ctx *BossContext) BehaviorResult {
	n.ActionNode.Execute(ctx)
	// 检查攻击冷却
	if ctx.CurrentTime.Sub(n.lastAttackTime) < n.attackInterval {
		n.SetFailed()
		return ResultFailure
	}
	if ctx.Target == nil || !ctx.Target.IsAlive() {
		n.SetFailed()
		return ResultFailure
	}

	// 技能id， 如果是配置未指定，则通过GetAvailableSkill获取一个可用的技能
	if n.skillId == 0 {
		skillId := int32(n.GetParamAsInt("skill_id", 0))
		if skillId == 0 {
			rules := n.GetParamAsString("rules", "")
			skillId = ctx.Boss.GetAvailableSkill(rules)
			n.skillId = skillId
		} else {
			if !ctx.Boss.CanUseSkill(skillId) {
				// 技能不可用
				n.SetFailed()
				return ResultFailure
			}
			n.skillId = skillId
		}
	}

	if n.skillId == 0 {
		logger.Errorf("SkillUsageNode: No available skill found")
		n.SetFailed()
		return ResultFailure
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
				logger.Errorf("SkillUsageNode: No available skill found in current phase")
				n.SetFailed()
				return ResultFailure
			}
		}
	}

	err := ctx.Boss.UseSkill(n.skillId, ctx.Target)
	if err != nil {
		logger.Errorf("SkillUsageNode: UseSkill failed: %v", err)
		n.SetFailed()
		return ResultFailure
	}
	n.lastAttackTime = ctx.CurrentTime
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

	n.Reset() // 重置下一次重新获取新的skillid
	return ResultSuccess
}

func (n *SkillUsageNode) Reset() {
	n.BaseBehaviorNode.Reset()
	n.skillId = 0
}

// EscapeCheckConditionNode 逃跑检查条件节点
type EscapeCheckConditionNode struct {
	*ActionNode
	escapeThreshold float64
}

func NewEscapeCheckConditionNode(name string) *EscapeCheckConditionNode {
	return &EscapeCheckConditionNode{
		ActionNode:      NewActionNode(name),
		escapeThreshold: 0.2, // 20%血量以下逃跑
	}
}

func (n *EscapeCheckConditionNode) Execute(ctx *BossContext) BehaviorResult {
	n.ActionNode.Execute(ctx)

	currentHP := float64(ctx.Boss.GetCurrentLife()) / float64(ctx.Boss.GetMaxLife())
	if currentHP > n.escapeThreshold {
		n.SetFailed()
		return ResultFailure
	}

	// 条件达成，请求转换到逃跑状态
	reason := fmt.Sprintf("Health below threshold: %.2f < %.2f", currentHP, n.escapeThreshold)
	ctx.RequestStateTransition(StateRetreat, reason, 100) // 高优先级

	n.SetComplete(ResultSuccess)
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
		// 这里奖励只允许触发一次
		return ResultSuccess
	}

	// 触发奖励逻辑
	logger.Infof("Boss %d death rewards triggered", ctx.Boss.GetID())
	// 这里可以添加具体的奖励触发逻辑
	// 例如：经验值、装备掉落、称号等

	n.rewardTriggered = true
	return ResultSuccess
}
