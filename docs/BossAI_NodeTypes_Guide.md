# Boss AI 节点类型配置指南

## 概述

为了提高代码的维护性和可读性，我们将JSON配置文件中使用的行为树节点名称提取到了统一的静态配置中。这个改进避免了硬编码字符串散布在代码中，减少了拼写错误，并便于维护。

## 节点类型常量

### 行为节点类型 (ActionNodeTypes)

在JSON配置文件的行为树中，可以使用以下行为节点类型：

- `"Use Healing Skill"` - 使用治疗技能
- `"Attack Target"` - 攻击目标
- `"Move To Position"` - 移动到指定位置
- `"Cast Skill"` - 释放技能
- `"Retreat"` - 撤退
- `"Patrol"` - 巡逻
- `"Wait"` - 等待
- `"Check Phase Triggers"` - 检查阶段触发器
- `"Update Phase State"` - 更新阶段状态
- `"Target Validation"` - 目标验证
- `"Patrol Behavior"` - 巡逻行为

### 条件节点类型 (ConditionNodeTypes)

在JSON配置文件的行为树中，可以使用以下条件节点类型：

- `"Low Health Check"` - 低血量检查
- `"Enemy In Range"` - 敌人在范围内
- `"Skill Available"` - 技能可用
- `"Combat State"` - 战斗状态
- `"Distance Check"` - 距离检查
- `"Phase Check"` - 阶段检查
- `"Emergency Response"` - 紧急响应
- `"Combat Actions"` - 战斗动作

### 装饰节点类型 (DecoratorNodeTypes)

在JSON配置文件的行为树中，可以使用以下装饰节点类型：

- `"Inverter"` - 取反装饰器
- `"Repeater"` - 重复装饰器
- `"Retry"` - 重试装饰器
- `"Timeout"` - 超时装饰器
- `"Cooldown"` - 冷却装饰器

## 使用示例

### JSON配置文件示例

```json
{
  "behavior_tree": {
    "root_node": {
      "type": 3,
      "name": "Boss AI Root",
      "children": [
        {
          "type": 1,
          "name": "Emergency Response",
          "children": [
            {
              "type": 0,
              "name": "Low Health Check",
              "params": {
                "health_threshold": 0.15
              }
            },
            {
              "type": 0,
              "name": "Use Healing Skill",
              "params": {
                "skill_id": 1007
              }
            }
          ]
        },
        {
          "type": 0,
          "name": "Attack Target"
        }
      ]
    }
  }
}
```

### 代码中使用常量

```go
// 在代码中使用节点类型常量而不是硬编码字符串
switch nodeConfig.Name {
case ActionNodeTypes.UseHealingSkill:
    // 处理治疗技能节点
    return NewUseSkillActionNode(nodeConfig.Name, skillID), nil

case ActionNodeTypes.AttackTarget:
    // 处理攻击目标节点
    return NewAttackActionNode(nodeConfig.Name), nil
}
```

## 验证功能

系统现在提供了自动验证功能：

### 配置验证

加载配置文件时，系统会自动验证所有节点类型是否有效：

```go
config, err := LoadBossConfigFromFile("config.json")
// 如果配置中有无效的节点类型，会在控制台输出警告信息
```

### 手动验证

```go
// 检查单个节点类型是否有效
if IsValidActionNodeType("Attack Target") {
    // 节点类型有效
}

// 获取所有有效的节点类型
allActionTypes := GetAllActionNodeTypes()

// 获取类似节点类型的建议
suggestions := GetNodeTypeHint("Attack")
// 返回: ["Attack Target"]
```

## 最佳实践

1. **始终使用常量**：在代码中引用节点类型时，使用 `ActionNodeTypes.XXX` 常量而不是硬编码字符串
2. **保持一致性**：JSON配置文件中的节点名称必须与常量中定义的完全匹配
3. **添加新节点类型**：
   - 在 `node_types.go` 中添加新的常量
   - 在对应的创建函数中添加处理逻辑
   - 更新文档

## 错误处理

如果配置文件中使用了无效的节点类型，系统会：

1. 在加载配置时输出警告信息
2. 创建通用节点作为回退方案
3. 继续运行，但可能无法实现预期的行为

建议在开发阶段仔细检查配置文件，确保所有节点类型都是有效的。

## 扩展指南

要添加新的节点类型：

1. 在 `node_types.go` 中的相应结构体中添加新常量
2. 更新 `GetAllXXXNodeTypes()` 函数
3. 在 `boss_ai_manager.go` 中的相应创建函数中添加处理逻辑
4. 在 `behavior_nodes.go` 中实现节点类型
5. 更新此文档

这个架构使得系统更加模块化和易于维护，同时保持了强类型检查和良好的错误处理。