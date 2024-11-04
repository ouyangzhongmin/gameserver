
class GridShow {
    constructor() {
        this.grid = null

        this.semiTransparentRed = 'rgba(255, 0, 0, 0.5)';
    }

    update(deltaTime, camera){
        if (!this.grid || this.grid.length === 0){
            return
        }
        this.startX = getGridXByPixel(camera.x);
        this.startY = getGridYByPixel(camera.y);
        this.endX = this.startX + Math.ceil(camera.width/GRID_WIDTH)
        this.endY = this.startY + Math.ceil(camera.height/GRID_HEIGHT)
        if (this.endY >= this.grid.length){
            this.endY = this.grid.length - 1;
        }
        if (this.endX >= this.grid[0].length){
            this.endX = this.grid[0].length - 1;
        }
    }

    draw(ctx) {
        if (!this.grid || this.grid.length === 0){
            return
        }
        // console.log("gridShow.draw0::", this.startY, this.endY, this.endX , this.startX)
        for (let row = 0; row < (this.endY - this.startY); row++) {
            for (let col = 0; col < (this.endX - this.startX); col++) {
                ctx.strokeStyle = 'black';
                ctx.strokeRect(col * GRID_WIDTH, row * GRID_HEIGHT, GRID_WIDTH, GRID_HEIGHT);
                // console.log("gridShow.draw::", row+this.startY, col+this.startX, this.grid[row+this.startY][col+this.startX])
                if (this.grid[row+this.startY][col+this.startX] === 0) {
                    ctx.fillStyle = this.semiTransparentRed;
                    ctx.fillRect(col * GRID_WIDTH, row * GRID_HEIGHT, GRID_WIDTH, GRID_HEIGHT);
                }
            }
        }
    }

    show(grid){
        this.grid = grid;
        // console.log("gridShow.show:", grid)
    }

}