package bossai

import (
	"time"

	"github.com/ouyangzhongmin/gameserver/pkg/logger"
)

// 组合的节点定义
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

// ConditionNode 条件节点基类
type ConditionNode struct {
	*BaseBehaviorNode
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
			logger.Debugf("SequenceNode: %s-%d execute failed", n.name, n.currentChildIndex)
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
