package constants

const (
	KCurHero = "hero"
)

const (
	DEFAULT_SCENE  = 1
	DEFAULT_SCENE2 = 2

	ATTR_TYPE_STRENGTH = iota
	ATTR_TYPE_AGILITY
	ATTR_TYPE_INTELLIGENCE

	MONSTER_TYPE_NORMAL = 0
	MONSTER_TYPE_NPC    = 1
)

const (
	ENTITY_TYPE_HERO int = iota
	ENTITY_TYPE_MONSTER
	ENTITY_TYPE_SPELL
)

type ActionState int

const (
	ACTION_STATE_IDLE ActionState = iota
	ACTION_STATE_WALK
	ACTION_STATE_RUN
	ACTION_STATE_CHASE  //追击
	ACTION_STATE_ESCAPE //逃跑
	ACTION_STATE_ATTACK
	ACTION_STATE_DIE
)

type BEHAVIOR int

// AI行为状态
const (
	BEHAVIOR_STATE_IDLE BEHAVIOR = iota
	BEHAVIOR_STATE_ATTACK
	BEHAVIOR_STATE_ESCAPE
	BEHAVIOR_STATE_RETURN
	BEHAVIOR_STATE_FOLLOW
	BEHAVIOR_STATE_LOOP_WALKER
)

const (
	/**
	 * 英雄和怪物地图块的默认大小 将整个场景分割成很多个地图块, 减少广播时遍历的数量, 仅遍历相关的地图块中的英雄和怪物
	 * 英雄和怪物移动时都会检查有没有改变其所在的地图块 例: 地图块大小为5, 表示5x5格为一块, 若地图总共为12x12格,则总共会有3x3的地图块
	 *
	 * 0__________1__________2____ | |//////////|////| | |//////////|////| |
	 * |//////YX//|////| | |//////////|////| |__________|__________|____| |
	 * |//////////|////| | | | | | | | | | | | | |__________|__________|____| |
	 * | | | |__________|__________|____|
	 *
	 *
	 * 英雄所在地图块通过将坐标除以每块长度所得， 比如英雄坐标为 （8,3），则在地图块（1,0）中，储存在heroBlock[0][1]格中
	 * （x，y）对调，y先x后，猜测因为y通常较小，从而减少数组的数量
	 * 获取英雄一定范围内的怪物或英雄时，寻找所有受影响的地图块，然后遍历其中的英雄或怪物，检查它们与英雄之剑的距离
	 * 比如英雄坐标为（8,3），需要寻找3格以内的所有英雄。先找到x和y的上下限，找到它们所在的地图块，然后检查它们之间的所有地图块
	 * 上图中阴影部分标出了所有受影响的地图块，遍历4个地图块中的所有英雄，检查他们到（8,3）的距离是否为3以内
	 */
	SCENE_BLOCK_TO_SHIFT = 7
	SCENE_AOI_GRID_SIZE  = 60
)
