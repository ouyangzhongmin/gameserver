package bossai

import (
	"testing"
	"time"

	"github.com/ouyangzhongmin/gameserver/pkg/coord"
)

// MockBossEntity 模拟Boss实体
type MockBossEntity struct {
	id             int64
	position       coord.Vector3
	maxLife        int32
	currentLife    int32
	attackPower    int32
	isAlive        bool
	isInCombat     bool
	combatTarget   IEntity
	nearbyEntities []IEntity
}

func NewMockBossEntity(id int64) *MockBossEntity {
	return &MockBossEntity{
		id:             id,
		position:       coord.Vector3{X: 0, Y: 0, Z: 0},
		maxLife:        1000,
		currentLife:    1000,
		attackPower:    100,
		isAlive:        true,
		nearbyEntities: make([]IEntity, 0),
	}
}

func (m *MockBossEntity) GetID() int64          { return m.id }
func (m *MockBossEntity) GetPos() coord.Vector3 { return m.position }
func (m *MockBossEntity) GetEntityType() int    { return 1 }
func (m *MockBossEntity) IsAlive() bool         { return m.isAlive }
func (m *MockBossEntity) IsDestroyed() bool     { return !m.isAlive }

func (m *MockBossEntity) MoveTo(x, y, z coord.Coord) error {
	m.position.X = x
	m.position.Y = y
	m.position.Z = z
	return nil
}

func (m *MockBossEntity) Stop() error { return nil }

func (m *MockBossEntity) CanAttackTarget(target IEntity) bool {
	return target != nil && target.IsAlive()
}

func (m *MockBossEntity) IsInAttackRange(x, y coord.Coord) bool {
	distance := calculateDistance(
		float64(m.position.X), float64(m.position.Y),
		float64(x), float64(y),
	)
	return distance <= 50.0
}

func (m *MockBossEntity) GetMaxLife() int32     { return m.maxLife }
func (m *MockBossEntity) GetCurrentLife() int32 { return m.currentLife }
func (m *MockBossEntity) GetAttackPower() int32 { return m.attackPower }

func (m *MockBossEntity) CanUseSkill(skillID int32) bool               { return true }
func (m *MockBossEntity) UseSkill(skillID int32, target IEntity) error { return nil }
func (m *MockBossEntity) GetSkillCooldown(skillID int32) time.Duration { return 0 }

func (m *MockBossEntity) IsInCombat() bool               { return m.isInCombat }
func (m *MockBossEntity) GetCombatTarget() IEntity       { return m.combatTarget }
func (m *MockBossEntity) SetCombatTarget(target IEntity) { m.combatTarget = target }

func (m *MockBossEntity) GetEnemiesInRange(radius float64) []IEntity {
	return m.nearbyEntities
}

func (m *MockBossEntity) GetNearestEnemy() IEntity {
	if len(m.nearbyEntities) > 0 {
		return m.nearbyEntities[0]
	}
	return nil
}

func (m *MockBossEntity) SetLife(life int32) {
	m.currentLife = life
	if life <= 0 {
		m.isAlive = false
	}
}

func (m *MockBossEntity) SetInCombat(inCombat bool) {
	m.isInCombat = inCombat
}

func (m *MockBossEntity) AddNearbyEntity(entity IEntity) {
	m.nearbyEntities = append(m.nearbyEntities, entity)
}

// MockEntity 模拟普通实体
type MockEntity struct {
	id       int64
	position coord.Vector3
	isAlive  bool
}

func NewMockEntity(id int64) *MockEntity {
	return &MockEntity{
		id:       id,
		position: coord.Vector3{X: 100, Y: 100, Z: 0},
		isAlive:  true,
	}
}

func (m *MockEntity) GetID() int64          { return m.id }
func (m *MockEntity) GetPos() coord.Vector3 { return m.position }
func (m *MockEntity) GetEntityType() int    { return 0 }
func (m *MockEntity) IsAlive() bool         { return m.isAlive }
func (m *MockEntity) IsDestroyed() bool     { return !m.isAlive }

// 测试用例

func TestBossAIManagerInitialization(t *testing.T) {
	ai := NewBossAIManager()
	boss := NewMockBossEntity(1)

	if ai.isInitialized {
		t.Error("AI should not be initialized initially")
	}

	err := ai.Initialize(boss)
	if err != nil {
		t.Errorf("Failed to initialize AI: %v", err)
	}

	if !ai.isInitialized {
		t.Error("AI should be initialized after Initialize()")
	}

	if ai.boss != boss {
		t.Error("Boss entity not set correctly")
	}
}

func TestStateMachine(t *testing.T) {
	sm := NewStateMachine()

	// 测试添加状态
	idleState := NewIdleState()
	err := sm.AddState(idleState)
	if err != nil {
		t.Errorf("Failed to add state: %v", err)
	}

	// 测试重复添加同一状态
	err = sm.AddState(idleState)
	if err == nil {
		t.Error("Should not allow duplicate states")
	}

	// 测试设置初始状态
	boss := NewMockBossEntity(1)
	ctx := &BossContext{
		Boss:        boss,
		CurrentTime: time.Now(),
		SkillsUsed:  make(map[int32]int),
		CustomData:  make(map[string]interface{}),
	}

	err = sm.SetInitialState(StateIdle, ctx)
	if err != nil {
		t.Errorf("Failed to set initial state: %v", err)
	}

	currentState := sm.GetCurrentState()
	if currentState == nil {
		t.Error("Current state should not be nil")
	}

	if currentState.GetName() != "Idle" {
		t.Errorf("Expected Idle state, got %s", currentState.GetName())
	}
}

func TestSkillSystem(t *testing.T) {
	sm := NewSkillManager()

	// 创建测试技能
	skill := NewBossSkill(1, "Fireball")
	skill.SetCooldown(time.Second * 3)
	skill.SetRange(100.0)
	skill.SetCastTime(time.Millisecond * 500)

	// 添加技能
	err := sm.AddSkill(skill)
	if err != nil {
		t.Errorf("Failed to add skill: %v", err)
	}

	// 测试获取技能
	retrievedSkill, err := sm.GetSkill(1)
	if err != nil {
		t.Errorf("Failed to get skill: %v", err)
	}

	if retrievedSkill.GetName() != "Fireball" {
		t.Errorf("Expected Fireball, got %s", retrievedSkill.GetName())
	}

	// 测试技能冷却
	boss := NewMockBossEntity(1)
	ctx := &BossContext{
		Boss:        boss,
		CurrentTime: time.Now(),
		SkillsUsed:  make(map[int32]int),
		CustomData:  make(map[string]interface{}),
	}

	if !skill.CanUse(ctx) {
		t.Error("Skill should be usable initially")
	}

	// 使用技能
	err = skill.Cast(ctx, []IEntity{})
	if err != nil {
		t.Errorf("Failed to cast skill: %v", err)
	}

	if !skill.IsOnCooldown() {
		t.Error("Skill should be on cooldown after use")
	}

	if skill.CanUse(ctx) {
		t.Error("Skill should not be usable while on cooldown")
	}
}

func TestPhaseSystem(t *testing.T) {
	pm := NewPhaseManager()

	// 创建测试阶段
	phase1 := NewBossPhase(1, "Normal Phase")
	phase1.SetTriggerCondition(PhaseCondition{
		Type: CondHealthPercent,
		Params: map[string]interface{}{
			"threshold": 1.0,
			"operator":  "less_than",
		},
	})

	phase2 := NewBossPhase(2, "Enrage Phase")
	phase2.SetTriggerCondition(PhaseCondition{
		Type: CondHealthPercent,
		Params: map[string]interface{}{
			"threshold": 0.3,
			"operator":  "less_than",
		},
	})

	// 添加阶段
	err := pm.AddPhase(phase1)
	if err != nil {
		t.Errorf("Failed to add phase 1: %v", err)
	}

	err = pm.AddPhase(phase2)
	if err != nil {
		t.Errorf("Failed to add phase 2: %v", err)
	}

	// 设置阶段顺序
	err = pm.SetPhaseOrder([]int32{1, 2})
	if err != nil {
		t.Errorf("Failed to set phase order: %v", err)
	}

	// 测试阶段转换
	boss := NewMockBossEntity(1)
	boss.SetLife(1000) // 满血

	ctx := &BossContext{
		Boss:        boss,
		CurrentTime: time.Now(),
		SkillsUsed:  make(map[int32]int),
		CustomData:  make(map[string]interface{}),
	}

	// 设置初始阶段
	err = pm.SetInitialPhase(1, ctx)
	if err != nil {
		t.Errorf("Failed to set initial phase: %v", err)
	}

	currentPhase := pm.GetCurrentPhase()
	if currentPhase == nil {
		t.Error("Current phase should not be nil")
	}

	if currentPhase.GetName() != "Normal Phase" {
		t.Errorf("Expected Normal Phase, got %s", currentPhase.GetName())
	}

	// 降低血量，测试阶段转换
	boss.SetLife(200) // 20% 血量

	err = pm.Update(ctx, time.Millisecond*100)
	if err != nil {
		t.Errorf("Failed to update phase manager: %v", err)
	}

	// 应该转换到第二阶段
	currentPhase = pm.GetCurrentPhase()
	if currentPhase.GetName() != "Enrage Phase" {
		t.Errorf("Expected Enrage Phase after health drop, got %s", currentPhase.GetName())
	}
}

func TestBehaviorTree(t *testing.T) {
	// 创建简单的行为树
	tree := NewBehaviorTreeBuilder("Test Tree").
		Selector("Root Selector").
		Condition(NewHasTargetCondition()).
		Action(NewFindTargetNode(200.0)).
		End().
		Build()

	if tree.GetRootNode() == nil {
		t.Error("Behavior tree root node should not be nil")
	}

	// 测试执行
	boss := NewMockBossEntity(1)
	enemy := NewMockEntity(2)
	boss.AddNearbyEntity(enemy)

	ctx := &BossContext{
		Boss:          boss,
		Target:        nil,
		CurrentTime:   time.Now(),
		NearbyEnemies: []IEntity{enemy},
		SkillsUsed:    make(map[int32]int),
		CustomData:    make(map[string]interface{}),
		ActionHistory: make([]*AIAction, 0),
	}

	result := tree.Execute(ctx)
	if result != ResultSuccess && result != ResultFailure {
		t.Errorf("Unexpected behavior tree result: %v", result)
	}
}

func TestAIPlugins(t *testing.T) {
	plugin := NewCombatAnalysisPlugin()

	if plugin.GetName() != "CombatAnalysis" {
		t.Errorf("Expected CombatAnalysis, got %s", plugin.GetName())
	}

	if !plugin.IsEnabled() {
		t.Error("Plugin should be enabled by default")
	}

	// 测试插件初始化
	boss := NewMockBossEntity(1)
	ctx := &BossContext{
		Boss:        boss,
		CurrentTime: time.Now(),
		SkillsUsed:  make(map[int32]int),
		CustomData:  make(map[string]interface{}),
	}

	err := plugin.Initialize(ctx)
	if err != nil {
		t.Errorf("Failed to initialize plugin: %v", err)
	}

	// 测试插件更新
	err = plugin.Update(ctx, time.Millisecond*100)
	if err != nil {
		t.Errorf("Failed to update plugin: %v", err)
	}

	// 测试禁用插件
	plugin.SetEnabled(false)
	if plugin.IsEnabled() {
		t.Error("Plugin should be disabled")
	}
}

func TestFullAISystem(t *testing.T) {
	// 创建完整的AI系统测试
	ai := NewBossAIManager()
	boss := NewMockBossEntity(1)
	enemy := NewMockEntity(2)

	boss.AddNearbyEntity(enemy)
	boss.SetInCombat(true)
	boss.SetCombatTarget(enemy)

	// 创建基础配置
	config := &BossConfig{
		ID:   1,
		Name: "Test Boss",
		States: []StateConfig{
			{
				ID:   StateIdle,
				Name: "Idle",
			},
		},
		InitialState: StateIdle,
		Phases: []PhaseConfig{
			{
				ID:   1,
				Name: "Normal Phase",
				TriggerCondition: PhaseCondition{
					Type: CondHealthPercent,
					Params: map[string]interface{}{
						"threshold": 1.0,
						"operator":  "less_than",
					},
				},
			},
		},
		Skills: []SkillConfig{
			{
				ID:       1,
				Name:     "Basic Attack",
				Cooldown: time.Second * 2,
				Range:    100.0,
				CastTime: time.Millisecond * 500,
				Effects: []SkillEffect{
					{
						Type:   "damage",
						Value:  100.0,
						Target: "enemy",
					},
				},
			},
		},
		BehaviorTree: BehaviorTreeConfig{
			UpdateInterval: time.Millisecond * 100,
		},
		Plugins: []PluginConfig{
			{
				Name:     "CombatAnalysis",
				Type:     "combat_analysis",
				Enabled:  true,
				Priority: 1,
			},
		},
	}

	// 加载配置
	err := ai.LoadConfig(config)
	if err != nil {
		t.Errorf("Failed to load config: %v", err)
	}

	// 初始化AI
	err = ai.Initialize(boss)
	if err != nil {
		t.Errorf("Failed to initialize AI: %v", err)
	}

	// 启动AI
	err = ai.Start()
	if err != nil {
		t.Errorf("Failed to start AI: %v", err)
	}

	if !ai.IsRunning() {
		t.Error("AI should be running after start")
	}

	// 更新AI几次
	for i := 0; i < 5; i++ {
		err = ai.Update(time.Millisecond * 100)
		if err != nil {
			t.Errorf("Failed to update AI (iteration %d): %v", i, err)
		}
	}

	// 停止AI
	err = ai.Stop()
	if err != nil {
		t.Errorf("Failed to stop AI: %v", err)
	}

	if ai.IsRunning() {
		t.Error("AI should not be running after stop")
	}

	// 销毁AI
	err = ai.Destroy()
	if err != nil {
		t.Errorf("Failed to destroy AI: %v", err)
	}
}

func TestConditionEvaluation(t *testing.T) {
	boss := NewMockBossEntity(1)
	boss.SetLife(300) // 30% 血量

	ctx := &BossContext{
		Boss:        boss,
		CurrentTime: time.Now(),
		SkillsUsed:  make(map[int32]int),
		CustomData:  make(map[string]interface{}),
	}

	// 测试血量条件
	condition := PhaseCondition{
		Type: CondHealthPercent,
		Params: map[string]interface{}{
			"threshold": 0.5,
			"operator":  "less_than",
		},
	}

	phase := NewBossPhase(1, "Test Phase")
	result := phase.evaluateCondition(condition, ctx)
	if !result {
		t.Error("Health condition should be true (30% < 50%)")
	}

	// 测试复合条件 (AND)
	condition2 := PhaseCondition{
		Type: CondTargetCount,
		Params: map[string]interface{}{
			"min_count": 0,
		},
	}

	compoundCondition := PhaseCondition{
		Type:     CondHealthPercent,
		Operator: OpAnd,
		SubConds: []PhaseCondition{condition, condition2},
	}

	result = phase.evaluateCondition(compoundCondition, ctx)
	if !result {
		t.Error("Compound AND condition should be true")
	}
}

// 性能测试
func BenchmarkAIUpdate(b *testing.B) {
	ai := NewBossAIManager()
	boss := NewMockBossEntity(1)

	config := &BossConfig{
		ID:           1,
		Name:         "Benchmark Boss",
		InitialState: StateIdle,
		States: []StateConfig{
			{ID: StateIdle, Name: "Idle"},
		},
		Phases: []PhaseConfig{
			{
				ID:   1,
				Name: "Normal Phase",
				TriggerCondition: PhaseCondition{
					Type: CondHealthPercent,
					Params: map[string]interface{}{
						"threshold": 1.0,
						"operator":  "less_than",
					},
				},
			},
		},
		Skills:       []SkillConfig{},
		BehaviorTree: BehaviorTreeConfig{},
		Plugins:      []PluginConfig{},
	}

	ai.LoadConfig(config)
	ai.Initialize(boss)
	ai.Start()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ai.Update(time.Millisecond * 16) // 约60FPS
	}
}
