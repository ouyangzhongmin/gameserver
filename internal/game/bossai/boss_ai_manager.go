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

	// 配置插件
	if err := ai.configurePlugins(); err != nil {
		return fmt.Errorf("failed to configure plugins: %w", err)
	}

	logger.Debugf("Configuration applied for Boss %s", ai.config.Name)
	return nil
}

// configureStateMachine 配置状态机
func (ai *BossAIManager) configureStateMachine() error {
	// 根据配置创建状态
	for _, stateConfig := range ai.config.States {
		state, err := ai.createStateFromConfig(stateConfig)
		if err != nil {
			logger.Errorf("Failed to create state %s: %v", stateConfig.Name, err)
			continue
		}

		// 添加到状态机
		if err := ai.stateMachine.AddState(state); err != nil {
			logger.Errorf("Failed to add state %s: %v", state.GetName(), err)
		}

		logger.Debugf("Added state: %s (ID: %d)", stateConfig.Name, stateConfig.ID)
	}

	// 设置状态转换关系
	for _, stateConfig := range ai.config.States {
		state := ai.stateMachine.GetState(stateConfig.ID)
		if state == nil {
			continue
		}

		// 配置状态转换
		for _, transition := range stateConfig.Transitions {
			if baseState, ok := state.(*BaseBossState); ok {
				baseState.AddTransition(transition.ToState, transition)
				logger.Debugf("Added transition from %s to %d (priority: %d)",
					stateConfig.Name, transition.ToState, transition.Priority)
			}
		}
	}

	// 设置初始状态
	if err := ai.stateMachine.SetInitialState(ai.config.InitialState, ai.context); err != nil {
		return err
	}

	return nil
}

// createStateFromConfig 根据配置创建状态
func (ai *BossAIManager) createStateFromConfig(config StateConfig) (IBossState, error) {
	// 根据状态ID和名称创建对应状态
	var state IBossState

	switch config.ID {
	case StateIdle:
		state = NewIdleState()
	case StateChase:
		state = NewChaseState()
	case StateAttack:
		state = NewAttac		// default:
		// 	// 创建通用状态
		// 	state = NewBaseBossState(config.ID, config.Name)
	}

	if state == nil {
		return nil, fmt.Errorf("failed to create state for ID %d", config.ID)
	}

	// 应用 modifiers
	if baseState, ok := state.(*BaseBossState); ok {
		for key, value := range config.Modifiers {
			baseState.SetModifier(key, value)
		}
		// 设置行为配置
		baseState.SetBehaviorConfigs(config.Behaviors)
		// 创建状态对应的行为树
		behaviorTree, err := ai.createBehaviorTreeForState(config.Name, config.Behaviors)
		if err != nil {
			logger.Errorf("Failed to create behavior tree for state %s: %v", config.Name, err)
		} else {
			baseState.SetBehaviorTree(behaviorTree)
		}
	}

	return state, nil
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

	// 更新附近实体（使用更大的搜索半径以包含更多敌人）
	// 这样后续的BossContext方法可以从中筛选不同范围的敌人
	entites := ai.boss.GetEntitesInRange(10)
	ai.context.NearbyEnemies = make([]IEntity, 0)
	ai.context.NearbyAllies = make([]IEntity, 0)
	for _, entity := range entites {
		if ai.context.Boss.IsEnemy(entity) { //敌人列表
			ai.context.NearbyEnemies = append(ai.context.NearbyEnemies, entity)
		} else if ai.context.Boss.IsAlly(entity) { // 如果有盟友系统
			ai.context.NearbyAllies = append(ai.context.NearbyAllies, entity)
		}
	}

	// 更新位置
	ai.context.Position = ai.boss.GetPos()

	// 注意：不再主动从 Monster 系统获取目标
	// Boss AI 系统应该通过自己的决策逻辑（状态机、行为树等）来选择和设置目标
	// 如果当前目标已死亡，则清除目标，由 AI 系统自行决定下一个目标
	if ai.context.Target != nil && !ai.context.Target.IsAlive() {
		ai.context.Target = nil
		// 同时清除 Monster 系统的目标
		ai.boss.SetCombatTarget(nil)
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

// createBehaviorTreeForState 从行为名称列表创建行为树
func (ai *BossAIManager) createBehaviorTreeForState(stateName string, behaviors []string) (*BehaviorTree, error) {
	if len(behaviors) == 0 {
		return nil, nil
	}

	treeName := fmt.Sprintf("%s_BehaviorTree", stateName)
	tree := NewBehaviorTree(treeName)

	// 创建顶层选择节点，让状态可以执行多种行为
	rootSelector := NewSelectorNode(fmt.Sprintf("%s_Root", stateName))

	// 为每个行为创建节点
	for _, behaviorName := range behaviors {
// createNodeFromBehaviorName 从行为名称创建节点
func (ai *BossAIManager) createNodeFromBehaviorName(behaviorName string) (IBehaviorNode, error) {
	switch behaviorName {
	case "scan_enemies":
		return NewScanEnemiesActionNode("Scan Enemies"), nil
	case "random_move":
		return NewRandomMoveActionNode("Random Move"), nil
	case "random_speech":
		return NewRandomSpeechActionNode("Random Speech"), nil
	case "random_action":
		return NewRandomChanceConditionNode("Random Action Chance"), nil
	case "chase_target":
		return NewChaseTargetActionNode("Chase Target"), nil
	case "basic_attack":
		return NewBasicAttackActionNode("Basic Attack"), nil
	case "return_spawn":
		return NewReturnToSpawnActionNode("Return To Spawn"), nil
	case "auto_recover":
		return NewAutoRecoverActionNode("Auto Recover"), nil
	case "trigger_reward":
		return NewTriggerRewardActionNode("Trigger Reward"), nil
	default:
		return nil, fmt.Errorf("unknown behavior name: %bsystems中的状态机更新处理
	return nil
}
iggerRewardActionNode("Trigger Reward"), nil
	default:
		return nil, fmt.Errorf("unknown behavior name: %s", behaviorName)
	}
}

// executeAI 执行AI决策 - 新架构中主要由状态机管理
func (ai *BossAIManager) executeAI() error {
	// 新架构中，行为由状态机中的各个状态的行为树执行
	// 这里不再直接管理行为，由updateSubsystems中的状态机更新处理
	return nil
}

	case "return_spawn":
		return NewReturnToSpawnActionNode("Return To Spawn"), nil
	case "auto_recover":
		return NewAutoRecoverActionNode("Auto Recover"), nil
	case "trigger_reward":
		return NewTriggerRewardActionNode("Trigger Reward"), nil
	default:
		return nil, fmt.Errorf("unknown behavior name: %s", behaviorName)
	}
}
// executeAI 执行AI决策 - 新架构中主要由状态机管理
func (ai *BossAIManager) executeAI() error {
	// 新架构中，行为由状态机中的各个状态的行为树执行
	// 这里不再直接管理行为，由updateSubsystems中的状态机更新处理
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
		ai.debugInfo.PewCheckAttackRangeConditionNode("Range Check"), nil

	// Attack状态的行为
	case "basic_attack":
		return NewBasicAttackActionNode("Basic Attack"), nil
	case "combo_attack":
		return NewComboAttackSequenceNode("Combo Attack"), nil
	case "skill_usage":
		return NewSkillUsageConditionNode("Skill Usage"), nil
	case "use_item":
		return NewUseItemActionNode("Use Item"), nil
	case "escape_check":
		return NewEscapeCheckConditionNode("Escape Check"), nil

	// Retreat状态的行为
	case "return_to_spawn":
		return NewReturnToSpawnActionNode("Return To Spawn"), nil
	case "auto_recover":
		return NewAutoRecoverActionNode("Auto Recover"), nil

	// Dying状态的行为
	case "trigger_reward":
		return NewTriggerRewardActionNode("Trigger Reward"), nil

	// 通用行为
	case "enemy_detection":
		return NewEnemyDetectionConditionNode("Enemy Detection"), nil
	case "patrol_movement":
		return NewPatrolMovementActionNode("Patrol Movement"), nil
	case "target_validation":
		return NewTargetValidationConditionNode("Target Validation"), nil

	default:
		logger.Warnf("Unknown behavior name: %s", behaviorName)
		return nil, fmt.Errorf("unknown behavior: %s", behaviorName)
	}
}
