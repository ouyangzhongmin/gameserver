package game

import (
	"fmt"
	"time"

	"github.com/ouyangzhongmin/gameserver/pkg/fileutil"

	"github.com/ouyangzhongmin/gameserver/pkg/shape"

	"github.com/ouyangzhongmin/gameserver/constants"
	"github.com/ouyangzhongmin/gameserver/db/model"
	"github.com/ouyangzhongmin/gameserver/internal/game/bossai"
	"github.com/ouyangzhongmin/gameserver/internal/game/object"
	"github.com/ouyangzhongmin/gameserver/pkg/coord"
	"github.com/ouyangzhongmin/gameserver/pkg/logger"
)

// BossAIManagerAdapter 兼容IAiManager接口的Boss AI适配器
// 实现了IAiManager接口，可以直接设置到Monster中
type BossAIManagerAdapter struct {
	monster    *Monster
	bossEntity *BossEntityAdapter
	aiManager  *bossai.BossAIManager
	config     *bossai.BossConfig
	lastUpdate time.Time
	updateRate time.Duration
	isActive   bool
	// 数据库中的ai配置
	aiData *model.Aiconfig
}

// NewBossAIManagerAdapter 创建兼容IAiManager的Boss AI适配器
func NewBossAIManagerAdapter(monster *Monster, aidata *model.Aiconfig, configPath string) (*BossAIManagerAdapter, error) {
	adapter := &BossAIManagerAdapter{
		monster:    monster,
		aiManager:  bossai.NewBossAIManager(),
		updateRate: time.Millisecond * 300, // 30Hz更新频率
		aiData:     aidata,
	}

	// 创建Boss实体适配器
	adapter.bossEntity = NewBossEntityAdapter(monster, aidata, adapter.aiManager)

	// 加载配置（如果提供）
	if configPath != "" {
		config, err := bossai.LoadBossConfigFromFile(configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load boss config: %w", err)
		}
		adapter.config = config

		err = adapter.aiManager.LoadConfig(config)
		if err != nil {
			return nil, fmt.Errorf("failed to apply config: %w", err)
		}
	}

	// 初始化AI管理器
	err := adapter.aiManager.Initialize(adapter.bossEntity)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize AI: %w", err)
	}

	// 启动AI
	err = adapter.aiManager.Start()
	if err != nil {
		return nil, fmt.Errorf("failed to start AI: %w", err)
	}

	adapter.isActive = true
	adapter.lastUpdate = time.Now()

	return adapter, nil
}

// 实现IAiManager接口

// update 更新AI（实现IAiManager接口）
func (a *BossAIManagerAdapter) update(curMilliSecond int64, elapsedTime int64) error {
	if !a.isActive {
		return nil
	}

	currentTime := time.Unix(0, curMilliSecond*int64(time.Millisecond))

	// 检查更新频率
	if currentTime.Sub(a.lastUpdate) < a.updateRate {
		return nil
	}

	deltaTime := currentTime.Sub(a.lastUpdate)
	a.lastUpdate = currentTime

	// 更新Boss AI
	return a.aiManager.Update(deltaTime)
}

// onBeenAttacked 处理受到攻击事件（实现IAiManager接口）
func (a *BossAIManagerAdapter) onBeenAttacked(target IMovableEntity) {
	if a.isActive {
		// 使用适配器封装攻击者
		attackerAdapter := NewEntityAdapter(target)
		a.aiManager.OnDamageReceived(attackerAdapter, 0) // 伤害值需要从其他地方获取
	}
}

// GetAiData 获取AI数据（实现IAiManager接口）
func (a *BossAIManagerAdapter) GetAiData() interface{} {
	if a.aiData != nil {
		return a.aiData
	}
	// 返回调试信息作为AI数据
	return a.aiManager.GetDebugInfo()
}

// GetOwner 获取AI拥有者（实现IAiManager接口）
func (a *BossAIManagerAdapter) GetOwner() IMovableEntity {
	return a.monster
}

// clear 清理AI（实现IAiManager接口）
func (a *BossAIManagerAdapter) clear() {
	if a.isActive {
		a.aiManager.Stop()
		a.aiManager.Destroy()
		a.isActive = false
	}
}

// Boss AI专用方法

// GetBossAIManager 获取底层的Boss AI管理器
func (a *BossAIManagerAdapter) GetBossAIManager() *bossai.BossAIManager {
	return a.aiManager
}

// GetDebugInfo 获取Boss AI调试信息
func (a *BossAIManagerAdapter) GetDebugInfo() *bossai.AIDebugInfo {
	return a.aiManager.GetDebugInfo()
}

// SetDebugEnabled 设置调试模式
func (a *BossAIManagerAdapter) SetDebugEnabled(enabled bool) {
	a.aiManager.SetDebugEnabled(enabled)
}

// TransitionTo 强制转换到指定状态
func (a *BossAIManagerAdapter) TransitionTo(stateID string) error {
	return a.aiManager.TransitionTo(stateID)
}

// AddPlugin 添加AI插件
func (a *BossAIManagerAdapter) AddPlugin(plugin bossai.IAIPlugin) error {
	return a.aiManager.AddPlugin(plugin)
}

// RemovePlugin 移除AI插件
func (a *BossAIManagerAdapter) RemovePlugin(pluginName string) error {
	return a.aiManager.RemovePlugin(pluginName)
}

// Pause 暂停AI
func (a *BossAIManagerAdapter) Pause() {
	a.aiManager.Pause()
}

// Resume 恢复AI
func (a *BossAIManagerAdapter) Resume() {
	a.aiManager.Resume()
}

// IsRunning 检查AI是否在运行
func (a *BossAIManagerAdapter) IsRunning() bool {
	return a.aiManager.IsRunning()
}

// BossEntityAdapter 将Monster适配为IBossEntity接口
type BossEntityAdapter struct {
	monster *Monster
	// 数据库中的ai配置
	aiData     *model.Aiconfig
	aiManager  *bossai.BossAIManager
	isInCombat bool
}

func NewBossEntityAdapter(monster *Monster, aiData *model.Aiconfig, aiManager *bossai.BossAIManager) *BossEntityAdapter {
	return &BossEntityAdapter{
		monster:   monster,
		aiData:    aiData,
		aiManager: aiManager,
	}
}

func (b *BossEntityAdapter) GetID() int64 {
	return b.monster.GetID()
}

func (b *BossEntityAdapter) GetPos() coord.Vector3 {
	return b.monster.GetPos()
}

func (b *BossEntityAdapter) GetEntityType() int {
	return b.monster.GetEntityType()
}

func (b *BossEntityAdapter) IsAlive() bool {
	return b.monster.IsAlive()
}

func (b *BossEntityAdapter) IsDestroyed() bool {
	return b.monster.IsDestroyed()
}

func (b *BossEntityAdapter) WalkTo(x, y, z coord.Coord) error {
	b.Walk()
	return b.monster.MoveTo(x, y, 0)
}

func (b *BossEntityAdapter) EscapeTo(x, y, z coord.Coord) error {
	b.Escape()
	return b.monster.MoveTo(x, y, 0)
}

func (b *BossEntityAdapter) ChaseTo(x, y, z coord.Coord) error {
	b.Chase()
	return b.monster.MoveTo(x, y, 0)
}

func (b *BossEntityAdapter) Stop() error {
	b.monster.Stop()
	return nil
}

func (b *BossEntityAdapter) CanAttackTarget(target bossai.IEntity) bool {
	if entityAdapter, ok := target.(*EntityAdapter); ok {
		return b.monster.CanAttackTarget(entityAdapter.entity)
	}
	return false
}

func (b *BossEntityAdapter) IsInAttackRange(x, y coord.Coord) bool {
	return b.monster.IsInAttackRange(x, y)
}

func (b *BossEntityAdapter) GetMaxLife() int32 {
	return int32(b.monster.MaxLife)
}

func (b *BossEntityAdapter) GetCurrentLife() int32 {
	return int32(b.monster.Life)
}

func (b *BossEntityAdapter) GetAttackDuration() int {
	return b.monster.GetAttackDuration()
}

// 复用Monster的移动速度计算
func (b *BossEntityAdapter) GetStepTime() int {
	return b.monster.getStepTime()
}

// 获取出生点
func (b *BossEntityAdapter) GetBornPos() coord.Vector3 {
	return b.monster.bornPos
}

// 获取最大的追击距离
func (b *BossEntityAdapter) GetMaxChaseDist() int {
	if b.aiData != nil {
		return b.aiData.ChaseRange
	}
	return 20
}

func (b *BossEntityAdapter) CanUseSkill(skillID int32) bool {
	// 复用Monster的技能检查逻辑
	spell := b.monster.GetSpell(int(skillID))
	if spell == nil {
		logger.Println("Invalid spell ID")
		return false
	}
	return spell.CurCdTime <= 0 && b.monster.Mana >= spell.Data.Mana
}

// 获取技能是否还在cd中
func (b *BossEntityAdapter) IsSkillInCD(skillID int32) bool {
	spell := b.monster.GetSpell(int(skillID))
	if spell == nil {
		logger.Println("Invalid spell ID")
		return true
	}
	return spell.CurCdTime > 0
}

// 获取指定类型的可用技能
func (b *BossEntityAdapter) GetAvailableSkill(rules string) int32 {
	// 首先尝试从当前阶段获取可用技能
	if spell := b.GetAvailableSkillFromPhase(); spell != nil {
		return int32(spell.SpellId)
	}

	// 如果当前阶段没有限制或没有可用技能，则使用默认逻辑
	spell := b.monster.GetCanUseSpell(0)
	if spell == nil {
		return 0
	}
	return int32(spell.SpellId)
}

// 从当前阶段获取可用技能
func (b *BossEntityAdapter) GetAvailableSkillFromPhase() *object.SpellObject {
	// 获取当前阶段
	if currentPhase := b.aiManager.GetCurrentPhase(); currentPhase != nil {
		// 检查阶段是否有限制可用技能
		availableSkills := currentPhase.GetAvailableSkills()
		if len(availableSkills) > 0 {
			// 从阶段可用技能中选择一个可用的技能
			for _, skillID := range availableSkills {
				if b.CanUseSkill(skillID) {
					// 通过Monster获取实际的技能对象
					return b.monster.GetSpell(int(skillID))
				}
			}
		}
	}
	return nil
}

// 复用Monster的技能范围检查
func (b *BossEntityAdapter) IsInSkillAttackRange(skillID int32, x, y coord.Coord) bool {
	spell := b.monster.GetSpell(int(skillID))
	if spell == nil {
		logger.Println("Invalid spell ID")
		return false
	}
	return b.monster.IsInSpellAttackRange(spell, x, y)
}

func (b *BossEntityAdapter) UseSkill(skillID int32, target bossai.IEntity) error {
	// 复用Monster的技能释放逻辑
	spell := b.monster.GetSpell(int(skillID))
	if spell == nil {
		logger.Println("Invalid spell ID")
		return fmt.Errorf("skill %d not available", skillID)
	}
	if entityAdapter, ok := target.(*EntityAdapter); ok {
		err := b.monster.SpellAttack(spell, entityAdapter.entity)
		if err != nil {
			return err
		}
		// 记录已使用技能
		b.aiManager.OnSkillUsed(skillID, true)
		return nil
	}

	return fmt.Errorf("skill:%d use invalid target", skillID)
}

// 复用Monster的基础攻击功能
func (b *BossEntityAdapter) DoAttackTarget(target bossai.IEntity) error {
	if entityAdapter, ok := target.(*EntityAdapter); ok {
		b.monster.doAttackTarget(entityAdapter.entity)
		return nil
	}
	return fmt.Errorf("invalid target type")
}

// 复用Monster的位置计算功能
func (b *BossEntityAdapter) GetCanAttackPos(target bossai.IEntity, offset int) (coord.Vector3, error) {
	if entityAdapter, ok := target.(*EntityAdapter); ok {
		return b.monster.GetCanAttackPos(entityAdapter.entity, offset)
	}
	return coord.Vector3{}, fmt.Errorf("invalid target type")
}

func (b *BossEntityAdapter) GetRandomPos(radius int) (coord.Vector3, error) {
	rx, ry, err := b.monster.scene.GetRandomXY(b.GetMovableRect(), 20)
	if err != nil {
		return coord.Vector3{}, err
	}
	return coord.Vector3{X: rx, Y: ry, Z: 0}, nil
}

func (b *BossEntityAdapter) GetMovableRect() shape.Rect {
	return b.monster.GetMovableRect()
}

func (b *BossEntityAdapter) IsInCombat() bool {
	// 检查是否在战斗中
	return b.isInCombat
}

func (b *BossEntityAdapter) SetInCombat(val bool) {
	b.isInCombat = val
}

func (b *BossEntityAdapter) GetCombatTarget() bossai.IEntity {
	// 获取当前目标
	if b.monster.aimgr != nil {
		// 这里需要适配原有AI系统的目标获取
		// 暂时返回nil，需要根据实际情况调整
	}
	return nil
}

func (b *BossEntityAdapter) SetCombatTarget(target bossai.IEntity) {
	// 设置战斗目标
	if entityAdapter, ok := target.(*EntityAdapter); ok {
		// 这里需要调用原有系统的目标设置方法
		// 可以通过AI管理器设置目标
		_ = entityAdapter // 暂时忽略，需要根据实际情况调整
	}
}

func (b *BossEntityAdapter) IsEnemy(entity bossai.IEntity) bool {
	return entity.GetEntityType() == constants.ENTITY_TYPE_HERO && entity.IsAlive() && !entity.IsDestroyed()
}

// 是否盟友
func (b *BossEntityAdapter) IsAlly(entity bossai.IEntity) bool {
	return entity.GetEntityType() == constants.ENTITY_TYPE_MONSTER
}

func (b *BossEntityAdapter) GetEntitesInRange(radius float64) []bossai.IEntity {
	// 获取范围内的实体
	entities := b.monster.scene.getEntitiesByRange(
		b.monster.GetPos().X,
		b.monster.GetPos().Y,
		coord.Coord(radius),
	)

	result := make([]bossai.IEntity, 0, len(entities))
	for _, entity := range entities {
		// 使用适配器封装原有实体
		if entity == b.monster {
			continue
		}
		// if !b.monster.CanAttackTarget(entity) {
		// 	continue
		// }
		result = append(result, NewEntityAdapter(entity))
	}

	return result
}

// EntityAdapter 将原有游戏实体适配为Boss AI系统的接口
type EntityAdapter struct {
	entity IMovableEntity
}

func NewEntityAdapter(entity IMovableEntity) *EntityAdapter {
	return &EntityAdapter{
		entity: entity,
	}
}

func (e *EntityAdapter) GetID() int64 {
	return e.entity.GetID()
}

func (e *EntityAdapter) GetPos() coord.Vector3 {
	return e.entity.GetPos()
}

func (e *EntityAdapter) GetEntityType() int {
	return e.entity.GetEntityType()
}

func (e *EntityAdapter) IsAlive() bool {
	// 根据实体类型判断
	switch entity := e.entity.(type) {
	case *Hero:
		return entity.IsAlive()
	case *Monster:
		return entity.IsAlive()
	default:
		return true
	}
}

func (e *EntityAdapter) IsDestroyed() bool {
	return e.entity.IsDestroyed()
}

// 添加Monster状态控制方法的适配
func (b *BossEntityAdapter) Idle() {
	b.monster.Idle()
}

func (b *BossEntityAdapter) Walk() {
	b.monster.Walk()
}

func (b *BossEntityAdapter) Run() {
	b.monster.Run()
}

func (b *BossEntityAdapter) Chase() {
	b.monster.Chase()
}

func (b *BossEntityAdapter) Escape() {
	b.monster.Escape()
}

func (b *BossEntityAdapter) AttackAction() {
	b.monster.AttackAction()
}

func (b *BossEntityAdapter) Die() {
	b.monster.Die()
}

func (b *BossEntityAdapter) IsIdle() bool {
	return b.monster.GetState() == constants.ACTION_STATE_IDLE
}
func (b *BossEntityAdapter) IsWalking() bool {
	return b.monster.GetState() == constants.ACTION_STATE_WALK
}
func (b *BossEntityAdapter) IsRunning() bool {
	return b.monster.GetState() == constants.ACTION_STATE_RUN
}
func (b *BossEntityAdapter) IsAttacking() bool {
	return b.monster.GetState() == constants.ACTION_STATE_ATTACK
}
func (b *BossEntityAdapter) IsChasing() bool {
	return b.monster.GetState() == constants.ACTION_STATE_CHASE
}
func (b *BossEntityAdapter) IsEscaping() bool {
	return b.monster.GetState() == constants.ACTION_STATE_ESCAPE
}
func (b *BossEntityAdapter) IsDied() bool {
	return b.monster.GetState() == constants.ACTION_STATE_DIE
}

// OnPhaseEnter 当进入新阶段时调用
func (b *BossEntityAdapter) OnPhaseEnter(phase bossai.IBossPhase) {
	// 获取阶段修饰器
	modifiers := phase.GetStateModifiers()

	// 创建阶段修饰器对象
	phaseModifier := &PhaseModifier{
		SkillCooldownMultiplier:  1.0,
		AttackCooldownMultiplier: 1.0,
		DamageMultiplier:         1.0,
		DefenseMultiplier:        1.0,
		Transform:                "",
	}

	// 应用修饰器
	if skillCooldownMultiplier, ok := modifiers["skill_cooldown_multiplier"].(float64); ok {
		phaseModifier.SkillCooldownMultiplier = skillCooldownMultiplier
	}
	if attackCooldownMultiplier, ok := modifiers["attack_cooldown_multiplier"].(float64); ok {
		phaseModifier.AttackCooldownMultiplier = attackCooldownMultiplier // 普通攻击
	}
	if damageMultiplier, ok := modifiers["damage_multiplier"].(float64); ok {
		phaseModifier.DamageMultiplier = damageMultiplier
	}
	if defenseMultiplier, ok := modifiers["defense_multiplier"].(float64); ok {
		phaseModifier.DefenseMultiplier = defenseMultiplier
	}
	if transform, ok := modifiers["transform"].(string); ok {
		phaseModifier.Transform = transform
	}

	// 设置阶段修饰器到Monster
	b.monster.SetPhaseModifier(phaseModifier)

	// 如果有变身，则应用变身
	if phaseModifier.Transform != "" {
		b.monster.ApplyTransform(phaseModifier.Transform)
	}
}

// CreateExampleBossMonster 创建示例Boss怪物
func CreateExampleBossMonster(scene *Scene) *Monster {
	// 创建基础怪物数据
	monsterData := &model.Monster{
		Id:         1001,
		Name:       "Fire Dragon Lord",
		Level:      50,
		BaseAttack: 800,
		BaseLife:   50000,
		// 其他属性...
	}

	// 创建Monster实例
	monster := NewMonster(monsterData, 0)

	// 启用Boss AI
	bossAI, err := NewBossAIManagerAdapter(monster, nil, fileutil.FindResourcePth("configs/boss_fire_dragon_lord.json"))
	if err != nil {
		logger.Errorln(err)
		return monster
	}
	monster.SetAiData(bossAI)

	// 添加到场景
	scene.addMonster(monster)

	logger.Debugf("Created Boss monster with advanced AI: %s", monster._name)
	return monster
}
