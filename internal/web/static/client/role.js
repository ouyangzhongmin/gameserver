
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
        this.life = data.life;
        this.mana = data.mana;
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
    }

    update(deltaTime, camera){
        if (this.isAutoMoving){
            this.updateMove(deltaTime);
        }
        this.screenX = this.x - camera.x;
        this.screenY = this.y - camera.y;

        if (this.isSelfHero){
            this.updateSyncMovePaths();
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
        ctx.font = '14px Arial';
        ctx.textAlign = 'center';
        ctx.fillText(this.name, this.screenX, this.screenY - this.radius - 10);

        // ����Ѫ��������
        // const barWidth = this.radius * 1.6;
        // const barHeight = 3;
        //
        // // HP������
        // ctx.fillStyle = 'darkred';
        // ctx.fillRect(this.screenX - barWidth / 2, this.screenY - this.radius - 15, barWidth, barHeight);
        // // HP��
        // ctx.fillStyle = 'red';
        // ctx.fillRect(this.screenX - barWidth / 2, this.screenY - this.radius - 15, barWidth * (this.data.life / this.data.max_life), barHeight);
        //
        // // MP������
        // ctx.fillStyle = 'darkblue';
        // ctx.fillRect(this.screenX - barWidth / 2, this.screenY - this.radius - 35, barWidth, barHeight);
        // // MP��
        // ctx.fillStyle = 'blue';
        // ctx.fillRect(this.screenX - barWidth / 2, this.screenY - this.radius - 35, barWidth * (this.data.mana / this.data.max_mana), barHeight);
        //
        // // ����HP��MP�ı�
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

    // paths中的数据paths[i][0]为y轴, paths[i][1]为X轴
    moveByPaths(paths, stepTime){
        // console.log(this.name , "moveByPaths:", paths)
        if (!paths || paths.length === 0){
            console.log("移动路径为空")
            return
        }
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

    stand(){
        super.stand()
        this._paths = null;
        this._pathIndex = 0;
        this._lastPathIndex = -1;
        this.isAutoMoving = false;

        if (this.isSelfHero){
            Global.nano.request("SceneManager.HeroMoveStop", {
                uid:this.id,
                pos_x: getGridXByPixel(this.x),
                pos_y: getGridYByPixel(this.y),
                pos_z: 0 ,
            })
        }
    }

    invokeWalkCompleted(success){

    }

    updateSyncMovePaths() {
        if(!this._needSyncPaths || !this.isSelfHero) {
            return
        }
        const paths = this.getSyncPaths();
        if(!paths || paths.length === 0){
            return
        }
        Global.nano.request("SceneManager.HeroMove", {
            uid:this.id,
            trace_paths: paths,
        })
    }

    getSyncPaths() { //本类在C2SSyncComponent前面刷新
        if(this._needSyncPaths) {
            this._needSyncPaths = false;
            this._syncPaths = null;
            //本地开始走向_pathIndex这个目标点
            if(this._pathIndex === 0) {
                this._syncPaths = this._paths.slice(this._pathIndex , this._pathIndex + this._syncPathsLen + 1);
                if(this._syncPaths && this._syncPaths.length > 0) {
                    //判断当前所占格子是否已经在路径中，没有则添加
                    let selfGridX = getGridXByPixel(this.x);
                    let selfGridY = getGridYByPixel(this.y);
                    if(Math.abs(this._syncPaths[0][1] - selfGridX) > 5 || Math.abs(this._syncPaths[0][0] - selfGridY) > 5)
                    {
                        this._syncPaths = [[selfGridY, selfGridX ]];
                    }
                }
            } else {
                this._syncPaths = this._paths.slice(this._pathIndex - 1 , this._pathIndex + this._syncPathsLen + 1);
            }

            if(this.isIndexOf(this._syncPaths , this._lastSyncPaths)) {
                console.log(this.name + ' 当前要同步的路径在上一条同步的路径中被包含了，不需要再同步');
                return null;
            }
            if(this._syncPaths && Global.block.isBlock(this._syncPaths[0][1] , this._syncPaths[0][0])) {
                let selfGridX = getGridXByPixel(this.x);
                let selfGridY = getGridYByPixel(this.y);
                console.log(this.name, "第一格不可以行走")
                this._syncPaths[0][0] = selfGridY;
                this._syncPaths[0][1] = selfGridX;
            }
            this._lastSyncPaths = this._syncPaths;
            return this._syncPaths;
        }
        return null;
    }

    /**
     * 当前的路径是否在上一条路径中已经被包含了
     * @param syncPaths
     * @param lastSyncPaths
     * @return
     */
    isIndexOf(syncPaths  , lastSyncPaths ) 
    {
        if(!syncPaths || !lastSyncPaths)
            return false;

        const len  = syncPaths.length;
        const lastLen  = lastSyncPaths.length;
        if(len > lastLen)
            return false;
        let j = 0;
        for(j  = 0 ; j < lastLen ; j++)
        {
            if(lastSyncPaths[j][0] === syncPaths[0][0] && lastSyncPaths[j][1] === syncPaths[0][1])
                break;
        }
        for(let i  = 0 ; i < len ; i++)
        {
            if(j + i >= lastLen)
                return false;
            if(lastSyncPaths[j + i][0] !== syncPaths[i][0] || lastSyncPaths[j + i][1] !== syncPaths[i][1])
                return false;
        }
        return true;
    }
}