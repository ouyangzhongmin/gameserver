package game

import (
	"fmt"
	"time"

	"github.com/ouyangzhongmin/gameserver/constants"
	"github.com/ouyangzhongmin/gameserver/db/model"
	"github.com/ouyangzhongmin/gameserver/internal/game/bossai"
	"github.com/ouyangzhongmin/gameserver/internal/game/object"
	"github.com/ouyangzhongmin/gameserver/pkg/coord"
	"github.com/ouyangzhongmin/gameserver/pkg/logger"
	"github.com/ouyangzhongmin/gameserver/pkg/shape"
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

	// 用于与原有系统兼容的AI数据
	aiData interface{}
}

// NewBossAIManagerAdapter 创建兼容IAiManager的Boss AI适配器
func NewBossAIManagerAdapter(monster *Monster, configPath string) (*BossAIManagerAdapter, error) {
	adapter := &BossAIManagerAdapter{
		monster:    monster,
		aiManager:  bossai.NewBossAIManager(),
		updateRate: time.Millisecond * 300, // 30Hz更新频率
	}

	// 创建Boss实体适配器
	adapter.bossEntity = NewBossEntityAdapter(monster)

	// 加载配置（如果提供）
	if configPath != "" {
		config, err := bossai.LoadBossConfigFromFile(configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load boss config: %w", err)
		}
		adapter.config = config
		adapter.aiData = config // 用作AI数据返回

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
func (a *BossAIManagerAdapter) TransitionTo(stateID int32) error {
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
}

func NewBossEntityAdapter(monster *Monster) *BossEntityAdapter {
	return &BossEntityAdapter{
		monster: monster,
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

func (b *BossEntityAdapter) MoveTo(x, y, z coord.Coord) error {
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

func (b *BossEntityAdapter) GetAttackPower() int32 {
	return int32(b.monster.GetAttack())
}

func (b *BossEntityAdapter) GetSpawnPosition() coord.Vector3 {
	return b.monster.bornPos
}

func (b *BossEntityAdapter) GetPatrolPoints() []coord.Vector3 {
	// 从预制路径中获取巡逻点
	if b.monster.preparePaths != nil && len(b.monster.preparePaths.Paths) > 0 {
		var points []coord.Vector3
		for _, path := range b.monster.preparePaths.Paths {
			points = append(points, coord.Vector3{
				X: coord.Coord(path.Sx),
				Y: coord.Coord(path.Sy),
				Z: 0,
			})
			points = append(points, coord.Vector3{
				X: coord.Coord(path.Ex),
				Y: coord.Coord(path.Ey),
				Z: 0,
			})
		}
		return points
	}
	return []coord.Vector3{}
}

func (b *BossEntityAdapter) CanUseSkill(skillID int32) bool {
	// 复用Monster的技能检查逻辑
	return b.monster.GetCanUseSpell(0) != nil
}

func (b *BossEntityAdapter) UseSkill(skillID int32, target bossai.IEntity) error {
	// 复用Monster的技能释放逻辑
	spell := b.monster.GetCanUseSpell(0)
	if spell != nil {
		if entityAdapter, ok := target.(*EntityAdapter); ok {
			return b.monster.SpellAttack(spell, entityAdapter.entity)
		}
	}
	return fmt.Errorf("skill %d not available or invalid target", skillID)
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

// 复用Monster的移动速度计算
func (b *BossEntityAdapter) GetStepTime() int {
	return b.monster.getStepTime()
}

// 复用Monster的技能范围检查
func (b *BossEntityAdapter) IsInSpellAttackRange(spell interface{}, x, y coord.Coord) bool {
	if spellObj, ok := spell.(*object.SpellObject); ok {
		return b.monster.IsInSpellAttackRange(spellObj, x, y)
	}
	return false
}

// 获取指定类型的可用技能
func (b *BossEntityAdapter) GetCanUseSpell(spellType int) interface{} {
	return b.monster.GetCanUseSpell(spellType)
}

func (b *BossEntityAdapter) GetSkillCooldown(skillID int32) time.Duration {
	// 获取技能冷却时间
	return time.Duration(b.monster.Data.AttackDuration) * time.Millisecond
}

func (b *BossEntityAdapter) IsInCombat() bool {
	// 检查是否在战斗中
	return b.monster.State == constants.ACTION_STATE_ATTACK ||
		b.monster.State == constants.ACTION_STATE_CHASE
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

func (b *BossEntityAdapter) GetEnemiesInRange(radius float64) []bossai.IEntity {
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
		if !b.monster.CanAttackTarget(entity) {
			continue
		}
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

func (e *EntityAdapter) TakeDamage(damage int32, attacker interface{}) {
	// 复用Monster的伤害处理逻辑
	switch entity := e.entity.(type) {
	case *Hero:
		entity.onBeenHurt(int64(damage))
		// 根据攻击者类型进行处理
		if entityAdapter, ok := attacker.(*EntityAdapter); ok {
			entity.onBeenAttacked(entityAdapter.entity)
		} else if bossAdapter, ok := attacker.(*BossEntityAdapter); ok {
			entity.onBeenAttacked(bossAdapter.monster)
		} else if movableEntity, ok := attacker.(IMovableEntity); ok {
			entity.onBeenAttacked(movableEntity)
		}
	case *Monster:
		entity.onBeenHurt(int64(damage))
		// 根据攻击者类型进行处理
		if entityAdapter, ok := attacker.(*EntityAdapter); ok {
			entity.onBeenAttacked(entityAdapter.entity)
		} else if bossAdapter, ok := attacker.(*BossEntityAdapter); ok {
			entity.onBeenAttacked(bossAdapter.monster)
		} else if movableEntity, ok := attacker.(IMovableEntity); ok {
			entity.onBeenAttacked(movableEntity)
		}
	}
}

func (b *BossEntityAdapter) GetNearestEnemy() bossai.IEntity {
	// 获取最近的敌人
	entities := b.GetEnemiesInRange(10.0)
	if len(entities) > 0 {
		var dist float64 = 10000000
		var enemy bossai.IEntity = nil
		if entities != nil && len(entities) > 0 {
			for _, e := range entities {

				tmpDist := shape.CalculateDistance(float64(b.monster.GetPos().X), float64(b.monster.GetPos().Y), float64(e.GetPos().X), float64(e.GetPos().Y))
				if tmpDist < dist {
					dist = tmpDist
					enemy = e
				}
			}
		}
		return enemy
	}
	return nil
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

// 使用示例函数

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
	err := monster.EnableBossAI("configs/boss_fire_dragon_lord.json")
	if err != nil {
		logger.Errorf("Failed to enable Boss AI: %v", err)
		// 如果Boss AI启用失败，可以回退到普通AI
		// monster.SetAiData(newMonsterAi(monster, defaultAiConfig))
		return monster
	}

	// 添加到场景
	scene.addMonster(monster)

	logger.Debugf("Created Boss monster with advanced AI: %s", monster._name)
	return monster
}

// 在场景更新中集成Boss AI - 这个函数是可选的，因为Monster.update会自动调用aimgr.update
func (s *Scene) updateBossAI(curMilliSecond int64, elapsedTime int64) {
	// 更新所有Boss怪物的AI
	// 注意：这个函数是可选的，因为Scene.update -> Monster.update -> aimgr.update 会自动处理
	s.monsters.Range(func(key, value interface{}) bool {
		monster := value.(*Monster)

		// 检查是否是Boss AI并提供额外的状态监控
		if monster.IsBossAI() {
			bossAI := monster.GetBossAI()
			if bossAI != nil && !bossAI.IsRunning() {
				logger.Warnf("Boss AI for monster %d is not running", monster.GetID())
				// 可以在这里添加重启逻辑或其他处理
			}
		}

		return true
	})
}
