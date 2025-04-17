//----------------------------------------------------------------------------//
// Imports                                                                    //
//----------------------------------------------------------------------------//

import { Vector3 } from './vector3';
import { Vector4 } from './vector4';
import { Matrix } from './matrix';

//----------------------------------------------------------------------------//
// Types                                                                      //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////

export interface Vector2Object {
	x: number;
	y: number;
}

////////////////////////////////////////////////////////////////////////////////

export type Vector2Params = Parameters<typeof Vector2.standardize>;
export type Vector2Return = ReturnType<typeof Vector2.standardize>;

//----------------------------------------------------------------------------//
// Vector2                                                                    //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////

export class Vector2 {
	//----------------------------------------------------------------------------//
	// Constants                                                                  //
	//----------------------------------------------------------------------------//

	////////////////////////////////////////////////////////////////////////////////

	public static readonly Zero  = new Vector2(0, 0);
	public static readonly UnitX = new Vector2(1, 0);
	public static readonly UnitY = new Vector2(0, 1);

	//----------------------------------------------------------------------------//
	// Fields                                                                     //
	//----------------------------------------------------------------------------//

	////////////////////////////////////////////////////////////////////////////////

	private readonly _x: number;
	private readonly _y: number;

	//----------------------------------------------------------------------------//
	// Constructor                                                                //
	//----------------------------------------------------------------------------//

	////////////////////////////////////////////////////////////////////////////////

	public constructor(...args: Vector2Params) {

		const v = Vector2.standardize(...args);

		this._x = v.x;
		this._y = v.y;
	}

	//----------------------------------------------------------------------------//
	// Methods                                                                    //
	//----------------------------------------------------------------------------//

	////////////////////////////////////////////////////////////////////////////////

	public isZero(): boolean {
		return this._x === 0 && this._y === 0;
	}

	////////////////////////////////////////////////////////////////////////////////

	public setX(x: number): Vector2 {
		return new Vector2(x, this._y);
	}

	////////////////////////////////////////////////////////////////////////////////

	public setY(y: number): Vector2 {
		return new Vector2(this._x, y);
	}

	////////////////////////////////////////////////////////////////////////////////

	public toArray(): [number, number] {
		return [
			this._x,
			this._y,
		];
	}

	////////////////////////////////////////////////////////////////////////////////

	public toObject(): Vector2Object {
		return {
			x: this._x,
			y: this._y,
		};
	}

	////////////////////////////////////////////////////////////////////////////////

	public toString(): string {
		return `[${this._x.toFixed(2)}, ${this._y.toFixed(2)}]`;
	}

	////////////////////////////////////////////////////////////////////////////////

	public normalize(): Vector2 {

		const magnitude = Math.sqrt(this._x ** 2 + this._y ** 2);

		// The default case
		if (magnitude === 0) {
			return Vector2.Zero;
		}

		return new Vector2(
			this._x / magnitude,
			this._y / magnitude,
		);
	}

	////////////////////////////////////////////////////////////////////////////////

	public distance(value: Vector2): number {

		const dx = this._x - value._x;
		const dy = this._y - value._y;

		return Math.sqrt(dx * dx + dy * dy);
	}

	////////////////////////////////////////////////////////////////////////////////

	public distanceSq(value: Vector2): number {

		const dx = this._x - value._x;
		const dy = this._y - value._y;

		return dx * dx + dy * dy;
	}

	////////////////////////////////////////////////////////////////////////////////

	public dot(value: Vector2): number {

		return this._x * value._x + this._y * value._y;
	}

	////////////////////////////////////////////////////////////////////////////////

	public reflect(normal: Vector2): Vector2 {

		const dot = 2 * (this._x * normal._x + this._y * normal._y);

		return new Vector2(
			this._x - dot * normal._x,
			this._y - dot * normal._y,
		);
	}

	////////////////////////////////////////////////////////////////////////////////

	public length(): number {

		return Math.sqrt(this._x ** 2 + this._y ** 2);
	}

	////////////////////////////////////////////////////////////////////////////////

	public lengthSq(): number {

		return this._x ** 2 + this._y ** 2;
	}

	//----------------------------------------------------------------------------//
	// Static                                                                     //
	//----------------------------------------------------------------------------//

	////////////////////////////////////////////////////////////////////////////////

	public static standardize(
		ax?: number | Vector2 | Vector2Object | [number, number],
		ay?: number,
	): Vector2Object {

		if (ax instanceof Vector2) {
			return { x: ax._x, y: ax._y };
		}

		if (Array.isArray(ax)) {
			const [x, y] = ax;
			return {
				x: typeof x === 'number' ? x : 0,
				y: typeof y === 'number' ? y : 0,
			};
		}

		if (typeof ax === 'object' && ax !== null) {
			return {
				x: typeof ax.x === 'number' ? ax.x : 0,
				y: typeof ax.y === 'number' ? ax.y : 0,
			};
		}

		if (typeof ax === 'number') {
			if (typeof ay === 'number') {
				return { x: ax, y: ay };
			}

			return { x: ax, y: ax };
		}

		return { x: 0, y: 0 };
	}

	////////////////////////////////////////////////////////////////////////////////

	public static transformVector2(matrix: Matrix, value: Vector2): Vector2 {

		return new Vector2(
			matrix.m11 * value.x + matrix.m21 * value.y + matrix.m41,
			matrix.m12 * value.x + matrix.m22 * value.y + matrix.m42,
		);
	}

	////////////////////////////////////////////////////////////////////////////////

	public static transformVector3(matrix: Matrix, value: Vector3): Vector2 {

		return new Vector2(
			matrix.m11 * value.x + matrix.m21 * value.y + matrix.m31 * value.z + matrix.m41,
			matrix.m12 * value.x + matrix.m22 * value.y + matrix.m32 * value.z + matrix.m42,
		);
	}

	////////////////////////////////////////////////////////////////////////////////

	public static transformVector4(matrix: Matrix, value: Vector4): Vector2 {

		return new Vector2(
			matrix.m11 * value.x + matrix.m21 * value.y + matrix.m31 * value.z + matrix.m41 * value.w,
			matrix.m12 * value.x + matrix.m22 * value.y + matrix.m32 * value.z + matrix.m42 * value.w,
		);
	}

	//----------------------------------------------------------------------------//
	// Accessors                                                                  //
	//----------------------------------------------------------------------------//

	////////////////////////////////////////////////////////////////////////////////

	public get x(): number {
		return this._x;
	}

	////////////////////////////////////////////////////////////////////////////////

	public get y(): number {
		return this._y;
	}

	//----------------------------------------------------------------------------//
	// Operators                                                                  //
	//----------------------------------------------------------------------------//

	////////////////////////////////////////////////////////////////////////////////

	public add(...args: Vector2Params): Vector2 {

		const v = Vector2.standardize(...args);

		return new Vector2(
			this._x + v.x,
			this._y + v.y,
		);
	}

	////////////////////////////////////////////////////////////////////////////////

	public sub(...args: Vector2Params): Vector2 {

		const v = Vector2.standardize(...args);

		return new Vector2(
			this._x - v.x,
			this._y - v.y,
		);
	}

	////////////////////////////////////////////////////////////////////////////////

	public mul(...args: Vector2Params): Vector2 {

		const v = Vector2.standardize(...args);

		return new Vector2(
			this._x * v.x,
			this._y * v.y,
		);
	}

	////////////////////////////////////////////////////////////////////////////////

	public div(...args: Vector2Params): Vector2 {

		const v = Vector2.standardize(...args);

		return new Vector2(
			this._x / v.x,
			this._y / v.y,
		);
	}

	////////////////////////////////////////////////////////////////////////////////

	public neg(): Vector2 {

		return new Vector2(
			-this._x,
			-this._y,
		);
	}

	////////////////////////////////////////////////////////////////////////////////

	public compare(...args: Vector2Params): number {

		const v = Vector2.standardize(...args);

		if (this._x !== v.x) {
			return this._x < v.x ? -1 : 1;
		}

		if (this._y !== v.y) {
			return this._y < v.y ? -1 : 1;
		}

		return 0;
	}

	////////////////////////////////////////////////////////////////////////////////

	public lt(...args: Vector2Params): boolean {
		return this.compare(...args) < 0;
	}

	////////////////////////////////////////////////////////////////////////////////

	public gt(...args: Vector2Params): boolean {
		return this.compare(...args) > 0;
	}

	////////////////////////////////////////////////////////////////////////////////

	public le(...args: Vector2Params): boolean {
		return this.compare(...args) <= 0;
	}

	////////////////////////////////////////////////////////////////////////////////

	public ge(...args: Vector2Params): boolean {
		return this.compare(...args) >= 0;
	}

	////////////////////////////////////////////////////////////////////////////////

	public eq(...args: Vector2Params): boolean {

		const v = Vector2.standardize(...args);
		return this._x === v.x && this._y === v.y;
	}

	////////////////////////////////////////////////////////////////////////////////

	public ne(...args: Vector2Params): boolean {

		const v = Vector2.standardize(...args);
		return this._x !== v.x || this._y !== v.y;
	}
}
