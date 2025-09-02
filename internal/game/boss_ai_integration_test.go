package game

import (
	"testing"
	"time"

	"github.com/ouyangzhongmin/gameserver/db/model"
	"github.com/ouyangzhongmin/gameserver/internal/game/bossai"
)

// TestBossAIManagerAdapter 测试Boss AI适配器
func TestBossAIManagerAdapter(t *testing.T) {
	// 创建测试怪物
	monsterData := &model.Monster{
		Id:         1001,
		Name:       "Test Boss",
		Level:      50,
		BaseAttack: 800,
		BaseLife:   50000,
	}

	monster := NewMonster(monsterData, 0)

	// 测试创建Boss AI适配器
	adapter, err := NewBossAIManagerAdapter(monster, "")
	if err != nil {
		t.Errorf("Failed to create Boss AI adapter: %v", err)
		return
	}

	// 测试IAiManager接口实现
	if adapter.GetOwner() != monster {
		t.Error("GetOwner should return the monster")
	}

	// 测试AI数据
	aiData := adapter.GetAiData()
	if aiData == nil {
		t.Error("GetAiData should not return nil")
	}

	// 测试更新
	err = adapter.update(time.Now().UnixMilli(), 100)
	if err != nil {
		t.Errorf("Update failed: %v", err)
	}

	// 测试清理
	adapter.clear()
	if adapter.IsRunning() {
		t.Error("AI should not be running after clear")
	}
}

// TestMonsterBossAIIntegration 测试Monster的Boss AI集成
func TestMonsterBossAIIntegration(t *testing.T) {
	// 创建测试怪物
	monsterData := &model.Monster{
		Id:         1002,
		Name:       "Test Dragon",
		Level:      60,
		BaseAttack: 1000,
		BaseLife:   80000,
	}

	monster := NewMonster(monsterData, 0)

	// 测试启用Boss AI（没有配置文件）
	err := monster.EnableBossAI("")
	if err != nil {
		t.Errorf("Failed to enable Boss AI: %v", err)
		return
	}

	// 验证Boss AI已启用
	if !monster.IsBossAI() {
		t.Error("Monster should have Boss AI enabled")
	}

	// 获取Boss AI管理器
	bossAI := monster.GetBossAI()
	if bossAI == nil {
		t.Error("Should be able to get Boss AI manager")
	}

	// 测试Boss AI是否在运行
	if !bossAI.IsRunning() {
		t.Error("Boss AI should be running")
	}

	// 测试调试模式
	monster.SetBossAIDebug(true)

	// 测试状态转换（这个会失败因为没有对应的状态，但不应该崩溃）
	err = monster.TransitionBossAIState(int32(bossai.StateAttack))
	// 这里期望有错误，因为没有初始化完整的状态机
	if err == nil {
		t.Log("State transition succeeded or failed gracefully")
	}

	// 测试清理
	if monster.aimgr != nil {
		monster.aimgr.clear()
	}
}

// TestBossEntityAdapter 测试Boss实体适配器
func TestBossEntityAdapter(t *testing.T) {
	// 创建测试怪物
	monsterData := &model.Monster{
		Id:         1003,
		Name:       "Test Adapter Boss",
		Level:      70,
		BaseAttack: 1200,
		BaseLife:   100000,
	}

	monster := NewMonster(monsterData, 0)
	adapter := NewBossEntityAdapter(monster)

	// 测试基础接口
	if adapter.GetID() != monster.GetID() {
		t.Error("Adapter ID should match monster ID")
	}

	if adapter.GetEntityType() != monster.GetEntityType() {
		t.Error("Adapter entity type should match monster")
	}

	// 测试生命状态
	if !adapter.IsAlive() {
		t.Error("Adapter should report monster as alive")
	}

	// 测试属性获取
	if adapter.GetMaxLife() != int32(monster.MaxLife) {
		t.Errorf("Max life mismatch: expected %d, got %d", monster.MaxLife, adapter.GetMaxLife())
	}

	if adapter.GetCurrentLife() != int32(monster.Life) {
		t.Errorf("Current life mismatch: expected %d, got %d", monster.Life, adapter.GetCurrentLife())
	}

	if adapter.GetAttackPower() != int32(monster.Data.BaseAttack) {
		t.Errorf("Attack power mismatch: expected %d, got %d", monster.Data.BaseAttack, adapter.GetAttackPower())
	}

	// 测试战斗状态
	if adapter.IsInCombat() {
		t.Error("Monster should not be in combat initially")
	}

	// 测试技能检查
	canUseSkill := adapter.CanUseSkill(1)
	t.Logf("Can use skill: %v", canUseSkill)

	// 测试范围内实体获取（没有场景时应该返回空）
	entities := adapter.GetEntitiesInRange(100.0)
	if entities == nil {
		t.Error("GetEntitiesInRange should not return nil")
	}
}

// TestEntityAdapter 测试实体适配器
func TestEntityAdapter(t *testing.T) {
	// 创建测试怪物
	monsterData := &model.Monster{
		Id:         1004,
		Name:       "Test Entity",
		Level:      40,
		BaseAttack: 600,
		BaseLife:   30000,
	}

	monster := NewMonster(monsterData, 0)
	adapter := NewEntityAdapter(monster)

	// 测试基础接口
	if adapter.GetID() != monster.GetID() {
		t.Error("Entity adapter ID should match monster ID")
	}

	if !adapter.IsAlive() {
		t.Error("Entity adapter should report monster as alive")
	}

	if adapter.IsDestroyed() {
		t.Error("Entity adapter should not report monster as destroyed")
	}
}

// BenchmarkBossAIUpdate 基准测试Boss AI更新性能
func BenchmarkBossAIUpdate(b *testing.B) {
	// 创建测试怪物
	monsterData := &model.Monster{
		Id:         1005,
		Name:       "Benchmark Boss",
		Level:      80,
		BaseAttack: 1500,
		BaseLife:   150000,
	}

	monster := NewMonster(monsterData, 0)

	// 启用Boss AI
	err := monster.EnableBossAI("")
	if err != nil {
		b.Fatalf("Failed to enable Boss AI: %v", err)
	}

	bossAI := monster.GetBossAI()
	if bossAI == nil {
		b.Fatal("Boss AI should not be nil")
	}

	// 基准测试更新性能
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := bossAI.update(time.Now().UnixMilli(), 16) // 模拟60FPS
		if err != nil {
			b.Errorf("Update failed: %v", err)
		}
	}
}
