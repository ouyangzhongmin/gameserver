package bossai

import (
	"fmt"
	"strings"
)

// NodeTypeConstants 行为树节点类型常量
// 用于统一管理JSON配置文件中使用的节点名称
type NodeTypeConstants struct{}

// ActionNodeTypes 行为节点类型常量
var ActionNodeTypes = struct {
	UseHealingSkill    string
	AttackTarget       string
	MoveToPosition     string
	CastSkill          string
	Retreat            string
	Patrol             string
	Wait               string
	CheckPhase         string
	UpdatePhase        string
	TargetValidation   string
	PatrolBehavior     string
	SelectOptimalSkill string
	ExecuteAction      string

	// 新增的节点类型
	RandomMove          string
	PatrolMove          string
	RandomSpeech        string
	RandomAction        string
	ScanEnemies         string
	ChaseTarget         string
	CheckAttackRange    string
	BasicAttack         string
	UseItem             string
	UseSkillByCondition string
	FleeEscape          string
	ReturnToSpawn       string
	AutoRecover         string
	TriggerReward       string
}{
	UseHealingSkill:    "Use Healing Skill",
	AttackTarget:       "Attack Target",
	MoveToPosition:     "Move To Position",
	CastSkill:          "Cast Skill",
	Retreat:            "Retreat",
	Patrol:             "Patrol",
	Wait:               "Wait",
	CheckPhase:         "Check Phase Triggers",
	UpdatePhase:        "Update Phase State",
	TargetValidation:   "Target Validation",
	PatrolBehavior:     "Patrol Behavior",
	SelectOptimalSkill: "Select Optimal Skill",
	ExecuteAction:      "Execute Action",

	// 新增的节点类型
	RandomMove:          "Random Move",
	PatrolMove:          "Patrol Move",
	RandomSpeech:        "Random Speech",
	RandomAction:        "Random Action",
	ScanEnemies:         "Scan Enemies",
	ChaseTarget:         "Chase Target",
	CheckAttackRange:    "Check Attack Range",
	BasicAttack:         "Basic Attack",
	UseItem:             "Use Item",
	UseSkillByCondition: "Use Skill By Condition",
	FleeEscape:          "Flee Escape",
	ReturnToSpawn:       "Return To Spawn",
	AutoRecover:         "Auto Recover",
	TriggerReward:       "Trigger Reward",
}

// ConditionNodeTypes 条件节点类型常量
var ConditionNodeTypes = struct {
	LowHealthCheck    string
	EnemyInRange      string
	SkillAvailable    string
	CombatState       string
	DistanceCheck     string
	PhaseCheck        string
	EmergencyResponse string
	CombatActions     string

	// 新增的条件节点类型
	LowManaCheck    string
	NeedHealCheck   string
	NeedEscapeCheck string
	TargetDeadCheck string
	ItemAvailable   string
	RandomChance    string
	TimeCondition   string
}{
	LowHealthCheck:    "Low Health Check",
	EnemyInRange:      "Enemy In Range",
	SkillAvailable:    "Skill Available",
	CombatState:       "Combat State",
	DistanceCheck:     "Distance Check",
	PhaseCheck:        "Phase Check",
	EmergencyResponse: "Emergency Response",
	CombatActions:     "Combat Actions",

	// 新增的条件节点类型
	LowManaCheck:    "Low Mana Check",
	NeedHealCheck:   "Need Heal Check",
	NeedEscapeCheck: "Need Escape Check",
	TargetDeadCheck: "Target Dead Check",
	ItemAvailable:   "Item Available",
	RandomChance:    "Random Chance",
	TimeCondition:   "Time Condition",
}

// DecoratorNodeTypes 装饰节点类型常量
var DecoratorNodeTypes = struct {
	Inverter string
	Repeater string
	Retry    string
	Timeout  string
	Cooldown string
}{
	Inverter: "Inverter",
	Repeater: "Repeater",
	Retry:    "Retry",
	Timeout:  "Timeout",
	Cooldown: "Cooldown",
}

// GetAllActionNodeTypes 获取所有行为节点类型
func GetAllActionNodeTypes() []string {
	return []string{
		ActionNodeTypes.UseHealingSkill,
		ActionNodeTypes.AttackTarget,
		ActionNodeTypes.MoveToPosition,
		ActionNodeTypes.CastSkill,
		ActionNodeTypes.Retreat,
		ActionNodeTypes.Patrol,
		ActionNodeTypes.Wait,
		ActionNodeTypes.CheckPhase,
		ActionNodeTypes.UpdatePhase,
		ActionNodeTypes.TargetValidation,
		ActionNodeTypes.PatrolBehavior,
		ActionNodeTypes.SelectOptimalSkill,
		ActionNodeTypes.ExecuteAction,
		// 新增的节点类型
		ActionNodeTypes.RandomMove,
		ActionNodeTypes.PatrolMove,
		ActionNodeTypes.RandomSpeech,
		ActionNodeTypes.RandomAction,
		ActionNodeTypes.ScanEnemies,
		ActionNodeTypes.ChaseTarget,
		ActionNodeTypes.CheckAttackRange,
		ActionNodeTypes.BasicAttack,
		ActionNodeTypes.UseItem,
		ActionNodeTypes.UseSkillByCondition,
		ActionNodeTypes.FleeEscape,
		ActionNodeTypes.ReturnToSpawn,
		ActionNodeTypes.AutoRecover,
		ActionNodeTypes.TriggerReward,
	}
}

// GetAllConditionNodeTypes 获取所有条件节点类型
func GetAllConditionNodeTypes() []string {
	return []string{
		ConditionNodeTypes.LowHealthCheck,
		ConditionNodeTypes.EnemyInRange,
		ConditionNodeTypes.SkillAvailable,
		ConditionNodeTypes.CombatState,
		ConditionNodeTypes.DistanceCheck,
		ConditionNodeTypes.PhaseCheck,
		ConditionNodeTypes.EmergencyResponse,
		ConditionNodeTypes.CombatActions,
		// 新增的条件节点类型
		ConditionNodeTypes.LowManaCheck,
		ConditionNodeTypes.NeedHealCheck,
		ConditionNodeTypes.NeedEscapeCheck,
		ConditionNodeTypes.TargetDeadCheck,
		ConditionNodeTypes.ItemAvailable,
		ConditionNodeTypes.RandomChance,
		ConditionNodeTypes.TimeCondition,
	}
}

// GetAllDecoratorNodeTypes 获取所有装饰节点类型
func GetAllDecoratorNodeTypes() []string {
	return []string{
		DecoratorNodeTypes.Inverter,
		DecoratorNodeTypes.Repeater,
		DecoratorNodeTypes.Retry,
		DecoratorNodeTypes.Timeout,
		DecoratorNodeTypes.Cooldown,
	}
}

// IsValidActionNodeType 检查是否为有效的行为节点类型
func IsValidActionNodeType(nodeType string) bool {
	validTypes := GetAllActionNodeTypes()
	for _, validType := range validTypes {
		if validType == nodeType {
			return true
		}
	}
	return false
}

// IsValidConditionNodeType 检查是否为有效的条件节点类型
func IsValidConditionNodeType(nodeType string) bool {
	validTypes := GetAllConditionNodeTypes()
	for _, validType := range validTypes {
		if validType == nodeType {
			return true
		}
	}
	return false
}

// IsValidDecoratorNodeType 检查是否为有效的装饰节点类型
func IsValidDecoratorNodeType(nodeType string) bool {
	validTypes := GetAllDecoratorNodeTypes()
	for _, validType := range validTypes {
		if validType == nodeType {
			return true
		}
	}
	return false
}

// ValidateNodeTypesInConfig 验证配置文件中的节点类型
func ValidateNodeTypesInConfig(config *BossConfig) []string {
	var errors []string

	// 验证行为树配置中的节点类型
	if config.BehaviorTree.RootNode.Name != "" {
		validationErrors := validateBehaviorNodeConfig(config.BehaviorTree.RootNode)
		errors = append(errors, validationErrors...)
	}

	return errors
}

// validateBehaviorNodeConfig 递归验证行为节点配置
func validateBehaviorNodeConfig(nodeConfig BehaviorNodeConfig) []string {
	var errors []string

	// 验证当前节点类型
	switch nodeConfig.Type {
	case NodeTypeAction:
		if !IsValidActionNodeType(nodeConfig.Name) {
			errors = append(errors, fmt.Sprintf("Invalid action node type: %s", nodeConfig.Name))
		}
	case NodeTypeCondition:
		if !IsValidConditionNodeType(nodeConfig.Name) {
			errors = append(errors, fmt.Sprintf("Invalid condition node type: %s", nodeConfig.Name))
		}
	case NodeTypeDecorator:
		if !IsValidDecoratorNodeType(nodeConfig.Name) {
			errors = append(errors, fmt.Sprintf("Invalid decorator node type: %s", nodeConfig.Name))
		}
	}

	// 递归验证子节点
	for _, child := range nodeConfig.Children {
		childErrors := validateBehaviorNodeConfig(child)
		errors = append(errors, childErrors...)
	}

	return errors
}

// GetNodeTypeHint 获取节点类型提示（用于错误提示和自动完成）
func GetNodeTypeHint(partialName string) []string {
	var suggestions []string

	// 搜索所有节点类型中包含部分名称的
	allTypes := append(GetAllActionNodeTypes(), GetAllConditionNodeTypes()...)
	allTypes = append(allTypes, GetAllDecoratorNodeTypes()...)

	for _, nodeType := range allTypes {
		if contains(nodeType, partialName) {
			suggestions = append(suggestions, nodeType)
		}
	}

	return suggestions
}

// contains 检查字符串是否包含子字符串（不区分大小写）
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			strings.Contains(strings.ToLower(s), strings.ToLower(substr)))
}
