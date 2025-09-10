package bossai

import (
	"fmt"
	"sync"
	"time"

	"github.com/ouyangzhongmin/gameserver/pkg/logger"
)

// BossPhase Boss阶段实现
type BossPhase struct {
	id               int32
	name             string
	description      string
	triggerCondition PhaseCondition

	// 阶段配置
	availableSkills []int32
	stateModifiers  map[string]interface{}

	// 阶段状态
	isActive  bool
	enterTime time.Time
	duration  time.Duration

	// 阶段事件
	onEnterHandlers  []PhaseEventHandler
	onExitHandlers   []PhaseEventHandler
	onUpdateHandlers []PhaseUpdateHandler

	// 阶段特殊逻辑
	customLogic map[string]interface{}

	mutex sync.RWMutex
}

type PhaseEventHandler func(phase *BossPhase, ctx *BossContext) error
type PhaseUpdateHandler func(phase *BossPhase, ctx *BossContext, deltaTime time.Duration) error

// NewBossPhase 创建Boss阶段
func NewBossPhase(id int32, name string) *BossPhase {
	return &BossPhase{
		id:               id,
		name:             name,
		availableSkills:  make([]int32, 0),
		stateModifiers:   make(map[string]interface{}),
		onEnterHandlers:  make([]PhaseEventHandler, 0),
		onExitHandlers:   make([]PhaseEventHandler, 0),
		onUpdateHandlers: make([]PhaseUpdateHandler, 0),
		customLogic:      make(map[string]interface{}),
	}
}

func (p *BossPhase) GetID() int32 {
	return p.id
}

func (p *BossPhase) GetName() string {
	return p.name
}

func (p *BossPhase) GetTriggerCondition() PhaseCondition {
	return p.triggerCondition
}

func (p *BossPhase) SetTriggerCondition(condition PhaseCondition) {
	p.triggerCondition = condition
}

func (p *BossPhase) SetDescription(description string) {
	p.description = description
}

func (p *BossPhase) SetAvailableSkills(skillIDs []int32) {
	p.availableSkills = make([]int32, len(skillIDs))
	copy(p.availableSkills, skillIDs)
}

func (p *BossPhase) AddAvailableSkill(skillID int32) {
	p.availableSkills = append(p.availableSkills, skillID)
}

func (p *BossPhase) GetAvailableSkills() []int32 {
	return p.availableSkills
}

func (p *BossPhase) SetStateModifier(key string, value interface{}) {
	p.stateModifiers[key] = value
}

func (p *BossPhase) GetStateModifiers() map[string]interface{} {
	return p.stateModifiers
}

func (p *BossPhase) AddOnEnterHandler(handler PhaseEventHandler) {
	p.onEnterHandlers = append(p.onEnterHandlers, handler)
}

func (p *BossPhase) AddOnExitHandler(handler PhaseEventHandler) {
	p.onExitHandlers = append(p.onExitHandlers, handler)
}

func (p *BossPhase) AddOnUpdateHandler(handler PhaseUpdateHandler) {
	p.onUpdateHandlers = append(p.onUpdateHandlers, handler)
}

func (p *BossPhase) SetCustomLogic(key string, value interface{}) {
	p.customLogic[key] = value
}

func (p *BossPhase) GetCustomLogic(key string) interface{} {
	return p.customLogic[key]
}

// OnEnter 进入阶段
func (p *BossPhase) OnEnter(ctx *BossContext) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.isActive = true
	p.enterTime = ctx.CurrentTime

	logger.Debugf("Boss %d entering phase: %s", ctx.Boss.GetID(), p.name)

	// 执行进入处理器
	for _, handler := range p.onEnterHandlers {
		if err := handler(p, ctx); err != nil {
			logger.Errorf("Phase %s enter handler failed: %v", p.name, err)
			continue
		}
	}

	// 应用状态修改器
	p.applyStateModifiers(ctx)

	return nil
}

// OnExit 退出阶段
func (p *BossPhase) OnExit(ctx *BossContext) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if !p.isActive {
		return nil
	}

	p.duration = ctx.CurrentTime.Sub(p.enterTime)
	p.isActive = false

	logger.Debugf("Boss %d exiting phase: %s (duration: %v)", ctx.Boss.GetID(), p.name, p.duration)

	// 执行退出处理器
	for _, handler := range p.onExitHandlers {
		if err := handler(p, ctx); err != nil {
			logger.Errorf("Phase %s exit handler failed: %v", p.name, err)
			continue
		}
	}

	// 移除状态修改器
	p.removeStateModifiers(ctx)

	return nil
}

// Update 更新阶段
func (p *BossPhase) Update(ctx *BossContext, deltaTime time.Duration) error {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	if !p.isActive {
		return nil
	}

	// 执行更新处理器
	for _, handler := range p.onUpdateHandlers {
		if err := handler(p, ctx, deltaTime); err != nil {
			logger.Errorf("Phase %s update handler failed: %v", p.name, err)
			continue
		}
	}

	return nil
}

// IsActive 检查阶段是否激活
func (p *BossPhase) IsActive() bool {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	return p.isActive
}

// GetTimeInPhase 获取在阶段中的时间
func (p *BossPhase) GetTimeInPhase(currentTime time.Time) time.Duration {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	if !p.isActive {
		return p.duration
	}

	return currentTime.Sub(p.enterTime)
}

// CanTrigger 检查是否可以触发阶段
func (p *BossPhase) CanTrigger(ctx *BossContext) bool {
	return p.evaluateCondition(p.triggerCondition, ctx)
}

// evaluateCondition 评估条件
func (p *BossPhase) evaluateCondition(condition PhaseCondition, ctx *BossContext) bool {
	switch condition.Type {
	case CondHealthPercent:
		threshold, ok := condition.Params["threshold"].(float64)
		if !ok {
			return false
		}

		currentPercent := float64(ctx.Boss.GetCurrentLife()) / float64(ctx.Boss.GetMaxLife())
		operator, ok := condition.Params["operator"].(string)
		if !ok {
			operator = "less_than"
		}

		switch operator {
		case "less_than":
			return currentPercent < threshold
		case "greater_than":
			return currentPercent > threshold
		case "equal":
			return currentPercent == threshold
		}

	case CondTimeElapsed:
		duration, ok := condition.Params["duration"].(time.Duration)
		if !ok {
			return false
		}

		return ctx.CombatTime >= duration

	case CondSkillUsed:
		skillID, ok := condition.Params["skill_id"].(int32)
		if !ok {
			return false
		}

		count, exists := ctx.SkillsUsed[skillID]
		if !exists {
			return false
		}

		minCount, ok := condition.Params["min_count"].(int)
		if !ok {
			minCount = 1
		}

		return count >= minCount

	case CondTargetCount:
		minCount, ok := condition.Params["min_count"].(int)
		if !ok {
			return false
		}

		return len(ctx.NearbyEnemies) >= minCount

	case CondCustomScript:
		// 自定义脚本逻辑
		scriptName, ok := condition.Params["script"].(string)
		if !ok {
			return false
		}

		return p.evaluateCustomScript(scriptName, ctx)
	}

	// 处理复合条件
	if len(condition.SubConds) > 0 {
		results := make([]bool, len(condition.SubConds))
		for i, subCond := range condition.SubConds {
			results[i] = p.evaluateCondition(subCond, ctx)
		}

		switch condition.Operator {
		case OpAnd:
			for _, result := range results {
				if !result {
					return false
				}
			}
			return true
		case OpOr:
			for _, result := range results {
				if result {
					return true
				}
			}
			return false
		case OpNot:
			if len(results) > 0 {
				return !results[0]
			}
		}
	}

	return false
}

// evaluateCustomScript 评估自定义脚本
func (p *BossPhase) evaluateCustomScript(scriptName string, ctx *BossContext) bool {
	logger.Debugf("Custom script evaluation for phase %s: %s", p.name, scriptName)
	// 这里可以集成脚本引擎
	return false
}

// applyStateModifiers 应用状态修改器
func (p *BossPhase) applyStateModifiers(ctx *BossContext) {
	for key, value := range p.stateModifiers {
		logger.Debugf("Applying phase modifier %s: %v", key, value)

		// 将修改器应用到Boss上下文
		if ctx.CustomData == nil {
			ctx.CustomData = make(map[string]interface{})
		}

		phaseModKey := fmt.Sprintf("phase_%d_%s", p.id, key)
		ctx.CustomData[phaseModKey] = value
	}

	// 调用Boss实体的OnPhaseEnter方法
	if ctx.Boss != nil {
		ctx.Boss.OnPhaseEnter(p)
	}
}

// removeStateModifiers 移除状态修改器
func (p *BossPhase) removeStateModifiers(ctx *BossContext) {
	for key := range p.stateModifiers {
		phaseModKey := fmt.Sprintf("phase_%d_%s", p.id, key)
		if ctx.CustomData != nil {
			delete(ctx.CustomData, phaseModKey)
		}
	}
}

// GetStatistics 获取阶段统计
func (p *BossPhase) GetStatistics() PhaseStatistics {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	stats := PhaseStatistics{
		ID:          p.id,
		Name:        p.name,
		Description: p.description,
		IsActive:    p.isActive,
		Duration:    p.duration,
	}

	if p.isActive {
		stats.TimeInPhase = time.Since(p.enterTime)
	} else {
		stats.TimeInPhase = p.duration
	}

	return stats
}

type PhaseStatistics struct {
	ID          int32         `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	IsActive    bool          `json:"is_active"`
	TimeInPhase time.Duration `json:"time_in_phase"`
	Duration    time.Duration `json:"duration"`
}

// PhaseManager 阶段管理器
type PhaseManager struct {
	phases        map[int32]*BossPhase
	phaseOrder    []int32 // 阶段检查顺序
	currentPhase  *BossPhase
	previousPhase *BossPhase
	context       *BossContext

	// 阶段转换
	isTransitioning     bool
	transitionStartTime time.Time
	transitionDuration  time.Duration

	// 统计信息
	phaseTransitions int64
	totalPhaseTime   time.Duration

	mutex sync.RWMutex
}

// NewPhaseManager 创建阶段管理器
func NewPhaseManager() *PhaseManager {
	return &PhaseManager{
		phases:             make(map[int32]*BossPhase),
		phaseOrder:         make([]int32, 0),
		transitionDuration: time.Millisecond * 500,
	}
}

// AddPhase 添加阶段
func (pm *PhaseManager) AddPhase(phase *BossPhase) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	if _, exists := pm.phases[phase.id]; exists {
		return fmt.Errorf("phase %d already exists", phase.id)
	}

	pm.phases[phase.id] = phase
	pm.phaseOrder = append(pm.phaseOrder, phase.id)

	return nil
}

// RemovePhase 移除阶段
func (pm *PhaseManager) RemovePhase(phaseID int32) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	phase, exists := pm.phases[phaseID]
	if !exists {
		return fmt.Errorf("phase %d not found", phaseID)
	}

	// 如果是当前阶段，不能删除
	if pm.currentPhase != nil && pm.currentPhase.id == phaseID {
		return fmt.Errorf("cannot remove current phase %d", phaseID)
	}

	// 如果阶段正在运行，先退出
	if phase.IsActive() {
		phase.OnExit(pm.context)
	}

	delete(pm.phases, phaseID)

	// 从顺序列表中移除
	for i, id := range pm.phaseOrder {
		if id == phaseID {
			pm.phaseOrder = append(pm.phaseOrder[:i], pm.phaseOrder[i+1:]...)
			break
		}
	}

	return nil
}

// GetPhase 获取阶段
func (pm *PhaseManager) GetPhase(phaseID int32) (*BossPhase, error) {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	phase, exists := pm.phases[phaseID]
	if !exists {
		return nil, fmt.Errorf("phase %d not found", phaseID)
	}

	return phase, nil
}

// GetCurrentPhase 获取当前阶段
func (pm *PhaseManager) GetCurrentPhase() *BossPhase {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	return pm.currentPhase
}

// GetPreviousPhase 获取前一个阶段
func (pm *PhaseManager) GetPreviousPhase() *BossPhase {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	return pm.previousPhase
}

// SetInitialPhase 设置初始阶段
func (pm *PhaseManager) SetInitialPhase(phaseID int32, ctx *BossContext) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	phase, exists := pm.phases[phaseID]
	if !exists {
		return fmt.Errorf("initial phase %d not found", phaseID)
	}

	pm.context = ctx
	pm.currentPhase = phase

	// 进入初始阶段
	if err := phase.OnEnter(ctx); err != nil {
		return fmt.Errorf("failed to enter initial phase: %w", err)
	}

	// 更新上下文
	ctx.CurrentPhase = phase
	ctx.IsPhaseChanged = true

	logger.Debugf("Boss %d set initial phase: %s", ctx.Boss.GetID(), phase.name)

	return nil
}

// Update 更新阶段管理器
func (pm *PhaseManager) Update(ctx *BossContext, deltaTime time.Duration) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	pm.context = ctx

	// 检查阶段转换
	if !pm.isTransitioning {
		if err := pm.checkPhaseTransitions(ctx); err != nil {
			return fmt.Errorf("phase transition check failed: %w", err)
		}
	} else {
		// 处理转换过程
		if ctx.CurrentTime.Sub(pm.transitionStartTime) >= pm.transitionDuration {
			pm.isTransitioning = false
		}
	}

	// 更新当前阶段
	if pm.currentPhase != nil && pm.currentPhase.IsActive() {
		if err := pm.currentPhase.Update(ctx, deltaTime); err != nil {
			return fmt.Errorf("phase update failed: %w", err)
		}
	}

	return nil
}

// checkPhaseTransitions 检查阶段转换
func (pm *PhaseManager) checkPhaseTransitions(ctx *BossContext) error {
	// 按顺序检查所有阶段
	for _, phaseID := range pm.phaseOrder {
		phase := pm.phases[phaseID]

		// 跳过当前阶段
		if pm.currentPhase != nil && phase.id == pm.currentPhase.id {
			continue
		}

		// 检查是否可以触发
		if phase.CanTrigger(ctx) {
			return pm.TransitionTo(phaseID, ctx)
		}
	}

	return nil
}

// TransitionTo 转换到指定阶段
func (pm *PhaseManager) TransitionTo(phaseID int32, ctx *BossContext) error {
	targetPhase, exists := pm.phases[phaseID]
	if !exists {
		return fmt.Errorf("target phase %d not found", phaseID)
	}

	// 如果已经在目标阶段
	if pm.currentPhase != nil && pm.currentPhase.id == phaseID {
		return nil
	}

	logger.Debugf("Boss %d transitioning to phase: %s", ctx.Boss.GetID(), targetPhase.name)

	// 开始转换过程
	pm.isTransitioning = true
	pm.transitionStartTime = ctx.CurrentTime
	pm.phaseTransitions++

	// 退出当前阶段
	if pm.currentPhase != nil {
		if err := pm.currentPhase.OnExit(ctx); err != nil {
			return fmt.Errorf("failed to exit current phase: %w", err)
		}

		pm.previousPhase = pm.currentPhase
	}

	// 进入新阶段
	pm.currentPhase = targetPhase
	if err := targetPhase.OnEnter(ctx); err != nil {
		return fmt.Errorf("failed to enter new phase: %w", err)
	}

	// 更新上下文
	ctx.PreviousPhase = pm.previousPhase
	ctx.CurrentPhase = pm.currentPhase
	ctx.IsPhaseChanged = true

	return nil
}

// ForceTransition 强制转换阶段
func (pm *PhaseManager) ForceTransition(phaseID int32, ctx *BossContext) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	return pm.TransitionTo(phaseID, ctx)
}

// IsTransitioning 检查是否正在转换
func (pm *PhaseManager) IsTransitioning() bool {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	return pm.isTransitioning
}

// SetPhaseOrder 设置阶段检查顺序
func (pm *PhaseManager) SetPhaseOrder(phaseOrder []int32) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	// 验证所有阶段都存在
	for _, phaseID := range phaseOrder {
		if _, exists := pm.phases[phaseID]; !exists {
			return fmt.Errorf("phase %d not found", phaseID)
		}
	}

	pm.phaseOrder = make([]int32, len(phaseOrder))
	copy(pm.phaseOrder, phaseOrder)

	return nil
}

// GetAllPhases 获取所有阶段
func (pm *PhaseManager) GetAllPhases() map[int32]*BossPhase {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	// 返回副本
	result := make(map[int32]*BossPhase)
	for id, phase := range pm.phases {
		result[id] = phase
	}

	return result
}

// GetStatistics 获取阶段管理器统计
func (pm *PhaseManager) GetStatistics() PhaseManagerStatistics {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	stats := PhaseManagerStatistics{
		PhaseTransitions: pm.phaseTransitions,
		TotalPhaseTime:   pm.totalPhaseTime,
		IsTransitioning:  pm.isTransitioning,
		TotalPhases:      len(pm.phases),
	}

	if pm.currentPhase != nil {
		stats.CurrentPhaseID = pm.currentPhase.id
		stats.CurrentPhaseName = pm.currentPhase.name
		stats.TimeInCurrentPhase = pm.currentPhase.GetTimeInPhase(time.Now())
	}

	if pm.previousPhase != nil {
		stats.PreviousPhaseID = pm.previousPhase.id
		stats.PreviousPhaseName = pm.previousPhase.name
	}

	return stats
}

type PhaseManagerStatistics struct {
	PhaseTransitions   int64         `json:"phase_transitions"`
	TotalPhaseTime     time.Duration `json:"total_phase_time"`
	IsTransitioning    bool          `json:"is_transitioning"`
	TotalPhases        int           `json:"total_phases"`
	CurrentPhaseID     int32         `json:"current_phase_id"`
	CurrentPhaseName   string        `json:"current_phase_name"`
	TimeInCurrentPhase time.Duration `json:"time_in_current_phase"`
	PreviousPhaseID    int32         `json:"previous_phase_id"`
	PreviousPhaseName  string        `json:"previous_phase_name"`
}

// Reset 重置阶段管理器
func (pm *PhaseManager) Reset() {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	// 退出当前阶段
	if pm.currentPhase != nil && pm.context != nil {
		pm.currentPhase.OnExit(pm.context)
	}

	pm.currentPhase = nil
	pm.previousPhase = nil
	pm.context = nil
	pm.isTransitioning = false
	pm.phaseTransitions = 0
	pm.totalPhaseTime = 0

	// 重置所有阶段
	for _, phase := range pm.phases {
		if phase.IsActive() && pm.context != nil {
			phase.OnExit(pm.context)
		}
	}
}

// Destroy 销毁阶段管理器
func (pm *PhaseManager) Destroy() {
	pm.Reset()
	pm.phases = nil
}
