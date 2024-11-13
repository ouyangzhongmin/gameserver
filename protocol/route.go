package protocol

const (
	// 服务器将多条消息合并进入一条数据发给前端， 需要前端做特殊处理
	OnMergeMessages = "OnMergeMessages"

	//client用的
	OnEnterScene          = "OnEnterScene"
	OnEnterView           = "OnEnterView"
	OnExitView            = "OnExitView"
	OnHeroMoveTrace       = "OnHeroMoveTrace"
	OnHeroMoveStopped     = "OnHeroMoveStopped"
	OnMonsterMoveTrace    = "OnMonsterMoveTrace"
	OnMonsterMoveStopped  = "OnMonsterMoveStopped"
	OnMonsterCommonAttack = "OnMonsterCommonAttack"
	OnReleaseSpell        = "OnReleaseSpell"
	OnLifeChanged         = "OnLifeChanged"
	OnManaChanged         = "OnManaChanged"
	OnEntityDie           = "OnEntityDie"
	OnBufferAdd           = "OnBufferAdd"
	OnBufferRemove        = "OnBufferRemove"

	OnTextMessage = "OnTextMessage"
)
