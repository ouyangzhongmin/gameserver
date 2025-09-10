package bossai

import (
	"fmt"
	"sync"
	"time"

	"github.com/ouyangzhongmin/gameserver/pkg/logger"
)

// 基础节点定义（用于支持复杂行为树）
// NodeExecuteState 节点执行状态
type NodeExecuteState int

const (
	NodeStateIdle     NodeExecuteState = iota // 空闲状态
	NodeStateRunning                          // 运行中
	NodeStateComplete                         // 已完成
	NodeStateFailed                           // 失败
)

// BaseBehaviorNode 基础行为树节点
type BaseBehaviorNode struct {
	name     string
	nodeType BehaviorNodeType
	children []IBehaviorNode
	parent   IBehaviorNode

	// 节点执行状态管理
	executeState NodeExecuteState // 节点执行状态
	lastResult   BehaviorResult   // 最后执行结果
	executeCount int64            // 执行次数
	startTime    time.Time        // 开始执行时间

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
	// 如果节点已完成，直接返回结果
	if n.executeState == NodeStateComplete {
		return n.lastResult
	}

	// 如果节点失败，直接返回失败结果
	if n.executeState == NodeStateFailed {
		return ResultFailure
	}

	// 开始执行节点
	if n.executeState == NodeStateIdle {
		n.executeState = NodeStateRunning
		n.startTime = ctx.CurrentTime
		n.executeCount++
	}

	n.lastExecuteTime = ctx.CurrentTime
	start := time.Now()

	defer func() {
		n.totalExecuteTime += time.Since(start)
	}()

	// 基础节点不执行任何操作
	n.executeState = NodeStateFailed
	n.lastResult = ResultFailure
	return ResultFailure
}

func (n *BaseBehaviorNode) Reset() {
	n.executeState = NodeStateIdle
	n.lastResult = ResultFailure

	// 递归重置子节点
	for _, child := range n.children {
		child.Reset()
	}
}

func (n *BaseBehaviorNode) SetParams(val map[string]interface{}) {
	if val != nil {
		for key, value := range val {
			n.params[key] = value
		}
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

// CheckStateAndBeginExecution 检查状态并开始执行（通用方法）
func (n *BaseBehaviorNode) CheckStateAndBeginExecution(ctx *BossContext) BehaviorResult {
	// 如果节点已完成，直接返回结果
	if n.executeState == NodeStateComplete {
		return n.lastResult
	}

	// 如果节点失败，直接返回失败结果
	if n.executeState == NodeStateFailed {
		return ResultFailure
	}

	// 开始执行节点
	if n.executeState == NodeStateIdle {
		n.executeState = NodeStateRunning
		n.startTime = ctx.CurrentTime
		n.executeCount++
	}

	n.lastExecuteTime = ctx.CurrentTime
	return ResultRunning // 表示需要继续执行
}

// SetComplete 设置节点完成状态
func (n *BaseBehaviorNode) SetComplete(result BehaviorResult) {
	n.executeState = NodeStateComplete
	n.lastResult = result
}

// SetFailed 设置节点失败状态
func (n *BaseBehaviorNode) SetFailed() {
	n.executeState = NodeStateFailed
	n.lastResult = ResultFailure
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
