
// Monster class definition
class SpellEntity{
    constructor(spellData, target, parent) {
        this.data = spellData
        this.id = this.data.id;
        this.target = target;
        this.parent = parent;
        this.targetX = getPixelXByGrid(this.data.target_pos.x)
        this.targetY = getPixelYByGrid(this.data.target_pos.y);
        this.speed = getSpeed(this.data.fly_step_time)
        this.radius = 5;
        this.setPos(getPixelXByGrid(this.data.pos_x),getPixelYByGrid(this.data.pos_y),0)
    }

    update(deltaTime, camera){
        if (this.target){
            //跟随目标
            this.targetX = getPixelXByGrid(this.target.x)
            this.targetY = getPixelYByGrid(this.target.y);
        }
        this.updateMove(deltaTime);
        this.screenX = this.x - camera.x;
        this.screenY = this.y - camera.y;

    }

    draw(ctx) {
        ctx.beginPath();
        ctx.arc(this.screenX, this.screenY, this.radius, 0, Math.PI * 2, false);
        ctx.fillStyle = "black";
        ctx.fill();
        // ���Ʊ߿�
        ctx.lineWidth = 1;
        ctx.strokeStyle = 'black';
        ctx.stroke();
    }

    updateMove(deltaTime){
        const tempTargetX = this.targetX;
        const tempTargetY = this.targetY;
        const tempDist = getDistance(this.x , this.y , tempTargetX , tempTargetY);
        const tempAngle = getAngle(this.x , this.y , tempTargetX , tempTargetY);
        const tempActualSpeed = this.speed * deltaTime;
        if(tempDist <= Math.floor(tempActualSpeed + 0.5)) {
            this.setPos(tempTargetX, tempTargetY, 0)
            this.destroy()
        }else
        {
            this.speedX = cosD(tempAngle) * tempActualSpeed;
            this.speedY = sinD(tempAngle) * tempActualSpeed;
            this.setPos(this.x +this.speedX, this.y + this.speedY , 0)
        }
    }

    setPos(x,y,z){
        // console.log(this.name, "setPos::", x, y , z)
        this.x = x;
        this.y = y;
        this.z = z;
    }

    destroy(){
        if (this.parent){
            this.parent.removeSpell(this.data.id)
            this.parent = null
        }
    }
}