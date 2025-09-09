package bossai

import (
	"fmt"
	"sync"
	"time"

	"github.com/ouyangzhongmin/gameserver/pkg/logger"
)

// StateMachine Boss状态机
type StateMachine struct {
	states        map[BossStateID]IBossState
	currentState  IBossState
	previousState IBossState
	context       *BossContext
	mutex         sync.RWMutex

	// 性能监控
	stateChangeCount int64
	lastUpdateTime   time.Time
	updateDuration   time.Duration

	// 配置
	enableLogging   bool
	maxStateHistory int
	stateHistory    []StateHistoryEntry
}

type StateHistoryEntry struct {
	StateID   BossStateID
	StateName string
	EnterTime time.Time
	ExitTime  time.Time
	Duration  time.Duration
}

// NewStateMachine 创建新的状态机
func NewStateMachine() *StateMachine {
	return &StateMachine{
		states:          make(map[BossStateID]IBossState),
		enableLogging:   true,
		maxStateHistory: 100,
		stateHistory:    make([]StateHistoryEntry, 0),
	}
}

// AddState 添加状态
func (sm *StateMachine) AddState(state IBossState) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	stateID := BossStateID(state.GetID())

	if _, exists := sm.states[stateID]; exists {
		return fmt.Errorf("state %d already exists", stateID)
	}

	sm.states[stateID] = state

	if sm.enableLogging {
		logger.Debugf("Added state: %s (ID: %d)", state.GetName(), stateID)
	}

	return nil
}

// GetState 获取指定状态
func (sm *StateMachine) GetState(stateID BossStateID) IBossState {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	return sm.states[stateID]
}

// RemoveState 移除状态
func (sm *StateMachine) RemoveState(stateID BossStateID) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	state, exists := sm.states[stateID]
	if !exists {
		return fmt.Errorf("state %d not found", stateID)
	}

	// 如果是当前状态，不能删除
	if sm.currentState != nil && BossStateID(sm.currentState.GetID()) == stateID {
		return fmt.Errorf("cannot remove current state %d", stateID)
	}

	delete(sm.states, stateID)

	if sm.enableLogging {
		logger.Debugf("Removed state: %s (ID: %d)", state.GetName(), stateID)
	}

	return nil
}

// SetInitialState 设置初始状态
func (sm *StateMachine) SetInitialState(stateID BossStateID, ctx *BossContext) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	state, exists := sm.states[stateID]
	if !exists {
		return fmt.Errorf("initial state %d not found", stateID)
	}

	sm.context = ctx
	sm.currentState = state

	// 进入初始状态
	if err := state.OnEnter(ctx); err != nil {
		return fmt.Errorf("failed to enter initial state: %w", err)
	}

	// 更新上下文
	ctx.CurrentState = state
	ctx.IsStateChanged = true

	// 记录历史
	sm.addToHistory(BossStateID(state.GetID()), state.GetName(), ctx.CurrentTime)

	if sm.enableLogging {
		logger.Debugf("Boss %d set initial state: %s", ctx.Boss.GetID(), state.GetName())
	}

	return nil
}

// Update 更新状态机
func (sm *StateMachine) Update(ctx *BossContext, deltaTime time.Duration) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	start := time.Now()
	defer func() {
		sm.updateDuration = time.Since(start)
		sm.lastUpdateTime = ctx.CurrentTime
	}()

	if sm.currentState == nil {
		return fmt.Errorf("no current state set")
	}

	// 更新上下文
	sm.context = ctx
	ctx.CurrentState = sm.currentState
	ctx.PreviousState = sm.previousState
	ctx.DeltaTime = deltaTime

	// 检查状态转换请求
	if sm.checkStateTransitionRequest(ctx) {
		// 如果处理了状态转换请求，直接返回
		return nil
	}

	// 检查状态转换
	if err := sm.checkStateTransitions(ctx); err != nil {
		return fmt.Errorf("state transition failed: %w", err)
	}

	// 更新当前状态
	if err := sm.currentState.OnUpdate(ctx, deltaTime); err != nil {
		return fmt.Errorf("state update failed: %w", err)
	}

	return nil
}

// checkStateTransitionRequest 检查并处理状态转换请求
func (sm *StateMachine) checkStateTransitionRequest(ctx *BossContext) bool {
	request := ctx.GetStateTransitionRequest()
	if request == nil {
		return false
	}

	// 清除请求
	ctx.ClearStateTransitionRequest()

	// 检查目标状态是否存在
	targetState, exists := sm.states[request.TargetState]
	if !exists {
		logger.Warnf("State transition request to non-existent state: %d", request.TargetState)
		return false
	}

	// 如果已经在目标状态，无需转换
	if sm.currentState.GetID() == int32(request.TargetState) {
		return false
	}

	// 执行状态转换
	if err := sm.ForceTransition(request.TargetState, ctx); err != nil {
		logger.Errorf("Failed to transition to state %d: %v", request.TargetState, err)
		return false
	}

	logger.Debugf("Boss %d transitioned to %s due to request: %s",
		ctx.Boss.GetID(), targetState.GetName(), request.Reason)

	return true
}

// checkStateTransitions 检查状态转换
func (sm *StateMachine) checkStateTransitions(ctx *BossContext) error {
	nextStateID := sm.currentState.GetNextState(ctx)
	currentStateID := sm.currentState.GetID()

	// 如果状态需要改变
	if nextStateID != currentStateID {
		if err := sm.TransitionTo(BossStateID(nextStateID), ctx); err != nil {
			return err
		}
	}

	return nil
}

// TransitionTo 转换到指定状态
func (sm *StateMachine) TransitionTo(stateID BossStateID, ctx *BossContext) error {
	targetState, exists := sm.states[stateID]
	if !exists {
		return fmt.Errorf("target state %d not found", stateID)
	}

	// 检查是否可以转换
	if !sm.currentState.CanTransitionTo(int32(stateID), ctx) {
		if sm.enableLogging {
			logger.Debugf("Boss %d cannot transition from %s to %s",
				ctx.Boss.GetID(), sm.currentState.GetName(), targetState.GetName())
		}
		return nil // 不是错误，只是条件不满足
	}

	// 退出当前状态
	if err := sm.currentState.OnExit(ctx); err != nil {
		return fmt.Errorf("failed to exit current state: %w", err)
	}

	// 更新历史记录
	sm.updateHistory(ctx.CurrentTime)

	// 切换状态
	sm.previousState = sm.currentState
	sm.currentState = targetState
	sm.stateChangeCount++

	// 进入新状态
	if err := targetState.OnEnter(ctx); err != nil {
		return fmt.Errorf("failed to enter new state: %w", err)
	}

	// 更新上下文
	ctx.PreviousState = sm.previousState
	ctx.CurrentState = sm.currentState
	ctx.IsStateChanged = true

	// 记录新状态历史
	sm.addToHistory(stateID, targetState.GetName(), ctx.CurrentTime)

	if sm.enableLogging {
		logger.Debugf("Boss %d transitioned from %s to %s",
			ctx.Boss.GetID(), sm.previousState.GetName(), targetState.GetName())
	}

	return nil
}

// ForceTransition 强制转换状态（不检查条件）
func (sm *StateMachine) ForceTransition(stateID BossStateID, ctx *BossContext) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	targetState, exists := sm.states[stateID]
	if !exists {
		return fmt.Errorf("target state %d not found", stateID)
	}

	if sm.currentState != nil {
		// 退出当前状态
		if err := sm.currentState.OnExit(ctx); err != nil {
			logger.Errorf("Failed to exit current state during force transition: %v", err)
		}

		// 更新历史记录
		sm.updateHistory(ctx.CurrentTime)

		sm.previousState = sm.currentState
	}

	// 切换状态
	sm.currentState = targetState
	sm.stateChangeCount++

	// 进入新状态
	if err := targetState.OnEnter(ctx); err != nil {
		return fmt.Errorf("failed to enter new state: %w", err)
	}

	// 更新上下文
	ctx.PreviousState = sm.previousState
	ctx.CurrentState = sm.currentState
	ctx.IsStateChanged = true

	// 记录新状态历史
	sm.addToHistory(stateID, targetState.GetName(), ctx.CurrentTime)

	if sm.enableLogging {
		logger.Debugf("Boss %d force transitioned to %s", ctx.Boss.GetID(), targetState.GetName())
	}

	return nil
}

// GetCurrentState 获取当前状态
func (sm *StateMachine) GetCurrentState() IBossState {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	return sm.currentState
}

// GetPreviousState 获取前一个状态
func (sm *StateMachine) GetPreviousState() IBossState {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	return sm.previousState
}

// GetAllStates 获取所有状态
func (sm *StateMachine) GetAllStates() map[BossStateID]IBossState {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	// 返回副本以避免并发修改
	result := make(map[BossStateID]IBossState)
	for id, state := range sm.states {
		result[id] = state
	}

	return result
}

// IsInState 检查是否在指定状态
func (sm *StateMachine) IsInState(stateID BossStateID) bool {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	return sm.currentState != nil && BossStateID(sm.currentState.GetID()) == stateID
}

// addToHistory 添加到历史记录
func (sm *StateMachine) addToHistory(stateID BossStateID, stateName string, enterTime time.Time) {
	entry := StateHistoryEntry{
		StateID:   stateID,
		StateName: stateName,
		EnterTime: enterTime,
	}

	sm.stateHistory = append(sm.stateHistory, entry)

	// 限制历史记录长度
	if len(sm.stateHistory) > sm.maxStateHistory {
		sm.stateHistory = sm.stateHistory[1:]
	}
}

// updateHistory 更新最后一条历史记录的退出时间
func (sm *StateMachine) updateHistory(exitTime time.Time) {
	if len(sm.stateHistory) > 0 {
		lastEntry := &sm.stateHistory[len(sm.stateHistory)-1]
		lastEntry.ExitTime = exitTime
		lastEntry.Duration = exitTime.Sub(lastEntry.EnterTime)
	}
}

// GetStateHistory 获取状态历史
func (sm *StateMachine) GetStateHistory() []StateHistoryEntry {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	// 返回副本
	result := make([]StateHistoryEntry, len(sm.stateHistory))
	copy(result, sm.stateHistory)

	return result
}

// GetStatistics 获取状态机统计信息
func (sm *StateMachine) GetStatistics() StateMachineStats {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	stats := StateMachineStats{
		StateChangeCount: sm.stateChangeCount,
		LastUpdateTime:   sm.lastUpdateTime,
		UpdateDuration:   sm.updateDuration,
		CurrentStateID:   -1,
		CurrentStateName: "",
		TotalStates:      len(sm.states),
	}

	if sm.currentState != nil {
		stats.CurrentStateID = int32(sm.currentState.GetID())
		stats.CurrentStateName = sm.currentState.GetName()
	}

	// 计算状态持续时间统计
	if len(sm.stateHistory) > 0 {
		totalDuration := time.Duration(0)
		for _, entry := range sm.stateHistory {
			if !entry.ExitTime.IsZero() {
				totalDuration += entry.Duration
			}
		}
		stats.AverageStateDuration = totalDuration / time.Duration(len(sm.stateHistory))
	}

	return stats
}

type StateMachineStats struct {
	StateChangeCount     int64         `json:"state_change_count"`
	LastUpdateTime       time.Time     `json:"last_update_time"`
	UpdateDuration       time.Duration `json:"update_duration"`
	CurrentStateID       int32         `json:"current_state_id"`
	CurrentStateName     string        `json:"current_state_name"`
	TotalStates          int           `json:"total_states"`
	AverageStateDuration time.Duration `json:"average_state_duration"`
}

// Reset 重置状态机
func (sm *StateMachine) Reset() {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	if sm.currentState != nil && sm.context != nil {
		sm.currentState.OnExit(sm.context)
	}

	sm.currentState = nil
	sm.previousState = nil
	sm.context = nil
	sm.stateChangeCount = 0
	sm.stateHistory = sm.stateHistory[:0] // 清空但保持容量
}

// Destroy 销毁状态机
func (sm *StateMachine) Destroy() {
	sm.Reset()
	sm.states = nil
}
