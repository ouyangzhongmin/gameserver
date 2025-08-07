
class CellsShow{
    constructor() {
        this.cells = []
    }

    update(deltaTime, camera){
        if (!this.isShow()){
            return
        }
        this.startX = camera.x;
        this.endX = this.startX + camera.width;
        /**
         * [
         *     {
         *         "cell_id": 10001,
         *         "scene_id": 1,
         *         "bounds": {
         *             "X": 0,
         *             "Y": 0,
         *             "Width": 250,
         *             "Height": 188
         *         },
         *         "edge_size": 30,
         *     }
         * ]
         */
        for (let i =0; i < this.cells.length; i++) {
            const cell = this.cells[i]
            // 计算cell在屏幕上的位置（地图坐标转屏幕坐标）
            const cellX = getPixelXByGrid(cell.bounds.X)
            const cellWidth = cell.bounds.Width * GRID_WIDTH;

            if ((cellX < this.endX && cellX > this.startX ) ||
                (cellX + cellWidth < this.endX && cellX + cellWidth > this.startX)) {
                // 在屏幕内 计算绘制范围
                const cellStartX = Math.max(this.startX, cellX)
                cell.draw = {}
                cell.draw.x = cellStartX - this.startX;
                cell.draw.y = 0;
                cell.draw.width = Math.min(cellStartX + cellWidth, this.endX) - cellStartX;
                cell.draw.height = camera.height;
            }else{
                cell.draw = null;
            }
        }
    }

    draw(ctx) {
        if (!this.isShow()){
            return
        }

        for (let i =0; i < this.cells.length; i++) {
            const cell = this.cells[i]
            if (cell.draw){
                // 根据cell_id生成唯一颜色（使用黄金角分割确保颜色多样性）
                const hue = (cell.cell_id * 137.508) % 360; // 137.508°是黄金角
                ctx.fillStyle = `hsla(${hue}, 70%, 50%, 0.7)`;
                // 绘制cell区域
                ctx.fillRect(cell.draw.x, cell.draw.y, cell.draw.width, cell.draw.height);

                // 绘制cell边界（黑色边框）
                ctx.strokeStyle = '#000000';
                ctx.lineWidth = 1;
                ctx.strokeRect(cell.draw.x, cell.draw.y, cell.draw.width, cell.draw.height);
            }
        }
    }

    show(cells){
        this.cells = cells;
    }

    isShow(){
        return this.cells && this.cells.length > 0
    }

}