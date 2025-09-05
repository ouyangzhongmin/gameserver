package bossai

import (
	"math/rand"
	"time"

	"github.com/ouyangzhongmin/gameserver/pkg/coord"
	"github.com/ouyangzhongmin/gameserver/pkg/logger"
)

var rnd = rand.New(rand.NewSource(time.Now().UnixNano()))

// ===================== 新增的行为节点实现 =====================

// RandomMoveActionNode 随机移动节点
type RandomMoveActionNode struct {
	*ActionNode
	moveRadius   float64
	lastMoveTime time.Time
	moveInterval time.Duration
}

func NewRandomMoveActionNode(name string) *RandomMoveActionNode {
	return &RandomMoveActionNode{
		ActionNode:   NewActionNode(name),
		moveRadius:   100.0,
		moveInterval: time.Second * 5, // 默认5秒移动一次
	}
}

func (n *RandomMoveActionNode) Execute(ctx *BossContext) BehaviorResult {
	n.ActionNode.Execute(ctx) // 调用基类统计

	// 检查移动时间间隔
	if ctx.CurrentTime.Sub(n.lastMoveTime) < n.moveInterval {
		return ResultSuccess
	}
	if ctx.Boss.IsIdle() {
		return ResultSuccess
	}

	// 获取移动半径参数
	if radius := n.GetParamAsFloat64("radius", 0); radius > 0 {
		n.moveRadius = radius
	}
	if interval := n.GetParam("interval"); interval != nil {
		if intervalStr, ok := interval.(string); ok {
			if parsed, err := time.ParseDuration(intervalStr); err == nil {
				n.moveInterval = parsed
			}
		}
	}

	// 随机移动
	bossPos := ctx.Boss.GetPos()
	randomX := float64(bossPos.X) + (2*rnd.Float64()-1)*n.moveRadius
	randomY := float64(bossPos.Y) + (2*rnd.Float64()-1)*n.moveRadius

	err := ctx.Boss.MoveTo(coord.Coord(randomX), coord.Coord(randomY), bossPos.Z)
	if err != nil {
		logger.Debugf("Random move failed: %v", err)
		return ResultFailure
	}

	n.lastMoveTime = ctx.CurrentTime
	return ResultSuccess
}

// PatrolMoveActionNode 巡逻移动节点
type PatrolMoveActionNode struct {
	*ActionNode
	patrolPoints    []Position
	currentPointIdx int
}

func NewPatrolMoveActionNode(name string) *PatrolMoveActionNode {
	return &PatrolMoveActionNode{
		ActionNode:      NewActionNode(name),
		patrolPoints:    make([]Position, 0),
		currentPointIdx: 0,
	}
}

func (n *PatrolMoveActionNode) Execute(ctx *BossContext) BehaviorResult {
	n.ActionNode.Execute(ctx) // 调用基类统计

	// 获取巡逻点配置
	if points := n.GetParam("patrol_points"); points != nil {
		if pointsSlice, ok := points.([]interface{}); ok {
			n.patrolPoints = make([]Position, 0, len(pointsSlice))
			for _, point := range pointsSlice {
				if pointMap, ok := point.(map[string]interface{}); ok {
					if x, okX := pointMap["x"].(float64); okX {
						if y, okY := pointMap["y"].(float64); okY {
							z := 0.0
							if zVal, okZ := pointMap["z"].(float64); okZ {
								z = zVal
							}
							n.patrolPoints = append(n.patrolPoints, Position{X: x, Y: y, Z: z})
						}
					}
				}
			}
		}
	}

	if len(n.patrolPoints) == 0 {
		return ResultFailure
	}

	// 移动到目标点
	target := n.patrolPoints[n.currentPointIdx]
	bossPos := ctx.Boss.GetPos()
	distance := calculateDistance(
		float64(bossPos.X), float64(bossPos.Y),
		target.X, target.Y,
	)

	if distance < 5.0 { // 到达目标点
		n.currentPointIdx = (n.currentPointIdx + 1) % len(n.patrolPoints)
		return ResultSuccess
	}

	err := ctx.Boss.MoveTo(coord.Coord(target.X), coord.Coord(target.Y), coord.Coord(target.Z))
	if err != nil {
		return ResultFailure
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
	return &RandomSpeechActionNode{
		ActionNode:     NewActionNode(name),
		speechTexts:    make([]string, 0),
		speechInterval: time.Second * 10, // 默认10秒检查一次
		speechChance:   0.3,              // 默认30%概率
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
		scanRadius:   150.0,
		scanInterval: time.Second * 2, // 默认2秒扫描一次
	}
}

func (n *ScanEnemiesActionNode) Execute(ctx *BossContext) BehaviorResult {
	n.ActionNode.Execute(ctx) // 调用基类统计

	// 检查扫描时间间隔
	if ctx.CurrentTime.Sub(n.lastScanTime) < n.scanInterval {
		return ResultSuccess
	}

	// 获取扫描半径参数
	if radius := n.GetParamAsFloat64("radius", 0); radius > 0 {
		n.scanRadius = radius
	}

	// 从已缓存的NearbyEnemies中查找指定范围内的敌人
	enemies := ctx.GetEnemiesInRange(n.scanRadius)
	if len(enemies) > 0 {
		// 找到敌人，设置目标
		nearest := ctx.GetNearestEnemyInRange(n.scanRadius)
		if nearest != nil {
			ctx.Target = nearest
			ctx.Boss.SetCombatTarget(nearest)
			logger.Debugf("Boss %d found enemy %d in range", ctx.Boss.GetID(), nearest.GetID())
		}
	}

	n.lastScanTime = ctx.CurrentTime
	return ResultSuccess
}

// ChaseTargetActionNode 追击目标节点
type ChaseTargetActionNode struct {
	*ActionNode
	maxChaseDistance float64
}

func NewChaseTargetActionNode(name string) *ChaseTargetActionNode {
	return &ChaseTargetActionNode{
		ActionNode:       NewActionNode(name),
		maxChaseDistance: 300.0,
	}
}

func (n *ChaseTargetActionNode) Execute(ctx *BossContext) BehaviorResult {
	n.ActionNode.Execute(ctx) // 调用基类统计

	if ctx.Target == nil || !ctx.Target.IsAlive() {
		return ResultFailure
	}

	// 获取追击距离参数
	if distance := n.GetParamAsFloat64("max_distance", 0); distance > 0 {
		n.maxChaseDistance = distance
	}

	// 检查距离
	bossPos := ctx.Boss.GetPos()
	targetPos := ctx.Target.GetPos()
	distance := calculateDistance(
		float64(bossPos.X), float64(bossPos.Y),
		float64(targetPos.X), float64(targetPos.Y),
	)

	if distance > n.maxChaseDistance {
		// 超出追击范围，放弃目标
		ctx.Target = nil
		ctx.Boss.SetCombatTarget(nil)
		return ResultFailure
	}

	// 追击目标
	err := ctx.Boss.MoveTo(targetPos.X, targetPos.Y, targetPos.Z)
	if err != nil {
		return ResultFailure
	}

	return ResultRunning
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
	n.ActionNode.Execute(ctx) // 调用基类统计

	if ctx.Target == nil || !ctx.Target.IsAlive() {
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

	return ResultSuccess
}

// 新增的条件节点

// RandomChanceConditionNode 随机概率条件节点
type RandomChanceConditionNode struct {
	*ConditionNode
	chance float64
}

func NewRandomChanceConditionNode(name string) *RandomChanceConditionNode {
	return &RandomChanceConditionNode{
		ConditionNode: NewConditionNode(name),
		chance:        0.5, // 默认50%概率
	}
}

func (n *RandomChanceConditionNode) Execute(ctx *BossContext) BehaviorResult {
	n.ConditionNode.Execute(ctx) // 调用基类统计

	// 获取概率参数
	if chance := n.GetParamAsFloat64("chance", 0); chance > 0 {
		n.chance = chance
	}

	if rnd.Float64() < n.chance {
		return ResultSuccess
	}

	return ResultFailure
}

// ===================== 缺失的节点实现 =====================

// ResetCombatTargetActionNode 重置战斗目标节点
type ResetCombatTargetActionNode struct {
	*ActionNode
}

func NewResetCombatTargetActionNode(name string) *ResetCombatTargetActionNode {
	return &ResetCombatTargetActionNode{
		ActionNode: NewActionNode(name),
	}
}

func (n *ResetCombatTargetActionNode) Execute(ctx *BossContext) BehaviorResult {
	n.ActionNode.Execute(ctx)
	ctx.Target = nil
	ctx.Boss.SetCombatTarget(nil)
	return ResultSuccess
}

// ComboAttackSequenceNode 连击攻击序列节点
type ComboAttackSequenceNode struct {
	*SequenceNode
	comboCount    int
	maxCombo      int
	lastAttack    time.Time
	comboCooldown time.Duration
}

func NewComboAttackSequenceNode(name string) *ComboAttackSequenceNode {
	return &ComboAttackSequenceNode{
		SequenceNode:  NewSequenceNode(name),
		maxCombo:      3,
		comboCooldown: time.Second * 2,
	}
}

func (n *ComboAttackSequenceNode) Execute(ctx *BossContext) BehaviorResult {
	if ctx.Target == nil || !ctx.Target.IsAlive() {
		return ResultFailure
	}

	// 检查连击冷却
	if ctx.CurrentTime.Sub(n.lastAttack) > n.comboCooldown {
		n.comboCount = 0
	}

	if n.comboCount >= n.maxCombo {
		return ResultFailure
	}

	// 执行攻击
	err := ctx.Boss.DoAttackTarget(ctx.Target)
	if err != nil {
		return ResultFailure
	}

	n.comboCount++
	n.lastAttack = ctx.CurrentTime
	return ResultSuccess
}

// SkillUsageConditionNode 技能使用条件节点
type SkillUsageConditionNode struct {
	*ConditionNode
	healthThreshold float64
	manaThreshold   float64
}

func NewSkillUsageConditionNode(name string) *SkillUsageConditionNode {
	return &SkillUsageConditionNode{
		ConditionNode:   NewConditionNode(name),
		healthThreshold: 0.5,
		manaThreshold:   0.3,
	}
}

func (n *SkillUsageConditionNode) Execute(ctx *BossContext) BehaviorResult {
	n.ConditionNode.Execute(ctx)

	// 检查是否满足使用技能的条件
	currentHP := float64(ctx.Boss.GetCurrentLife()) / float64(ctx.Boss.GetMaxLife())
	if currentHP < n.healthThreshold {
		return ResultSuccess
	}

	return ResultFailure
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
	if currentHP < n.escapeThreshold {
		return ResultSuccess
	}

	return ResultFailure
}

// ReturnToSpawnActionNode 返回出生点节点
type ReturnToSpawnActionNode struct {
	*ActionNode
	spawnPoint Position
}

func NewReturnToSpawnActionNode(name string) *ReturnToSpawnActionNode {
	return &ReturnToSpawnActionNode{
		ActionNode: NewActionNode(name),
		spawnPoint: Position{X: 0, Y: 0, Z: 0},
	}
}

func (n *ReturnToSpawnActionNode) Execute(ctx *BossContext) BehaviorResult {
	n.ActionNode.Execute(ctx)

	// 获取出生点参数
	if spawnX := n.GetParamAsFloat64("spawn_x", 0); spawnX != 0 {
		n.spawnPoint.X = spawnX
	}
	if spawnY := n.GetParamAsFloat64("spawn_y", 0); spawnY != 0 {
		n.spawnPoint.Y = spawnY
	}

	bossPos := ctx.Boss.GetPos()
	distance := calculateDistance(
		float64(bossPos.X), float64(bossPos.Y),
		n.spawnPoint.X, n.spawnPoint.Y,
	)

	if distance < 10.0 {
		return ResultSuccess // 已到达出生点
	}

	err := ctx.Boss.MoveTo(
		coord.Coord(n.spawnPoint.X),
		coord.Coord(n.spawnPoint.Y),
		coord.Coord(n.spawnPoint.Z),
	)
	if err != nil {
		return ResultFailure
	}

	return ResultRunning
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

// EnemyDetectionConditionNode 敌人检测条件节点
type EnemyDetectionConditionNode struct {
	*ConditionNode
	detectionRadius float64
}

func NewEnemyDetectionConditionNode(name string) *EnemyDetectionConditionNode {
	return &EnemyDetectionConditionNode{
		ConditionNode:   NewConditionNode(name),
		detectionRadius: 150.0,
	}
}

func (n *EnemyDetectionConditionNode) Execute(ctx *BossContext) BehaviorResult {
	n.ConditionNode.Execute(ctx)

	// 获取检测半径参数
	if radius := n.GetParamAsFloat64("radius", 0); radius > 0 {
		n.detectionRadius = radius
	}

	// 检测敌人
	enemies := ctx.GetEnemiesInRange(n.detectionRadius)
	if len(enemies) > 0 {
		return ResultSuccess
	}

	return ResultFailure
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

// TargetValidationConditionNode 目标验证条件节点
type TargetValidationConditionNode struct {
	*ConditionNode
}

func NewTargetValidationConditionNode(name string) *TargetValidationConditionNode {
	return &TargetValidationConditionNode{
		ConditionNode: NewConditionNode(name),
	}
}

func (n *TargetValidationConditionNode) Execute(ctx *BossContext) BehaviorResult {
	n.ConditionNode.Execute(ctx)

	// 验证目标是否有效
	if ctx.Target != nil && ctx.Target.IsAlive() {
		return ResultSuccess
	}

	return ResultFailure
}

// LowManaCheckConditionNode 低魔法值检查条件节点
type LowManaCheckConditionNode struct {
	*ConditionNode
	threshold float64
}

func NewLowManaCheckConditionNode(name string) *LowManaCheckConditionNode {
	return &LowManaCheckConditionNode{
		ConditionNode: NewConditionNode(name),
		threshold:     0.3, // 默认30%魔法值
	}
}

func (n *LowManaCheckConditionNode) Execute(ctx *BossContext) BehaviorResult {
	n.ConditionNode.Execute(ctx) // 调用基类统计

	// 获取阈值参数
	if threshold := n.GetParamAsFloat64("threshold", 0); threshold > 0 {
		n.threshold = threshold
	}

	// 这里需要Boss实体提供GetCurrentMana和GetMaxMana方法
	// 暂时使用生命值作为替代示例
	currentPercent := float64(ctx.Boss.GetCurrentLife()) / float64(ctx.Boss.GetMaxLife())

	if currentPercent < n.threshold {
		return ResultSuccess
	}

	return ResultFailure
}
