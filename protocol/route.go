package protocol

const (
	//RPC用的

	//client用的
	OnEnterScene          = "OnEnterScene"
	OnEnterView           = "OnEnterView"
	OnExitView            = "OnExitView"
	OnHeroMoveTrace       = "OnHeroMoveTrace"
	OnHeroMoveStopped     = "OnHeroMoveStopped"
	OnMonsterMoveTrace    = "OnMonsterMoveTrace"
	OnMonsterMoveStopped  = "OnMonsterMoveStopped"
	OnMonsterCommonAttack = "OnMonsterCommonAttack"
	OnLifeChanged         = "OnLifeChanged"
	OnManaChanged         = "OnManaChanged"
	OnEntityDie           = "OnEntityDie"
	OnBufferAdd           = "OnBufferAdd"
	OnBufferRemove        = "OnBufferRemove"

	OnTextMessage = "OnTextMessage"
)

var (
	mergeMsgRouteMap = make(map[string]int)
)

func init() {
	//mergeMsgRouteMap[OnEnterScene] = 1
}

func IsMergeMsgRoute(route string) bool {
	_, ok := mergeMsgRouteMap[route]
	return ok
}
