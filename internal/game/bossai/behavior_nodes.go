package bossai

import (
	"fmt"
	"time"

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

	// 执行攻击逻辑
	logger.Debugf("Boss %d attacking target %d", ctx.Boss.GetID(), ctx.Target.GetID())

	// 这里可以调用具体的攻击方法
	// 暂时返回成功
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

	// 搜索附近敌人
	enemies := ctx.Boss.GetEntitiesInRange(n.searchRadius)

	if len(enemies) == 0 {
		return ResultFailure
	}

	// 找最近的敌人
	nearest := ctx.Boss.GetNearestEnemy()
	if nearest != nil {
		ctx.Target = nearest
		ctx.Boss.SetCombatTarget(nearest)
		return ResultSuccess
	}

	return ResultFailure
}
