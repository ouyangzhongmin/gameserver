
// Monster class definition
class Monster extends Role{
    constructor(monsterData) {
        super(monsterData, "red")
    }

    update(deltaTime, camera){
        super.update(deltaTime, camera)
    }

    draw(ctx) {
        super.draw(ctx)
    }
}