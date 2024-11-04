class SceneMap {
	constructor(mapName) {
		this.mapName = mapName;
		this.mapRows = 0; //多少行
		this.mapCols = 0; //多少列
		this.tileWidth = 200; //每块高度
		this.tileHeight = 200; //每块宽度

		this.tiles = new Map(); // 用于存储已加载的瓦片图像
		this.visibleTiles = new Set(); // 用于跟踪当前可视的瓦片
		this.toReleaseTiles = new Map();

		this.lastCameraPosition = { x: 0, y: 0 }; // 上一次英雄的位置
		this.cameraWidth = -1;
		this.cameraHeight = -1;
		this.inited = false;
	}

	initMap(mapRows,mapCols,tileWidth,tileHeight){
		this.mapRows = mapRows;
		this.mapCols = mapCols;
		this.tileWidth = tileWidth;
		this.tileHeight = tileHeight;

		this.mapWidth = this.mapCols * this.tileWidth;
		this.mapHeight = this.mapRows * this.tileHeight;
		this.inited = true;
	}

	// 更新可视瓦片，并清除不再可视的瓦片（实际上只是停止绘制它们）
	updateVisibleTiles(camera, nowts) {
		const x = camera.x, y = camera.y, cameraWidth = camera.width, cameraHeight = camera.height;

		if (this.lastCameraPosition.x === x && this.lastCameraPosition.y === y &&
			this.cameraWidth === cameraWidth && this.cameraHeight === cameraHeight){
			//没移动的时候就不刷新了
			return
		}
		const visibleTilesX = Math.ceil(cameraWidth / this.tileWidth) + 1; // 加1以确保边缘瓦片也被考虑在内
		const visibleTilesY = Math.ceil(cameraHeight / this.tileHeight) + 1;

		// 计算新的可视瓦片范围
		// 计算可视区域的起始瓦片坐标
		this.startX = Math.floor(x / this.tileWidth);
		this.startY = Math.floor(y / this.tileHeight);

		// 确保起始坐标不会越界
		this.startX = Math.max(0, this.startX);
		this.startY = Math.max(0, this.startY);

		this.endX = Math.min(this.mapCols, this.startX + visibleTilesX);
		this.endY = Math.min(this.mapRows, this.startY + visibleTilesY);

		// 创建一个新的集合来存储新的可视瓦片
		const newVisibleTiles = new Set();
		for (let y = this.startY; y < this.endY; y++) {
			for (let x = this.startX; x < this.endX; x++) {
				const tileName = `${this.mapName}_${x}_${y}`;
				if (!this.tiles.has(tileName)) {
					const img = new Image();
					img.src = `assets/map/${this.mapName}/pic/${tileName}.jpg`;
					// 这里可以添加异步加载的逻辑，但为了简化，我们假设图像立即加载完成
					this.tiles.set(tileName, img);
				}
				newVisibleTiles.add(`${x},${y}`);
				if (this.toReleaseTiles.has(tileName)){
					this.toReleaseTiles.delete(tileName)
				}
			}
		}

		// 找出不再可视的瓦片（即那些在新集合中不存在的旧瓦片）
		// 并从this.visibleTiles中移除它们（实际上我们不需要显式移除，因为稍后我们会用新集合替换它）
		// 但这里为了演示，我们还是遍历一下旧集合
		this.visibleTiles.forEach(tileKey => {
			if (!newVisibleTiles.has(tileKey)) {
				const [x, y] = tileKey.split(',').map(Number);
				const tileName = `${this.mapName}_${x}_${y}`;
				// 这里可以执行任何需要在瓦片不再可视时进行的操作
				// 但在Canvas上下文中，通常不需要显式卸载图像，因为图像对象仍在内存中
				// 我们只是简单地不再绘制这个瓦片
				if (this.tiles.has(tileName)) {
					// 注意：我们实际上没有从内存中“清除”不再可视的瓦片图像对象
					// 如果真的需要释放内存，你可以考虑在不再需要某个图像时将其设置为null
					// 并确保没有其他引用指向它，这样垃圾回收机制就会回收它的内存
					//todo 把带释放的位图id放入到一个set中，可以记录个时间戳，然后遍历时间戳到达时间则释放掉
					this.toReleaseTiles.set({tileName, ts: nowts})
				}
			}
		});
		// console.log("newVisibleTiles:", newVisibleTiles, this.startX, this.startY, this.endX, this.endY, cameraWidth, cameraHeight)
		// 更新this.visibleTiles为新的可视瓦片集合
		this.visibleTiles = newVisibleTiles;
		this.lastCameraPosition.x = x;
		this.lastCameraPosition.y = y;
		this.cameraWidth = cameraWidth;
		this.cameraHeight = cameraHeight;

	}

	update(deltaTime, camera) {
		if (!this.inited){
			return;
		}
		const nowts = Date.now()
		// 更新可视区域
		this.updateVisibleTiles(camera, nowts);


		//释放不需要的图片
		this.toReleaseTiles.forEach (function(value, tileName) {
			if (value && (value.ts + 20000 < nowts) ){
				const img = this.tiles.get(tileName);
				if (img) {
					img.destroy()
				}
				this.tiles.delete(tileName)
				this.toReleaseTiles.delete(tileName)
			}
		})
	}

	// 绘制场景（可视区域内的瓦片）
	draw(ctx) {
		if (!this.inited){
			return;
		}
		// 遍历可视瓦片并绘制它们
		this.visibleTiles.forEach(tileKey => {
			const [x, y] = tileKey.split(',').map(Number);
			const tileName = `${this.mapName}_${x}_${y}`;
			const img = this.tiles.get(tileName);
			if (img) {
				const drawX = x* this.tileWidth - this.lastCameraPosition.x ; // 计算瓦片在画布上的x坐标
				const drawY = y* this.tileHeight - this.lastCameraPosition.y ; // 计算瓦片在画布上的y坐标
				// console.log("img:", drawX, drawY)
				ctx.drawImage(img, drawX, drawY, this.tileWidth, this.tileHeight);
			}
		});

	}
}
