
// Hero class definition
class Hero extends Role{
    constructor(heroData, isSelfHero) {
        super(heroData, "blue")
        this.isSelfHero = isSelfHero;

    }

	update(deltaTime, camera){
		super.update(deltaTime, camera)

        if (this.isSelfHero){
            this.updateSyncMovePaths();
        }
	}

    draw(ctx) {
        super.draw(ctx)

    }

    stand(){
        super.stand()
        if (this.isSelfHero){
            Global.nano.request("SceneManager.HeroMoveStop", {
                uid:this.id,
                pos_x: getGridXByPixel(this.x),
                pos_y: getGridYByPixel(this.y),
                pos_z: 0 ,
            })
        }
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

    getSyncPaths() {
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