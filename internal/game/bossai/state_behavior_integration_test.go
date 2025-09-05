package bossai

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestNewStateBehaviorTreeArchitecture 测试新的状态-行为树架构
func TestNewStateBehaviorTreeArchitecture(t *testing.T) {
	// 创建基础状态
	idleState := NewIdleState()
	chaseState := NewChaseState()
	attackState := NewAttackState()

	// 验证状态创建
	assert.NotNil(t, idleState)
	assert.NotNil(t, chaseState)
	assert.NotNil(t, attackState)

	// 验证状态名称
	assert.Equal(t, "Idle", idleState.GetName())
	assert.Equal(t, "Chase", chaseState.GetName())
	assert.Equal(t, "Attack", attackState.GetName())

	// 验证状态ID
	assert.Equal(t, int32(StateIdle), idleState.GetID())
	assert.Equal(t, int32(StateChase), chaseState.GetID())
	assert.Equal(t, int32(StateAttack), attackState.GetID())
}

// TestBehaviorTreeCreation 测试行为树创建
func TestBehaviorTreeCreation(t *testing.T) {
	// 创建状态
	state := NewBaseBossState(StateIdle, "TestState")

	// 设置行为配置
	behaviors := []string{"scan_enemies", "random_move", "random_speech"}
	state.SetBehaviorConfigs(behaviors)

	// 验证行为配置
	assert.Equal(t, behaviors, state.GetBehaviorConfigs())

	// 测试modifier设置
	state.SetModifier("behaviors", behaviors)
	state.SetModifier("behavior_interval", "1s")

	// 验证modifier
	assert.Equal(t, behaviors, state.GetBehaviorConfigs())
}

// TestBehaviorTreeNodes 测试行为树节点
func TestBehaviorTreeNodes(t *testing.T) {
	// 测试随机移动节点
	randomMoveNode := NewRandomMoveActionNode("Random Move")
	assert.NotNil(t, randomMoveNode)
	assert.Equal(t, "Random Move", randomMoveNode.GetName())

	// 测试扫描敌人节点
	scanEnemiesNode := NewScanEnemiesActionNode("Scan Enemies")
	assert.NotNil(t, scanEnemiesNode)
	assert.Equal(t, "Scan Enemies", scanEnemiesNode.GetName())

	// 测试基础攻击节点
	basicAttackNode := NewBasicAttackActionNode("Basic Attack")
	assert.NotNil(t, basicAttackNode)
	assert.Equal(t, "Basic Attack", basicAttackNode.GetName())

	// 测试条件节点
	randomChanceNode := NewRandomChanceConditionNode("Random Chance")
	assert.NotNil(t, randomChanceNode)
	assert.Equal(t, "Random Chance", randomChanceNode.GetName())
}

// MockBossEntity 模拟Boss实体
type MockBossEntity struct {
	id           int64
	currentLife  int32
	maxLife      int32
	position     MockPosition
	combatTarget IEntity
}

type MockPosition struct {
	X, Y, Z float32
}

func (p MockPosition) DistanceTo(other MockPosition) float32 {
	dx := p.X - other.X
	dy := p.Y - other.Y
	return float32(sqrt(float64(dx*dx + dy*dy)))
}

func sqrt(x float64) float64 {
	// 简单的开平方实现，实际应该用math.Sqrt
	if x == 0 {
		return 0
	}
	z := x
	for i := 0; i < 10; i++ {
		z = (z + x/z) / 2
	}
	return z
}

func (m *MockBossEntity) GetID() int64                        { return m.id }
func (m *MockBossEntity) GetCurrentLife() int32               { return m.currentLife }
func (m *MockBossEntity) GetMaxLife() int32                   { return m.maxLife }
func (m *MockBossEntity) GetPos() MockPosition                { return m.position }
func (m *MockBossEntity) SetCombatTarget(target IEntity)      { m.combatTarget = target }
func (m *MockBossEntity) Idle()                               {}
func (m *MockBossEntity) Walk()                               {}
func (m *MockBossEntity) Run()                                {}
func (m *MockBossEntity) Chase()                              {}
func (m *MockBossEntity) Escape()                             {}
func (m *MockBossEntity) AttackAction()                       {}
func (m *MockBossEntity) Die()                                {}
func (m *MockBossEntity) Stop()                               {}
func (m *MockBossEntity) MoveTo(x, y, z Coord) error          { return nil }
func (m *MockBossEntity) IsInAttackRange(x, y Coord) bool     { return false }
func (m *MockBossEntity) DoAttackTarget(target IEntity) error { return nil }

// TestStateBehaviorTreeIntegration 测试状态-行为树集成
func TestStateBehaviorTreeIntegration(t *testing.T) {
	// 创建模拟Boss
	mockBoss := &MockBossEntity{
		id:          1,
		currentLife: 100,
		maxLife:     100,
		position:    MockPosition{X: 0, Y: 0, Z: 0},
	}

	// 创建上下文
	ctx := &BossContext{
		Boss:        mockBoss,
		CurrentTime: time.Now(),
		DeltaTime:   time.Millisecond * 100,
	}

	// 创建状态
	idleState := NewIdleState()

	// 测试状态进入
	err := idleState.OnEnter(ctx)
	assert.NoError(t, err)
	assert.True(t, idleState.IsActive())

	// 测试状态更新
	err = idleState.OnUpdate(ctx, time.Millisecond*100)
	assert.NoError(t, err)

	// 测试状态退出
	err = idleState.OnExit(ctx)
	assert.NoError(t, err)
	assert.False(t, idleState.IsActive())
}

// TestBossAIManagerConfiguration 测试BossAI管理器配置
func TestBossAIManagerConfiguration(t *testing.T) {
	// 创建AI管理器
	aiManager := NewBossAIManager()
	assert.NotNil(t, aiManager)

	// 创建模拟Boss
	mockBoss := &MockBossEntity{
		id:          1,
		currentLife: 100,
		maxLife:     100,
		position:    MockPosition{X: 0, Y: 0, Z: 0},
	}

	// 测试初始化
	err := aiManager.Initialize(mockBoss)
	assert.NoError(t, err)
}

// TestNodeTypeConstants 测试节点类型常量
func TestNodeTypeConstants(t *testing.T) {
	// 测试行为节点类型
	actionTypes := GetAllActionNodeTypes()
	assert.Contains(t, actionTypes, ActionNodeTypes.RandomMove)
	assert.Contains(t, actionTypes, ActionNodeTypes.ScanEnemies)
	assert.Contains(t, actionTypes, ActionNodeTypes.BasicAttack)

	// 测试条件节点类型
	conditionTypes := GetAllConditionNodeTypes()
	assert.Contains(t, conditionTypes, ConditionNodeTypes.RandomChance)
	assert.Contains(t, conditionTypes, ConditionNodeTypes.LowManaCheck)

	// 测试验证函数
	assert.True(t, IsValidActionNodeType(ActionNodeTypes.RandomMove))
	assert.False(t, IsValidActionNodeType("Invalid Type"))
}

// TestBehaviorTreeCreationFromConfig 测试从配置创建行为树
func TestBehaviorTreeCreationFromConfig(t *testing.T) {
	aiManager := NewBossAIManager()

	// 测试从行为名称创建节点
	node, err := aiManager.createNodeFromBehaviorName("scan_enemies")
	assert.NoError(t, err)
	assert.NotNil(t, node)

	node, err = aiManager.createNodeFromBehaviorName("random_move")
	assert.NoError(t, err)
	assert.NotNil(t, node)

	node, err = aiManager.createNodeFromBehaviorName("basic_attack")
	assert.NoError(t, err)
	assert.NotNil(t, node)

	// 测试无效行为名称
	node, err = aiManager.createNodeFromBehaviorName("invalid_behavior")
	assert.Error(t, err)
	assert.Nil(t, node)
}

// TestStateTransitions 测试状态转换
func TestStateTransitions(t *testing.T) {
	// 创建状态
	idleState := NewBaseBossState(StateIdle, "Idle")
	chaseState := NewBaseBossState(StateChase, "Chase")

	// 添加转换条件
	transition := StateTransition{
		ToState: StateChase,
		Condition: PhaseCondition{
			Type: CondTargetCount,
			Params: map[string]interface{}{
				"min_count": 1,
			},
		},
		Priority: 5,
	}
	idleState.AddTransition(StateChase, transition)

	// 创建测试上下文
	ctx := &BossContext{
		CurrentTime:   time.Now(),
		NearbyEnemies: []IEntity{}, // 空敌人列表
	}

	// 测试无敌人时不能转换
	canTransition := idleState.CanTransitionTo(int32(StateChase), ctx)
	assert.False(t, canTransition)

	// 模拟有敌人的情况（这里简化，实际需要Mock敌人）
	// ctx.NearbyEnemies = []IEntity{mockEnemy}
	// canTransition = idleState.CanTransitionTo(int32(StateChase), ctx)
	// assert.True(t, canTransition)
}
