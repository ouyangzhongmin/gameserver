
if (isPrdEnv()) {
    //改成自己的地址
    Global.WSAddr = "47.99.**.**"
    HttpBaseUrl = "http://127.0.0.1:12307"
}

function isPrdEnv(){
    // let test = true
    // if(test){
    //     return true
    // }
    var path = window.location.host;
    if (path.includes('47.99.**.**') || path.includes('**.**.com')) {
        return true
    }
    return false
}
