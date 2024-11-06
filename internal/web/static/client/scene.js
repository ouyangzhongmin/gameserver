

class Scene {
    constructor(canvas) {
        this.monsters = {}
		this.heros = {}
		this.deads = []
		this.selfHero = new Hero(1, 'HeroSelf', 200, 150, 150, 100, 'green');
		this.selfHero.isSelfHero = true;
		this.sceneId = 0
		this.sceneName = "";
		this.camera = new Camera(canvas.width, canvas.height);
		this.map = null;
		this.gridShow = new GridShow();
    }

	resize(w, h){
		this.camera.resize(w, h)
	}

	changeScene(data){
		console.log("scene.change::", data)
		this.sceneData = data.scene;
		this.doorList = data.doors;
		this.sceneId = this.sceneData.scene_id;
		this.sceneName = this.sceneData.scene_name;
		this.mapFile = this.sceneData.map_file;
		this.map = new SceneMap(this.mapFile);
		this.loadBlock();

		this.createSelfHero();

		// let rpos = generateRandomPoint(1920,1024)
		// let m1 = new Monster(1, 'Dragon', rpos.x, rpos.y, 150, 100, 'red')
		// rpos = generateRandomPoint(1920,1024)
        // let m2 = new Monster(2, 'Goblin', rpos.x, rpos.y, 80, 60, 'red')
		// rpos = generateRandomPoint(1920,1024)
        // let m3 = new Monster(3, 'Wizard', rpos.x, rpos.y, 100, 150, 'red')
		// this.monsters[m1.id] = m1
		// this.monsters[m2.id] = m2
		// this.monsters[m3.id] = m3
		//
        // rpos = generateRandomPoint(1920,1024)
		// let h2 = new Hero(2, 'Hero2', rpos.x, rpos.y, 150, 100, 'blue')
		// rpos = generateRandomPoint(1920,1024)
        // let h3 = new Hero(3, 'Hero3', rpos.x, rpos.y, 80, 60, 'blue')
		// rpos = generateRandomPoint(1920,1024)
        // let h4 = new Hero(4, 'Hero4', rpos.x, rpos.y, 100, 150, 'blue')
		// this.heros[h2.id] = h2
		// this.heros[h3.id] = h3
		// this.heros[h4.id] = h4
		// this.heros[this.selfHero.id] = this.selfHero
	}

	loadBlock(){
		const that = this;
		const xmlUrl = `assets/map/${this.mapFile}/${this.mapFile}.xml`;
		console.log("loadXml:::", xmlUrl)
		Global.http.fetchFileAsByteArray(xmlUrl).then(byteArray => {
			if (byteArray) {
				const xmlString = new TextDecoder().decode(byteArray);
				// console.log('xml content as string:', xmlString);

				const parser = new DOMParser();
				const xmlDoc = parser.parseFromString(xmlString, "text/xml");

				const sceneElement = xmlDoc.getElementsByTagName("scene")[0];
				const mapRows = sceneElement.getAttribute('mapRows');
				const mapCols = sceneElement.getAttribute('mapCols');
				const tileWidth = sceneElement.getAttribute('tileWidth');
				const tileHeight = sceneElement.getAttribute('tileHeight');
				console.log("initMap::", mapRows, mapCols, tileWidth, tileHeight )
				that.map.initMap(mapRows, mapCols, tileWidth, tileHeight);
				that.camera.setMaxViewWH(this.map.mapWidth, this.map.mapHeight);
			}
		});

		const fileUrl = `assets/map/${this.mapFile}/${this.mapFile}.block`;
		console.log("loadBlock:::", fileUrl)
		Global.http.fetchFileAsByteArray(fileUrl).then(byteArray => {
			if (byteArray) {
				console.log('File read as byte array:', byteArray);
				try {
					// DataView的字节序默认是大端序（big-endian）
					const dataView = new DataView(byteArray.buffer, byteArray.byteOffset, byteArray.length); // 从索引1开始，读取2个字节
					const cols = dataView.getUint16(0);
					const rows = dataView.getUint16(2);
					// let grid = new Grid({
					// 	col:cols,                  // col
					// 	row:rows,                   // row
					// 	render:function(){       // Optional, this method is triggered when the grid point changes
					// 		// console.log(this);
					// 	}
					// });
					let grid = [rows]
					console.log("grid::", grid)
					let offset = 4;
					for (let j = 0; j < rows; j++ ){
						let grid2 = [cols]
						grid[j] = grid2
						for (let i = 0; i < cols; i++){
							const b = dataView.getInt8(offset ++)
							if (b === 0){ //编辑器是127为可以走，0为不能走
								//console.log("isblock:::", i, j, b)
								grid2[i] = 0
							}else{
								grid2[i] = 1
							}
						}
					}
					Global.block.setGrid(grid);
					//需要显示的时候打开
					// that.gridShow.show(grid)
					console.log('init map blocks:', rows, cols);
					// 将Uint8Array转换为ArrayBuffer
					// const compressedData  = this.decryptByteArray(byteArray)
					// console.log('File read as byte array2:', compressedData );
					// const decompressedData = pako.inflate(compressedData );

					// 将ArrayBuffer转换回字符串（假设原始数据是字符串）
					// const decoder = new TextDecoder("utf-8");
					// const decompressedString = decoder.decode(decompressedData);

				} catch (err) {
					console.error('new grid失败:', err);
				}
			}
		});
	}

	update(deltaTime, width, height) {
		this.camera.update(deltaTime);
		if (this.map){
			this.map.update(deltaTime, this.camera)
		}
		if (this.gridShow){
			this.gridShow.update(deltaTime, this.camera);
		}
		for (let key in this.heros) {
			this.heros[key].update(deltaTime, this.camera);
		}
		for (let key in this.monsters) {
			this.monsters[key].update(deltaTime, this.camera);
		}
		if (this.deads.length > 0){
			let ts = Date.now()
			for (let i = this.deads.length -1;i >= 0 ; i--) {
				this.deads[i].update(deltaTime, this.camera);
				if (this.deads[i].deadTimeStamp + 20*1000 < ts){
					//清理掉
					this.deads.splice(i, 1)
				}
			}
		}
	}

    draw(ctx) {
		//console.log("scene.draw")
		if (this.map){
			// this.map.draw(ctx);
		}
		if (this.gridShow){
			this.gridShow.draw(ctx);
		}
		for (let key in this.heros) {
			this.heros[key].draw(ctx)
		}
		for (let key in this.monsters) {
			this.monsters[key].draw(ctx)
		}
		if (this.deads.length > 0){
			for (let i = this.deads.length -1;i >= 0 ; i--) {
				this.deads[i].draw(ctx);
			}
		}
	}

	onMouseUp(event) {
        const rect = canvas.getBoundingClientRect();
        const x = event.clientX - rect.left;
        const y = event.clientY - rect.top;
		if(this.selfHero.isDead()){
			this.selfHero.bubble("已死亡，无法移动", 1000, "red", true)
			return
		}

		const paths = this.findPath(this.selfHero.x, this.selfHero.y, x + this.camera.x, y + this.camera.y)
        this.selfHero.moveByPaths(paths);
    }

	onMousemove(event){
		// 获取鼠标的X和Y坐标
		if (this.gridShow.isShow()){
			const rect = canvas.getBoundingClientRect();
			const x = event.clientX - rect.left;
			const y = event.clientY - rect.top;
			console.log("cur grid::", getGridYByPixel(y + this.camera.y), getGridXByPixel(x + this.camera.x))
		}
	}

	findPath(sx,sy, ex,ey){
		const start = {x:getGridXByPixel(sx), y:getGridYByPixel(sy)} ;
		const end = {x:getGridXByPixel(ex), y:getGridYByPixel(ey)} ;
		//console.log("findPath::", start, end, "paths::", paths)
		return Global.block.findPath(start, end);
	}

	createSelfHero(){
		console.log("createSelfHero::", Global.selfHeroData)
		this.selfHero = new Hero(Global.selfHeroData, true);
		this.heros[this.selfHero.id] = this.selfHero;
		this.camera.focusAt(this.selfHero)
	}

	addHero(heroData){
		console.log("addHero:", heroData)
		let hero = new Hero(heroData, false);
		this.heros[hero.id] = hero;
	}

	removeHero(heroId){
		delete this.heros[heroId]
	}

	addMonster(monsterData){
		let monster = new Monster(monsterData);
		this.monsters[monster.id] = monster;
		console.log("addMonster:", monsterData, "当前数量:", Object.keys(this.monsters).length)
	}

	removeMonster(monsterId){
		delete this.monsters[monsterId]
	}

	heroMove(heroId, tracePaths, stepTime){
		const hero = this.heros[heroId]
		if (!hero){
			console.error(`hero:${heroId}不存在`)
			return;
		}
		hero.moveByPaths(tracePaths, stepTime);
	}

	heroMoveStop(heroId, posX, posY, posZ){
		const hero = this.heros[heroId]
		if (!hero){
			console.error(`hero:${heroId}不存在`)
			return;
		}
		hero.stand();
		hero.x = getPixelXByGrid(posX)
		hero.y = getPixelYByGrid(posY)
	}

	monsterMove(monserId, tracePaths, stepTime){
		//服务器上A*的路径tracePaths[[y,x]] 前面的是Y轴的数据，后面的是X轴的数据
		const monster = this.monsters[monserId]
		if (!monster){
			console.error(`monster:${monserId}不存在`)
			return;
		}
		monster.moveByPaths(tracePaths, stepTime);
	}

	monsterMoveStop(monserId, posX, posY, posZ){
		const monster = this.monsters[monserId]
		if (!monster){
			console.error(`monster:${monserId}不存在`)
			return;
		}
		monster.stand();
		monster.x = getPixelXByGrid(posX)
		monster.y = getPixelYByGrid(posY)
	}

	lifeChanged(data){
		const entityType = data.entity_type;
		if (entityType === 0){
			//hero
			const hero = this.heros[data.id]
			if (!hero){
				console.error(`hero:${data.id}不存在`)
				return;
			}
			hero.lifeChanged(data.damage, data.life, data.max_life)
		}else if(entityType === 1){
			//monster
			const monster = this.monsters[data.id]
			if (!monster){
				console.error(`monster:${data.id}不存在`)
				return;
			}
			monster.lifeChanged(data.damage, data.life, data.max_life)
		}
	}

	manaChanged(data){
		const entityType = data.entity_type;
		if (entityType === 0){
			//hero
			const hero = this.heros[data.id]
			if (!hero){
				console.error(`hero:${data.id}不存在`)
				return;
			}
			hero.manaChanged(data.cost, data.mana, data.max_mana)
		}else if(entityType === 1){
			//monster
			const monster = this.monsters[data.id]
			if (!monster){
				console.error(`monster:${data.id}不存在`)
				return;
			}
			monster.manaChanged(data.cost, data.mana, data.max_mana)
		}
	}

	entityDie(data){
		const entityType = data.entity_type;
		if (entityType === 0){
			const hero = this.heros[data.id]
			if (!hero){
				console.error(`hero:${data.id}不存在`)
				return;
			}
			hero.die()
			if (hero.id !== this.selfHero.id){
				this.removeHero(hero.id)
				// todo 死亡残留10S再删除
				hero.id = hero.id*10000 + Date.now()
				this.deads.push(hero)
			}
		}else if(entityType === 1){
			//monster
			const monster = this.monsters[data.id]
			if (!monster){
				console.error(`monster:${data.id}不存在`)
				return;
			}
			this.removeMonster(monster.id)
			// todo 死亡残留10S再删除
			monster.id = monster.id*10000 + Date.now()
			monster.die()
			this.deads.push(monster)
		}
	}

	heroTextMessage(heroId, msg){
		const hero = this.heros[heroId]
		if (!hero){
			console.error(`hero:${heroId}不存在`)
			return;
		}
		hero.bubble(msg, 3000, "white", false)
	}
}
