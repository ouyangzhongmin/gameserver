package bossai

import (
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/ouyangzhongmin/gameserver/pkg/coord"
	"github.com/ouyangzhongmin/gameserver/pkg/logger"
)

// BossSkill Boss技能实现
type BossSkill struct {
	id          int32
	name        string
	description string
	cooldown    time.Duration
	range_      float64
	castTime    time.Duration
	manaCost    int32

	// 冷却状态
	lastUsedTime time.Time
	isOnCooldown bool
	mutex        sync.RWMutex

	// 条件检查
	conditions []SkillCondition

	// 效果配置
	effects []SkillEffect

	// 目标选择
	targetType TargetType
	maxTargets int

	// 统计信息
	useCount     int64
	totalDamage  int64
	successCount int64

	// 自定义逻辑
	castHandler   SkillCastHandler
	effectHandler SkillEffectHandler
}

type TargetType int

const (
	TargetSelf TargetType = iota
	TargetEnemy
	TargetAlly
	TargetGround
	TargetAll
)

type SkillCondition struct {
	Type     ConditionType          `json:"type"`
	Params   map[string]interface{} `json:"params"`
	Operator LogicalOperator        `json:"operator,omitempty"`
	SubConds []SkillCondition       `json:"sub_conditions,omitempty"`
}

type SkillCastHandler func(skill *BossSkill, ctx *BossContext, targets []IEntity) error
type SkillEffectHandler func(effect SkillEffect, caster IEntity, target IEntity) error

// NewBossSkill 创建Boss技能
func NewBossSkill(id int32, name string) *BossSkill {
	return &BossSkill{
		id:         id,
		name:       name,
		cooldown:   time.Second * 5,
		range_:     100.0,
		castTime:   time.Second * 1,
		conditions: make([]SkillCondition, 0),
		effects:    make([]SkillEffect, 0),
		targetType: TargetEnemy,
		maxTargets: 1,
	}
}

func (s *BossSkill) GetID() int32 {
	return s.id
}

func (s *BossSkill) GetName() string {
	return s.name
}

func (s *BossSkill) GetCooldown() time.Duration {
	return s.cooldown
}

func (s *BossSkill) GetRange() float64 {
	return s.range_
}

func (s *BossSkill) GetCastTime() time.Duration {
	return s.castTime
}

func (s *BossSkill) SetCooldown(cooldown time.Duration) {
	s.cooldown = cooldown
}

func (s *BossSkill) SetRange(range_ float64) {
	s.range_ = range_
}

func (s *BossSkill) SetCastTime(castTime time.Duration) {
	s.castTime = castTime
}

func (s *BossSkill) SetManaCost(cost int32) {
	s.manaCost = cost
}

func (s *BossSkill) SetTargetType(targetType TargetType, maxTargets int) {
	s.targetType = targetType
	s.maxTargets = maxTargets
}

func (s *BossSkill) AddCondition(condition SkillCondition) {
	s.conditions = append(s.conditions, condition)
}

func (s *BossSkill) AddEffect(effect SkillEffect) {
	s.effects = append(s.effects, effect)
}

func (s *BossSkill) SetCastHandler(handler SkillCastHandler) {
	s.castHandler = handler
}

func (s *BossSkill) SetEffectHandler(handler SkillEffectHandler) {
	s.effectHandler = handler
}

// CanUse 检查是否可以使用技能
func (s *BossSkill) CanUse(ctx *BossContext) bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// 检查冷却
	if s.IsOnCooldown() {
		return false
	}

	// 检查蓝量（如果Boss有蓝量系统）
	// 这里暂时跳过，可以根据需要扩展

	// 检查条件
	for _, condition := range s.conditions {
		if !s.evaluateCondition(condition, ctx) {
			return false
		}
	}

	return true
}

// GetTargets 获取技能目标
func (s *BossSkill) GetTargets(ctx *BossContext) []IEntity {
	targets := make([]IEntity, 0)

	switch s.targetType {
	case TargetSelf:
		targets = append(targets, ctx.Boss)

	case TargetEnemy:
		if ctx.Target != nil {
			// 检查目标是否在技能范围内
			distance := s.calculateDistance(ctx.Boss.GetPos(), ctx.Target.GetPos())
			if distance <= s.range_ {
				targets = append(targets, ctx.Target)
			}
		}

		// 如果可以选择多个目标
		if s.maxTargets > 1 {
			for _, enemy := range ctx.NearbyEnemies {
				if len(targets) >= s.maxTargets {
					break
				}

				if enemy == ctx.Target {
					continue // 主要目标已经添加
				}

				distance := s.calculateDistance(ctx.Boss.GetPos(), enemy.GetPos())
				if distance <= s.range_ {
					targets = append(targets, enemy)
				}
			}
		}

	case TargetAlly:
		for _, ally := range ctx.NearbyAllies {
			if len(targets) >= s.maxTargets {
				break
			}

			distance := s.calculateDistance(ctx.Boss.GetPos(), ally.GetPos())
			if distance <= s.range_ {
				targets = append(targets, ally)
			}
		}

	case TargetAll:
		// 所有在范围内的实体
		allEntities := append(ctx.NearbyEnemies, ctx.NearbyAllies...)
		for _, entity := range allEntities {
			if len(targets) >= s.maxTargets {
				break
			}

			distance := s.calculateDistance(ctx.Boss.GetPos(), entity.GetPos())
			if distance <= s.range_ {
				targets = append(targets, entity)
			}
		}
	}

	return targets
}

// Cast 施放技能
func (s *BossSkill) Cast(ctx *BossContext, targets []IEntity) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if !s.CanUse(ctx) {
		return fmt.Errorf("skill %s cannot be used", s.name)
	}

	s.useCount++

	// 开始冷却
	s.StartCooldown()

	// 记录动作
	action := &AIAction{
		Type:      ActionCastSkill,
		SkillID:   s.id,
		Timestamp: ctx.CurrentTime,
		Priority:  7,
	}

	if len(targets) > 0 {
		action.TargetID = targets[0].GetID()
	}

	ctx.LastAction = action
	if ctx.ActionHistory == nil {
		ctx.ActionHistory = make([]*AIAction, 0)
	}
	ctx.ActionHistory = append(ctx.ActionHistory, action)

	// 更新技能使用统计
	if _, exists := ctx.SkillsUsed[s.id]; !exists {
		ctx.SkillsUsed[s.id] = 0
	}
	ctx.SkillsUsed[s.id]++

	logger.Debugf("Boss %d casting skill %s on %d targets", ctx.Boss.GetID(), s.name, len(targets))

	// 执行自定义施法逻辑
	if s.castHandler != nil {
		if err := s.castHandler(s, ctx, targets); err != nil {
			logger.Errorf("Skill %s cast handler failed: %v", s.name, err)
			return err
		}
	}

	// 应用技能效果
	s.successCount++
	return s.applyEffects(ctx.Boss, targets)
}

// StartCooldown 开始冷却
func (s *BossSkill) StartCooldown() {
	s.lastUsedTime = time.Now()
	s.isOnCooldown = true
}

// IsOnCooldown 检查是否在冷却中
func (s *BossSkill) IsOnCooldown() bool {
	if !s.isOnCooldown {
		return false
	}

	if time.Since(s.lastUsedTime) >= s.cooldown {
		s.isOnCooldown = false
		return false
	}

	return true
}

// GetRemainingCooldown 获取剩余冷却时间
func (s *BossSkill) GetRemainingCooldown() time.Duration {
	if !s.isOnCooldown {
		return 0
	}

	elapsed := time.Since(s.lastUsedTime)
	if elapsed >= s.cooldown {
		s.isOnCooldown = false
		return 0
	}

	return s.cooldown - elapsed
}

// evaluateCondition 评估技能条件
func (s *BossSkill) evaluateCondition(condition SkillCondition, ctx *BossContext) bool {
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

		return s.evaluateCustomScript(scriptName, ctx)
	}

	// 处理复合条件
	if len(condition.SubConds) > 0 {
		results := make([]bool, len(condition.SubConds))
		for i, subCond := range condition.SubConds {
			results[i] = s.evaluateCondition(subCond, ctx)
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
func (s *BossSkill) evaluateCustomScript(scriptName string, ctx *BossContext) bool {
	// 这里可以实现脚本引擎集成
	logger.Debugf("Custom script evaluation for skill %s: %s", s.name, scriptName)
	return false
}

// applyEffects 应用技能效果
func (s *BossSkill) applyEffects(caster IEntity, targets []IEntity) error {
	for _, target := range targets {
		for _, effect := range s.effects {
			if err := s.applyEffect(effect, caster, target); err != nil {
				logger.Errorf("Failed to apply effect %s: %v", effect.Type, err)
				continue
			}
		}
	}

	return nil
}

// applyEffect 应用单个效果
func (s *BossSkill) applyEffect(effect SkillEffect, caster IEntity, target IEntity) error {
	// 如果有自定义效果处理器
	if s.effectHandler != nil {
		return s.effectHandler(effect, caster, target)
	}

	// 默认效果处理
	switch effect.Type {
	case "damage":
		damage, ok := effect.Value.(float64)
		if !ok {
			return fmt.Errorf("invalid damage value")
		}

		logger.Debugf("Skill %s dealing %.0f damage to %d", s.name, damage, target.GetID())
		s.totalDamage += int64(damage)
		// 这里可以调用实际的伤害计算逻辑

	case "heal":
		heal, ok := effect.Value.(float64)
		if !ok {
			return fmt.Errorf("invalid heal value")
		}

		logger.Debugf("Skill %s healing %.0f to %d", s.name, heal, target.GetID())
		// 这里可以调用实际的治疗逻辑

	case "buff":
		buffType, ok := effect.Value.(string)
		if !ok {
			return fmt.Errorf("invalid buff type")
		}

		logger.Debugf("Skill %s applying buff %s to %d", s.name, buffType, target.GetID())
		// 这里可以调用实际的buff系统

	case "debuff":
		debuffType, ok := effect.Value.(string)
		if !ok {
			return fmt.Errorf("invalid debuff type")
		}

		logger.Debugf("Skill %s applying debuff %s to %d", s.name, debuffType, target.GetID())
		// 这里可以调用实际的debuff系统

	case "stun":
		duration := effect.Duration
		if duration == 0 {
			return fmt.Errorf("invalid stun duration")
		}

		logger.Debugf("Skill %s stunning %d for %v", s.name, target.GetID(), duration)
		// 这里可以调用实际的眩晕逻辑

	default:
		logger.Warnf("Unknown skill effect type: %s", effect.Type)
	}

	return nil
}

// calculateDistance 计算距离
func (s *BossSkill) calculateDistance(pos1, pos2 coord.Vector3) float64 {
	dx := float64(pos2.X - pos1.X)
	dy := float64(pos2.Y - pos1.Y)
	dz := float64(pos2.Z - pos1.Z)

	return math.Sqrt(dx*dx + dy*dy + dz*dz)
}

// GetStatistics 获取技能统计
func (s *BossSkill) GetStatistics() SkillStatistics {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	stats := SkillStatistics{
		ID:           s.id,
		Name:         s.name,
		UseCount:     s.useCount,
		SuccessCount: s.successCount,
		TotalDamage:  s.totalDamage,
		IsOnCooldown: s.IsOnCooldown(),
		Cooldown:     s.cooldown,
	}

	if s.useCount > 0 {
		stats.SuccessRate = float64(s.successCount) / float64(s.useCount)
		stats.AverageDamage = float64(s.totalDamage) / float64(s.useCount)
	}

	if s.IsOnCooldown() {
		stats.RemainingCooldown = s.GetRemainingCooldown()
	}

	return stats
}

type SkillStatistics struct {
	ID                int32         `json:"id"`
	Name              string        `json:"name"`
	UseCount          int64         `json:"use_count"`
	SuccessCount      int64         `json:"success_count"`
	SuccessRate       float64       `json:"success_rate"`
	TotalDamage       int64         `json:"total_damage"`
	AverageDamage     float64       `json:"average_damage"`
	IsOnCooldown      bool          `json:"is_on_cooldown"`
	RemainingCooldown time.Duration `json:"remaining_cooldown"`
	Cooldown          time.Duration `json:"cooldown"`
}

// SkillManager 技能管理器
type SkillManager struct {
	skills     map[int32]*BossSkill
	skillOrder []int32 // 技能优先级顺序
	mutex      sync.RWMutex
}

// NewSkillManager 创建技能管理器
func NewSkillManager() *SkillManager {
	return &SkillManager{
		skills:     make(map[int32]*BossSkill),
		skillOrder: make([]int32, 0),
	}
}

// AddSkill 添加技能
func (sm *SkillManager) AddSkill(skill *BossSkill) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	if _, exists := sm.skills[skill.id]; exists {
		return fmt.Errorf("skill %d already exists", skill.id)
	}

	sm.skills[skill.id] = skill
	sm.skillOrder = append(sm.skillOrder, skill.id)

	return nil
}

// RemoveSkill 移除技能
func (sm *SkillManager) RemoveSkill(skillID int32) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	if _, exists := sm.skills[skillID]; !exists {
		return fmt.Errorf("skill %d not found", skillID)
	}

	delete(sm.skills, skillID)

	// 从优先级列表中移除
	for i, id := range sm.skillOrder {
		if id == skillID {
			sm.skillOrder = append(sm.skillOrder[:i], sm.skillOrder[i+1:]...)
			break
		}
	}

	return nil
}

// GetSkill 获取技能
func (sm *SkillManager) GetSkill(skillID int32) (*BossSkill, error) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	skill, exists := sm.skills[skillID]
	if !exists {
		return nil, fmt.Errorf("skill %d not found", skillID)
	}

	return skill, nil
}

// GetAvailableSkills 获取可用技能列表
func (sm *SkillManager) GetAvailableSkills(ctx *BossContext) []*BossSkill {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	available := make([]*BossSkill, 0)

	for _, skillID := range sm.skillOrder {
		skill := sm.skills[skillID]
		if skill.CanUse(ctx) {
			available = append(available, skill)
		}
	}

	return available
}

// GetBestSkill 获取最佳技能（根据优先级和条件）
func (sm *SkillManager) GetBestSkill(ctx *BossContext) *BossSkill {
	available := sm.GetAvailableSkills(ctx)

	if len(available) == 0 {
		return nil
	}

	// 返回第一个可用技能（按优先级排序）
	return available[0]
}

// SetSkillPriority 设置技能优先级
func (sm *SkillManager) SetSkillPriority(skillOrder []int32) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	// 验证所有技能都存在
	for _, skillID := range skillOrder {
		if _, exists := sm.skills[skillID]; !exists {
			return fmt.Errorf("skill %d not found", skillID)
		}
	}

	sm.skillOrder = make([]int32, len(skillOrder))
	copy(sm.skillOrder, skillOrder)

	return nil
}

// GetAllSkills 获取所有技能
func (sm *SkillManager) GetAllSkills() map[int32]*BossSkill {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	// 返回副本
	result := make(map[int32]*BossSkill)
	for id, skill := range sm.skills {
		result[id] = skill
	}

	return result
}

// UpdateCooldowns 更新所有技能冷却
func (sm *SkillManager) UpdateCooldowns() {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	for _, skill := range sm.skills {
		skill.IsOnCooldown() // 这会自动更新冷却状态
	}
}

// Reset 重置所有技能冷却
func (sm *SkillManager) Reset() {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	for _, skill := range sm.skills {
		skill.isOnCooldown = false
	}
}

// GetStatistics 获取所有技能统计
func (sm *SkillManager) GetStatistics() map[int32]SkillStatistics {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	stats := make(map[int32]SkillStatistics)
	for id, skill := range sm.skills {
		stats[id] = skill.GetStatistics()
	}

	return stats
}
