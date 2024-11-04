
// Hero class definition
class Hero extends Role{
    constructor(heroData, isSelfHero) {
        super(heroData, "blue")
        this.isSelfHero = isSelfHero;
        this.bubbleMsg = null
        this.bubbleDeltaTime = 0
    }

	update(deltaTime, camera){
		super.update(deltaTime, camera)
        if (this.bubbleMsg){
            this.bubbleDeltaTime -= deltaTime
            if (this.bubbleDeltaTime <= 0){
                this.bubbleMsg = null
            }
        }
	}

    draw(ctx) {
        super.draw(ctx)

        if(this.bubbleMsg){
            ctx.fillStyle = 'white';
            ctx.fillText(this.bubbleMsg, this.screenX, this.screenY - this.radius - 30);
        }
    }

    bubble(msg){
        this.bubbleMsg = msg
        this.bubbleDeltaTime = 5000
    }
}