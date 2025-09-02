package bossai

import (
	"encoding/json"
	"io/ioutil"
	"time"

	"github.com/ouyangzhongmin/gameserver/pkg/logger"
)

// ExampleUsage 演示如何使用Boss AI系统
// 注意：这个示例需要在实际游戏环境中运行，这里仅作演示
func ExampleUsage() {
	logger.Debugln("=== Boss AI System Example ===")

	// 1. 创建AI管理器
	_ = NewBossAIManager() // 在实际使用中需要保存这个实例

	// 在实际使用中，你需要：
	// - 创建真实的Boss实体
	// - 加载配置文件
	// - 初始化AI系统
	// - 在游戏主循环中更新AI

	logger.Debugln("Boss AI system example completed (see integration guide in README)")
	logger.Debugln("=== Boss AI System Example Complete ===")
}

// LoadBossConfigFromFile 从文件加载Boss配置
func LoadBossConfigFromFile(filename string) (*BossConfig, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config BossConfig
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	// 转换时间字符串
	err = parseTimeStrings(&config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

// parseTimeStrings 解析配置中的时间字符串
// 注意：由于SkillConfig中的Cooldown和CastTime已经定义为time.Duration类型，
// 这个函数主要用于处理从JSON文件加载时可能出现的字符串转换
func parseTimeStrings(config *BossConfig) error {
	// 由于类型定义已经是time.Duration，无需额外转换
	// 如果需要从JSON字符串转换，应该在JSON unmarshal阶段处理
	return nil
}

// CreateCustomBossAI 创建自定义Boss AI示例
func CreateCustomBossAI() *BossAIManager {
	logger.Debugln("Creating custom Boss AI...")

	// 1. 创建AI管理器
	aiManager := NewBossAIManager()

	// 2. 手动配置状态机
	setupCustomStateMachine(aiManager)

	// 3. 手动配置技能系统
	setupCustomSkills(aiManager)

	// 4. 手动配置阶段系统
	setupCustomPhases(aiManager)

	// 5. 手动配置行为树
	setupCustomBehaviorTrees(aiManager)

	// 6. 添加插件
	setupCustomPlugins(aiManager)

	logger.Debugln("Custom Boss AI created")
	return aiManager
}

func setupCustomStateMachine(aiManager *BossAIManager) {
	// 添加自定义状态
	idleState := NewIdleState()
	chaseState := NewChaseState()
	attackState := NewAttackState()

	aiManager.stateMachine.AddState(idleState)
	aiManager.stateMachine.AddState(chaseState)
	aiManager.stateMachine.AddState(attackState)

	logger.Debugln("Custom state machine configured")
}

func setupCustomSkills(aiManager *BossAIManager) {
	// 创建自定义技能
	fireballSkill := NewBossSkill(1, "Fireball")
	fireballSkill.SetCooldown(time.Second * 3)
	fireballSkill.SetRange(150.0)
	fireballSkill.SetCastTime(time.Millisecond * 800)
	fireballSkill.AddEffect(SkillEffect{
		Type:   "damage",
		Value:  300.0,
		Target: "enemy",
	})

	healSkill := NewBossSkill(2, "Self Heal")
	healSkill.SetCooldown(time.Second * 10)
	healSkill.SetRange(0.0)
	healSkill.SetCastTime(time.Millisecond * 1500)
	healSkill.AddCondition(SkillCondition{
		Type: CondHealthPercent,
		Params: map[string]interface{}{
			"threshold": 0.5,
			"operator":  "less_than",
		},
	})
	healSkill.AddEffect(SkillEffect{
		Type:   "heal",
		Value:  1000.0,
		Target: "self",
	})

	aiManager.skillManager.AddSkill(fireballSkill)
	aiManager.skillManager.AddSkill(healSkill)
	aiManager.skillManager.SetSkillPriority([]int32{2, 1}) // 治疗优先

	logger.Debugln("Custom skills configured")
}

func setupCustomPhases(aiManager *BossAIManager) {
	// 正常阶段
	normalPhase := NewBossPhase(1, "Normal Phase")
	normalPhase.SetTriggerCondition(PhaseCondition{
		Type: CondHealthPercent,
		Params: map[string]interface{}{
			"threshold": 1.0,
			"operator":  "less_than",
		},
	})
	normalPhase.SetAvailableSkills([]int32{1})

	// 狂暴阶段
	enragePhase := NewBossPhase(2, "Enrage Phase")
	enragePhase.SetTriggerCondition(PhaseCondition{
		Type: CondHealthPercent,
		Params: map[string]interface{}{
			"threshold": 0.3,
			"operator":  "less_than",
		},
	})
	enragePhase.SetAvailableSkills([]int32{1, 2})
	enragePhase.SetStateModifier("damage_multiplier", 1.5)
	enragePhase.SetStateModifier("attack_speed", 1.3)

	aiManager.phaseManager.AddPhase(normalPhase)
	aiManager.phaseManager.AddPhase(enragePhase)
	aiManager.phaseManager.SetPhaseOrder([]int32{1, 2})

	logger.Debugln("Custom phases configured")
}

func setupCustomBehaviorTrees(aiManager *BossAIManager) {
	// 创建自定义行为树
	customTree := NewBehaviorTreeBuilder("Custom Combat").
		Selector("Main Strategy").
		// 紧急治疗
		Sequence("Emergency Heal").
		Condition(NewHealthPercentCondition(0.3, "less_than")).
		Action(NewActionNode("Use Heal Skill")).
		End().
		// 正常战斗
		Sequence("Normal Combat").
		Condition(NewHasTargetCondition()).
		Selector("Combat Actions").
		Sequence("Ranged Attack").
		Action(NewMoveToTargetNode(120.0)). // 保持距离
		Action(NewActionNode("Cast Fireball")).
		End().
		Sequence("Melee Attack").
		Action(NewMoveToTargetNode(30.0)).
		Action(NewAttackTargetNode()).
		End().
		End().
		End().
		// 寻找目标
		Action(NewFindTargetNode(200.0)).
		End().
		Build()

	aiManager.behaviorTrees["Custom"] = customTree

	logger.Debugln("Custom behavior trees configured")
}

func setupCustomPlugins(aiManager *BossAIManager) {
	// 添加战斗分析插件
	combatPlugin := NewCombatAnalysisPlugin()
	combatPlugin.SetConfig(map[string]interface{}{
		"analyze_interval_ms": 1500,
	})

	aiManager.AddPlugin(combatPlugin)

	logger.Debugln("Custom plugins configured")
}

// DemoLLMIntegration 演示LLM集成
func DemoLLMIntegration() {
	logger.Debugln("=== LLM Integration Demo ===")

	// 注意：这需要有效的API密钥才能正常工作
	apiKey := "your_openai_api_key_here"

	// 创建LLM提供者
	llmProvider := NewOpenAIProvider(apiKey)
	llmProvider.Configure(map[string]interface{}{
		"model":       "gpt-3.5-turbo",
		"max_tokens":  800,
		"temperature": 0.7,
	})

	// 创建LLM插件
	llmPlugin := NewLLMPlugin(llmProvider)
	llmPlugin.SetConfig(map[string]interface{}{
		"update_interval_ms":  5000,
		"max_strategy_age_ms": 20000,
	})

	// 创建AI管理器并添加LLM插件
	_ = NewBossAIManager() // 在实际使用中需要保存并使用这个实例
	// aiManager.AddPlugin(llmPlugin) // 在实际使用中取消注释

	_ = llmPlugin // 忽略未使用变量警告
	logger.Debugln("LLM plugin configured (requires valid API key to function)")
	logger.Debugln("=== LLM Integration Demo Complete ===")
}

// BenchmarkAIPerformance 性能测试示例
func BenchmarkAIPerformance() {
	logger.Debugln("=== AI Performance Benchmark ===")

	_ = NewBossAIManager() // 在实际使用中需要保存并配置这个实例

	// 注意：在实际基准测试中，你需要提供真实的Boss实体
	// 这里只演示基准测试的结构

	logger.Debugln("Performance benchmark completed")
	logger.Debugln("=== Performance Benchmark Complete ===")
}
