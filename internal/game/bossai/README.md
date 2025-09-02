# Boss AI 系统

## 概述

这是一套为MMRPG游戏设计的高性能、高扩展性Boss AI系统，基于状态机、行为树、技能系统、阶段系统等多个子系统构建，支持动态逻辑注入和大语言模型集成。

## 特性

### 🧠 **智能决策**
- **状态机系统**：管理Boss的各种行为状态（空闲、巡逻、追击、攻击、狂暴等）
- **行为树系统**：提供复杂的决策逻辑，支持条件、顺序、选择、并行等节点类型
- **智能目标选择**：基于威胁值、距离、血量等因素智能选择攻击目标

### ⚔️ **战斗系统**
- **技能系统**：完整的技能管理，包括冷却、条件检查、效果应用
- **阶段系统**：根据血量等条件触发不同战斗阶段，每个阶段有独特行为
- **伤害与治疗**：支持多种技能效果类型

### 🔧 **高度可扩展**
- **插件系统**：支持动态加载AI插件，实现自定义逻辑
- **配置驱动**：通过JSON配置文件定义Boss行为，无需修改代码
- **接口设计**：清晰的接口分离，便于扩展和测试

### 🤖 **AI集成**
- **大语言模型支持**：集成OpenAI等LLM，实现动态策略生成
- **战斗分析插件**：实时分析战斗情况，提供智能建议
- **自适应行为**：基于战斗反馈动态调整行为策略

### 🚀 **高性能**
- **并发安全**：使用读写锁保证并发安全
- **性能监控**：内置性能统计和调试信息
- **优化执行**：支持可配置的更新频率和行为深度限制

## 系统架构

### 接口设计

为了避免重复定义和保持与现有系统的兼容性，Boss AI系统直接复用了原有游戏系统中的核心接口：

- `IEntity`: 复用 `internal/game/igame.go` 中的定义
- `IMovableEntity`: 复用 `internal/game/igame.go` 中的定义
- `IBossEntity`: 扩展 `IMovableEntity`，添加Boss专用功能

这种设计确保了：
✅ 避免接口重复定义
✅ 与现有系统完全兼容
✅ 代码维护性更好
✅ 类型转换更安全

```
Boss AI Manager (主管理器)
├── State Machine (状态机)
│   ├── Idle State (空闲状态)
│   ├── Patrol State (巡逻状态)  
│   ├── Chase State (追击状态)
│   ├── Attack State (攻击状态)
│   └── Custom States (自定义状态)
├── Behavior Tree (行为树)
│   ├── Action Nodes (行为节点)
│   ├── Condition Nodes (条件节点)
│   ├── Composite Nodes (组合节点)
│   └── Decorator Nodes (装饰节点)
├── Skill System (技能系统)
│   ├── Skill Manager (技能管理器)
│   ├── Cooldown Management (冷却管理)
│   └── Effect System (效果系统)
├── Phase System (阶段系统)
│   ├── Phase Manager (阶段管理器)
│   ├── Trigger Conditions (触发条件)
│   └── Phase Behaviors (阶段行为)
└── Plugin System (插件系统)
    ├── Combat Analysis Plugin (战斗分析插件)
    ├── LLM Plugin (大模型插件)
    └── Custom Plugins (自定义插件)
```

## 快速开始

### 1. 基础使用

```go
// 创建AI管理器
aiManager := bossai.NewBossAIManager()

// 创建Boss实体
boss := NewMockBossEntity(1001)

// 初始化AI
err := aiManager.Initialize(boss)
if err != nil {
    panic(err)
}

// 启动AI
aiManager.Start()

// 在游戏主循环中更新
for {
    err := aiManager.Update(time.Millisecond * 16) // 60FPS
    if err != nil {
        log.Printf("AI update failed: %v", err)
    }
}
```

### 2. 配置文件使用

```go
// 从配置文件加载
config, err := bossai.LoadBossConfigFromFile("configs/boss_fire_dragon_lord.json")
if err != nil {
    panic(err)
}

// 应用配置
err = aiManager.LoadConfig(config)
if err != nil {
    panic(err)
}
```

### 3. 集成到现有系统

```go
// 在Monster中启用Boss AI
monster := NewMonster(monsterData, 0)
err := monster.EnableBossAI("configs/boss_config.json")
if err != nil {
    log.Printf("Failed to enable Boss AI: %v", err)
}
```

## 配置说明

### 基础配置结构

```json
{
  "id": 1001,
  "name": "Fire Dragon Lord",
  "description": "强大的火龙Boss",
  
  "states": [...],          // 状态配置
  "initial_state": 0,       // 初始状态
  "phases": [...],          // 阶段配置  
  "skills": [...],          // 技能配置
  "behavior_tree": {...},   // 行为树配置
  "plugins": [...],         // 插件配置
  "llm_config": {...}       // LLM配置
}
```

### 状态配置

```json
{
  "id": 0,
  "name": "Idle", 
  "transitions": [
    {
      "to_state": 1,
      "condition": {
        "type": 3,
        "params": {"min_count": 1}
      },
      "priority": 5
    }
  ],
  "modifiers": {
    "movement_speed": 1.0
  }
}
```

### 技能配置

```json
{
  "id": 1001,
  "name": "Fire Breath",
  "cooldown": "3s",
  "range": 150.0,
  "cast_time": "1.5s",
  "conditions": [...],
  "effects": [
    {
      "type": "damage",
      "value": 500.0,
      "target": "enemy"
    }
  ]
}
```

### 阶段配置

```json
{
  "id": 1,
  "name": "Normal Phase",
  "trigger_condition": {
    "type": 0,
    "params": {
      "threshold": 0.7,
      "operator": "less_than"
    }
  },
  "available_skills": [1001, 1002],
  "behavior_tree": "normal_combat"
}
```

## 插件开发

### 创建自定义插件

```go
type CustomPlugin struct {
    *bossai.BaseAIPlugin
    // 自定义字段
}

func NewCustomPlugin() *CustomPlugin {
    base := bossai.NewBaseAIPlugin("Custom", "1.0.0")
    return &CustomPlugin{
        BaseAIPlugin: base,
    }
}

func (p *CustomPlugin) Initialize(ctx *bossai.BossContext) error {
    // 初始化逻辑
    return p.BaseAIPlugin.Initialize(ctx)
}

func (p *CustomPlugin) Update(ctx *bossai.BossContext, deltaTime time.Duration) error {
    // 更新逻辑
    return p.BaseAIPlugin.Update(ctx, deltaTime)
}

func (p *CustomPlugin) SuggestAction(ctx *bossai.BossContext) *bossai.AIAction {
    // 决策逻辑
    return nil
}
```

### 注册插件

```go
plugin := NewCustomPlugin()
aiManager.AddPlugin(plugin)
```

## LLM集成

### 配置OpenAI

```json
{
  "llm_config": {
    "provider": "openai",
    "model": "gpt-3.5-turbo", 
    "api_key": "your_api_key_here",
    "endpoint": "https://api.openai.com/v1/chat/completions",
    "update_interval": "10s",
    "max_tokens": 1000,
    "temperature": 0.7,
    "enabled": true
  }
}
```

### 使用LLM插件

```go
// 创建LLM提供者
provider := bossai.NewOpenAIProvider("your_api_key")
provider.Configure(map[string]interface{}{
    "model": "gpt-3.5-turbo",
    "max_tokens": 800,
    "temperature": 0.7,
})

// 创建LLM插件
llmPlugin := bossai.NewLLMPlugin(provider)
aiManager.AddPlugin(llmPlugin)
```

## 性能优化

### 配置建议

```json
{
  "performance_config": {
    "max_behavior_depth": 10,      // 行为树最大深度
    "max_actions_per_frame": 3,    // 每帧最大动作数
    "enable_profiling": true,      // 启用性能分析
    "memory_limit": 104857600,     // 内存限制(100MB)
    "cpu_threshold": 0.8           // CPU阈值
  }
}
```

### 更新频率调整

```go
// 设置更新频率 (默认100ms)
aiManager.SetUpdateRate(time.Millisecond * 200) // 5Hz更新
```

## 调试功能

### 启用调试模式

```go
aiManager.SetDebugEnabled(true)
```

### 获取调试信息

```go
debugInfo := aiManager.GetDebugInfo()
fmt.Printf("Current State: %s\n", debugInfo.CurrentState)
fmt.Printf("Current Phase: %s\n", debugInfo.CurrentPhase)
fmt.Printf("Active Plugins: %v\n", debugInfo.ActivePlugins)
```

### 性能监控

```go
stats := aiManager.GetStatistics()
fmt.Printf("Update Time: %v\n", stats.UpdateTime)
fmt.Printf("Frame Rate: %.1f FPS\n", stats.FrameRate)
```

## 测试

### 运行单元测试

```bash
cd internal/game/bossai
go test -v
```

### 性能测试

```bash
go test -bench=. -benchmem
```

### 示例使用

```go
// 查看examples.go文件中的完整示例
bossai.ExampleUsage()                 // 基础使用示例
bossai.DemoLLMIntegration()          // LLM集成示例  
bossai.BenchmarkAIPerformance()      // 性能测试示例
```

## 文件结构

```
internal/game/bossai/
├── interfaces.go          // 核心接口定义
├── types.go              // 类型定义
├── states.go             // 状态实现
├── state_machine.go      // 状态机管理器
├── behavior_nodes.go     // 行为树节点
├── behavior_tree.go      // 行为树管理器
├── skill_system.go       // 技能系统
├── phase_system.go       // 阶段系统
├── plugin_system.go      // 插件系统
├── boss_ai_manager.go    // 主管理器
├── boss_ai_test.go       // 单元测试
└── examples.go           // 使用示例

configs/
└── boss_fire_dragon_lord.json  // 示例配置文件

internal/game/
└── boss_ai_integration.go      // 系统集成适配器
```

## 扩展指南

### 添加新状态

1. 实现`IBossState`接口
2. 在配置文件中定义状态
3. 配置状态转换条件

### 添加新行为节点

1. 实现`IBehaviorNode`接口
2. 继承`BaseBehaviorNode`
3. 实现`Execute`方法

### 添加新技能效果

1. 在技能配置中定义新效果类型
2. 在`applyEffect`方法中处理新类型
3. 可选：实现自定义效果处理器

### 添加新插件

1. 继承`BaseAIPlugin`
2. 实现必要的接口方法
3. 注册到AI管理器

## 注意事项

### 性能考虑

- Boss AI系统设计为高性能，但大量Boss同时运行时需注意CPU使用
- LLM插件会产生网络延迟，建议适当调整更新频率
- 行为树深度过大会影响性能，建议控制在10层以内

### 线程安全

- 所有公开接口都是线程安全的
- 自定义插件需要注意线程安全问题
- 避免在回调函数中长时间阻塞

### 内存管理

- 系统会自动管理大部分资源
- 长期运行需要注意内存泄漏
- 及时清理不再使用的AI实例

## 贡献

欢迎提交Issue和Pull Request！

## 许可证

MIT License

---

**Happy Coding! 🎮**