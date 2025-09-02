# Boss AI 系统集成使用指南

## 概述

Boss AI系统现在已经完全兼容原有的 `IAiManager` 接口，可以直接通过 `monster.SetAiData()` 设置到Monster中，实现无缝集成。

## 核心架构

### 1. 兼容性设计

```go
// BossAIManagerAdapter 实现了 IAiManager 接口
type BossAIManagerAdapter struct {
    monster      *Monster
    bossEntity   *BossEntityAdapter
    aiManager    *bossai.BossAIManager
    // ...
}

// 实现 IAiManager 接口的所有方法
func (a *BossAIManagerAdapter) update(curMilliSecond int64, elapsedTime int64) error
func (a *BossAIManagerAdapter) onBeenAttacked(target IMovableEntity)
func (a *BossAIManagerAdapter) GetAiData() interface{}
func (a *BossAIManagerAdapter) GetOwner() IMovableEntity
func (a *BossAIManagerAdapter) clear()
```

### 2. 适配器模式

```go
// BossEntityAdapter 将 Monster 适配为 IBossEntity 接口
type BossEntityAdapter struct {
    monster *Monster
}

// EntityAdapter 将原有实体适配为 Boss AI 系统接口
type EntityAdapter struct {
    entity IMovableEntity
}
```

## 使用方法

### 1. 基础使用 - 启用Boss AI

```go
// 创建普通Monster
monsterData := &model.Monster{
    Id:         1001,
    Name:       "Fire Dragon Lord",
    Level:      50,
    BaseAttack: 800,
    BaseLife:   50000,
}

monster := NewMonster(monsterData, 0)

// 启用Boss AI（可以提供配置文件路径或留空使用默认配置）
err := monster.EnableBossAI("configs/boss_fire_dragon_lord.json")
if err != nil {
    log.Printf("Failed to enable Boss AI: %v", err)
    // 可以回退到普通AI
    // monster.SetAiData(newMonsterAi(monster, defaultAiConfig))
}

// 添加到场景
scene.addMonster(monster)
```

### 2. 检查和管理Boss AI

```go
// 检查是否使用Boss AI
if monster.IsBossAI() {
    log.Printf("Monster %d is using Boss AI", monster.GetID())
    
    // 获取Boss AI管理器
    bossAI := monster.GetBossAI()
    if bossAI != nil {
        // 启用调试模式
        monster.SetBossAIDebug(true)
        
        // 强制转换状态
        err := monster.TransitionBossAIState(int32(bossai.StateAttack))
        if err != nil {
            log.Printf("State transition failed: %v", err)
        }
        
        // 获取调试信息
        debugInfo := bossAI.GetDebugInfo()
        log.Printf("Current state: %s", debugInfo.CurrentState)
        
        // 暂停/恢复AI
        bossAI.Pause()
        bossAI.Resume()
    }
}
```

### 3. 在原有系统中的无缝集成

```go
// 原有的Monster更新逻辑完全不需要修改
func (m *Monster) update(curMilliSecond int64, elapsedTime int64) error {
    // ... 原有逻辑 ...
    
    // Boss AI会通过IAiManager接口自动被调用
    if m.aimgr != nil {
        err = m.aimgr.update(curMilliSecond, elapsedTime)
        if err != nil {
            logger.Errorln("aimgr update err:", err)
        }
    }
    
    // ... 原有逻辑 ...
}

// 原有的攻击事件处理也无需修改
func (m *Monster) onBeenAttacked(attacker IMovableEntity) {
    if m.aimgr != nil {
        m.aimgr.onBeenAttacked(attacker)  // Boss AI会自动处理
    }
    // ... 原有逻辑 ...
}
```

### 4. 配置文件示例

```json
{
  "id": 1001,
  "name": "Fire Dragon Lord",
  "description": "强大的火龙Boss",
  
  "states": [
    {
      "id": 0,
      "name": "Idle",
      "transitions": [
        {
          "to_state": 2,
          "condition": {
            "type": 3,
            "params": {"min_count": 1}
          },
          "priority": 5
        }
      ]
    }
  ],
  "initial_state": 0,
  
  "skills": [
    {
      "id": 1001,
      "name": "Fire Breath",
      "cooldown": "3s",
      "range": 150.0,
      "cast_time": "1.5s",
      "effects": [
        {
          "type": "damage",
          "value": 500.0,
          "target": "enemy"
        }
      ]
    }
  ],
  
  "phases": [
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
      "available_skills": [1001]
    }
  ]
}
```

### 5. 高级功能 - 添加AI插件

```go
// 在启用Boss AI后添加插件
bossAI := monster.GetBossAI()
if bossAI != nil {
    // 添加战斗分析插件
    combatPlugin := bossai.NewCombatAnalysisPlugin()
    combatPlugin.SetConfig(map[string]interface{}{
        "analyze_interval_ms": 1500,
    })
    
    err := bossAI.AddPlugin(combatPlugin)
    if err != nil {
        log.Printf("Failed to add plugin: %v", err)
    }
    
    // 添加LLM插件（需要API密钥）
    if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" {
        provider := bossai.NewOpenAIProvider(apiKey)
        llmPlugin := bossai.NewLLMPlugin(provider)
        bossAI.AddPlugin(llmPlugin)
    }
}
```

## 性能特点

### 1. 高效更新
- Boss AI适配器会自动控制更新频率（默认10Hz）
- 只有在需要时才会更新底层AI系统
- 与原有Monster更新循环完美集成

### 2. 内存优化
- 适配器模式避免了重复的数据拷贝
- 延迟加载配置和资源
- 自动清理和资源管理

### 3. 并发安全
- 所有Boss AI操作都是线程安全的
- 与原有游戏系统的并发模型兼容

## 迁移指南

### 从普通AI迁移到Boss AI

```go
// 原有代码
monster.SetAiData(newMonsterAi(monster, aiConfig))

// 新代码 - 直接替换
err := monster.EnableBossAI("path/to/boss_config.json")
if err != nil {
    // 失败时回退到原有AI
    monster.SetAiData(newMonsterAi(monster, aiConfig))
}
```

### 原有AI配置的兼容性

```go
// 可以根据怪物等级或类型选择AI系统
func initMonsterAI(monster *Monster, aiConfig *model.Aiconfig) {
    if monster.Grade >= 3 { // Boss级别怪物
        err := monster.EnableBossAI("configs/boss_ai.json")
        if err == nil {
            return
        }
    }
    
    // 普通怪物或Boss AI启用失败时使用原有AI
    monster.SetAiData(newMonsterAi(monster, aiConfig))
}
```

## 调试和监控

### 1. 调试信息

```go
if monster.IsBossAI() {
    bossAI := monster.GetBossAI()
    debugInfo := bossAI.GetDebugInfo()
    
    log.Printf("Boss %d AI状态:", monster.GetID())
    log.Printf("  当前状态: %s", debugInfo.CurrentState)
    log.Printf("  当前阶段: %s", debugInfo.CurrentPhase)
    log.Printf("  活跃插件: %v", debugInfo.ActivePlugins)
    log.Printf("  更新时间: %v", debugInfo.PerformanceMetrics.UpdateTime)
}
```

### 2. 性能监控

```go
// 在场景更新中添加性能监控
func (s *Scene) updateBossAI(curMilliSecond int64, elapsedTime int64) {
    s.monsters.Range(func(key, value interface{}) bool {
        monster := value.(*Monster)
        
        if monster.IsBossAI() {
            bossAI := monster.GetBossAI()
            if bossAI != nil && !bossAI.IsRunning() {
                logger.Warnf("Boss AI for monster %d is not running", monster.GetID())
                // 可以在这里添加重启逻辑
            }
        }
        
        return true
    })
}
```

## 总结

通过实现 `IAiManager` 接口的适配器模式，Boss AI系统现在可以：

1. **无缝集成**：直接通过 `monster.EnableBossAI()` 启用
2. **完全兼容**：与原有系统API完全兼容，无需修改现有代码
3. **渐进迁移**：可以逐步将重要怪物从普通AI迁移到Boss AI
4. **向下兼容**：Boss AI启用失败时可以自动回退到原有AI
5. **高性能**：优化的更新频率和资源管理

这种设计既保持了系统的稳定性，又提供了强大的扩展能力。