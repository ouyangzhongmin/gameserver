
class Block {
    constructor() {
        this.grid = null
    }

    setGrid(grid){
        this.grid = grid;
    }

    isBlock(x, y) {
        if (!this.grid){
            return false;
        }
        return this.grid[y][x] === 0
    }

    // 返回的paths中的数据paths[i][0]为y轴, paths[i][1]为X轴
    findPath(start, end){
        let graph = new Graph(this.grid, { diagonal: true })
        const startp = graph.grid[start.y][start.x];
        const endp = graph.grid[end.y][end.x];
        const paths = astar.search(graph, startp, endp, { heuristic: astar.heuristics.diagonal });
        let newPaths = []
        if (paths){
            for(let i = 0; i < paths.length; i++){
                //这个Astar中返回的X为坐标系的Y轴，对应二维数组的1第一维
                newPaths.push([paths[i].x, paths[i].y])
            }
        }
        console.log("findPath:", start, end, startp, endp, "paths:::",  paths, ",newPaths:::",  newPaths)
        return newPaths;
    }
}
