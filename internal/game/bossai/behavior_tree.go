package bossai

import (
	"fmt"
	"sync"
	"time"

	"github.com/ouyangzhongmin/gameserver/pkg/logger"
)

// BehaviorTree 行为树
type BehaviorTree struct {
	name      string
	rootNode  IBehaviorNode
	context   *BossContext
	isRunning bool
	mutex     sync.RWMutex

	// 性能监控
	executeCount     int64
	totalExecuteTime time.Duration
	lastExecuteTime  time.Time
	maxExecuteTime   time.Duration

	// 配置
	updateInterval time.Duration
	maxDepth       int
	enableLogging  bool

	// 调试信息
	lastResult    BehaviorResult
	executionPath []string
}

// NewBehaviorTree 创建行为树
func NewBehaviorTree(name string) *BehaviorTree {
	return &BehaviorTree{
		name:           name,
		isRunning:      false,
		updateInterval: time.Millisecond * 100,
		maxDepth:       10,
		enableLogging:  false,
		executionPath:  make([]string, 0),
	}
}

// SetRootNode 设置根节点
func (bt *BehaviorTree) SetRootNode(node IBehaviorNode) {
	bt.mutex.Lock()
	defer bt.mutex.Unlock()

	bt.rootNode = node
}

// GetRootNode 获取根节点
func (bt *BehaviorTree) GetRootNode() IBehaviorNode {
	bt.mutex.RLock()
	defer bt.mutex.RUnlock()

	return bt.rootNode
}

// Execute 执行行为树
func (bt *BehaviorTree) Execute(ctx *BossContext) BehaviorResult {
	bt.mutex.Lock()
	defer bt.mutex.Unlock()

	if bt.rootNode == nil {
		return ResultFailure
	}

	bt.executeCount++
	start := time.Now()

	defer func() {
		duration := time.Since(start)
		bt.totalExecuteTime += duration
		bt.lastExecuteTime = ctx.CurrentTime

		if duration > bt.maxExecuteTime {
			bt.maxExecuteTime = duration
		}
	}()

	bt.context = ctx
	bt.isRunning = true
	bt.executionPath = bt.executionPath[:0] // 清空但保持容量

	result := bt.executeNode(bt.rootNode, ctx, 0)
	bt.lastResult = result
	bt.isRunning = false

	if bt.enableLogging {
		logger.Debugf("BehaviorTree %s executed, result: %v, path: %v",
			bt.name, result, bt.executionPath)
	}

	return result
}

// executeNode 执行节点（带深度检查）
func (bt *BehaviorTree) executeNode(node IBehaviorNode, ctx *BossContext, depth int) BehaviorResult {
	if depth > bt.maxDepth {
		logger.Errorf("BehaviorTree %s exceeded max depth %d", bt.name, bt.maxDepth)
		return ResultFailure
	}

	// 记录执行路径
	bt.executionPath = append(bt.executionPath, node.GetName())

	return node.Execute(ctx)
}

// Reset 重置行为树
func (bt *BehaviorTree) Reset() {
	bt.mutex.Lock()
	defer bt.mutex.Unlock()

	if bt.rootNode != nil {
		bt.rootNode.Reset()
	}

	bt.isRunning = false
	bt.lastResult = ResultFailure
	bt.executionPath = bt.executionPath[:0]
}

// IsRunning 检查是否正在运行
func (bt *BehaviorTree) IsRunning() bool {
	bt.mutex.RLock()
	defer bt.mutex.RUnlock()

	return bt.isRunning
}

// GetLastResult 获取最后执行结果
func (bt *BehaviorTree) GetLastResult() BehaviorResult {
	bt.mutex.RLock()
	defer bt.mutex.RUnlock()

	return bt.lastResult
}

// GetExecutionPath 获取执行路径
func (bt *BehaviorTree) GetExecutionPath() []string {
	bt.mutex.RLock()
	defer bt.mutex.RUnlock()

	// 返回副本
	path := make([]string, len(bt.executionPath))
	copy(path, bt.executionPath)
	return path
}

// GetStatistics 获取统计信息
func (bt *BehaviorTree) GetStatistics() BehaviorTreeStats {
	bt.mutex.RLock()
	defer bt.mutex.RUnlock()

	stats := BehaviorTreeStats{
		Name:             bt.name,
		ExecuteCount:     bt.executeCount,
		TotalExecuteTime: bt.totalExecuteTime,
		LastExecuteTime:  bt.lastExecuteTime,
		MaxExecuteTime:   bt.maxExecuteTime,
		IsRunning:        bt.isRunning,
		LastResult:       bt.lastResult,
	}

	if bt.executeCount > 0 {
		stats.AverageExecuteTime = bt.totalExecuteTime / time.Duration(bt.executeCount)
	}

	return stats
}

type BehaviorTreeStats struct {
	Name               string         `json:"name"`
	ExecuteCount       int64          `json:"execute_count"`
	TotalExecuteTime   time.Duration  `json:"total_execute_time"`
	AverageExecuteTime time.Duration  `json:"average_execute_time"`
	LastExecuteTime    time.Time      `json:"last_execute_time"`
	MaxExecuteTime     time.Duration  `json:"max_execute_time"`
	IsRunning          bool           `json:"is_running"`
	LastResult         BehaviorResult `json:"last_result"`
}

// SetConfiguration 设置配置
func (bt *BehaviorTree) SetConfiguration(config BehaviorTreeConfiguration) {
	bt.mutex.Lock()
	defer bt.mutex.Unlock()

	bt.updateInterval = config.UpdateInterval
	bt.maxDepth = config.MaxDepth
	bt.enableLogging = config.EnableLogging
}

type BehaviorTreeConfiguration struct {
	UpdateInterval time.Duration
	MaxDepth       int
	EnableLogging  bool
}

// BehaviorTreeBuilder 行为树构建器
type BehaviorTreeBuilder struct {
	name        string
	rootNode    IBehaviorNode
	currentNode IBehaviorNode
	nodeStack   []IBehaviorNode
}

// NewBehaviorTreeBuilder 创建行为树构建器
func NewBehaviorTreeBuilder(name string) *BehaviorTreeBuilder {
	return &BehaviorTreeBuilder{
		name:      name,
		nodeStack: make([]IBehaviorNode, 0),
	}
}

// Sequence 添加顺序节点
func (builder *BehaviorTreeBuilder) Sequence(name string) *BehaviorTreeBuilder {
	node := NewSequenceNode(name)
	return builder.addNode(node)
}

// Selector 添加选择节点
func (builder *BehaviorTreeBuilder) Selector(name string) *BehaviorTreeBuilder {
	node := NewSelectorNode(name)
	return builder.addNode(node)
}

// Parallel 添加并行节点
func (builder *BehaviorTreeBuilder) Parallel(name string, successThreshold, failureThreshold int) *BehaviorTreeBuilder {
	node := NewParallelNode(name, successThreshold, failureThreshold)
	return builder.addNode(node)
}

// Action 添加行为节点
func (builder *BehaviorTreeBuilder) Action(node IBehaviorNode) *BehaviorTreeBuilder {
	return builder.addNode(node)
}

// Condition 添加条件节点
func (builder *BehaviorTreeBuilder) Condition(node IBehaviorNode) *BehaviorTreeBuilder {
	return builder.addNode(node)
}

// Inverter 添加反转装饰节点
func (builder *BehaviorTreeBuilder) Inverter(name string) *BehaviorTreeBuilder {
	node := NewInverterNode(name)
	return builder.addNode(node)
}

// Repeater 添加重复装饰节点
func (builder *BehaviorTreeBuilder) Repeater(name string, maxRepeats int) *BehaviorTreeBuilder {
	node := NewRepeaterNode(name, maxRepeats)
	return builder.addNode(node)
}

// UntilFail 添加直到失败装饰节点
func (builder *BehaviorTreeBuilder) UntilFail(name string) *BehaviorTreeBuilder {
	node := NewUntilFailNode(name)
	return builder.addNode(node)
}

// Cooldown 添加冷却装饰节点
func (builder *BehaviorTreeBuilder) Cooldown(name string, cooldownTime time.Duration) *BehaviorTreeBuilder {
	node := NewCooldownNode(name, cooldownTime)
	return builder.addNode(node)
}

// End 结束当前复合节点
func (builder *BehaviorTreeBuilder) End() *BehaviorTreeBuilder {
	if len(builder.nodeStack) > 0 {
		builder.nodeStack = builder.nodeStack[:len(builder.nodeStack)-1]

		if len(builder.nodeStack) > 0 {
			builder.currentNode = builder.nodeStack[len(builder.nodeStack)-1]
		} else {
			builder.currentNode = nil
		}
	}

	return builder
}

// Build 构建行为树
func (builder *BehaviorTreeBuilder) Build() *BehaviorTree {
	tree := NewBehaviorTree(builder.name)
	tree.SetRootNode(builder.rootNode)
	return tree
}

// addNode 添加节点到当前复合节点
func (builder *BehaviorTreeBuilder) addNode(node IBehaviorNode) *BehaviorTreeBuilder {
	if builder.rootNode == nil {
		builder.rootNode = node
	}

	if builder.currentNode != nil {
		builder.currentNode.AddChild(node)
	}

	// 如果是复合节点，将其设置为当前节点
	switch node.GetType() {
	case NodeTypeSequence, NodeTypeSelector, NodeTypeParallel, NodeTypeDecorator:
		builder.currentNode = node
		builder.nodeStack = append(builder.nodeStack, node)
	}

	return builder
}

// BehaviorTreeManager 行为树管理器
type BehaviorTreeManager struct {
	trees map[string]*BehaviorTree
	mutex sync.RWMutex
}

// NewBehaviorTreeManager 创建行为树管理器
func NewBehaviorTreeManager() *BehaviorTreeManager {
	return &BehaviorTreeManager{
		trees: make(map[string]*BehaviorTree),
	}
}

// AddTree 添加行为树
func (btm *BehaviorTreeManager) AddTree(tree *BehaviorTree) error {
	btm.mutex.Lock()
	defer btm.mutex.Unlock()

	if _, exists := btm.trees[tree.name]; exists {
		return fmt.Errorf("behavior tree %s already exists", tree.name)
	}

	btm.trees[tree.name] = tree
	return nil
}

// RemoveTree 移除行为树
func (btm *BehaviorTreeManager) RemoveTree(name string) error {
	btm.mutex.Lock()
	defer btm.mutex.Unlock()

	tree, exists := btm.trees[name]
	if !exists {
		return fmt.Errorf("behavior tree %s not found", name)
	}

	tree.Reset()
	delete(btm.trees, name)
	return nil
}

// GetTree 获取行为树
func (btm *BehaviorTreeManager) GetTree(name string) (*BehaviorTree, error) {
	btm.mutex.RLock()
	defer btm.mutex.RUnlock()

	tree, exists := btm.trees[name]
	if !exists {
		return nil, fmt.Errorf("behavior tree %s not found", name)
	}

	return tree, nil
}

// ExecuteTree 执行指定的行为树
func (btm *BehaviorTreeManager) ExecuteTree(name string, ctx *BossContext) (BehaviorResult, error) {
	tree, err := btm.GetTree(name)
	if err != nil {
		return ResultFailure, err
	}

	return tree.Execute(ctx), nil
}

// GetAllTrees 获取所有行为树
func (btm *BehaviorTreeManager) GetAllTrees() map[string]*BehaviorTree {
	btm.mutex.RLock()
	defer btm.mutex.RUnlock()

	// 返回副本
	result := make(map[string]*BehaviorTree)
	for name, tree := range btm.trees {
		result[name] = tree
	}

	return result
}

// Reset 重置所有行为树
func (btm *BehaviorTreeManager) Reset() {
	btm.mutex.Lock()
	defer btm.mutex.Unlock()

	for _, tree := range btm.trees {
		tree.Reset()
	}
}

// Destroy 销毁管理器
func (btm *BehaviorTreeManager) Destroy() {
	btm.Reset()
	btm.trees = nil
}
