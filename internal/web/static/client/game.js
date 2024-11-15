
const canvas = document.getElementById('gameCanvas');
const ctx = canvas.getContext('2d');

// 游戏类定义
class Game {
    constructor() {
        this.nano = window.nano
        this.lastTime = Date.now();
        this.sceneLoaded = false; // 标记场景是否已加载
        canvas.width = window.innerWidth;
        canvas.height = window.innerHeight;
        this.scene = new Scene(canvas);
        this.backgroundColor = 'lightblue';
        this.connected = false;
        // this.uniqueId = localStorage.getItem("uniqueId")
        // if (!this.uniqueId || this.uniqueId === ''){
        //     this.uniqueId = generateUniqueId();
        //     console.log("生成唯一id:", this.uniqueId)
        //     localStorage.setItem("uniqueId", this.uniqueId)
        // }
        this.uniqueId = generateUniqueId();
        // this.uniqueId = "aa8f-839a-4a7d-b0fe-4f14-abaa";
        const that = this;
        canvas.addEventListener('mouseup', this.onMouseUp.bind(this));
        canvas.addEventListener('mousemove', this.onMousemove.bind(this));
        window.addEventListener('resize', function() {
            console.log('窗口尺寸变化了！当前宽度：', window.innerWidth, '；当前高度：', window.innerHeight);
            canvas.width = window.innerWidth;
            canvas.height = window.innerHeight;
            that.scene.resize(canvas.width, canvas.height);
        });
        this.startGameLoop();
        this.requestLogin();

    }

    requestLogin(){
        const nano = this.nano;
        const that = this;
        console.log("requestLogin:", that.uniqueId)
         Global.http.post("/v1/user/login/guest", {
            appId:'aagame',
            channelId:"yyb",
            imei: that.uniqueId,
            device: {
                imei :that.uniqueId,
                os: "14.1",     //os版本号
                model: "Iphone",  //硬件型号
                ip: "",    //内网IP
                remote: "", //外网IP
            }
        }).then(userInfo => {
            console.log("登录成功:", userInfo)
            Global.userInfo = userInfo;
            setTimeout(()=>{
                that.initNano()
            }, 1000)
        }).catch(err =>{
            console.error(err)
         })
    }

    initNano(){
        const that = this;
        
        console.log("initNano", Global.WSAddr, window.location.host)
        that.nano.init({host: Global.WSAddr, port: Global.WSPort, log: true}, function(){
            that.connected = true;
            console.log("connected success!!")
            window.showToast("connected success!!")
            if (!Global.userInfo.hero_list || Global.userInfo.hero_list.length === 0){
                console.log("当前账号没有hero , 调用创建英雄", Global.userInfo.uid);
                nano.request("Manager.CreateHero", {
                    uid: Global.userInfo.uid,
                    avatar: "https://img2.baidu.com/it/u=3171875674,3530712457&fm=253&fmt=auto&app=120&f=JPEG?w=800&h=800",
                    name: "",
                    attr_type: 0,// 0 力量，1敏捷，2智力
                }, function(data){
                    console.log("当前创建的英雄:", data)
                    Global.selfHeroData = data;
                });
            }else{
                const heroData = Global.userInfo.hero_list[0];
                Global.selfHeroData = heroData;
                console.log("选择英雄aaaa：", heroData, Global.userInfo.uid);
                nano.request("Manager.ChooseHero", {
                    uid: Global.userInfo.uid,
                    hero_id: heroData.id,
                    ip: "",
                }, function(data){
                    console.log("当前选择的英雄:", data)
                    Global.selfHeroData = data;
                });
            }

            nano.on("disconnect", function(data){
                that.connected = false;
                console.log(data);
                console.log("disconnected!!")
                window.showToast("disconnected!!")
            });
            nano.on("io-error", function(data){
                that.connected = false;
                console.log(data);
                console.log("connected error!!")
                window.showToast("connected error!!")
            });

            nano.on("notify", function(data){
                console.log(data);
            });

            //处理服务器合并的消息体
            nano.on("OnMergeMessages", function (msgs){
                console.log("OnMergeMessages", msgs);
                if(Array.isArray(msgs)) {
                    for(let i=0, l=msgs.length; i<l; i++) {
                        nano.emit(msgs[i].route, msgs[i].msg);
                    }
                }
            });

            nano.on("OnEnterScene", function(data){
                Global.selfHeroData = data.hero_data
                nano.request("SceneManager.HeroSetViewRange", {
                    hero_id: Global.selfHeroData.id,
                    width: Math.ceil(canvas.width/GRID_WIDTH),
                    height: Math.ceil(canvas.height/GRID_HEIGHT),
                }, function(data){
                    console.log("设置视野范围:", data)
                });
                that.OnEnterScene(data)
            })
            nano.on("OnEnterView", function(data){
                that.OnEnterView(data)
            })
            nano.on("OnExitView", function(data){
                that.OnExitView(data)
            })
            nano.on("OnHeroMoveTrace", function(data){
                that.OnHeroMoveTrace(data)
            })
            nano.on("OnHeroMoveStopped", function(data){
                that.OnHeroMoveStopped(data)
            })
            nano.on("OnMonsterMoveTrace", function(data){
                that.OnMonsterMoveTrace(data)
            })
            nano.on("OnMonsterMoveStopped", function(data){
                that.OnMonsterMoveStopped(data)
            })
            nano.on("OnMonsterCommonAttack", function(data){
                that.OnMonsterCommonAttack(data)
            })
            nano.on("OnLifeChanged", function(data){
                that.OnLifeChanged(data)
            })
            nano.on("OnManaChanged", function(data){
                that.OnManaChanged(data)
            })
            nano.on("OnEntityDie", function(data){
                that.OnEntityDie(data)
            })
            nano.on("OnReleaseSpell", function(data){
                that.OnReleaseSpell(data)
            })
            nano.on("OnBufferAdd", function(data){
                that.OnBufferAdd(data)
            })
            nano.on("OnBufferRemove", function(data){
                that.OnBufferRemove(data)
            })
            nano.on("OnTextMessage", function(data){
                that.OnTextMessage(data)
            })
        });
    }

    startGameLoop() {
        requestAnimationFrame((time) => {
            const deltaTime = time - this.lastTime;
            this.lastTime = time;

            // 更新游戏状态（例如怪物位置、动画等）
            this.scene.update(deltaTime, canvas.width, canvas.height);

            // 清除画布并绘制背景
            ctx.clearRect(0, 0, canvas.width, canvas.height);
            ctx.fillStyle = this.backgroundColor;
            ctx.fillRect(0, 0, canvas.width, canvas.height);

            // 绘制游戏场景（例如背景、地面等）
            this.scene.draw(ctx);

            // 继续游戏循环
            this.startGameLoop();
        });
    }

    // 处理鼠标点击事件
    onMouseUp(event) {
        this.scene.onMouseUp(event);
    }
    // 处理鼠标移动事件
    onMousemove(event) {
        this.scene.onMousemove(event);
    }

    OnEnterScene(data) {
        this.scene.changeScene(data);
        // 标记场景已加载
        this.sceneLoaded = true;
    }

    OnEnterView(data){
        console.log("OnEnterView:::", data)
        const entityType = data.entity_type;
        const entityData = data.data;
        const buffers = data.buffers;
        if (entityType === 0){
            //hero
            this.scene.addHero(entityData);
        }else if(entityType === 1){
            //monster
            this.scene.addMonster(entityData);
        }
    }

    OnExitView(data){
        console.log("OnExitView:::", data)
        const entityType = data.entity_type;
        const id = data.id;
        if (entityType === 0){
            //hero
            this.scene.removeHero(id)
        }else if(entityType === 1){
            //monster
            this.scene.removeMonster(id)
        }
    }

    OnHeroMoveTrace(data){
        console.log("OnHeroMoveTrace:::", data)
        this.scene.heroMove(data.id, data.trace_paths, data.step_time)
    }

    OnHeroMoveStopped(data){
        console.log("OnHeroMoveStopped:::", data)
        this.scene.heroMoveStop(data.id, data.pos_x, data.pos_y, data.pos_z)
    }

    OnMonsterMoveTrace(data){
        console.log("OnMonsterMoveTrace:::", data)
        this.scene.monsterMove(data.id, data.trace_paths, data.step_time)
    }

    OnMonsterMoveStopped(data){
        console.log("OnMonsterMoveStopped:::", data)
        this.scene.monsterMoveStop(data.id, data.pos_x, data.pos_y, data.pos_z)
    }

    OnMonsterCommonAttack(data){
        console.log("OnMonsterCommonAttack:::", data)
    }

    OnLifeChanged(data){
        console.log("OnLifeChanged:::", data)
        this.scene.lifeChanged(data)
    }

    OnManaChanged(data){
        console.log("OnManaChanged:::", data)
        this.scene.manaChanged(data)
    }

    OnEntityDie(data){
        console.log("OnEntityDie:::", data)
        this.scene.entityDie(data)
    }

    OnReleaseSpell(data){
        console.log("OnReleaseSpell:::", data)
        this.scene.OnReleaseSpell(data);
    }

    OnBufferAdd(data){
        console.log("OnBufferAdd:::", data)
    }

    OnBufferRemove(data){
        console.log("OnBufferRemove:::", data)
    }

    OnTextMessage(data){
        console.log("OnTextMessage:::", data)
        this.scene.heroTextMessage(data.hero_id, data.msg)
    }

    /**
     * {
     *   "resetSceneMonsters":1,
     *   "configs":[
     *    {
     *        "scene_id":1,
     *        "monster_id":1,
     *        "total": 10,
     *        "reborn": 60,
     *        "bornx": 50,
     *        "borny": 145,
     *        "bornz": 0,
     *        "a_range": 50
     *    },
     *    {
     *        "scene_id":1,
     *        "monster_id":2,
     *        "total": 100,
     *        "reborn": 60,
     *        "bornx": 120,
     *        "borny": 150,
     *        "bornz": 0,
     *        "a_range": 50
     *    }
     *   ]
     * }
     * @param msg
     */
    sendTextMsg(msg){
        if (msg.indexOf("add-") > -1){
            let cnt = parseInt(msg.replace("add-", ""))
            console.log("resetSceneMonsters::", cnt)
            let resetSceneMonsters = {
                "resetSceneMonsters":1,
                "configs":[
                    {
                        "scene_id":1,
                        "monster_id":1,
                        "total": parseInt(cnt/2),
                        "reborn": 60,
                        "bornx": 50,
                        "borny": 145,
                        "bornz": 0,
                        "a_range": 50
                    },
                    {
                        "scene_id":1,
                        "monster_id":2,
                        "total": parseInt(cnt/2),
                        "reborn": 60,
                        "bornx": 120,
                        "borny": 150,
                        "bornz": 0,
                        "a_range": 50
                    }
                ]
            }
            //动态怪物
           let configs = resetSceneMonsters//JSON.parse(resetSceneMonsters)
            nano.request("SceneManager.DynamicResetMonsters",configs, function(data){
                console.log("DynamicResetMonsters:", data)
            });
           return
        }
        if (msg.indexOf("scene-") > -1){
            let sceneId = parseInt(msg.replace("scene-", ""))
            console.log("请求切换场景:", sceneId)
            nano.request("Manager.HeroChangeScene",{
                scene_id: sceneId,
                uid : Global.selfHeroData.uid,
                hero_id: Global.selfHeroData.id,
            }, function(data){
                console.log("HeroChangeScene:", data)
            });
            return
        }
        nano.request("SceneManager.TextMessage", {
            hero_id: Global.selfHeroData.id,
            msg: msg,
        }, function(data){
            console.log("TextMessage:", data)
        });
        this.scene.heroTextMessage(Global.selfHeroData.id, msg)
    }
}

function testFastAstar(){
    // 使用示例
    var graph = new Graph([
        [1,1,1,1],
        [0,1,1,0],
        [0,0,1,1]
    ]);
    var start = graph.grid[0][0];
    var end = graph.grid[1][2];
    var path = astar.search(graph, start, end);
    console.log("paths::", path);
}
