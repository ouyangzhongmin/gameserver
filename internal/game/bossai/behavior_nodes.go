package bossai

import (
	"fmt"
	"math"
	"time"

	"github.com/ouyangzhongmin/gameserver/pkg/coord"
	"github.com/ouyangzhongmin/gameserver/pkg/logger"
)

// BaseBehaviorNode 基础行为树节点
type BaseBehaviorNode struct {
	name     string
	nodeType BehaviorNodeType
	children []IBehaviorNode
	parent   IBehaviorNode

	// 节点状态
	isRunning    bool
	lastResult   BehaviorResult
	executeCount int64

	// 性能监控
	totalExecuteTime time.Duration
	lastExecuteTime  time.Time

	// 配置参数
	params map[string]interface{}
}

// NewBaseBehaviorNode 创建基础行为节点
func NewBaseBehaviorNode(name string, nodeType BehaviorNodeType) *BaseBehaviorNode {
	return &BaseBehaviorNode{
		name:     name,
		nodeType: nodeType,
		children: make([]IBehaviorNode, 0),
		params:   make(map[string]interface{}),
	}
}

func (n *BaseBehaviorNode) GetName() string {
	return n.name
}

func (n *BaseBehaviorNode) GetType() BehaviorNodeType {
	return n.nodeType
}

func (n *BaseBehaviorNode) AddChild(child IBehaviorNode) error {
	n.children = append(n.children, child)
	return nil
}

func (n *BaseBehaviorNode) GetChildren() []IBehaviorNode {
	return n.children
}

func (n *BaseBehaviorNode) Execute(ctx *BossContext) BehaviorResult {
	n.executeCount++
	n.lastExecuteTime = ctx.CurrentTime
	start := time.Now()

	defer func() {
		n.totalExecuteTime += time.Since(start)
	}()

	// 基础节点不执行任何操作
	return ResultFailure
}

func (n *BaseBehaviorNode) Reset() {
	n.isRunning = false
	n.lastResult = ResultFailure

	// 递归重置子节点
	for _, child := range n.children {
		child.Reset()
	}
}

func (n *BaseBehaviorNode) SetParam(key string, value interface{}) {
	n.params[key] = value
}

func (n *BaseBehaviorNode) GetParam(key string) interface{} {
	return n.params[key]
}

func (n *BaseBehaviorNode) GetParamAsInt(key string, defaultValue int) int {
	if value, exists := n.params[key]; exists {
		if intValue, ok := value.(int); ok {
			return intValue
		}
	}
	return defaultValue
}

func (n *BaseBehaviorNode) GetParamAsFloat64(key string, defaultValue float64) float64 {
	if value, exists := n.params[key]; exists {
		if floatValue, ok := value.(float64); ok {
			return floatValue
		}
	}
	return defaultValue
}

func (n *BaseBehaviorNode) GetParamAsString(key string, defaultValue string) string {
	if value, exists := n.params[key]; exists {
		if strValue, ok := value.(string); ok {
			return strValue
		}
	}
	return defaultValue
}

func (n *BaseBehaviorNode) GetParamAsBool(key string, defaultValue bool) bool {
	if value, exists := n.params[key]; exists {
		if boolValue, ok := value.(bool); ok {
			return boolValue
		}
	}
	return defaultValue
}

// ActionNode 行为节点基类
type ActionNode struct {
	*BaseBehaviorNode
}

func NewActionNode(name string) *ActionNode {
	return &ActionNode{
		BaseBehaviorNode: NewBaseBehaviorNode(name, NodeTypeAction),
	}
}

// ConditionNode 条件节点基类
type ConditionNode struct {
	*BaseBehaviorNode
}

func NewConditionNode(name string) *ConditionNode {
	return &ConditionNode{
		BaseBehaviorNode: NewBaseBehaviorNode(name, NodeTypeCondition),
	}
}

// SequenceNode 顺序节点 - 所有子节点都成功才成功
type SequenceNode struct {
	*BaseBehaviorNode
	currentChildIndex int
}

func NewSequenceNode(name string) *SequenceNode {
	return &SequenceNode{
		BaseBehaviorNode:  NewBaseBehaviorNode(name, NodeTypeSequence),
		currentChildIndex: 0,
	}
}

func (n *SequenceNode) Execute(ctx *BossContext) BehaviorResult {
	n.BaseBehaviorNode.Execute(ctx) // 调用基类方法更新统计

	if len(n.children) == 0 {
		return ResultSuccess
	}

	for n.currentChildIndex < len(n.children) {
		child := n.children[n.currentChildIndex]
		result := child.Execute(ctx)

		switch result {
		case ResultSuccess:
			n.currentChildIndex++
			continue
		case ResultFailure:
			n.Reset()
			return ResultFailure
		case ResultRunning:
			return ResultRunning
		}
	}

	// 所有子节点都成功
	n.Reset()
	return ResultSuccess
}

func (n *SequenceNode) Reset() {
	n.currentChildIndex = 0
	n.BaseBehaviorNode.Reset()
}

// SelectorNode 选择节点 - 任一子节点成功就成功
type SelectorNode struct {
	*BaseBehaviorNode
	currentChildIndex int
}

func NewSelectorNode(name string) *SelectorNode {
	return &SelectorNode{
		BaseBehaviorNode:  NewBaseBehaviorNode(name, NodeTypeSelector),
		currentChildIndex: 0,
	}
}

func (n *SelectorNode) Execute(ctx *BossContext) BehaviorResult {
	n.BaseBehaviorNode.Execute(ctx)

	if len(n.children) == 0 {
		return ResultFailure
	}

	for n.currentChildIndex < len(n.children) {
		child := n.children[n.currentChildIndex]
		result := child.Execute(ctx)

		switch result {
		case ResultSuccess:
			n.Reset()
			return ResultSuccess
		case ResultFailure:
			n.currentChildIndex++
			continue
		case ResultRunning:
			return ResultRunning
		}
	}

	// 所有子节点都失败
	n.Reset()
	return ResultFailure
}

func (n *SelectorNode) Reset() {
	n.currentChildIndex = 0
	n.BaseBehaviorNode.Reset()
}

// ParallelNode 并行节点 - 同时执行所有子节点
type ParallelNode struct {
	*BaseBehaviorNode
	successThreshold int
	failureThreshold int
	successCount     int
	failureCount     int
}

func NewParallelNode(name string, successThreshold, failureThreshold int) *ParallelNode {
	return &ParallelNode{
		BaseBehaviorNode: NewBaseBehaviorNode(name, NodeTypeParallel),
		successThreshold: successThreshold,
		failureThreshold: failureThreshold,
	}
}

func (n *ParallelNode) Execute(ctx *BossContext) BehaviorResult {
	n.BaseBehaviorNode.Execute(ctx)

	if len(n.children) == 0 {
		return ResultSuccess
	}

	n.successCount = 0
	n.failureCount = 0
	runningCount := 0

	for _, child := range n.children {
		result := child.Execute(ctx)

		switch result {
		case ResultSuccess:
			n.successCount++
		case ResultFailure:
			n.failureCount++
		case ResultRunning:
			runningCount++
		}
	}

	// 检查成功条件
	if n.successCount >= n.successThreshold {
		n.Reset()
		return ResultSuccess
	}

	// 检查失败条件
	if n.failureCount >= n.failureThreshold {
		n.Reset()
		return ResultFailure
	}

	// 还有节点在运行
	if runningCount > 0 {
		return ResultRunning
	}

	// 既没达到成功条件也没达到失败条件
	return ResultRunning
}

func (n *ParallelNode) Reset() {
	n.successCount = 0
	n.failureCount = 0
	n.BaseBehaviorNode.Reset()
}

// DecoratorNode 装饰节点基类
type DecoratorNode struct {
	*BaseBehaviorNode
}

func NewDecoratorNode(name string) *DecoratorNode {
	return &DecoratorNode{
		BaseBehaviorNode: NewBaseBehaviorNode(name, NodeTypeDecorator),
	}
}

func (n *DecoratorNode) AddChild(child IBehaviorNode) error {
	if len(n.children) >= 1 {
		return fmt.Errorf("decorator node can only have one child")
	}
	return n.BaseBehaviorNode.AddChild(child)
}

// InverterNode 反转节点 - 反转子节点的结果
type InverterNode struct {
	*DecoratorNode
}

func NewInverterNode(name string) *InverterNode {
	return &InverterNode{
		DecoratorNode: NewDecoratorNode(name),
	}
}

func (n *InverterNode) Execute(ctx *BossContext) BehaviorResult {
	n.BaseBehaviorNode.Execute(ctx)

	if len(n.children) == 0 {
		return ResultFailure
	}

	result := n.children[0].Execute(ctx)

	switch result {
	case ResultSuccess:
		return ResultFailure
	case ResultFailure:
		return ResultSuccess
	case ResultRunning:
		return ResultRunning
	}

	return ResultFailure
}

// RepeaterNode 重复节点 - 重复执行子节点
type RepeaterNode struct {
	*DecoratorNode
	maxRepeats     int
	currentRepeats int
}

func NewRepeaterNode(name string, maxRepeats int) *RepeaterNode {
	return &RepeaterNode{
		DecoratorNode: NewDecoratorNode(name),
		maxRepeats:    maxRepeats,
	}
}

func (n *RepeaterNode) Execute(ctx *BossContext) BehaviorResult {
	n.BaseBehaviorNode.Execute(ctx)

	if len(n.children) == 0 {
		return ResultFailure
	}

	for n.currentRepeats < n.maxRepeats || n.maxRepeats <= 0 {
		result := n.children[0].Execute(ctx)

		switch result {
		case ResultSuccess:
			n.currentRepeats++
			n.children[0].Reset()
			if n.maxRepeats > 0 && n.currentRepeats >= n.maxRepeats {
				n.Reset()
				return ResultSuccess
			}
			continue
		case ResultFailure:
			n.Reset()
			return ResultFailure
		case ResultRunning:
			return ResultRunning
		}
	}

	n.Reset()
	return ResultSuccess
}

func (n *RepeaterNode) Reset() {
	n.currentRepeats = 0
	n.DecoratorNode.Reset()
}

// UntilFailNode 直到失败节点 - 重复执行直到失败
type UntilFailNode struct {
	*DecoratorNode
}

func NewUntilFailNode(name string) *UntilFailNode {
	return &UntilFailNode{
		DecoratorNode: NewDecoratorNode(name),
	}
}

func (n *UntilFailNode) Execute(ctx *BossContext) BehaviorResult {
	n.BaseBehaviorNode.Execute(ctx)

	if len(n.children) == 0 {
		return ResultFailure
	}

	result := n.children[0].Execute(ctx)

	switch result {
	case ResultSuccess:
		n.children[0].Reset()
		return ResultRunning // 继续执行
	case ResultFailure:
		return ResultSuccess // 失败了就成功
	case ResultRunning:
		return ResultRunning
	}

	return ResultFailure
}

// CooldownNode 冷却节点 - 在冷却时间内阻止执行
type CooldownNode struct {
	*DecoratorNode
	cooldownTime    time.Duration
	lastExecuteTime time.Time
}

func NewCooldownNode(name string, cooldownTime time.Duration) *CooldownNode {
	return &CooldownNode{
		DecoratorNode: NewDecoratorNode(name),
		cooldownTime:  cooldownTime,
	}
}

func (n *CooldownNode) Execute(ctx *BossContext) BehaviorResult {
	n.BaseBehaviorNode.Execute(ctx)

	if len(n.children) == 0 {
		return ResultFailure
	}

	// 检查冷却时间
	if ctx.CurrentTime.Sub(n.lastExecuteTime) < n.cooldownTime {
		return ResultFailure
	}

	result := n.children[0].Execute(ctx)

	if result == ResultSuccess || result == ResultFailure {
		n.lastExecuteTime = ctx.CurrentTime
	}

	return result
}

// 具体的行为节点实现

// AttackTargetNode 攻击目标节点
type AttackTargetNode struct {
	*ActionNode
}

func NewAttackTargetNode() *AttackTargetNode {
	return &AttackTargetNode{
		ActionNode: NewActionNode("AttackTarget"),
	}
}

func (n *AttackTargetNode) Execute(ctx *BossContext) BehaviorResult {
	n.BaseBehaviorNode.Execute(ctx)

	if ctx.Target == nil || !ctx.Target.IsAlive() {
		return ResultFailure
	}

	// 检查攻击范围
	if !ctx.Boss.IsInAttackRange(ctx.Target.GetPos().X, ctx.Target.GetPos().Y) {
		return ResultFailure
	}

	// 执行攻击逻辑 - 直接调用Boss接口的DoAttackTarget方法
	logger.Debugf("Boss %d attacking target %d", ctx.Boss.GetID(), ctx.Target.GetID())

	// 直接调用IBossEntity接口的DoAttackTarget方法
	err := ctx.Boss.DoAttackTarget(ctx.Target)
	if err != nil {
		logger.Errorf("Boss %d failed to attack target: %v", ctx.Boss.GetID(), err)
		// 如果DoAttackTarget失败，回退到基础攻击状态控制
		ctx.Boss.AttackAction()
		// 注意：具体的伤害计算和处理由Monster系统内部的doAttackTarget方法完成
		// 这里只设置攻击状态，不直接处理伤害
		logger.Debugf("Boss %d executed fallback attack action", ctx.Boss.GetID())
	}

	// 记录攻击动作
	if ctx.ActionHistory == nil {
		ctx.ActionHistory = make([]*AIAction, 0)
	}
	ctx.ActionHistory = append(ctx.ActionHistory, &AIAction{
		Type:     ActionAttack,
		Target:   ctx.Target,
		Executed: true,
	})

	return ResultSuccess
}

// MoveToTargetNode 移动到目标节点
type MoveToTargetNode struct {
	*ActionNode
	minDistance float64
}

func NewMoveToTargetNode(minDistance float64) *MoveToTargetNode {
	return &MoveToTargetNode{
		ActionNode:  NewActionNode("MoveToTarget"),
		minDistance: minDistance,
	}
}

func (n *MoveToTargetNode) Execute(ctx *BossContext) BehaviorResult {
	n.BaseBehaviorNode.Execute(ctx)

	if ctx.Target == nil {
		return ResultFailure
	}

	targetPos := ctx.Target.GetPos()
	currentPos := ctx.Boss.GetPos()

	// 计算距离
	distance := calculateDistance(
		float64(currentPos.X), float64(currentPos.Y),
		float64(targetPos.X), float64(targetPos.Y),
	)

	// 如果已经足够接近
	if distance <= n.minDistance {
		return ResultSuccess
	}

	// 移动到目标
	err := ctx.Boss.MoveTo(targetPos.X, targetPos.Y, targetPos.Z)
	if err != nil {
		logger.Errorf("Boss %d failed to move to target: %v", ctx.Boss.GetID(), err)
		return ResultFailure
	}

	return ResultRunning
}

// HasTargetCondition 有目标条件节点
type HasTargetCondition struct {
	*ConditionNode
}

func NewHasTargetCondition() *HasTargetCondition {
	return &HasTargetCondition{
		ConditionNode: NewConditionNode("HasTarget"),
	}
}

func (n *HasTargetCondition) Execute(ctx *BossContext) BehaviorResult {
	n.BaseBehaviorNode.Execute(ctx)

	if ctx.Target != nil && ctx.Target.IsAlive() {
		return ResultSuccess
	}

	return ResultFailure
}

// HealthPercentCondition 血量百分比条件节点
type HealthPercentCondition struct {
	*ConditionNode
	threshold float64
	operator  string // "less_than", "greater_than", "equal"
}

func NewHealthPercentCondition(threshold float64, operator string) *HealthPercentCondition {
	return &HealthPercentCondition{
		ConditionNode: NewConditionNode("HealthPercent"),
		threshold:     threshold,
		operator:      operator,
	}
}

func (n *HealthPercentCondition) Execute(ctx *BossContext) BehaviorResult {
	n.BaseBehaviorNode.Execute(ctx)

	currentPercent := float64(ctx.Boss.GetCurrentLife()) / float64(ctx.Boss.GetMaxLife())

	switch n.operator {
	case "less_than":
		if currentPercent < n.threshold {
			return ResultSuccess
		}
	case "greater_than":
		if currentPercent > n.threshold {
			return ResultSuccess
		}
	case "equal":
		if currentPercent == n.threshold {
			return ResultSuccess
		}
	}

	return ResultFailure
}

// FindTargetNode 寻找目标节点
type FindTargetNode struct {
	*ActionNode
	searchRadius float64
}

func NewFindTargetNode(searchRadius float64) *FindTargetNode {
	return &FindTargetNode{
		ActionNode:   NewActionNode("FindTarget"),
		searchRadius: searchRadius,
	}
}

func (n *FindTargetNode) Execute(ctx *BossContext) BehaviorResult {
	n.BaseBehaviorNode.Execute(ctx)

	// 使用BossContext的方法从已缓存的NearbyEnemies中查找
	nearest := ctx.GetNearestEnemyInRange(n.searchRadius)
	if nearest != nil {
		ctx.Target = nearest
		ctx.Boss.SetCombatTarget(nearest)
		return ResultSuccess
	}

	return ResultFailure
}

// calculateDistance 计算两点间距离
func calculateDistance(x1, y1, x2, y2 float64) float64 {
	dx := x2 - x1
	dy := y2 - y1
	return math.Sqrt(dx*dx + dy*dy)
}

// 扩展的行为节点实现

// UseSkillActionNode 使用技能行为节点
type UseSkillActionNode struct {
	*ActionNode
	skillID int32
}

func NewUseSkillActionNode(name string, skillID int32) *UseSkillActionNode {
	return &UseSkillActionNode{
		ActionNode: NewActionNode(name),
		skillID:    skillID,
	}
}

func (n *UseSkillActionNode) Execute(ctx *BossContext) BehaviorResult {
	n.BaseBehaviorNode.Execute(ctx)

	// 检查目标
	if ctx.Target == nil {
		return ResultFailure
	}

	// 获取技能管理器并使用技能
	if skillManager := ctx.GetSkillManager(); skillManager != nil {
		skill, err := skillManager.GetSkill(n.skillID)
		if err != nil {
			logger.Errorf("Skill %d not found: %v", n.skillID, err)
			return ResultFailure
		}

		// 检查技能是否可用（冷却、范围等）
		if !skill.IsAvailable(ctx) {
			logger.Debugf("Skill %d is not available", n.skillID)
			return ResultFailure
		}

		// 释放技能
		targets := skill.GetTargets(ctx)
		err = skill.Cast(ctx, targets)
		if err != nil {
			logger.Errorf("Failed to cast skill %d: %v", n.skillID, err)
			return ResultFailure
		}

		logger.Debugf("Boss %d successfully used skill %d on %d targets", ctx.Boss.GetID(), n.skillID, len(targets))

		// 记录技能使用
		if ctx.SkillsUsed == nil {
			ctx.SkillsUsed = make(map[int32]int)
		}
		ctx.SkillsUsed[n.skillID]++

		return ResultSuccess
	}

	// 如果没有技能管理器，使用基础攻击
	logger.Debugf("No skill manager available, using basic attack for skill %d", n.skillID)
	return NewAttackTargetNode().Execute(ctx)
}

// AttackActionNode 攻击行为节点
type AttackActionNode struct {
	*ActionNode
}

func NewAttackActionNode(name string) *AttackActionNode {
	return &AttackActionNode{
		ActionNode: NewActionNode(name),
	}
}

func (n *AttackActionNode) Execute(ctx *BossContext) BehaviorResult {
	return NewAttackTargetNode().Execute(ctx)
}

// MoveToActionNode 移动到指定位置行为节点
type MoveToActionNode struct {
	*ActionNode
	x, y, z float64
}

func NewMoveToActionNode(name string, x, y, z float64) *MoveToActionNode {
	return &MoveToActionNode{
		ActionNode: NewActionNode(name),
		x:          x,
		y:          y,
		z:          z,
	}
}

func (n *MoveToActionNode) Execute(ctx *BossContext) BehaviorResult {
	n.BaseBehaviorNode.Execute(ctx)

	err := ctx.Boss.MoveTo(coord.Coord(n.x), coord.Coord(n.y), coord.Coord(n.z))
	if err != nil {
		logger.Errorf("Boss %d failed to move to position (%.2f, %.2f, %.2f): %v", ctx.Boss.GetID(), n.x, n.y, n.z, err)
		return ResultFailure
	}

	return ResultRunning
}

// CastSkillActionNode 释放技能行为节点
type CastSkillActionNode struct {
	*ActionNode
	skillID int32
}

func NewCastSkillActionNode(name string, skillID int32) *CastSkillActionNode {
	return &CastSkillActionNode{
		ActionNode: NewActionNode(name),
		skillID:    skillID,
	}
}

func (n *CastSkillActionNode) Execute(ctx *BossContext) BehaviorResult {
	return NewUseSkillActionNode(n.name, n.skillID).Execute(ctx)
}

// RetreatActionNode 撤退行为节点
type RetreatActionNode struct {
	*ActionNode
}

func NewRetreatActionNode(name string) *RetreatActionNode {
	return &RetreatActionNode{
		ActionNode: NewActionNode(name),
	}
}

func (n *RetreatActionNode) Execute(ctx *BossContext) BehaviorResult {
	n.BaseBehaviorNode.Execute(ctx)

	// 实现撤退逻辑
	logger.Debugf("Boss %d initiating retreat", ctx.Boss.GetID())

	// 设置撤退状态
	ctx.Boss.Escape()

	// 清除当前目标
	ctx.Target = nil
	ctx.Boss.SetCombatTarget(nil)

	// 获取出生点位置
	spawnPos := ctx.Boss.GetSpawnPosition()
	currentPos := ctx.Boss.GetPos()

	// 计算到出生点的距离
	distance := calculateDistance(
		float64(currentPos.X), float64(currentPos.Y),
		float64(spawnPos.X), float64(spawnPos.Y),
	)

	// 如果已经在出生点附近，返回成功
	if distance < 10.0 {
		logger.Debugf("Boss %d successfully retreated to spawn point", ctx.Boss.GetID())
		return ResultSuccess
	}

	// 移动到出生点
	err := ctx.Boss.MoveTo(
		coord.Coord(spawnPos.X),
		coord.Coord(spawnPos.Y),
		coord.Coord(spawnPos.Z),
	)
	if err != nil {
		logger.Errorf("Boss %d failed to retreat to spawn point: %v", ctx.Boss.GetID(), err)
		return ResultFailure
	}

	// 记录撤退动作
	if ctx.ActionHistory == nil {
		ctx.ActionHistory = make([]*AIAction, 0)
	}
	ctx.ActionHistory = append(ctx.ActionHistory, &AIAction{
		Type:     ActionRetreat,
		Position: &spawnPos,
		Executed: true,
	})

	return ResultRunning
}

// PatrolActionNode 巡逻行为节点
type PatrolActionNode struct {
	*ActionNode
}

func NewPatrolActionNode(name string) *PatrolActionNode {
	return &PatrolActionNode{
		ActionNode: NewActionNode(name),
	}
}

func (n *PatrolActionNode) Execute(ctx *BossContext) BehaviorResult {
	n.BaseBehaviorNode.Execute(ctx)

	// 实现巡逻逻辑
	logger.Debugf("Boss %d starting patrol", ctx.Boss.GetID())

	// 设置巡逻状态
	ctx.Boss.Walk()

	// 获取巡逻点列表
	patrolPoints := ctx.Boss.GetPatrolPoints()
	if len(patrolPoints) == 0 {
		// 如果没有设置巡逻点，随机移动
		spawnPos := ctx.Boss.GetSpawnPosition()

		// 在出生点周围随机移动
		radius := 50.0
		angle := float64(ctx.CurrentTime.UnixNano()%360) * math.Pi / 180
		targetX := float64(spawnPos.X) + radius*math.Cos(angle)
		targetY := float64(spawnPos.Y) + radius*math.Sin(angle)

		err := ctx.Boss.MoveTo(
			coord.Coord(targetX),
			coord.Coord(targetY),
			spawnPos.Z,
		)
		if err != nil {
			logger.Errorf("Boss %d failed to patrol: %v", ctx.Boss.GetID(), err)
			return ResultFailure
		}

		logger.Debugf("Boss %d patrolling randomly around spawn point", ctx.Boss.GetID())
		return ResultRunning
	}

	// 使用巡逻点进行巡逻
	currentPatrolIndex := ctx.GetPatrolIndex()
	if currentPatrolIndex >= len(patrolPoints) {
		currentPatrolIndex = 0
		ctx.SetPatrolIndex(0)
	}

	targetPoint := patrolPoints[currentPatrolIndex]
	currentPos := ctx.Boss.GetPos()

	// 检查是否已到达当前巡逻点
	distance := calculateDistance(
		float64(currentPos.X), float64(currentPos.Y),
		float64(targetPoint.X), float64(targetPoint.Y),
	)

	if distance < 5.0 {
		// 到达当前点，移动到下一个巡逻点
		nextIndex := (currentPatrolIndex + 1) % len(patrolPoints)
		ctx.SetPatrolIndex(nextIndex)
		logger.Debugf("Boss %d reached patrol point %d, moving to next point %d", ctx.Boss.GetID(), currentPatrolIndex, nextIndex)
		return ResultSuccess
	}

	// 移动到当前巡逻点
	err := ctx.Boss.MoveTo(
		coord.Coord(targetPoint.X),
		coord.Coord(targetPoint.Y),
		coord.Coord(targetPoint.Z),
	)
	if err != nil {
		logger.Errorf("Boss %d failed to move to patrol point %d: %v", ctx.Boss.GetID(), currentPatrolIndex, err)
		return ResultFailure
	}

	logger.Debugf("Boss %d patrolling to point %d", ctx.Boss.GetID(), currentPatrolIndex)
	return ResultRunning
}

// WaitActionNode 等待行为节点
type WaitActionNode struct {
	*ActionNode
	duration  time.Duration
	startTime time.Time
	isWaiting bool
}

func NewWaitActionNode(name string, duration time.Duration) *WaitActionNode {
	return &WaitActionNode{
		ActionNode: NewActionNode(name),
		duration:   duration,
	}
}

func (n *WaitActionNode) Execute(ctx *BossContext) BehaviorResult {
	n.BaseBehaviorNode.Execute(ctx)

	if !n.isWaiting {
		n.startTime = ctx.CurrentTime
		n.isWaiting = true
		return ResultRunning
	}

	if ctx.CurrentTime.Sub(n.startTime) >= n.duration {
		n.isWaiting = false
		return ResultSuccess
	}

	return ResultRunning
}

func (n *WaitActionNode) Reset() {
	n.isWaiting = false
	n.ActionNode.Reset()
}

// GenericActionNode 通用行为节点
type GenericActionNode struct {
	*ActionNode
	params map[string]interface{}
}

func NewGenericActionNode(name string, params map[string]interface{}) *GenericActionNode {
	return &GenericActionNode{
		ActionNode: NewActionNode(name),
		params:     params,
	}
}

func (n *GenericActionNode) Execute(ctx *BossContext) BehaviorResult {
	n.BaseBehaviorNode.Execute(ctx)

	logger.Debugf("Executing generic action: %s", n.name)
	return ResultSuccess
}

// 扩展的条件节点实现

// HealthConditionNode 血量条件节点
type HealthConditionNode struct {
	*ConditionNode
	threshold float64
}

func NewHealthConditionNode(name string, threshold float64) *HealthConditionNode {
	return &HealthConditionNode{
		ConditionNode: NewConditionNode(name),
		threshold:     threshold,
	}
}

func (n *HealthConditionNode) Execute(ctx *BossContext) BehaviorResult {
	n.BaseBehaviorNode.Execute(ctx)

	currentPercent := float64(ctx.Boss.GetCurrentLife()) / float64(ctx.Boss.GetMaxLife())
	if currentPercent < n.threshold {
		return ResultSuccess
	}
	return ResultFailure
}

// EnemyInRangeConditionNode 敌人在范围内条件节点
type EnemyInRangeConditionNode struct {
	*ConditionNode
	rangeValue float64
}

func NewEnemyInRangeConditionNode(name string, rangeValue float64) *EnemyInRangeConditionNode {
	return &EnemyInRangeConditionNode{
		ConditionNode: NewConditionNode(name),
		rangeValue:    rangeValue,
	}
}

func (n *EnemyInRangeConditionNode) Execute(ctx *BossContext) BehaviorResult {
	n.BaseBehaviorNode.Execute(ctx)

	// 使用BossContext的方法从已缓存的NearbyEnemies中查找
	hasEnemies := ctx.HasEnemiesInRange(n.rangeValue)
	if hasEnemies {
		return ResultSuccess
	}
	return ResultFailure
}

// SkillAvailableConditionNode 技能可用条件节点
type SkillAvailableConditionNode struct {
	*ConditionNode
	skillID int32
}

func NewSkillAvailableConditionNode(name string, skillID int32) *SkillAvailableConditionNode {
	return &SkillAvailableConditionNode{
		ConditionNode: NewConditionNode(name),
		skillID:       skillID,
	}
}

func (n *SkillAvailableConditionNode) Execute(ctx *BossContext) BehaviorResult {
	n.BaseBehaviorNode.Execute(ctx)

	// 检查技能冷却、范围和其他条件
	if skillManager := ctx.GetSkillManager(); skillManager != nil {
		skill, err := skillManager.GetSkill(n.skillID)
		if err != nil {
			logger.Debugf("Skill %d not found", n.skillID)
			return ResultFailure
		}

		// 检查技能是否在冷却中
		if !skill.IsOffCooldown() {
			logger.Debugf("Skill %d is on cooldown", n.skillID)
			return ResultFailure
		}

		// 检查技能范围
		if ctx.Target != nil {
			bossPos := ctx.Boss.GetPos()
			targetPos := ctx.Target.GetPos()
			distance := calculateDistance(
				float64(bossPos.X), float64(bossPos.Y),
				float64(targetPos.X), float64(targetPos.Y),
			)

			if distance > skill.GetRange() {
				logger.Debugf("Target is out of range for skill %d (distance: %.2f, range: %.2f)",
					n.skillID, distance, skill.GetRange())
				return ResultFailure
			}
		}

		// 检查技能条件（如需要目标数量、血量限制等）
		if !skill.CheckConditions(ctx) {
			logger.Debugf("Skill %d conditions not met", n.skillID)
			return ResultFailure
		}

		logger.Debugf("Skill %d is available for use", n.skillID)
		return ResultSuccess
	}

	// 如果没有技能管理器，默认可用
	logger.Debugf("No skill manager available, assuming skill %d is available", n.skillID)
	return ResultSuccess
}

// CombatStateConditionNode 战斗状态条件节点
type CombatStateConditionNode struct {
	*ConditionNode
}

func NewCombatStateConditionNode(name string) *CombatStateConditionNode {
	return &CombatStateConditionNode{
		ConditionNode: NewConditionNode(name),
	}
}

func (n *CombatStateConditionNode) Execute(ctx *BossContext) BehaviorResult {
	n.BaseBehaviorNode.Execute(ctx)

	if ctx.Boss.IsInCombat() {
		return ResultSuccess
	}
	return ResultFailure
}

// DistanceConditionNode 距离条件节点
type DistanceConditionNode struct {
	*ConditionNode
	distance float64
	operator string
}

func NewDistanceConditionNode(name string, distance float64, operator string) *DistanceConditionNode {
	return &DistanceConditionNode{
		ConditionNode: NewConditionNode(name),
		distance:      distance,
		operator:      operator,
	}
}

func (n *DistanceConditionNode) Execute(ctx *BossContext) BehaviorResult {
	n.BaseBehaviorNode.Execute(ctx)

	if ctx.Target == nil {
		return ResultFailure
	}

	bossPos := ctx.Boss.GetPos()
	targetPos := ctx.Target.GetPos()
	dist := calculateDistance(float64(bossPos.X), float64(bossPos.Y), float64(targetPos.X), float64(targetPos.Y))

	switch n.operator {
	case "less_than":
		if dist < n.distance {
			return ResultSuccess
		}
	case "greater_than":
		if dist > n.distance {
			return ResultSuccess
		}
	case "equal":
		if dist == n.distance {
			return ResultSuccess
		}
	}

	return ResultFailure
}

// PhaseConditionNode 阶段条件节点
type PhaseConditionNode struct {
	*ConditionNode
	phaseID int32
}

func NewPhaseConditionNode(name string, phaseID int32) *PhaseConditionNode {
	return &PhaseConditionNode{
		ConditionNode: NewConditionNode(name),
		phaseID:       phaseID,
	}
}

func (n *PhaseConditionNode) Execute(ctx *BossContext) BehaviorResult {
	n.BaseBehaviorNode.Execute(ctx)

	if ctx.CurrentPhase != nil && ctx.CurrentPhase.GetID() == n.phaseID {
		return ResultSuccess
	}
	return ResultFailure
}

// GenericConditionNode 通用条件节点
type GenericConditionNode struct {
	*ConditionNode
	params map[string]interface{}
}

func NewGenericConditionNode(name string, params map[string]interface{}) *GenericConditionNode {
	return &GenericConditionNode{
		ConditionNode: NewConditionNode(name),
		params:        params,
	}
}

func (n *GenericConditionNode) Execute(ctx *BossContext) BehaviorResult {
	n.BaseBehaviorNode.Execute(ctx)

	logger.Debugf("Executing generic condition: %s", n.name)
	return ResultSuccess // 简化处理，默认成功
}

// 扩展的装饰节点实现

// RetryNode 重试装饰节点
type RetryNode struct {
	*DecoratorNode
	maxRetries     int
	currentRetries int
}

func NewRetryNode(name string, child IBehaviorNode, maxRetries int) *RetryNode {
	node := &RetryNode{
		DecoratorNode: NewDecoratorNode(name),
		maxRetries:    maxRetries,
	}
	node.AddChild(child)
	return node
}

func (n *RetryNode) Execute(ctx *BossContext) BehaviorResult {
	n.BaseBehaviorNode.Execute(ctx)

	if len(n.children) == 0 {
		return ResultFailure
	}

	for n.currentRetries < n.maxRetries {
		result := n.children[0].Execute(ctx)

		switch result {
		case ResultSuccess:
			n.Reset()
			return ResultSuccess
		case ResultFailure:
			n.currentRetries++
			n.children[0].Reset()
			continue
		case ResultRunning:
			return ResultRunning
		}
	}

	n.Reset()
	return ResultFailure
}

func (n *RetryNode) Reset() {
	n.currentRetries = 0
	n.DecoratorNode.Reset()
}

// TimeoutNode 超时装饰节点
type TimeoutNode struct {
	*DecoratorNode
	timeout   time.Duration
	startTime time.Time
	isRunning bool
}

func NewTimeoutNode(name string, child IBehaviorNode, timeout time.Duration) *TimeoutNode {
	node := &TimeoutNode{
		DecoratorNode: NewDecoratorNode(name),
		timeout:       timeout,
	}
	node.AddChild(child)
	return node
}

func (n *TimeoutNode) Execute(ctx *BossContext) BehaviorResult {
	n.BaseBehaviorNode.Execute(ctx)

	if len(n.children) == 0 {
		return ResultFailure
	}

	if !n.isRunning {
		n.startTime = ctx.CurrentTime
		n.isRunning = true
	}

	// 检查超时
	if ctx.CurrentTime.Sub(n.startTime) >= n.timeout {
		n.Reset()
		return ResultFailure
	}

	result := n.children[0].Execute(ctx)
	if result != ResultRunning {
		n.Reset()
	}

	return result
}

func (n *TimeoutNode) Reset() {
	n.isRunning = false
	n.DecoratorNode.Reset()
}

// GenericDecoratorNode 通用装饰节点
type GenericDecoratorNode struct {
	*DecoratorNode
	params map[string]interface{}
}

func NewGenericDecoratorNode(name string, child IBehaviorNode, params map[string]interface{}) *GenericDecoratorNode {
	node := &GenericDecoratorNode{
		DecoratorNode: NewDecoratorNode(name),
		params:        params,
	}
	node.AddChild(child)
	return node
}

func (n *GenericDecoratorNode) Execute(ctx *BossContext) BehaviorResult {
	n.BaseBehaviorNode.Execute(ctx)

	if len(n.children) == 0 {
		return ResultFailure
	}

	logger.Debugf("Executing generic decorator: %s", n.name)
	return n.children[0].Execute(ctx)
}

// CheckPhaseTriggersActionNode 检查阶段触发器行为节点
type CheckPhaseTriggersActionNode struct {
	*ActionNode
}

func NewCheckPhaseTriggersActionNode(name string) *CheckPhaseTriggersActionNode {
	return &CheckPhaseTriggersActionNode{
		ActionNode: NewActionNode(name),
	}
}

func (n *CheckPhaseTriggersActionNode) Execute(ctx *BossContext) BehaviorResult {
	n.BaseBehaviorNode.Execute(ctx)

	// 检查阶段触发条件
	if phaseManager := ctx.GetPhaseManager(); phaseManager != nil {
		// 更新阶段管理器状态
		err := phaseManager.Update(ctx, 0)
		if err != nil {
			logger.Errorf("Failed to update phase manager: %v", err)
			return ResultFailure
		}
		return ResultSuccess
	}

	return ResultFailure
}

// UpdatePhaseStateActionNode 更新阶段状态行为节点
type UpdatePhaseStateActionNode struct {
	*ActionNode
}

func NewUpdatePhaseStateActionNode(name string) *UpdatePhaseStateActionNode {
	return &UpdatePhaseStateActionNode{
		ActionNode: NewActionNode(name),
	}
}

func (n *UpdatePhaseStateActionNode) Execute(ctx *BossContext) BehaviorResult {
	n.BaseBehaviorNode.Execute(ctx)

	// 更新当前阶段状态
	if ctx.CurrentPhase != nil {
		logger.Debugf("Boss %d updating phase state: %s", ctx.Boss.GetID(), ctx.CurrentPhase.GetName())
		return ResultSuccess
	}

	return ResultFailure
}

// TargetValidationActionNode 目标验证行为节点
type TargetValidationActionNode struct {
	*ActionNode
}

func NewTargetValidationActionNode(name string) *TargetValidationActionNode {
	return &TargetValidationActionNode{
		ActionNode: NewActionNode(name),
	}
}

func (n *TargetValidationActionNode) Execute(ctx *BossContext) BehaviorResult {
	n.BaseBehaviorNode.Execute(ctx)

	// 验证当前目标是否有效
	if ctx.Target == nil || !ctx.Target.IsAlive() {
		// 目标无效，清除目标
		ctx.Target = nil
		ctx.Boss.SetCombatTarget(nil)
		logger.Debugf("Boss %d target validation failed, clearing target", ctx.Boss.GetID())
		return ResultFailure
	}

	// 检查目标是否在有效范围内
	bossPos := ctx.Boss.GetPos()
	targetPos := ctx.Target.GetPos()
	distance := calculateDistance(
		float64(bossPos.X), float64(bossPos.Y),
		float64(targetPos.X), float64(targetPos.Y),
	)

	// 如果目标太远，可能需要放弃追击
	maxRange := 500.0 // 可配置的最大追击范围
	if distance > maxRange {
		logger.Debugf("Boss %d target too far (%.2f > %.2f), abandoning target", ctx.Boss.GetID(), distance, maxRange)
		ctx.Target = nil
		ctx.Boss.SetCombatTarget(nil)
		return ResultFailure
	}

	logger.Debugf("Boss %d target validation passed", ctx.Boss.GetID())
	return ResultSuccess
}

// PatrolBehaviorActionNode 巡逻行为节点
type PatrolBehaviorActionNode struct {
	*ActionNode
}

func NewPatrolBehaviorActionNode(name string) *PatrolBehaviorActionNode {
	return &PatrolBehaviorActionNode{
		ActionNode: NewActionNode(name),
	}
}

func (n *PatrolBehaviorActionNode) Execute(ctx *BossContext) BehaviorResult {
	// 直接调用已有的巡逻节点实现
	return NewPatrolActionNode(n.name).Execute(ctx)
}

// EmergencyResponseConditionNode 紧急响应条件节点
type EmergencyResponseConditionNode struct {
	*ConditionNode
}

func NewEmergencyResponseConditionNode(name string) *EmergencyResponseConditionNode {
	return &EmergencyResponseConditionNode{
		ConditionNode: NewConditionNode(name),
	}
}

func (n *EmergencyResponseConditionNode) Execute(ctx *BossContext) BehaviorResult {
	n.BaseBehaviorNode.Execute(ctx)

	// 检查紧急情况
	// 1. 低血量检查
	healthPercent := float64(ctx.Boss.GetCurrentLife()) / float64(ctx.Boss.GetMaxLife())
	if healthPercent < 0.2 { // 20%以下血量
		logger.Debugf("Boss %d emergency: low health (%.1f%%)", ctx.Boss.GetID(), healthPercent*100)
		return ResultSuccess
	}

	// 2. 被多个敌人围攻
	enemiesNearby := ctx.GetEnemiesInRange(100.0)
	if len(enemiesNearby) >= 3 {
		logger.Debugf("Boss %d emergency: surrounded by %d enemies", ctx.Boss.GetID(), len(enemiesNearby))
		return ResultSuccess
	}

	// 3. 持续受到大量伤害
	if ctx.DamageReceived > int32(ctx.Boss.GetMaxLife()/4) { // 受到超过25%最大血量的伤害
		logger.Debugf("Boss %d emergency: heavy damage received (%d)", ctx.Boss.GetID(), ctx.DamageReceived)
		return ResultSuccess
	}

	return ResultFailure
}

// CombatActionsConditionNode 战斗动作条件节点
type CombatActionsConditionNode struct {
	*ConditionNode
}

func NewCombatActionsConditionNode(name string) *CombatActionsConditionNode {
	return &CombatActionsConditionNode{
		ConditionNode: NewConditionNode(name),
	}
}

func (n *CombatActionsConditionNode) Execute(ctx *BossContext) BehaviorResult {
	n.BaseBehaviorNode.Execute(ctx)

	// 检查是否可以执行战斗动作
	// 1. 有有效目标
	if ctx.Target == nil || !ctx.Target.IsAlive() {
		return ResultFailure
	}

	// 2. 在战斗状态
	if !ctx.Boss.IsInCombat() {
		return ResultFailure
	}

	// 3. 不在特殊状态（如眩晕、死亡等）
	if ctx.CurrentState != nil {
		stateID := ctx.CurrentState.GetID()
		if stateID == int32(StateStunned) || stateID == int32(StateDying) {
			return ResultFailure
		}
	}

	logger.Debugf("Boss %d combat actions available", ctx.Boss.GetID())
	return ResultSuccess
}

// SelectOptimalSkillActionNode 选择最优技能行为节点
type SelectOptimalSkillActionNode struct {
	*ActionNode
}

func NewSelectOptimalSkillActionNode(name string) *SelectOptimalSkillActionNode {
	return &SelectOptimalSkillActionNode{
		ActionNode: NewActionNode(name),
	}
}

func (n *SelectOptimalSkillActionNode) Execute(ctx *BossContext) BehaviorResult {
	n.BaseBehaviorNode.Execute(ctx)

	// 选择最优技能的逻辑
	// 这里可以根据当前情况选择最适合的技能
	logger.Debugf("Boss %d selecting optimal skill", ctx.Boss.GetID())

	// 简化实现：选择第一个可用技能
	if skillManager := ctx.GetSkillManager(); skillManager != nil {
		availableSkills := skillManager.GetAvailableSkills(ctx)
		if len(availableSkills) > 0 {
			// 选择第一个可用技能
			selectedSkill := availableSkills[0]
			logger.Debugf("Boss %d selected skill %d", ctx.Boss.GetID(), selectedSkill.GetID())
			return ResultSuccess
		}
	}

	return ResultFailure
}

// ExecuteActionActionNode 执行动作行为节点
type ExecuteActionActionNode struct {
	*ActionNode
}

func NewExecuteActionActionNode(name string) *ExecuteActionActionNode {
	return &ExecuteActionActionNode{
		ActionNode: NewActionNode(name),
	}
}

func (n *ExecuteActionActionNode) Execute(ctx *BossContext) BehaviorResult {
	n.BaseBehaviorNode.Execute(ctx)

	// 执行动作的逻辑
	// 这里可以执行之前选择的动作
	logger.Debugf("Boss %d executing action", ctx.Boss.GetID())

	// 简化实现：执行基本攻击
	if ctx.Target != nil && ctx.Target.IsAlive() {
		if err := ctx.Boss.DoAttackTarget(ctx.Target); err != nil {
			logger.Errorf("Boss execute action failed: %v", err)
			return ResultFailure
		}
		return ResultSuccess
	}

	return ResultFailure
}
