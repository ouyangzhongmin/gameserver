package bossai

import (
	"math/rand"

	"github.com/ouyangzhongmin/gameserver/pkg/logger"
)

// 集合节点，所有节点都会执行
type SetsNode struct {
	*BaseBehaviorNode
}

func NewSetsNode(name string) *SetsNode {
	return &SetsNode{
		BaseBehaviorNode: NewBaseBehaviorNode(name, NodeTypeSets),
	}
}

func (n *SetsNode) Execute(ctx *BossContext) BehaviorResult {
	// 开始执行
	if n.executeState == NodeStateIdle {
		n.executeState = NodeStateRunning
		n.executeCount++
		n.startTime = ctx.CurrentTime
	}

	n.lastExecuteTime = ctx.CurrentTime

	if len(n.children) == 0 {
		n.SetComplete(ResultSuccess)
		return ResultSuccess
	}

	isChildRunning := false
	failCnt := 0
	// 循环执行子节点
	for i := 0; i < len(n.children); i++ {
		child := n.children[i]
		result := child.Execute(ctx)
		//logger.Debugf("SetsNode: %s child:%s execute result:%d", n.name, child.GetName(), result)
		if result == ResultRunning {
			isChildRunning = true
		} else if result == ResultFailure {
			failCnt++
		}
	}

	if isChildRunning {
		n.executeState = NodeStateRunning
		return ResultRunning
	}
	n.Reset() // 这里重置表示本轮运行已结束了
	if failCnt == len(n.children) {
		// 全部都失败了
		return ResultFailure
	}
	return ResultSuccess
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
	// 开始执行
	if n.executeState == NodeStateIdle {
		n.executeState = NodeStateRunning
		n.currentChildIndex = 0
		n.executeCount++
		n.startTime = ctx.CurrentTime
	}

	n.lastExecuteTime = ctx.CurrentTime

	if len(n.children) == 0 {
		n.SetComplete(ResultSuccess)
		return ResultSuccess
	}

	// 循环执行子节点
	for n.currentChildIndex < len(n.children) {
		child := n.children[n.currentChildIndex]
		result := child.Execute(ctx)

		switch result {
		case ResultSuccess:
			// 当前子节点成功, 转到下一个
			n.currentChildIndex++
			continue
		case ResultFailure:
			// 子节点失败，整个序列失败
			logger.Debugf("SequenceNode: %s-%d execute failed", n.name, n.currentChildIndex)
			n.Reset() // 这里需要重置下轮从新开始
			return ResultFailure
		case ResultRunning:
			// 子节点运行中，继续等待
			return ResultRunning
		}
	}

	// 所有子节点都成功 本轮执行结束了
	n.Reset() // 重置下轮从新开始
	return ResultSuccess
}

func (n *SequenceNode) Reset() {
	n.BaseBehaviorNode.Reset() // 调用基类的Reset
	n.currentChildIndex = 0
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
	// 开始执行
	if n.executeState == NodeStateIdle {
		n.executeState = NodeStateRunning
		n.currentChildIndex = 0
		n.executeCount++
		n.startTime = ctx.CurrentTime
	}

	if len(n.children) == 0 {
		n.SetFailed()
		return ResultFailure
	}

	// 循环执行子节点
	for n.currentChildIndex < len(n.children) {
		child := n.children[n.currentChildIndex]
		result := child.Execute(ctx)

		switch result {
		case ResultSuccess:
			// 找到一个成功的子节点，整个选择节点成功
			n.lastResult = ResultSuccess
			n.Reset()
			return ResultSuccess
		case ResultFailure:
			// 子节点失败，重置子节点并试下一个
			n.currentChildIndex++
			continue
		case ResultRunning:
			// 子节点运行中，继续等待
			return ResultRunning
		}
	}
	// 如果执行到最后都失败了则重置重新开始执行
	n.Reset()
	return ResultFailure
}

func (n *SelectorNode) Reset() {
	n.BaseBehaviorNode.Reset()
	n.currentChildIndex = 0
}

// RepeaterNode 重复节点 - 重复执行子节点
type RepeaterNode struct {
	*BaseBehaviorNode
	maxRepeats        int
	currentRepeats    int
	currentChildIndex int
}

func NewRepeaterNode(name string, maxRepeats int) *RepeaterNode {
	return &RepeaterNode{
		BaseBehaviorNode:  NewBaseBehaviorNode(name, NodeTypeDecorator),
		maxRepeats:        maxRepeats,
		currentChildIndex: 0,
	}
}

func (n *RepeaterNode) Execute(ctx *BossContext) BehaviorResult {
	if n.executeState == NodeStateIdle {
		n.executeState = NodeStateRunning
		n.executeCount++
		n.currentRepeats = 0
		n.currentChildIndex = 0
		n.startTime = ctx.CurrentTime
		if n.maxRepeats < 1 {
			n.maxRepeats = 1
		}
	}

	if len(n.children) == 0 {
		n.executeState = NodeStateFailed
		n.lastResult = ResultFailure
		return ResultFailure
	}

	// 检查是否达到最大重复次数
	if n.maxRepeats > 0 && n.currentRepeats >= n.maxRepeats {
		n.Reset() // 重置等待下一轮可以继续执行
		return ResultSuccess
	}

	// 执行子节点
	child := n.children[n.currentChildIndex]
	result := child.Execute(ctx)

	switch result {
	case ResultSuccess:
		// 子节点执行成功，重置并增加重复次数
		child.Reset()
		n.currentChildIndex++
		if n.currentChildIndex >= len(n.children) {
			n.currentRepeats++
			// 检查是否达到最大重复次数
			if n.maxRepeats > 0 && n.currentRepeats >= n.maxRepeats {
				n.Reset() // 本轮正常结束了
				return ResultSuccess
			}
		}
		// 继续下一个子节点执行
		return ResultRunning
	case ResultFailure:
		// 子节点失败，重复节点失败
		n.Reset()
		return ResultFailure
	case ResultRunning:
		// 子节点运行中，继续等待
		return ResultRunning
	}

	n.Reset()
	return ResultFailure
}

func (n *RepeaterNode) Reset() {
	n.BaseBehaviorNode.Reset()
	n.currentRepeats = 0
	n.currentChildIndex = 0
}

func (n *RepeaterNode) SetParam(key string, value interface{}) {
	// 支持设置最大重复次数
	n.BaseBehaviorNode.SetParam(key, value)
	if key == "max_repeats" {
		if maxRepeats, ok := value.(int); ok {
			n.maxRepeats = maxRepeats
		}
	}
}

// RandomNode 随机选择节点 - 根据权重随机执行子节点
type RandomNode struct {
	*BaseBehaviorNode
	weights      []int         // 子节点的权重
	selectedNode IBehaviorNode // 当前选中的子节点
}

func NewRandomNode(name string) *RandomNode {
	return &RandomNode{
		BaseBehaviorNode: NewBaseBehaviorNode(name, NodeTypeSelector), // 使用Selector类型
		weights:          make([]int, 0),
	}
}

func (n *RandomNode) AddChild(child IBehaviorNode) error {
	err := n.BaseBehaviorNode.AddChild(child)
	if err == nil {
		// 默认权重为100
		n.weights = append(n.weights, 100)
	}
	return err
}

func (n *RandomNode) GetChildren() []IBehaviorNode {
	return n.BaseBehaviorNode.GetChildren()
}

// SetChildWeight 设置子节点权重
func (n *RandomNode) SetChildWeight(index int, weight int) {
	if index >= 0 && index < len(n.weights) {
		n.weights[index] = weight
	}
}

func (n *RandomNode) Execute(ctx *BossContext) BehaviorResult {
	if len(n.children) == 0 {
		return ResultFailure
	}

	// 开始执行 - 随机选择一个子节点
	if n.executeState == NodeStateIdle {
		n.executeState = NodeStateRunning

		// 计算总权重
		totalWeight := 0
		for _, weight := range n.weights {
			totalWeight += weight
		}

		if totalWeight <= 0 {
			n.executeState = NodeStateFailed
			n.lastResult = ResultFailure
			return ResultFailure
		}

		// 随机选择一个子节点
		randomValue := rand.Intn(totalWeight)
		currentWeight := 0

		for i, weight := range n.weights {
			currentWeight += weight
			if randomValue < currentWeight {
				n.selectedNode = n.children[i]
				logger.Debugf("RandomNode %s selected child %d: %s", n.name, i, n.selectedNode.GetName())
				break
			}
		}

		// 如果没有选中任何节点，选择第一个
		if n.selectedNode == nil && len(n.children) > 0 {
			n.selectedNode = n.children[0]
		}

		if n.selectedNode == nil {
			return ResultFailure
		}
	}

	// 执行当前选中的子节点
	if n.selectedNode != nil {
		result := n.selectedNode.Execute(ctx)

		switch result {
		case ResultSuccess:
			n.Reset()
			return ResultSuccess
		case ResultFailure:
			n.Reset()
			return ResultFailure
		case ResultRunning:
			// 继续执行当前选中的节点
			return ResultRunning
		}
	}

	n.Reset()
	return ResultFailure
}

func (n *RandomNode) Reset() {
	n.BaseBehaviorNode.Reset()
	n.selectedNode = nil
}
