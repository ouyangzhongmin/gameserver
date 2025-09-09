# 逃跑检查机制说明

## 概述

本系统实现了基于行为树的逃跑检查机制，允许Boss在特定条件下（如血量过低）主动请求转换到逃跑状态。

## 核心组件

### 1. EscapeCheckConditionNode (逃跑检查条件节点)

这是一个行为树条件节点，用于检查Boss是否应该逃跑。当条件满足时，它会通过BossContext请求状态机转换到Retreat（逃跑）状态。

#### 配置参数

- `escape_threshold`: 逃跑血量阈值（0.0-1.0），默认为0.2（20%血量）

#### 使用示例

```json
{
  "type": "node",
  "node": "Escape Check",
  "params": {
    "escape_threshold": 0.3
  }
}
```

### 2. 状态转换机制

系统通过以下步骤实现状态转换：

1. **条件检查**: EscapeCheckConditionNode检查Boss当前血量
2. **请求转换**: 当条件满足时，节点通过`ctx.RequestStateTransition()`方法请求转换到逃跑状态
3. **处理请求**: StateMachine在每次更新时检查转换请求
4. **执行转换**: 如果存在请求，状态机强制转换到目标状态

### 3. 配置示例

在Boss配置文件中，可以在任何状态的行为树中添加逃跑检查：

```json
{
  "id": 3,
  "name": "Attack",
  "behaviors": {
    "type": "selector",
    "children": [
      {
        "type": "sequence",
        "children": [
          {
            "type": "node",
            "node": "Escape Check",
            "params": {
              "escape_threshold": 0.15
            }
          }
        ]
      },
      {
        "type": "node",
        "node": "Basic Attack"
      }
    ]
  }
}
```

## 工作流程

1. **行为树执行**: 在Attack状态中，行为树首先执行Escape Check节点
2. **条件评估**: 如果Boss血量低于15%，EscapeCheckConditionNode返回Success
3. **请求发送**: 节点通过上下文发送状态转换请求到Retreat状态
4. **状态转换**: 状态机在下一次更新时处理请求，强制转换到Retreat状态
5. **逃跑执行**: 在Retreat状态中，Boss执行返回出生点和自动恢复行为

## 优势

- **灵活配置**: 可以为不同状态设置不同的逃跑阈值
- **优先级控制**: 逃跑请求具有高优先级，确保紧急情况下能及时响应
- **解耦设计**: 行为树负责决策，状态机负责执行，符合设计原则
- **易于扩展**: 可以轻松添加其他触发逃跑的条件

## 注意事项

1. 确保Retreat状态在Boss配置中正确定义
2. 逃跑阈值应根据Boss的设计合理设置
3. 可以通过调整优先级来控制逃跑请求的重要性