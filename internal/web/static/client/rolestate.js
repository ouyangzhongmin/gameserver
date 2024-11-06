

class RoleState {
	constructor(){
		//状态机制
		this._state = State.STAND ;
		this._stateChaned = false;

		this.fightable = true;
		this.movable = true;

		this._unMovableTicked = 0;
		this._unFightableTicked = 0;

		this.speedX = 0;
		this.speedY = 0;
		this.deadTimeStamp = 0
	}

	setMovable(value){
		if (!value)
			this._unMovableTicked++;
		else
			this._unMovableTicked--;
		if (value)
		{
			if (this._unMovableTicked <= 0)
			{
				this.movable = value;
				this._unMovableTicked = 0;
			}
		}
		else
		{
			this.movable = value;
			stand();
		}
	}

	setFightable(value)
	{
		if (!value)
			this._unFightableTicked++;
		else
			this._unFightableTicked--;
		if (value)
		{
			if (this._unFightableTicked <= 0)
			{
				this.fightable = value;
				this._unFightableTicked = 0;
			}
		}
		else{
			this.fightable = value;
		}
	}

	getState()
	{
		return this._state;
	}

	setState(value)
	{
		this._state = value;
		this._stateChaned = true;
	}

	isStateChanged()
	{
		return this._stateChaned;
	}

	resetStateChanged()
	{
		this._stateChaned = false;
	}

	beenAttack() {
		if(!this.canBeenAttack())return;
		this.resetSpeed();
		this.setState(State.BEENATTACK);
	}

	walk()
	{
		if(!this.canWalk())return;
		this.setState(State.WALK);
	}

	attack()
	{
		if(!this.canAttack())return false;
		this.resetSpeed();
		this.setState(State.ATTACK);
		return true;
	}

	sit()
	{
		if(!this.canSit())return;
		this.resetSpeed();
		this.setState(State.SIT);
	}

	beatBack()
	{
		if(!this.canBeatBack())return;
		this.resetSpeed();
		this.setState(State.BEATBACK);
	}

	special()
	{
		this.setState(State.SPECIAL);
	}

	hit()
	{
		this.resetSpeed();
		this.setState(State.HIT);
	}

	stall()
	{
		if(!this.canStall())return;
		this.resetSpeed();
		this.setState(State.STALL);
	}

	castspell()
	{
		if(!this.canCastspell())return;
		this.resetSpeed();
		this.setState(State.CASTSPELL);
	}

	resetSpeed(){
		this.speedX = 0;
		this.speedY = 0;
	}

	stand()
	{
		if(!this.canStand())return;
		this.setState(State.STAND);
		this.resetSpeed()
	}


	die()
	{
		this.resetSpeed();
		this.setState(State.DEAD);
		this.deadTimeStamp = Date.now()
	}

	jump()
	{
		if(!this.canJump())
			return;
		this.resetSpeed();
		this.setState(State.JUMP);
	}

	canWalk()
	{
		return this.movable && !this.isAttack() && !this.isDead() &&
			!this.isBeenAttack() && !this.isCastSpell() && !this.isJump() &&
			!this.isBlink() && !this.isRush() && !this.isHit();
	}

	canAttack()
	{
		return  this.fightable() && !this.isCastSpell() &&!this.isHit() && !this.isBeenAttack() &&
			!this.isDead() && !this.isJump() && !this.isBeatBack() && !this.isBlink() && !this.isRush();
	}

	canCastspell()
	{
		return  this.fightable() &&!this.isHit() && !this.isBeenAttack() &&
			!this.isDead() && !this.isJump() && !this.isBeatBack() && !this.isBlink() && !this.isRush();
	}

	canJump()
	{
		return this.canWalk();
	}

	canBeenAttack()
	{
		return  !this.isDead();
	}

	canHit()
	{
		return !this.isDead();
	}

	canStand()
	{
		return  !this.isDead();
	}

	canSit()
	{
		return !this.isAttack() && !this.isHit() && !this.isDead() && !this.isCastSpell() &&
			!this.isBeatBack() && !this.isBeenAttack() && !this.isStall() && !this.isJump()&& !this.isBlink() && !this.isRush();
	}

	canBeatBack()
	{
		return !this.isDead() && !this.isStall();
	}

	canStall()
	{
		return !this.isDead() && !this.isAttack() && !this.isCastSpell() && !this.isBeatBack() && !this.isJump();
	}

	canRush()
	{
		return this.canAttack();
	}

	canBlink()
	{
		return this.canAttack() && this.movable;
	}

	//被击晕
	isStun()
	{
		return !this.movable && !this.fightable;
	}

	isRush()
	{
		return this._state === State.RUSH;
	}

	isBlink()
	{
		return this._state === State.BLINK;
	}

	/**
	 * 正在受击
	 * @return
	 */
	isHit()
	{
		return this._state === State.HIT;
	}

	isWalk()
	{
		return this._state === State.WALK;
	}

	isStand()
	{
		return this._state === State.STAND;
	}

	isAttack()
	{
		return this._state === State.ATTACK  ;
	}

	isCastSpell()
	{
		return this._state === State.CASTSPELL;
	}

	isDead()
	{
		return this._state === State.DEAD;
	}

	isBeenAttack()
	{
		return this._state === State.BEENATTACK;
	}

	isSit()
	{
		return this._state === State.SIT;
	}

	isBeatBack()
	{
		return this._state === State.BEATBACK;
	}

	isStall()
	{
		return this._state === State.STALL;
	}

	isJump()
	{
		return this._state === State.JUMP;
	}
}