

var HttpBaseUrl = "http://127.0.0.1:12307"
if (isPrdEnv()) {
    HttpBaseUrl = "http://47.99.180.185:12307"
}
const GRID_WIDTH = 16;
const GRID_HEIGHT = 16;

const Direction = {
    NONE :  -1,
    S  :  0,
    W  :  1,
    E  :  2,
    N  :  3,
    SW :  4,
    SE :  5,
    NW :  6,
    NE :  7,
}

//状态
const State = {
    STAND: 0,
    WALK: 1,
    //连招中每一招对应一种状态
    ATTACK: 2,
    CASTSPELL: 3,
    DEAD: 4,
    BEENATTACK: 5, //被攻击
    SIT: 6, //打坐
    BEATBACK: 7, //击退
    STALL: 8, //摆摊
    JUMP: 9, //跳
    RUSH: 10, //冲锋
    BLINK: 11, //闪烁
    SPECIAL: 12,
    HIT: 13, //受击 与BEENATTACK有歧义
}

const Global = {
    block : new Block(),
    nano : window.nano,
    http: new HttpClient(HttpBaseUrl),
    userInfo: null,
    selfHeroData: null,
}

function isPrdEnv(){
    let a = true
    if(a){
        return true
    }
    var path = window.location.host;
    if (path.includes('47.99.180.185') || path.includes('jsmx-test.zhiyun-tech.com')) {
       return true
    }
    return false
}

function getPixelXByGrid(gridX) {
    return gridX * GRID_WIDTH + GRID_WIDTH/2;
}

function getPixelYByGrid(gridY ) {
    return gridY * GRID_HEIGHT + GRID_HEIGHT/2;
}

function getGridXByPixel(pixelX ) {
    return Math.floor(pixelX / GRID_WIDTH);
}

function getGridYByPixel(pixelY ) {
    return Math.floor(pixelY / GRID_HEIGHT);
}

function getStepTime(speed ) {
    return Math.floor(GRID_WIDTH / speed);
}

function getSpeed(stepTime ) {
    return Number((GRID_WIDTH / stepTime).toFixed(3));
}

/**
 * 根据角度获取方向
 * @param rotation
 * @return
 */
function getDirection(rotation ) {
    if (rotation <= 22 || rotation >= 338) {
        return Direction.E;
    }

    if (rotation >= 23 && rotation <= 67) {
        return Direction.SE;
    }

    if (rotation >= 68 && rotation <= 112) {
        return Direction.S;
    }

    if (rotation >= 113 && rotation <= 157) {
        return Direction.SW;
    }

    if (rotation >= 158 && rotation <= 202) {
        return Direction.W;
    }

    if (rotation >= 203 && rotation <= 247) {
        return Direction.NW;
    }

    if (rotation >= 248 && rotation <= 292) {
        return Direction.N;
    }

    if (rotation >= 293 && rotation <= 337) {
        return Direction.NE;
    }

    return -1;
}

function getAngelByDir(dir)
{
    if(dir === Direction.S)
    {
        return 90;
    }
    else if(dir === Direction.W)
    {
        return 180;
    }
    else if(dir === Direction.E)
    {
        return 0;
    }
    else if(dir === Direction.N)
    {
        return 270;
    }
    else if(dir === Direction.SW)
    {
        return 135;
    }
    else if(dir === Direction.SE)
    {
        return 45;
    }
    else if(dir === Direction.NW)
    {
        return 225;
    }
    else if(dir === Direction.NE)
    {
        return 315;
    }
    return 0;
}

/**
 * 取得一个方向的反向
 * @param dir
 * @return
 *
 */
function getReverseDirByDir(dir ) 
{
    switch(dir)
    {
        case Direction.E:
            return Direction.W;
        case Direction.SE:
            return Direction.NW;
        case Direction.S:
            return Direction.N;
        case Direction.NE:
            return Direction.SW;
        case Direction.N:
            return Direction.S;
        case Direction.W:
            return Direction.E;
        case Direction.NW:
            return Direction.SE;
        case Direction.SW:
            return Direction.NE;
    }
    return -1;
}

/**
 * 计算亮点之间的角度
 * @param Ax	A点x坐标
 * @param Ay	A点y坐标
 * @param Bx	B点x坐标
 * @param By	B点y坐标
 */
function getAngle(Ax , Ay , Bx , By ) {
    const tempXDistance  = Bx - Ax;
    const tempYDistance  = By - Ay;

    let rotation  = Math.round(Math.atan2(tempYDistance, tempXDistance) * 57.33);
    rotation = (rotation + 360)%360;

    return rotation;
}

/**
 * 计算两点之间 的弧度
 * @param Ax
 * @param Ay
 * @param Bx
 * @param By
 * @return
 *
 */
function getRotate(Ax , Ay , Bx , By ) {
    const tempXDistance  = Bx - Ax;
    const tempYDistance  = By - Ay;
    return Math.atan2(tempYDistance, tempXDistance);
}

/**
 * 根据角度得出弧度
 * @param angle
 * @return
 *
 */
function getRotateByAngle(angle ) {
    return angle*Math.PI/180;
}

/**
 * 根据弧度得到角度
 * @param rotate
 * @return
 */
function getAngleByRotate(rotate)
{
    return Math.round(rotate * 180 / Math.PI);
}

/**
 * 计算两点距离
 * @param Ax	A点x坐标
 * @param Ay	A点y坐标
 * @param Bx	B点x坐标
 * @param By	B点y坐标
 */
function getDistance(Ax , Ay , Bx , By ) {
    return (Math.sqrt(((Ax - Bx) * (Ax - Bx)) + ((Ay - By) * (Ay - By))));
}

function angleToRadian(angle) {
    return angle*(Math.PI/180);
}

/**
 * 弧度转换为角度
 * @param radian
 * @return
 *
 */
function radianToAngle(radian) {
    return fixAngle(radian*(180/Math.PI));
}

/**
 * 修正角度在360度以内
 */
function fixAngle(angle) {
    return (angle + 360 ) % 360;
}

/**
 * 以角度为单位计算三角函数值
 * @param angle
 * @return
 *
 */
function sinD(angle) {
    return Math.sin(angleToRadian(angle));
}

function cosD(angle) {
    return Math.cos(angleToRadian(angle));
}

function tanD(angle) {
    return Math.tan(angleToRadian(angle));
}

/**
 * 返回的值是以角度为单位
 * @param radian
 * @return
 *
 */
function asinD(radian) {
    return radianToAngle(Math.acos(radian));
}
function acosD(radian) {
    return radianToAngle(Math.acos(radian));
}
function atanD(radian) {
    return radianToAngle(Math.acos(radian));
}
function atan2D(y, x) {
    return radianToAngle(Math.atan2(y, x));
}

function decryptByteArray(bytes )
{
    const key = "fucku";
    let flag = 0;

    let newBytes =  new Uint8Array(bytes.length);
    const len = bytes.length;
    for(let i = 0 ;i < len ;i++ ,flag++)
    {
        if(flag >= key.length)
        {
            flag = 0;
        }
        newBytes[i] = bytes[i] - key.charCodeAt(flag);
    }
    return newBytes;
}

function generateRandomPoint(xRange, yRange) {
    return {
        x: generateRandomInt(0, xRange),
        y: generateRandomInt(0, yRange)
    };
}

function generateRandomInt(min, max) {
    return Math.floor(Math.random() * (max - min + 1)) + min;
}

// 这是一个简单的UUID v4生成方法
function generateUniqueId() {
    const uniqueId = 'xxxx-xxxx-4xxx-yxxx-xxxx-yyyy'.replace(/[xy]/g, function(c) {
        const r = Math.random() * 16 | 0, v = c === 'x' ? r : (r & 0x3 | 0x8);
        return v.toString(16);
    });
    return uniqueId
}

//16进制颜色合并alfpha转换为rgba
// console.log(hexToRGBA("#FF5733", 0.5)); // 输出: rgba(255, 87, 51, 0.5)
function hexToRGBA(hex, alpha = 1) {
    // 确保传入的是一个有效的十六进制颜色
    if (!/#[0-9A-F]{6}$/i.test(hex)) {
        throw new Error('Invalid HEX color.');
    }

    // 从十六进制字符串中提取颜色分量
    const r = parseInt(hex.slice(1, 3), 16);
    const g = parseInt(hex.slice(3, 5), 16);
    const b = parseInt(hex.slice(5, 7), 16);

    // 返回RGBA字符串
    return `rgba(${r}, ${g}, ${b}, ${alpha})`;
}

// 颜色值名称转换为rgba
// console.log(colorNameToRGBA("red", 0.5)); // 输出: rgba(255, 0, 0, 0.5)
function colorNameToRGBA(colorName, alpha = 1) {
    const colors = {
        red: [255, 0, 0],
        green: [0, 255, 0],
        blue: [0, 0, 255],
        white: [255, 255, 255],
        black: [0, 0, 0],
        // 可以添加更多颜色...
    };

    const rgb = colors[colorName.toLowerCase()];
    if (!rgb) {
        throw new Error('Unsupported color name.');
    }

    return `rgba(${rgb[0]}, ${rgb[1]}, ${rgb[2]}, ${alpha})`;
}





