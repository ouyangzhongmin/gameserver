package bossai

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"strconv"
	"time"
)

// LoadBossConfigFromFile 从文件加载Boss配置
func LoadBossConfigFromFile(filename string) (*BossConfig, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var raw map[string]interface{}
	err = json.Unmarshal(data, &raw)
	if err != nil {
		return nil, err
	}

	// 处理时间字符串
	if err := processTimeStrings(raw); err != nil {
		return nil, err
	}

	// 处理数值类型
	if err := parseSkillNumbers(raw); err != nil {
		return nil, err
	}

	// 将raw转换为BossConfig
	bytes, err := json.Marshal(raw)
	if err != nil {
		return nil, err
	}

	var config BossConfig
	err = json.Unmarshal(bytes, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

// processTimeStrings 处理配置中的时间字符串
// 将map中的时间字符串转换为time.Duration类型
func processTimeStrings(config map[string]interface{}) error {
	// 处理行为树更新间隔
	if behaviorTree, ok := config["behavior_tree"].(map[string]interface{}); ok {
		if intervalStr, ok := behaviorTree["update_interval"].(string); ok {
			duration, err := time.ParseDuration(intervalStr)
			if err != nil {
				return fmt.Errorf("failed to parse behavior tree update interval: %v", err)
			}
			behaviorTree["update_interval"] = duration
		}
	}

	// 处理技能配置
	if skills, ok := config["skills"].([]interface{}); ok {
		for _, v := range skills {
			if skill, ok := v.(map[string]interface{}); ok {
				// 处理冷却时间
				if cooldownStr, ok := skill["cooldown"].(string); ok {
					duration, err := time.ParseDuration(cooldownStr)
					if err != nil {
						return fmt.Errorf("failed to parse cooldown: %v", err)
					}
					skill["cooldown"] = duration
				}

				// 处理施法时间
				if castTimeStr, ok := skill["cast_time"].(string); ok {
					duration, err := time.ParseDuration(castTimeStr)
					if err != nil {
						return fmt.Errorf("failed to parse cast time: %v", err)
					}
					skill["cast_time"] = duration
				}

				// 处理效果中的持续时间
				if effects, ok := skill["effects"].([]interface{}); ok {
					for _, e := range effects {
						if effect, ok := e.(map[string]interface{}); ok {
							if durationStr, ok := effect["duration"].(string); ok {
								duration, err := time.ParseDuration(durationStr)
								if err != nil {
									return fmt.Errorf("failed to parse effect duration: %v", err)
								}
								effect["duration"] = duration
							}
						}
					}
				}
			}
		}
	}

	// 处理状态配置
	if states, ok := config["states"].([]interface{}); ok {
		for _, v := range states {
			if state, ok := v.(map[string]interface{}); ok {
				// 处理modifiers中的持续时间
				if modifiers, ok := state["modifiers"].(map[string]interface{}); ok {
					for key, value := range modifiers {
						if key == "duration" || key == "cooldown" || key == "cast_time" || key == "move_speed" {
							if str, ok := value.(string); ok {
								duration, err := time.ParseDuration(str)
								if err != nil {
									return fmt.Errorf("failed to parse %s duration: %v", key, err)
								}
								modifiers[key] = duration
							}
						}
					}
				}
			}
		}
	}

	// 处理阶段配置
	if phases, ok := config["phases"].([]interface{}); ok {
		for _, v := range phases {
			if phase, ok := v.(map[string]interface{}); ok {
				// 处理modifiers中的持续时间
				if modifiers, ok := phase["modifiers"].(map[string]interface{}); ok {
					for key, value := range modifiers {
						if key == "duration" || key == "cooldown" || key == "cast_time" || key == "move_speed" {
							if str, ok := value.(string); ok {
								duration, err := time.ParseDuration(str)
								if err != nil {
									return fmt.Errorf("failed to parse %s duration: %v", key, err)
								}
								modifiers[key] = duration
							}
						}
					}
				}
			}
		}
	}

	// 处理LLM配置
	if llmConfig, ok := config["llm_config"].(map[string]interface{}); ok {
		if intervalStr, ok := llmConfig["update_interval"].(string); ok {
			duration, err := time.ParseDuration(intervalStr)
			if err != nil {
				return fmt.Errorf("failed to parse LLM update interval: %v", err)
			}
			llmConfig["update_interval"] = duration
		}
	}

	return nil
}

// parseSkillNumbers 解析技能配置中的数值类型
// 将JSON中的数值字符串转换为float64类型
func parseSkillNumbers(config map[string]interface{}) error {
	// 处理技能配置
	if skills, ok := config["skills"].([]interface{}); ok {
		for _, v := range skills {
			if skill, ok := v.(map[string]interface{}); ok {
				// 处理范围字段
				if rangeStr, ok := skill["range"].(string); ok {
					value, err := strconv.ParseFloat(rangeStr, 64)
					if err != nil {
						return fmt.Errorf("failed to parse range: %v", err)
					}
					skill["range"] = value
				}

				// 处理效果值中的数值
				if effects, ok := skill["effects"].([]interface{}); ok {
					for _, e := range effects {
						if effect, ok := e.(map[string]interface{}); ok {
							if valueStr, ok := effect["value"].(string); ok {
								if value, err := strconv.ParseFloat(valueStr, 64); err == nil {
									effect["value"] = value
								} else {
									// 如果转换失败，保留原始字符串
									effect["value"] = valueStr
								}
							}
						}
					}
				}
			}
		}
	}

	// 处理状态配置
	if states, ok := config["states"].([]interface{}); ok {
		for _, v := range states {
			if state, ok := v.(map[string]interface{}); ok {
				// 处理modifiers中的数值字段
				if modifiers, ok := state["modifiers"].(map[string]interface{}); ok {
					for key, value := range modifiers {
						if key == "attack_speed" || key == "damage_multiplier" || key == "defense_multiplier" || key == "skill_cooldown_reduction" {
							if str, ok := value.(string); ok {
								if num, err := strconv.ParseFloat(str, 64); err == nil {
									modifiers[key] = num
								} else {
									// 如果转换失败，保留原始字符串
									modifiers[key] = str
								}
							}
						}
					}
				}
			}
		}
	}

	// 处理阶段配置
	if phases, ok := config["phases"].([]interface{}); ok {
		for _, v := range phases {
			if phase, ok := v.(map[string]interface{}); ok {
				// 处理modifiers中的数值字段
				if modifiers, ok := phase["modifiers"].(map[string]interface{}); ok {
					for key, value := range modifiers {
						if key == "attack_speed" || key == "damage_multiplier" || key == "defense_multiplier" || key == "skill_cooldown_reduction" {
							if str, ok := value.(string); ok {
								if num, err := strconv.ParseFloat(str, 64); err == nil {
									modifiers[key] = num
								} else {
									// 如果转换失败，保留原始字符串
									modifiers[key] = str
								}
							}
						}
					}
				}
			}
		}
	}

	return nil
}

// calculateDistance 计算两点间距离
func calculateDistance(x1, y1, x2, y2 float64) float64 {
	dx := x2 - x1
	dy := y2 - y1
	return math.Sqrt(dx*dx + dy*dy)
}
