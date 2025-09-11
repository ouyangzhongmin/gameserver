package game

import (
	"testing"
	"time"

	"github.com/ouyangzhongmin/gameserver/internal/game/bossai"
	"github.com/ouyangzhongmin/gameserver/pkg/coord"
)

// MockMonster 模拟Monster用于测试
type MockMonster struct {
	ID            int64
	Name          string
	Life          int64
	MaxLife       int64
	Attack        int64
	Defense       int64
	Pos           coord.Vector3
	State         int
	isAlive       bool
	phaseModifier *PhaseModifier
}

func NewMockMonster() *MockMonster {
	return &MockMonster{
		ID:      1001,
		Name:    "Test Boss",
		Life:    10000,
		MaxLife: 10000,
		Attack:  1000,
		Defense: 500,
		Pos:     coord.Vector3{X: 0, Y: 0, Z: 0},
		isAlive: true,
	}
}

func (m *MockMonster) GetID() int64 {
	return m.ID
}

func (m *MockMonster) GetPos() coord.Vector3 {
	return m.Pos
}

func (m *MockMonster) GetEntityType() int {
	return 1 // ENTITY_TYPE_MONSTER
}

func (m *MockMonster) IsDestroyed() bool {
	return false
}

func (m *MockMonster) IsAlive() bool {
	return m.isAlive
}

func (m *MockMonster) MoveTo(x, y, z coord.Coord) error {
	m.Pos = coord.Vector3{X: x, Y: y, Z: z}
	return nil
}

func (m *MockMonster) Stop() error {
	return nil
}

func (m *MockMonster) CanAttackTarget(target bossai.IEntity) bool {
	return target.GetEntityType() == 2 // ENTITY_TYPE_HERO
}

func (m *MockMonster) IsInAttackRange(x, y coord.Coord) bool {
	return true
}

func (m *MockMonster) GetMaxLife() int32 {
	return int32(m.MaxLife)
}

func (m *MockMonster) GetCurrentLife() int32 {
	return int32(m.Life)
}

func (m *MockMonster) GetAttackDuration() int {
	return 2000
}

func (m *MockMonster) CanUseSkill(skillID int32) bool {
	return true
}

func (m *MockMonster) IsSkillInCD(skillID int32) bool {
	return false
}

func (m *MockMonster) GetAvailableSkill(rules string) int32 {
	return 1001
}

func (m *MockMonster) IsInSkillAttackRange(skillID int32, x, y coord.Coord) bool {
	return true
}

func (m *MockMonster) UseSkill(skillID int32, target bossai.IEntity) error {
	return nil
}

func (m *MockMonster) DoAttackTarget(target bossai.IEntity) error {
	return nil
}

func (m *MockMonster) GetCanAttackPos(target bossai.IEntity, offset int) (coord.Vector3, error) {
	return m.Pos, nil
}

func (m *MockMonster) GetStepTime() int {
	return 100
}

func (m *MockMonster) IsInCombat() bool {
	return true
}

func (m *MockMonster) GetCombatTarget() bossai.IEntity {
	return nil
}

func (m *MockMonster) SetCombatTarget(target bossai.IEntity) {
}

func (m *MockMonster) GetEntitesInRange(radius float64) []bossai.IEntity {
	return []bossai.IEntity{}
}

func (m *MockMonster) IsEnemy(entity bossai.IEntity) bool {
	return entity.GetEntityType() == 2 // ENTITY_TYPE_HERO
}

func (m *MockMonster) IsAlly(entity bossai.IEntity) bool {
	return entity.GetEntityType() == 1 // ENTITY_TYPE_MONSTER
}

func (m *MockMonster) Idle() {
	m.State = 0
}

func (m *MockMonster) Walk() {
	m.State = 1
}

func (m *MockMonster) Run() {
	m.State = 2
}

func (m *MockMonster) Chase() {
	m.State = 3
}

func (m *MockMonster) Escape() {
	m.State = 4
}

func (m *MockMonster) AttackAction() {
	m.State = 5
}

func (m *MockMonster) Die() {
	m.isAlive = false
	m.State = 6
}

func (m *MockMonster) IsIdle() bool {
	return m.State == 0
}

func (m *MockMonster) IsWalking() bool {
	return m.State == 1
}

func (m *MockMonster) IsRunning() bool {
	return m.State == 2
}

func (m *MockMonster) IsChasing() bool {
	return m.State == 3
}

func (m *MockMonster) IsEscaping() bool {
	return m.State == 4
}

func (m *MockMonster) IsAttacking() bool {
	return m.State == 5
}

func (m *MockMonster) IsDied() bool {
	return m.State == 6
}

func (m *MockMonster) GetCurrentPhase() bossai.IBossPhase {
	return nil
}

func (m *MockMonster) OnPhaseEnter(phase bossai.IBossPhase) {
	// 获取阶段修饰器
	modifiers := phase.GetStateModifiers()

	// 创建阶段修饰器对象
	phaseModifier := &PhaseModifier{
		SkillCooldownMultiplier: 1.0,
		DamageMultiplier:        1.0,
		DefenseMultiplier:       1.0,
		Transform:               "",
	}

	// 应用修饰器
	if skillCooldownMultiplier, ok := modifiers["skill_cooldown_multiplier"].(float64); ok {
		phaseModifier.SkillCooldownMultiplier = skillCooldownMultiplier
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

	// 设置阶段修饰器
	m.phaseModifier = phaseModifier

	// 如果有变身，则应用变身
	if phaseModifier.Transform != "" {
		m.ApplyTransform(phaseModifier.Transform)
	}
}

func (m *MockMonster) SetPhaseModifier(modifier *PhaseModifier) {
	m.phaseModifier = modifier
}

func (m *MockMonster) GetPhaseModifier() *PhaseModifier {
	return m.phaseModifier
}

func (m *MockMonster) ApplyTransform(transform string) {
	// 应用变身
}

// TestBossAIManager 测试Boss AI管理器
func TestBossAIManager(t *testing.T) {
	// 创建模拟Monster
	monster := NewMockMonster()

	// 创建Boss AI管理器
	aiManager := bossai.NewBossAIManager()

	// 初始化Boss AI管理器
	err := aiManager.Initialize(monster)
	if err != nil {
		t.Fatalf("Failed to initialize Boss AI manager: %v", err)
	}

	// 创建测试配置
	config := &bossai.BossConfig{
		ID:           1001,
		Name:         "Test Boss",
		InitialState: bossai.StateIdle,
		States: []bossai.StateConfig{
			{
				ID: bossai.StateIdle,
				Transitions: []bossai.StateTransition{
					{
						ToState: bossai.StateChase,
						Condition: bossai.TriggerCondition{
							Type: bossai.CondTargetCount,
							Params: map[string]interface{}{
								"min_count": 1,
							},
						},
						Priority: 5,
					},
				},
				Behaviors: []interface{}{"scan_enemies", "random_move"},
				Modifiers: map[string]interface{}{
					"behavior_interval": "1s",
				},
			},
			{
				ID: bossai.StateChase,
				Transitions: []bossai.StateTransition{
					{
						ToState: bossai.StateAttack,
						Condition: bossai.TriggerCondition{
							Type: bossai.CondCustomScript,
							Params: map[string]interface{}{
								"script": "in_attack_range",
							},
						},
						Priority: 10,
					},
				},
				Behaviors: []interface{}{},
				Modifiers: map[string]interface{}{
					"behavior_interval": "300ms",
				},
			},
			{
				ID: bossai.StateAttack,
				Transitions: []bossai.StateTransition{
					{
						ToState: bossai.StateChase,
						Condition: bossai.TriggerCondition{
							Type: bossai.CondCustomScript,
							Params: map[string]interface{}{
								"script": "target_out_of_attack_range",
							},
						},
						Priority: 8,
					},
				},
				Behaviors: []interface{}{"basic_attack", "use_skill"},
				Modifiers: map[string]interface{}{
					"attack_cooldown": "1.5s",
				},
			},
		},
		Phases: []bossai.PhaseConfig{
			{
				ID:   1,
				Name: "Normal Phase",
				TriggerCondition: bossai.TriggerCondition{
					Type: bossai.CondHealthPercent,
					Params: map[string]interface{}{
						"threshold": 1.0,
						"operator":  "less_than",
					},
				},
				AvailableSkills: []int32{1001, 1002, 1003},
				Modifiers: map[string]interface{}{
					"skill_cooldown_multiplier": 1.0,
					"movement_speed":            1.0,
					"transform":                 "变身1",
				},
			},
			{
				ID:   2,
				Name: "Aggressive Phase",
				TriggerCondition: bossai.TriggerCondition{
					Type: bossai.CondHealthPercent,
					Params: map[string]interface{}{
						"threshold": 0.7,
						"operator":  "less_than",
					},
				},
				AvailableSkills: []int32{1001, 1002, 1003, 1004, 1005},
				Modifiers: map[string]interface{}{
					"skill_cooldown_multiplier": 0.8,
					"aggression_level":          "high",
					"transform":                 "变身2",
				},
			},
			{
				ID:   3,
				Name: "Desperate Phase",
				TriggerCondition: bossai.TriggerCondition{
					Type: bossai.CondHealthPercent,
					Params: map[string]interface{}{
						"threshold": 0.3,
						"operator":  "less_than",
					},
				},
				AvailableSkills: []int32{1001, 1002, 1003, 1004, 1005, 1006, 1007},
				Modifiers: map[string]interface{}{
					"skill_cooldown_multiplier": 0.5,
					"damage_multiplier":         1.3,
					"defense_multiplier":        1.3,
					"transform":                 "变身3",
				},
			},
		},
	}

	// 加载配置
	err = aiManager.LoadConfig(config)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// 启动AI
	err = aiManager.Start()
	if err != nil {
		t.Fatalf("Failed to start AI: %v", err)
	}

	// 检查初始状态
	currentState := aiManager.GetCurrentState()
	if currentState == nil {
		t.Error("Current state should not be nil")
	}

	// 检查初始阶段
	currentPhase := aiManager.GetCurrentPhase()
	if currentPhase == nil {
		t.Error("Current phase should not be nil")
	}

	// 更新AI
	err = aiManager.Update(time.Millisecond * 100)
	if err != nil {
		t.Errorf("Failed to update AI: %v", err)
	}

	// 检查是否运行中
	if !aiManager.IsRunning() {
		t.Error("AI should be running")
	}

	// 停止AI
	err = aiManager.Stop()
	if err != nil {
		t.Errorf("Failed to stop AI: %v", err)
	}

	// 检查是否已停止
	if aiManager.IsRunning() {
		t.Error("AI should be stopped")
	}
}

// TestPhaseModifiers 测试阶段修饰器
func TestPhaseModifiers(t *testing.T) {
	// 创建模拟Monster
	monster := NewMockMonster()

	// 创建阶段
	phase := bossai.NewBossPhase(1, "Test Phase")
	phase.SetStateModifier("skill_cooldown_multiplier", 0.5)
	phase.SetStateModifier("damage_multiplier", 1.5)
	phase.SetStateModifier("defense_multiplier", 1.2)
	phase.SetStateModifier("transform", "变身测试")

	// 创建上下文
	ctx := &bossai.BossContext{
		Boss:        monster,
		CurrentTime: time.Now(),
	}

	// 进入阶段
	err := phase.OnEnter(ctx)
	if err != nil {
		t.Errorf("Failed to enter phase: %v", err)
	}

	// 检查Monster是否正确接收了阶段修饰器
	modifier := monster.GetPhaseModifier()
	if modifier == nil {
		t.Error("Phase modifier should not be nil")
	}

	if modifier.SkillCooldownMultiplier != 0.5 {
		t.Errorf("Expected skill cooldown multiplier 0.5, got %f", modifier.SkillCooldownMultiplier)
	}

	if modifier.DamageMultiplier != 1.5 {
		t.Errorf("Expected damage multiplier 1.5, got %f", modifier.DamageMultiplier)
	}

	if modifier.DefenseMultiplier != 1.2 {
		t.Errorf("Expected defense multiplier 1.2, got %f", modifier.DefenseMultiplier)
	}

	if modifier.Transform != "变身测试" {
		t.Errorf("Expected transform '变身测试', got %s", modifier.Transform)
	}
}
