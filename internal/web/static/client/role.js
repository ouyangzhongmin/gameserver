
// Hero class definition
class Role extends RoleState{
    constructor(data, color) {
        super();
        this.data = data;
        this.id = data.id
        this.name = data.name;
        this.screenX = 0;
        this.screenY = 0;
        this.x = 0;
        this.y = 0;
        this.z = 0;
        this.setPos(getPixelXByGrid(+data.pos_x),getPixelYByGrid(+data.pos_y), 0)
        this.color = color;
        this.radius = 15;

        this.isSelfHero = false;

        this.isAutoMoving = false;

        this._paths = null;
        this._pathIndex = 0;
        this._lastPathIndex = -1;

        this.dir = Direction.E;
        this.stepTime = 200;
        this.speed = getSpeed(this.stepTime);

        this._needSyncPaths = false;
        this._nextSyncPathIndex = 0;
        this._syncPathsLen = 5;

        this._correctTargetX = -1;
        this._correctTargetY = -1;
        this._isCorrecting = false;

        if (this.data.life <= 0){
            this.die();
        }

        /**
         * {
         *     msg:"aaa",
         *     duration:3000,
         *     color: 'red',
         *     offsetx: 0,//每次会位移坐标
         *     offsety: 0,
         *     alpha: 1, //从不透明渐变到透明
         *     flowup: 1, //向上流动
         * }
         * @type {*[]}
         */
        this.bubbleMsgs = []
    }

    update(deltaTime, camera){
        if (this.isAutoMoving){
            this.updateMove(deltaTime);
        }
        if (this._isCorrecting){
            this.updateCorrectPos(deltaTime);
        }
        this.screenX = this.x - camera.x;
        this.screenY = this.y - camera.y;

        if (this.bubbleMsgs){
            for (let i = this.bubbleMsgs.length -1 ;i >= 0 ; i--){
                let bmsg = this.bubbleMsgs[i]
                bmsg.duration -= deltaTime
                if (bmsg.flowup === 1){
                    bmsg.offsetx += 0.1
                    bmsg.offsety += 0.4
                }
                bmsg.alpha -= parseFloat((1 / bmsg.duration).toFixed(3))
                if (bmsg.alpha < 0){
                    bmsg.alpha = 0
                }
                if ( bmsg.duration <= 0){
                    //删除对象
                    this.bubbleMsgs.splice(i, 1)
                }
            }
        }
    }

    draw(ctx) {
        // Draw the circle representing the monster
        //console.log("entity.draw::", this.name, this.screenX, this.screenY)
        ctx.beginPath();
        ctx.arc(this.screenX, this.screenY, this.radius, 0, Math.PI * 2, false);
        ctx.fillStyle = this.color;
        ctx.fill();
        // ���Ʊ߿�
        ctx.lineWidth = this.borderWidth;
        ctx.strokeStyle = 'black';
        ctx.stroke();

        // ��������
        ctx.fillStyle = 'black';
        ctx.font = '11px Arial';
        ctx.textAlign = 'center';
        ctx.fillText(this.name+`【${this.data.level}级】`, this.screenX, this.screenY - this.radius - 5);

        if (this.bubbleMsgs){
            //绘制冒泡
            for (let i = 0 ;i < this.bubbleMsgs.length ; i++){
                let bmsg = this.bubbleMsgs[i]
                let rgba = colorNameToRGBA(bmsg.color, bmsg.alpha)
                ctx.fillStyle = rgba;
                // console.log("rgba::", rgba)
                ctx.fillText(bmsg.msg, this.screenX + bmsg.offsetx, this.screenY - this.radius - 25 - i*5 - bmsg.offsety);
            }
        }
        // ����Ѫ��������
        const barWidth = this.radius * 2;
        const barHeight = 3;

        // MP������
        ctx.fillStyle = 'darkblue';
        ctx.fillRect(this.screenX - barWidth / 2, this.screenY - this.radius - 20, barWidth, barHeight);
        // MP��
        ctx.fillStyle = 'blue';
        ctx.fillRect(this.screenX - barWidth / 2, this.screenY - this.radius - 20, barWidth * (this.data.mana / this.data.max_mana), barHeight);

        // HP������
        ctx.fillStyle = 'darkred';
        ctx.fillRect(this.screenX - barWidth / 2, this.screenY - this.radius - 25, barWidth, barHeight);
        // HP��
        ctx.fillStyle = 'red';
        ctx.fillRect(this.screenX - barWidth / 2, this.screenY - this.radius - 25, barWidth * (this.data.life / this.data.max_life), barHeight);



        // ����HP��MP�ı�
        // ctx.fillStyle = 'white';
        // ctx.fillText(`HP: ${this.data.life}`, this.screenX, this.screenY - this.radius - 30);
        // ctx.fillText(`MP: ${this.data.mana}`, this.screenX, this.screenY - this.radius - 50);
    }

    setPos(x,y,z){
        // console.log(this.name, "setPos::", x, y , z)
        this.x = x;
        this.y = y;
        this.z = z;
    }

    lifeChanged(damage, life, max_life){
        this.data.life = life;
        this.data.max_life = max_life;

        this.bubble(damage < 0 ? '+' + damage : "-" + damage, 1000, "red", true)
    }

    manaChanged(cost, mana, max_mana){
        this.data.mana = mana;
        this.data.max_mana = max_mana;
        //this.bubble('-'+cost, 1000, "blue", true)
    }

    // paths中的数据paths[i][0]为y轴, paths[i][1]为X轴
    moveByPaths(paths, stepTime){
        // console.log(this.name , "moveByPaths:", paths)
        if (!paths || paths.length === 0){
            console.log("移动路径为空")
            return
        }
        this.clearCorrectPos();
        this._paths = paths;
        this.isAutoMoving = true;
        this._lastPathIndex = -1;
        this._pathIndex = 0;
        this._nextSyncPathIndex = 0;
        this._needSyncPaths = false;
        if (stepTime > 0){
            this.speed = getSpeed(stepTime)
        }
    }

    updateMove(deltaTime){
        if (!this._paths || this._paths.length === 0){
            return;
        }
        // paths中的数据paths[i][0]为y轴, paths[i][1]为X轴
        const tempTargetX = getPixelXByGrid(this._paths[this._pathIndex][1]);
        const tempTargetY = getPixelYByGrid(this._paths[this._pathIndex][0]);
        const tempDist = getDistance(this.x , this.y , tempTargetX , tempTargetY);
        const tempAngle = getAngle(this.x , this.y , tempTargetX , tempTargetY);
        const tempActualSpeed = this.speed * deltaTime;
        if(this._lastPathIndex !== this._pathIndex)
        {
            if(tempDist > tempActualSpeed) {
                this.dir = getDirection(tempAngle);
            }
            this._lastPathIndex = this._pathIndex;
            if(this._pathIndex === 0)
            {
                this._needSyncPaths = true;
                this._nextSyncPathIndex = this._pathIndex + this._syncPathsLen;
            }
        }
        if(tempDist <= Math.floor(tempActualSpeed + 0.5))
        {
            this.setPos(tempTargetX,tempTargetY, 0 )
            this._pathIndex++;
            if(this._pathIndex > this._paths.length - 1)
            {
                this.stand();
                this.invokeWalkCompleted(true);
            }
            else
            {
                if( this._pathIndex >=  this._nextSyncPathIndex)
                {
                    this._needSyncPaths = true;
                    this._nextSyncPathIndex =  this._pathIndex +  this._syncPathsLen;
                }
            }
        }
        else
        {
            this.speedX = cosD(tempAngle) * tempActualSpeed;
            this.speedY = sinD(tempAngle) * tempActualSpeed;
            this.setPos(this.x +this.speedX, this.y + this.speedY , 0)
            this.walk();
        }
    }

    //修正位置
    correctPos(tx,ty){
        this._correctTargetX = tx;
        this._correctTargetY = ty;
        this._isCorrecting = true;
    }

    updateCorrectPos(deltaTime){
        if (!this._isCorrecting){
            return
        }
        const tempTargetX = this._correctTargetX;
        const tempTargetY = this._correctTargetY;
        const tempDist = getDistance(this.x , this.y , tempTargetX , tempTargetY);
        const tempAngle = getAngle(this.x , this.y , tempTargetX , tempTargetY);
        const correctSpeed = getSpeed(150)
        const tempActualSpeed = correctSpeed * deltaTime;
        if(tempDist <= Math.floor(tempActualSpeed + 0.5)) {
            this.setPos(tempTargetX, tempTargetY, 0)
            this.stand();
        }else
        {
            this.speedX = cosD(tempAngle) * tempActualSpeed;
            this.speedY = sinD(tempAngle) * tempActualSpeed;
            this.setPos(this.x +this.speedX, this.y + this.speedY , 0)
            this.walk();
        }
    }

    stand(){
        super.stand()
        this.clearMovePaths();
        this.clearCorrectPos();
    }

    clearMovePaths(){
        this._paths = null;
        this._pathIndex = 0;
        this._lastPathIndex = -1;
        this.isAutoMoving = false;
    }

    clearCorrectPos(){
        this._isCorrecting = false;
        this._correctTargetX = -1;
        this._correctTargetY = -1;
    }

    die()
    {
        super.die()
        this.name += "【dead】";
        this.clearMovePaths();
        this.clearCorrectPos();
    }

    invokeWalkCompleted(success){

    }

    // 冒泡
    // flowup bool 是否向上流动
    bubble(msg, duration, color, flowup){
        if (!flowup){
            this.bubbleMsgs.unshift({
                msg: msg,
                duration:duration,
                color: color,
                offsetx: 0,//每次会位移坐标
                offsety: 0,
                alpha: 1, //从不透明渐变到透明
                flowup: 0,
            })
        }else{
            this.bubbleMsgs.unshift({
                msg: msg,
                duration:duration,
                color: color,
                offsetx: 0,//每次会位移坐标
                offsety: 0,
                alpha: 1, //从不透明渐变到透明
                flowup: 1, //向上流动
            })
        }
    }
}