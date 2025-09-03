package bossai

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/ouyangzhongmin/gameserver/pkg/logger"
)

// BossAIManager Boss AI主管理器
type BossAIManager struct {
	// 核心组件
	boss    IBossEntity
	context *BossContext

	// 子系统
	stateMachine  *StateMachine
	phaseManager  *PhaseManager
	skillManager  *SkillManager
	behaviorTrees map[string]*BehaviorTree
	plugins       map[string]IAIPlugin
	pluginOrder   []string

	// 配置
	config *BossConfig

	// 运行状态
	isInitialized bool
	isRunning     bool
	isPaused      bool

	// 性能监控
	updateCount     int64
	totalUpdateTime time.Duration
	lastUpdateTime  time.Time
	frameTime       time.Duration

	// 调试信息
	enableDebug bool
	debugInfo   *AIDebugInfo

	// 同步控制
	mutex sync.RWMutex
	// updateMutex sync.Mutex
}

// NewBossAIManager 创建Boss AI管理器
func NewBossAIManager() *BossAIManager {
	return &BossAIManager{
		stateMachine:  NewStateMachine(),
		phaseManager:  NewPhaseManager(),
		skillManager:  NewSkillManager(),
		behaviorTrees: make(map[string]*BehaviorTree),
		plugins:       make(map[string]IAIPlugin),
		pluginOrder:   make([]string, 0),
		context: &BossContext{
			CustomData:    make(map[string]interface{}),
			SkillsUsed:    make(map[int32]int),
			ActionHistory: make([]*AIAction, 0),
		},
		debugInfo: &AIDebugInfo{},
	}
}

// Initialize 初始化Boss AI
func (ai *BossAIManager) Initialize(boss IBossEntity) error {
	ai.mutex.Lock()
	defer ai.mutex.Unlock()

	if ai.isInitialized {
		return fmt.Errorf("AI manager already initialized")
	}

	ai.boss = boss

	// 初始化上下文
	ai.context.Boss = boss
	ai.context.CurrentTime = time.Now()

	// 如果有配置，应用配置
	if ai.config != nil {
		if err := ai.applyConfiguration(); err != nil {
			return fmt.Errorf("failed to apply configuration: %w", err)
		}
	}

	// 初始化插件
	for _, plugin := range ai.plugins {
		if err := plugin.Initialize(ai.context); err != nil {
			logger.Errorf("Failed to initialize plugin %s: %v", plugin.GetName(), err)
			continue
		}
	}

	ai.isInitialized = true
	logger.Debugf("BossAI initialized for entity %d", boss.GetID())

	return nil
}

// LoadConfig 加载配置
func (ai *BossAIManager) LoadConfig(config *BossConfig) error {
	ai.mutex.Lock()
	defer ai.mutex.Unlock()

	ai.config = config

	if ai.isInitialized {
		return ai.applyConfiguration()
	}

	return nil
}

// applyConfiguration 应用配置
func (ai *BossAIManager) applyConfiguration() error {
	if ai.config == nil {
		return fmt.Errorf("no configuration to apply")
	}

	// 配置状态机
	if err := ai.configureStateMachine(); err != nil {
		return fmt.Errorf("failed to configure state machine: %w", err)
	}

	// 配置阶段系统
	if err := ai.configurePhaseSystem(); err != nil {
		return fmt.Errorf("failed to configure phase system: %w", err)
	}

	// 配置技能系统
	if err := ai.configureSkillSystem(); err != nil {
		return fmt.Errorf("failed to configure skill system: %w", err)
	}

	// 配置行为树
	if err := ai.configureBehaviorTrees(); err != nil {
		return fmt.Errorf("failed to configure behavior trees: %w", err)
	}

	// 配置插件
	if err := ai.configurePlugins(); err != nil {
		return fmt.Errorf("failed to configure plugins: %w", err)
	}

	logger.Debugf("Configuration applied for Boss %s", ai.config.Name)
	return nil
}

// configureStateMachine 配置状态机
func (ai *BossAIManager) configureStateMachine() error {
	// 添加预定义状态
	states := []IBossState{
		NewIdleState(),
		NewPatrolState([]Position{}),
		NewChaseState(),
		NewAttackState(),
		NewRetreatState(Position{X: 0, Y: 0, Z: 0}), // 默认出生点
		NewStunnedState(time.Second * 3),            // 默认3秒眩晕
		NewDyingState(),
	}

	for _, state := range states {
		if err := ai.stateMachine.AddState(state); err != nil {
			logger.Errorf("Failed to add state %s: %v", state.GetName(), err)
		}
	}

	// 配置自定义状态
	for _, stateConfig := range ai.config.States {
		// 这里可以根据配置创建自定义状态
		logger.Debugf("State configuration: %s", stateConfig.Name)
	}

	// 设置初始状态
	if err := ai.stateMachine.SetInitialState(ai.config.InitialState, ai.context); err != nil {
		return err
	}

	return nil
}

// configurePhaseSystem 配置阶段系统
func (ai *BossAIManager) configurePhaseSystem() error {
	for _, phaseConfig := range ai.config.Phases {
		phase := NewBossPhase(phaseConfig.ID, phaseConfig.Name)
		phase.SetTriggerCondition(phaseConfig.TriggerCondition)
		phase.SetAvailableSkills(phaseConfig.AvailableSkills)
		phase.SetBehaviorTreeName(phaseConfig.BehaviorTree)

		// 应用修改器
		for key, value := range phaseConfig.Modifiers {
			phase.SetStateModifier(key, value)
		}

		if err := ai.phaseManager.AddPhase(phase); err != nil {
			logger.Errorf("Failed to add phase %s: %v", phase.GetName(), err)
		}
	}

	// 设置阶段检查顺序
	phaseOrder := make([]int32, 0, len(ai.config.Phases))
	for _, phaseConfig := range ai.config.Phases {
		phaseOrder = append(phaseOrder, phaseConfig.ID)
	}

	return ai.phaseManager.SetPhaseOrder(phaseOrder)
}

// configureSkillSystem 配置技能系统
func (ai *BossAIManager) configureSkillSystem() error {
	skillPriority := make([]int32, 0, len(ai.config.Skills))

	for _, skillConfig := range ai.config.Skills {
		skill := NewBossSkill(skillConfig.ID, skillConfig.Name)
		skill.SetCooldown(skillConfig.Cooldown)
		skill.SetRange(skillConfig.Range)
		skill.SetCastTime(skillConfig.CastTime)

		// 添加条件
		for _, condition := range skillConfig.Conditions {
			skill.AddCondition(SkillCondition{
				Type:     condition.Type,
				Params:   condition.Params,
				Operator: condition.Operator,
				SubConds: convertToSkillConditions(condition.SubConds),
			})
		}

		// 添加效果
		for _, effect := range skillConfig.Effects {
			skill.AddEffect(effect)
		}

		if err := ai.skillManager.AddSkill(skill); err != nil {
			logger.Errorf("Failed to add skill %s: %v", skill.GetName(), err)
		} else {
			skillPriority = append(skillPriority, skillConfig.ID)
		}
	}

	// 设置技能优先级
	return ai.skillManager.SetSkillPriority(skillPriority)
}

// configureBehaviorTrees 配置行为树
func (ai *BossAIManager) configureBehaviorTrees() error {
	// 创建默认行为树
	defaultTrees := map[string]*BehaviorTree{
		"BasicCombat":    CreateBasicCombatTree(),
		"AdvancedCombat": CreateAdvancedCombatTree(),
	}

	for name, tree := range defaultTrees {
		ai.behaviorTrees[name] = tree
	}

	// 根据配置创建自定义行为树
	if ai.config.BehaviorTree.RootNode.Name != "" {
		customTree, err := ai.buildBehaviorTreeFromConfig(ai.config.BehaviorTree)
		if err != nil {
			return fmt.Errorf("failed to build custom behavior tree: %w", err)
		}

		ai.behaviorTrees["Custom"] = customTree
	}

	return nil
}

// configurePlugins 配置插件
func (ai *BossAIManager) configurePlugins() error {
	for _, pluginConfig := range ai.config.Plugins {
		if !pluginConfig.Enabled {
			continue
		}

		var plugin IAIPlugin

		switch pluginConfig.Type {
		case "combat_analysis":
			plugin = NewCombatAnalysisPlugin()
		case "llm":
			if ai.config.LLMConfig != nil && ai.config.LLMConfig.Enabled {
				provider := NewOpenAIProvider(ai.config.LLMConfig.APIKey)
				provider.Configure(map[string]interface{}{
					"endpoint":    ai.config.LLMConfig.Endpoint,
					"model":       ai.config.LLMConfig.Model,
					"max_tokens":  ai.config.LLMConfig.MaxTokens,
					"temperature": ai.config.LLMConfig.Temperature,
					"timeout_ms":  int(ai.config.LLMConfig.UpdateInterval.Milliseconds()),
				})
				plugin = NewLLMPlugin(provider)
			}
		default:
			logger.Warnf("Unknown plugin type: %s", pluginConfig.Type)
			continue
		}

		if plugin != nil {
			plugin.SetConfig(pluginConfig.Config)
			plugin.SetPriority(pluginConfig.Priority)

			if err := ai.AddPlugin(plugin); err != nil {
				logger.Errorf("Failed to add plugin %s: %v", pluginConfig.Name, err)
			}
		}
	}

	return nil
}

// Update 更新Boss AI
func (ai *BossAIManager) Update(deltaTime time.Duration) error {
	// monster中的update是来自于scene的update,已确保是在同一条线程调用的，如果这里加锁会导致性能很差
	// ai.updateMutex.Lock()
	// defer ai.updateMutex.Unlock()

	if !ai.isInitialized || !ai.isRunning || ai.isPaused {
		return nil
	}

	start := time.Now()
	defer func() {
		ai.frameTime = time.Since(start)
		ai.totalUpdateTime += ai.frameTime
		ai.updateCount++
		ai.lastUpdateTime = time.Now()
	}()

	// 更新上下文
	ai.updateContext(deltaTime)

	// 更新子系统
	if err := ai.updateSubsystems(deltaTime); err != nil {
		return fmt.Errorf("subsystem update failed: %w", err)
	}

	// 更新插件
	if err := ai.updatePlugins(deltaTime); err != nil {
		return fmt.Errorf("plugin update failed: %w", err)
	}

	// 执行AI决策
	if err := ai.executeAI(); err != nil {
		return fmt.Errorf("AI execution failed: %w", err)
	}

	// 更新调试信息
	if ai.enableDebug {
		ai.updateDebugInfo()
	}

	return nil
}

// updateContext 更新AI上下文
func (ai *BossAIManager) updateContext(deltaTime time.Duration) {
	ai.context.CurrentTime = time.Now()
	ai.context.DeltaTime = deltaTime

	// 更新战斗时间
	if ai.boss.IsInCombat() {
		ai.context.CombatTime += deltaTime
	}

	// 更新附近实体
	ai.context.NearbyEnemies = ai.boss.GetEnemiesInRange(10.0)

	// 更新位置
	ai.context.Position = ai.boss.GetPos()

	// 更新目标
	if ai.context.Target == nil || !ai.context.Target.IsAlive() {
		ai.context.Target = ai.boss.GetCombatTarget()
	}

	// 重置状态变化标志
	ai.context.IsStateChanged = false
	ai.context.IsPhaseChanged = false
}

// updateSubsystems 更新子系统
func (ai *BossAIManager) updateSubsystems(deltaTime time.Duration) error {
	// 更新技能冷却
	ai.skillManager.UpdateCooldowns()

	// 更新阶段管理器
	if err := ai.phaseManager.Update(ai.context, deltaTime); err != nil {
		return fmt.Errorf("phase manager update failed: %w", err)
	}

	// 更新状态机
	if err := ai.stateMachine.Update(ai.context, deltaTime); err != nil {
		return fmt.Errorf("state machine update failed: %w", err)
	}

	return nil
}

// updatePlugins 更新插件
func (ai *BossAIManager) updatePlugins(deltaTime time.Duration) error {
	for _, pluginName := range ai.pluginOrder {
		plugin := ai.plugins[pluginName]
		if err := plugin.Update(ai.context, deltaTime); err != nil {
			logger.Errorf("Plugin %s update failed: %v", pluginName, err)
		}
	}

	return nil
}

// executeAI 执行AI决策
func (ai *BossAIManager) executeAI() error {
	// 收集所有插件的建议
	suggestions := make([]*AIAction, 0)

	for _, pluginName := range ai.pluginOrder {
		plugin := ai.plugins[pluginName]
		if suggestion := plugin.SuggestAction(ai.context); suggestion != nil {
			suggestions = append(suggestions, suggestion)
		}
	}

	// 选择最佳动作
	var bestAction *AIAction

	if len(suggestions) > 0 {
		// 按优先级排序
		sort.Slice(suggestions, func(i, j int) bool {
			return suggestions[i].Priority > suggestions[j].Priority
		})
		bestAction = suggestions[0]
	}

	// 如果没有插件建议，使用行为树
	if bestAction == nil {
		bestAction = ai.executeBehaviorTree()
	}

	// 让插件修改行为
	if bestAction != nil {
		for _, pluginName := range ai.pluginOrder {
			plugin := ai.plugins[pluginName]
			bestAction = plugin.ModifyBehavior(ai.context, bestAction)
		}

		// 执行动作
		return ai.executeAction(bestAction)
	}

	return nil
}

// executeBehaviorTree 执行行为树
func (ai *BossAIManager) executeBehaviorTree() *AIAction {
	// 根据当前阶段选择行为树
	var treeName string
	if ai.context.CurrentPhase != nil && ai.context.CurrentPhase.GetBehaviorTreeName() != "" {
		treeName = ai.context.CurrentPhase.GetBehaviorTreeName()
	} else {
		treeName = "BasicCombat" // 默认行为树
	}

	tree, exists := ai.behaviorTrees[treeName]
	if !exists {
		tree = ai.behaviorTrees["BasicCombat"] // 回退到基础行为树
	}

	if tree != nil {
		result := tree.Execute(ai.context)
		logger.Debugf("Behavior tree %s executed with result: %v", treeName, result)
	}

	// 行为树通过上下文间接产生动作，这里返回nil
	// 实际的动作会在行为树节点中直接执行
	return nil
}

// executeAction 执行动作
func (ai *BossAIManager) executeAction(action *AIAction) error {
	switch action.Type {
	case ActionAttack:
		// 执行攻击
		logger.Debugf("Boss %d executing attack action", ai.boss.GetID())

	case ActionCastSkill:
		// 使用技能
		skill, err := ai.skillManager.GetSkill(action.SkillID)
		if err != nil {
			return err
		}

		targets := skill.GetTargets(ai.context)
		return skill.Cast(ai.context, targets)

	case ActionMove:
		// 移动
		if action.Position != nil {
			return ai.boss.MoveTo(action.Position.X, action.Position.Y, action.Position.Z)
		}

	case ActionRetreat:
		// 撤退逻辑
		logger.Debugf("Boss %d retreating", ai.boss.GetID())

	default:
		logger.Warnf("Unknown action type: %v", action.Type)
	}

	// 记录动作
	ai.context.LastAction = action
	ai.context.ActionHistory = append(ai.context.ActionHistory, action)

	// 限制历史记录长度
	if len(ai.context.ActionHistory) > 100 {
		ai.context.ActionHistory = ai.context.ActionHistory[1:]
	}

	return nil
}

// AddPlugin 添加插件
func (ai *BossAIManager) AddPlugin(plugin IAIPlugin) error {
	ai.mutex.Lock()
	defer ai.mutex.Unlock()

	name := plugin.GetName()

	if _, exists := ai.plugins[name]; exists {
		return fmt.Errorf("plugin %s already exists", name)
	}

	ai.plugins[name] = plugin
	ai.pluginOrder = append(ai.pluginOrder, name)

	// 按优先级排序
	sort.Slice(ai.pluginOrder, func(i, j int) bool {
		return ai.plugins[ai.pluginOrder[i]].GetPriority() > ai.plugins[ai.pluginOrder[j]].GetPriority()
	})

	// 如果AI已初始化，立即初始化插件
	if ai.isInitialized {
		if err := plugin.Initialize(ai.context); err != nil {
			delete(ai.plugins, name)
			// 从排序列表中移除
			for i, n := range ai.pluginOrder {
				if n == name {
					ai.pluginOrder = append(ai.pluginOrder[:i], ai.pluginOrder[i+1:]...)
					break
				}
			}
			return fmt.Errorf("failed to initialize plugin: %w", err)
		}
	}

	logger.Debugf("Plugin %s added with priority %d", name, plugin.GetPriority())
	return nil
}

// RemovePlugin 移除插件
func (ai *BossAIManager) RemovePlugin(pluginName string) error {
	ai.mutex.Lock()
	defer ai.mutex.Unlock()

	plugin, exists := ai.plugins[pluginName]
	if !exists {
		return fmt.Errorf("plugin %s not found", pluginName)
	}

	// 清理插件
	if err := plugin.Cleanup(); err != nil {
		logger.Errorf("Plugin %s cleanup failed: %v", pluginName, err)
	}

	delete(ai.plugins, pluginName)

	// 从排序列表中移除
	for i, name := range ai.pluginOrder {
		if name == pluginName {
			ai.pluginOrder = append(ai.pluginOrder[:i], ai.pluginOrder[i+1:]...)
			break
		}
	}

	logger.Debugf("Plugin %s removed", pluginName)
	return nil
}

// GetCurrentState 获取当前状态
func (ai *BossAIManager) GetCurrentState() IBossState {
	return ai.stateMachine.GetCurrentState()
}

// TransitionTo 转换状态
func (ai *BossAIManager) TransitionTo(stateID int32) error {
	return ai.stateMachine.ForceTransition(BossStateID(stateID), ai.context)
}

// GetCurrentPhase 获取当前阶段
func (ai *BossAIManager) GetCurrentPhase() IBossPhase {
	return ai.phaseManager.GetCurrentPhase()
}

// CheckPhaseTransition 检查阶段转换
func (ai *BossAIManager) CheckPhaseTransition() error {
	return ai.phaseManager.Update(ai.context, 0)
}

// ExecuteBehavior 执行行为
func (ai *BossAIManager) ExecuteBehavior() error {
	return ai.executeAI()
}

// InterruptBehavior 中断行为
func (ai *BossAIManager) InterruptBehavior() error {
	// 重置所有行为树
	for _, tree := range ai.behaviorTrees {
		tree.Reset()
	}

	// 停止当前动作
	return ai.boss.Stop()
}

// OnDamageReceived 处理受到伤害事件
func (ai *BossAIManager) OnDamageReceived(attacker IEntity, damage int32) error {
	ai.context.DamageReceived += damage

	// 通知状态机和插件
	// 这里可以触发特定的AI反应

	return nil
}

// OnSkillUsed 处理技能使用事件
func (ai *BossAIManager) OnSkillUsed(skillID int32, success bool) error {
	if _, exists := ai.context.SkillsUsed[skillID]; !exists {
		ai.context.SkillsUsed[skillID] = 0
	}

	if success {
		ai.context.SkillsUsed[skillID]++
	}

	return nil
}

// OnTargetChanged 处理目标改变事件
func (ai *BossAIManager) OnTargetChanged(newTarget IEntity) error {
	ai.context.Target = newTarget
	return nil
}

// Start 启动AI
func (ai *BossAIManager) Start() error {
	ai.mutex.Lock()
	defer ai.mutex.Unlock()

	if !ai.isInitialized {
		return fmt.Errorf("AI not initialized")
	}

	ai.isRunning = true
	ai.isPaused = false

	logger.Debugf("BossAI started for entity %d", ai.boss.GetID())
	return nil
}

// Stop 停止AI
func (ai *BossAIManager) Stop() error {
	ai.mutex.Lock()
	defer ai.mutex.Unlock()

	ai.isRunning = false

	logger.Debugf("BossAI stopped for entity %d", ai.boss.GetID())
	return nil
}

// Pause 暂停AI
func (ai *BossAIManager) Pause() {
	ai.mutex.Lock()
	defer ai.mutex.Unlock()

	ai.isPaused = true
}

// Resume 恢复AI
func (ai *BossAIManager) Resume() {
	ai.mutex.Lock()
	defer ai.mutex.Unlock()

	ai.isPaused = false
}

// IsRunning 检查是否运行中
func (ai *BossAIManager) IsRunning() bool {
	ai.mutex.RLock()
	defer ai.mutex.RUnlock()

	return ai.isRunning && !ai.isPaused
}

// GetDebugInfo 获取调试信息
func (ai *BossAIManager) GetDebugInfo() *AIDebugInfo {
	ai.mutex.RLock()
	defer ai.mutex.RUnlock()

	return ai.debugInfo
}

// SetDebugEnabled 设置调试模式
func (ai *BossAIManager) SetDebugEnabled(enabled bool) {
	ai.mutex.Lock()
	defer ai.mutex.Unlock()

	ai.enableDebug = enabled
}

// updateDebugInfo 更新调试信息
func (ai *BossAIManager) updateDebugInfo() {
	ai.debugInfo.CurrentState = ""
	if ai.context.CurrentState != nil {
		ai.debugInfo.CurrentState = ai.context.CurrentState.GetName()
	}

	ai.debugInfo.CurrentPhase = ""
	if ai.context.CurrentPhase != nil {
		ai.debugInfo.CurrentPhase = ai.context.CurrentPhase.GetName()
	}

	ai.debugInfo.LastAction = ai.context.LastAction

	// 收集插件状态
	ai.debugInfo.ActivePlugins = make([]string, 0, len(ai.plugins))
	for name, plugin := range ai.plugins {
		if plugin.IsEnabled() {
			ai.debugInfo.ActivePlugins = append(ai.debugInfo.ActivePlugins, name)
		}
	}

	// 性能指标
	if ai.debugInfo.PerformanceMetrics == nil {
		ai.debugInfo.PerformanceMetrics = &PerformanceMetrics{}
	}

	ai.debugInfo.PerformanceMetrics.UpdateTime = ai.frameTime
	ai.debugInfo.PerformanceMetrics.FrameRate = 1.0 / ai.frameTime.Seconds()
}

// Destroy 销毁AI管理器
func (ai *BossAIManager) Destroy() error {
	ai.mutex.Lock()
	defer ai.mutex.Unlock()

	ai.isRunning = false

	// 清理插件
	for _, plugin := range ai.plugins {
		if err := plugin.Cleanup(); err != nil {
			logger.Errorf("Plugin %s cleanup failed: %v", plugin.GetName(), err)
		}
	}

	// 清理子系统
	ai.stateMachine.Destroy()
	ai.phaseManager.Destroy()

	// 重置行为树
	for _, tree := range ai.behaviorTrees {
		tree.Reset()
	}

	logger.Debugf("BossAI destroyed for entity %d", ai.boss.GetID())
	return nil
}

// 辅助函数

func convertToSkillConditions(conditions []PhaseCondition) []SkillCondition {
	result := make([]SkillCondition, len(conditions))
	for i, cond := range conditions {
		result[i] = SkillCondition{
			Type:     cond.Type,
			Params:   cond.Params,
			Operator: cond.Operator,
			SubConds: convertToSkillConditions(cond.SubConds),
		}
	}
	return result
}

func (ai *BossAIManager) buildBehaviorTreeFromConfig(config BehaviorTreeConfig) (*BehaviorTree, error) {
	if config.RootNode.Name == "" {
		// 如果没有配置根节点，返回基础战斗树
		logger.Debugf("No root node configured, using basic combat tree")
		return CreateBasicCombatTree(), nil
	}

	// 从配置构建行为树

	// 递归构建根节点
	rootNode, err := ai.buildNodeFromConfig(config.RootNode)
	if err != nil {
		return nil, fmt.Errorf("failed to build root node: %w", err)
	}

	tree := NewBehaviorTree("ConfiguredTree")
	tree.SetRootNode(rootNode)

	// 设置行为树配置
	if config.UpdateInterval > 0 {
		treeConfig := BehaviorTreeConfiguration{
			UpdateInterval: config.UpdateInterval,
			MaxDepth:       10,
			EnableLogging:  false,
		}
		tree.SetConfiguration(treeConfig)
	}

	logger.Debugf("Successfully built behavior tree from config")
	return tree, nil
}

// buildNodeFromConfig 从配置构建节点
func (ai *BossAIManager) buildNodeFromConfig(nodeConfig BehaviorNodeConfig) (IBehaviorNode, error) {
	switch nodeConfig.Type {
	case NodeTypeSequence:
		// 顺序节点
		node := NewSequenceNode(nodeConfig.Name)
		for _, childConfig := range nodeConfig.Children {
			child, err := ai.buildNodeFromConfig(childConfig)
			if err != nil {
				return nil, fmt.Errorf("failed to build child node for sequence %s: %w", nodeConfig.Name, err)
			}
			node.AddChild(child)
		}
		return node, nil

	case NodeTypeSelector:
		// 选择节点
		node := NewSelectorNode(nodeConfig.Name)
		for _, childConfig := range nodeConfig.Children {
			child, err := ai.buildNodeFromConfig(childConfig)
			if err != nil {
				return nil, fmt.Errorf("failed to build child node for selector %s: %w", nodeConfig.Name, err)
			}
			node.AddChild(child)
		}
		return node, nil

	case NodeTypeParallel:
		// 并行节点
		successThreshold := 1
		failureThreshold := 1

		if threshold, ok := nodeConfig.Params["success_threshold"].(float64); ok {
			successThreshold = int(threshold)
		}
		if threshold, ok := nodeConfig.Params["failure_threshold"].(float64); ok {
			failureThreshold = int(threshold)
		}

		node := NewParallelNode(nodeConfig.Name, successThreshold, failureThreshold)
		for _, childConfig := range nodeConfig.Children {
			child, err := ai.buildNodeFromConfig(childConfig)
			if err != nil {
				return nil, fmt.Errorf("failed to build child node for parallel %s: %w", nodeConfig.Name, err)
			}
			node.AddChild(child)
		}
		return node, nil

	case NodeTypeAction:
		// 行为节点 - 根据名称和参数创建具体的行为节点
		return ai.createActionNode(nodeConfig)

	case NodeTypeCondition:
		// 条件节点 - 根据名称和参数创建具体的条件节点
		return ai.createConditionNode(nodeConfig)

	case NodeTypeDecorator:
		// 装饰节点 - 根据名称创建具体的装饰节点
		return ai.createDecoratorNode(nodeConfig)

	default:
		return nil, fmt.Errorf("unknown node type: %d", nodeConfig.Type)
	}
}

// createActionNode 创建行为节点
func (ai *BossAIManager) createActionNode(nodeConfig BehaviorNodeConfig) (IBehaviorNode, error) {
	switch nodeConfig.Name {
	case "Use Healing Skill":
		// 治疗技能行为节点
		skillID := int32(1007) // 默认治疗技能ID
		if id, ok := nodeConfig.Params["skill_id"].(float64); ok {
			skillID = int32(id)
		}
		return NewUseSkillActionNode(nodeConfig.Name, skillID), nil

	case "Attack Target":
		// 攻击目标行为节点
		return NewAttackActionNode(nodeConfig.Name), nil

	case "Move To Position":
		// 移动到位置行为节点
		x, y, z := 0.0, 0.0, 0.0
		if pos, ok := nodeConfig.Params["position"].(map[string]interface{}); ok {
			if xVal, ok := pos["x"].(float64); ok {
				x = xVal
			}
			if yVal, ok := pos["y"].(float64); ok {
				y = yVal
			}
			if zVal, ok := pos["z"].(float64); ok {
				z = zVal
			}
		}
		return NewMoveToActionNode(nodeConfig.Name, x, y, z), nil

	case "Cast Skill":
		// 释放技能行为节点
		skillID := int32(1001) // 默认技能ID
		if id, ok := nodeConfig.Params["skill_id"].(float64); ok {
			skillID = int32(id)
		}
		return NewCastSkillActionNode(nodeConfig.Name, skillID), nil

	case "Retreat":
		// 撤退行为节点
		return NewRetreatActionNode(nodeConfig.Name), nil

	case "Patrol":
		// 巡逻行为节点
		return NewPatrolActionNode(nodeConfig.Name), nil

	case "Wait":
		// 等待行为节点
		duration := time.Second * 1 // 默认等待1秒
		if dur, ok := nodeConfig.Params["duration"].(string); ok {
			if parsed, err := time.ParseDuration(dur); err == nil {
				duration = parsed
			}
		}
		return NewWaitActionNode(nodeConfig.Name, duration), nil

	default:
		// 创建通用行为节点
		logger.Warnf("Unknown action node type: %s, creating generic action", nodeConfig.Name)
		return NewGenericActionNode(nodeConfig.Name, nodeConfig.Params), nil
	}
}

// createConditionNode 创建条件节点
func (ai *BossAIManager) createConditionNode(nodeConfig BehaviorNodeConfig) (IBehaviorNode, error) {
	switch nodeConfig.Name {
	case "Low Health Check":
		// 低血量检查条件节点
		threshold := 0.3 // 默认30%血量
		if t, ok := nodeConfig.Params["health_threshold"].(float64); ok {
			threshold = t
		}
		return NewHealthConditionNode(nodeConfig.Name, threshold), nil

	case "Enemy In Range":
		// 敌人在范围内条件节点
		rangeValue := 100.0 // 默认100单位范围
		if r, ok := nodeConfig.Params["range"].(float64); ok {
			rangeValue = r
		}
		return NewEnemyInRangeConditionNode(nodeConfig.Name, rangeValue), nil

	case "Skill Available":
		// 技能可用条件节点
		skillID := int32(1001) // 默认技能ID
		if id, ok := nodeConfig.Params["skill_id"].(float64); ok {
			skillID = int32(id)
		}
		return NewSkillAvailableConditionNode(nodeConfig.Name, skillID), nil

	case "Combat State":
		// 战斗状态条件节点
		return NewCombatStateConditionNode(nodeConfig.Name), nil

	case "Distance Check":
		// 距离检查条件节点
		distance := 50.0        // 默认距离
		operator := "less_than" // 默认操作符
		if d, ok := nodeConfig.Params["distance"].(float64); ok {
			distance = d
		}
		if op, ok := nodeConfig.Params["operator"].(string); ok {
			operator = op
		}
		return NewDistanceConditionNode(nodeConfig.Name, distance, operator), nil

	case "Phase Check":
		// 阶段检查条件节点
		phaseID := int32(1) // 默认阶段ID
		if id, ok := nodeConfig.Params["phase_id"].(float64); ok {
			phaseID = int32(id)
		}
		return NewPhaseConditionNode(nodeConfig.Name, phaseID), nil

	default:
		// 创建通用条件节点
		logger.Warnf("Unknown condition node type: %s, creating generic condition", nodeConfig.Name)
		return NewGenericConditionNode(nodeConfig.Name, nodeConfig.Params), nil
	}
}

// createDecoratorNode 创建装饰节点
func (ai *BossAIManager) createDecoratorNode(nodeConfig BehaviorNodeConfig) (IBehaviorNode, error) {
	if len(nodeConfig.Children) == 0 {
		return nil, fmt.Errorf("decorator node %s must have exactly one child", nodeConfig.Name)
	}

	// 递归构建子节点
	child, err := ai.buildNodeFromConfig(nodeConfig.Children[0])
	if err != nil {
		return nil, fmt.Errorf("failed to build child for decorator %s: %w", nodeConfig.Name, err)
	}

	switch nodeConfig.Name {
	case "Inverter":
		// 取反装饰节点
		node := NewInverterNode(nodeConfig.Name)
		node.AddChild(child)
		return node, nil

	case "Repeater":
		// 重复装饰节点
		count := 1 // 默认重复1次
		if c, ok := nodeConfig.Params["count"].(float64); ok {
			count = int(c)
		}
		node := NewRepeaterNode(nodeConfig.Name, count)
		node.AddChild(child)
		return node, nil

	case "Retry":
		// 重试装饰节点
		maxRetries := 3 // 默认最大重试3次
		if r, ok := nodeConfig.Params["max_retries"].(float64); ok {
			maxRetries = int(r)
		}
		return NewRetryNode(nodeConfig.Name, child, maxRetries), nil

	case "Timeout":
		// 超时装饰节点
		timeout := time.Second * 5 // 默认超时5秒
		if t, ok := nodeConfig.Params["timeout"].(string); ok {
			if parsed, err := time.ParseDuration(t); err == nil {
				timeout = parsed
			}
		}
		return NewTimeoutNode(nodeConfig.Name, child, timeout), nil

	case "Cooldown":
		// 冷却装饰节点
		cooldown := time.Second * 1 // 默认冷却1秒
		if c, ok := nodeConfig.Params["cooldown"].(string); ok {
			if parsed, err := time.ParseDuration(c); err == nil {
				cooldown = parsed
			}
		}
		node := NewCooldownNode(nodeConfig.Name, cooldown)
		node.AddChild(child)
		return node, nil

	default:
		// 创建通用装饰节点
		logger.Warnf("Unknown decorator node type: %s, creating generic decorator", nodeConfig.Name)
		return NewGenericDecoratorNode(nodeConfig.Name, child, nodeConfig.Params), nil
	}
}
