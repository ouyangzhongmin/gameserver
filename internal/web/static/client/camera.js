
class Camera{
	constructor(w, h) {
		this.target = null;
		this.x = 0;
		this.y = 0;
		this.maxWidth = 0;
		this.maxHeight = 0;
        this.resize(w,h);

		this.cameraCenterX = 0;
		this.cameraCenterY = 0;

		this._shake = new Shake()
	}

	resize(w,h){
		this.width = w;
		this.height = h;
		if (this.maxWidth  < w ){
			this.maxWidth = w
		}
		if (this.maxHeight < h){
			this.maxHeight = h
		}

		this.halfViewWidth  = this.width *0.5;
		this.halfViewHeight = this.height *0.5;
	}
    
	 //����Ķ���
	 focusAt(target){
		 this.target = target
	 }
      
	 //�л���ͼ��ʱ���������
     setMaxViewWH(mw, mh){
		this.maxWidth = mw;
		this.maxHeight = mh;
	 }

     //������
	 shake(t  , r ) {
		this._shake.shake(t  , r )
	 }

     update(deltaTime){
		 if (this.target){
			 if (isNaN(this.target.x) || isNaN(this.target.y)){
				 console.error("camera.target.x or y error::", this.target.name, this.target.x, this.target.y)
			 }
			 this.cameraCenterX = +this.target.x;
			 this.cameraCenterY = +this.target.y;

			 this.cameraCenterX = Math.max(this.cameraCenterX, 0 + this.halfViewWidth);
			 this.cameraCenterX = Math.min(this.cameraCenterX, this.maxWidth - this.halfViewWidth);
			 this.cameraCenterY = Math.max(this.cameraCenterY, 0 + this.halfViewHeight);
			 this.cameraCenterY = Math.min(this.cameraCenterY, this.maxHeight - this.halfViewHeight);

			 this._shake.updateShake();
			 this.x = this.cameraCenterX - this.halfViewWidth + this._shake.shakeOffsetX;
			 this.y = this.cameraCenterY - this.halfViewHeight + this._shake.shakeOffsetY;
			 //console.log("camera.update", this.cameraCenterX, this.halfViewWidth,)
		 }
	 }
}

class Shake{
   constructor() {
		this.shakeCouter = 0;
		this.shakeAmount = 10;
		this.radius = 3;
		this.needShake = false;
		this.shakeOffsetX = 0;
		this.shakeOffsetY = 0;
   }

   shake(t  , r ) {
		this.needShake = true;
		this.shakeCouter = 0;
		this.shakeAmount = t;
		this.radius = r;
		this.shakeOffsetX = 0;
		this.shakeOffsetY = 0;
	}

	updateShake(){
		if (!this.needShake)
			return;
		this.shakeCouter = this.shakeCouter + 1;
		if (this.shakeCouter < this.shakeAmount){
			if (this.shakeCouter % 2 == 0){
				this.shakeOffsetX =  this.radius;
				this.shakeOffsetY =  this.radius;
			}
			else if (this.shakeCouter % 2 == 1)
			{
				this.shakeOffsetX = - this.radius;
				this.shakeOffsetY = - this.radius;
			}
				this.radius = _radius * 0.95;
		}
		else{
			this.shakeOffsetX =  0;
			this.shakeOffsetY =  0;
			this.needShake = false;
		}
	}

}