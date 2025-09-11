package bossai

import (
	"time"

	"github.com/ouyangzhongmin/gameserver/pkg/logger"
	"github.com/ouyangzhongmin/gameserver/pkg/shape"
)

// evaluateCondition 评估条件
func evaluateCondition(condition TriggerCondition, ctx *BossContext) bool {
	if condition.Key != "" {
		// 优先判定key条件是否符合
		result := evaluateCustomScript(condition.Key, condition.Params, ctx)
		if result {
			return true
		}
	}

	// 如果上面的key不符合则判定子复合条件
	if len(condition.SubConds) > 0 {
		results := make([]bool, len(condition.SubConds))
		for i, subCond := range condition.SubConds {
			results[i] = evaluateCondition(subCond, ctx)
		}

		switch condition.Operator {
		case OpAnd:
			for _, result := range results {
				if !result {
					return false
				}
			}
			return true
		case OpOr:
			for _, result := range results {
				if result {
					return true
				}
			}
			return false
		case OpNot:
			if len(results) > 0 {
				return !results[0]
			}
		}
	}

	return false
}

// evaluateCustomScript 评估自定义脚本
func evaluateCustomScript(key string, params map[string]interface{}, ctx *BossContext) bool {
	// 这里可以实现脚本引擎或者预定义的逻辑
	switch key {
	case CondHealthPercent:
		threshold, ok := params["threshold"].(float64)
		if !ok {
			return false
		}
		currentPercent := float64(ctx.Boss.GetCurrentLife()) / float64(ctx.Boss.GetMaxLife())
		operator, ok := params["operator"].(string)
		if !ok {
			operator = "less_than"
		}
		switch operator {
		case "less_than":
			return currentPercent < threshold
		case "greater_than":
			return currentPercent > threshold
		case "equal":
			return currentPercent == threshold
		default:
			return false
		}

	case CondTimeElapsed:
		tmp, ok := params["duration"].(int)
		if !ok {
			return false
		}
		duration := time.Duration(tmp)
		return ctx.CurrentState != nil && ctx.CurrentState.GetTimeInState(ctx.CurrentTime) >= duration

	case CondTargetCount:
		minCount, ok := params["min_count"].(int)
		if !ok {
			return false
		}
		return len(ctx.NearbyEnemies) >= minCount
	case CondScannedEnemy:
		return ctx.Target != nil && ctx.Target.IsAlive()

	case CondMissEnemy:
		return ctx.Target == nil || !ctx.Target.IsAlive()

	case CondInAttackRange:
		if ctx.Target == nil {
			return false
		}
		return ctx.Boss.IsInAttackRange(ctx.Target.GetPos().X, ctx.Target.GetPos().Y)

	case CondOutOfAttackRange:
		if ctx.Target == nil {
			return true
		}
		return !ctx.Boss.IsInAttackRange(ctx.Target.GetPos().X, ctx.Target.GetPos().Y)

	case CondOutOfMovableRange:
		// 目标超出出生点最大可移动范围
		return !ctx.Boss.GetMovableRect().Contains(int64(ctx.Boss.GetPos().X), int64(ctx.Boss.GetPos().Y))

	case CondChaseTimeoutOrOutOfRange:
		// 检查追击超时或超出范围
		if ctx.Target == nil {
			return true
		}
		maxChaseTime := time.Second * 10           // 超出追击时间
		maxChaseDist := ctx.Boss.GetMaxChaseDist() // 超出追击范围
		if ctx.CurrentState != nil && ctx.CurrentState.GetTimeInState(ctx.CurrentTime) > maxChaseTime {
			return true
		}
		// 检查距离
		bossPos := ctx.Boss.GetPos()
		targetPos := ctx.Target.GetPos()
		distance := shape.CalculateDistance(
			float64(bossPos.X), float64(bossPos.Y),
			float64(targetPos.X), float64(targetPos.Y),
		)
		return distance > float64(maxChaseDist)

	case CondReachedBornPoint:
		// 回到了出生点
		bossPos := ctx.Boss.GetPos()
		bornPos := ctx.Boss.GetBornPos()
		distance := shape.CalculateDistance(
			float64(bossPos.X), float64(bossPos.Y),
			float64(bornPos.X), float64(bornPos.Y),
		)
		if distance < 1.0 {
			return true
		}
		return false

	case CondSkillCastComplete:
		// 这里需要检查技能释放是否完成
		// 暂时简单的时间检查
		return ctx.CurrentState != nil && ctx.CurrentState.GetTimeInState(ctx.CurrentTime) > time.Second*2

	case CondStunExpired:
		// 检查眩晕是否结束
		stunDuration := time.Second * 3 // 默认3秒
		return ctx.CurrentState != nil && ctx.CurrentState.GetTimeInState(ctx.CurrentTime) >= stunDuration

	default:
		logger.Debugf("Condition evaluation not implemented: %s", key)
		return false
	}
	return false
}
