//----------------------------------------------------------------------------//
// Imports                                                                    //
//----------------------------------------------------------------------------//

import { Vector2 } from './vector2';
import { Vector3 } from './vector3';
import { Matrix } from './matrix';

//----------------------------------------------------------------------------//
// Types                                                                      //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////

export interface Vector4Object {
	x: number;
	y: number;
	z: number;
	w: number;
}

////////////////////////////////////////////////////////////////////////////////

export type Vector4Params = Parameters<typeof Vector4.standardize>;
export type Vector4Return = ReturnType<typeof Vector4.standardize>;

//----------------------------------------------------------------------------//
// Vector4                                                                    //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////

export class Vector4 {
	//----------------------------------------------------------------------------//
	// Constants                                                                  //
	//----------------------------------------------------------------------------//

	////////////////////////////////////////////////////////////////////////////////

	public static readonly Zero     = new Vector4(0, 0, 0, 0);
	public static readonly UnitX    = new Vector4(1, 0, 0, 0);
	public static readonly UnitY    = new Vector4(0, 1, 0, 0);
	public static readonly UnitZ    = new Vector4(0, 0, 1, 0);
	public static readonly UnitW    = new Vector4(0, 0, 0, 1);
	public static readonly Identity = new Vector4(0, 0, 0, 1);

	//----------------------------------------------------------------------------//
	// Fields                                                                     //
	//----------------------------------------------------------------------------//

	////////////////////////////////////////////////////////////////////////////////

	private readonly _x: number;
	private readonly _y: number;
	private readonly _z: number;
	private readonly _w: number;

	//----------------------------------------------------------------------------//
	// Constructor                                                                //
	//----------------------------------------------------------------------------//

	////////////////////////////////////////////////////////////////////////////////

	public constructor(...args: Vector4Params) {

		const v = Vector4.standardize(...args);

		this._x = v.x;
		this._y = v.y;
		this._z = v.z;
		this._w = v.w;
	}

	//----------------------------------------------------------------------------//
	// Methods                                                                    //
	//----------------------------------------------------------------------------//

	////////////////////////////////////////////////////////////////////////////////

	public isZero(): boolean {
		return this._x === 0 && this._y === 0 && this._z === 0 && this._w === 0;
	}

	////////////////////////////////////////////////////////////////////////////////

	public setX(x: number): Vector4 {
		return new Vector4(x, this._y, this._z, this._w);
	}

	////////////////////////////////////////////////////////////////////////////////

	public setY(y: number): Vector4 {
		return new Vector4(this._x, y, this._z, this._w);
	}

	////////////////////////////////////////////////////////////////////////////////

	public setZ(z: number): Vector4 {
		return new Vector4(this._x, this._y, z, this._w);
	}

	////////////////////////////////////////////////////////////////////////////////

	public setW(w: number): Vector4 {
		return new Vector4(this._x, this._y, this._z, w);
	}

	////////////////////////////////////////////////////////////////////////////////

	public toArray(): [number, number, number, number] {
		return [
			this._x,
			this._y,
			this._z,
			this._w,
		];
	}

	////////////////////////////////////////////////////////////////////////////////

	public toObject(): Vector4Object {
		return {
			x: this._x,
			y: this._y,
			z: this._z,
			w: this._w,
		};
	}

	////////////////////////////////////////////////////////////////////////////////

	public toString(): string {
		return `[${this._x.toFixed(2)}, ${this._y.toFixed(2)}, ${this._z.toFixed(2)}, ${this._w.toFixed(2)}]`;
	}

	////////////////////////////////////////////////////////////////////////////////

	public normalize(): Vector4 {

		const magnitude = Math.sqrt(this._x ** 2 + this._y ** 2 + this._z ** 2 + this._w ** 2);

		// The default case
		if (magnitude === 0) {
			return Vector4.Zero;
		}

		return new Vector4(
			this._x / magnitude,
			this._y / magnitude,
			this._z / magnitude,
			this._w / magnitude,
		);
	}

	////////////////////////////////////////////////////////////////////////////////

	public distance(value: Vector4): number {

		const dx = this._x - value._x;
		const dy = this._y - value._y;
		const dz = this._z - value._z;
		const dw = this._w - value._w;

		return Math.sqrt(dx * dx + dy * dy + dz * dz + dw * dw);
	}

	////////////////////////////////////////////////////////////////////////////////

	public distanceSq(value: Vector4): number {

		const dx = this._x - value._x;
		const dy = this._y - value._y;
		const dz = this._z - value._z;
		const dw = this._w - value._w;

		return dx * dx + dy * dy + dz * dz + dw * dw;
	}

	////////////////////////////////////////////////////////////////////////////////

	public dot(value: Vector4): number {

		return this._x * value._x + this._y * value._y + this._z * value._z + this._w * value._w;
	}

	////////////////////////////////////////////////////////////////////////////////

	public length(): number {

		return Math.sqrt(this._x ** 2 + this._y ** 2 + this._z ** 2 + this._w ** 2);
	}

	////////////////////////////////////////////////////////////////////////////////

	public lengthSq(): number {

		return this._x ** 2 + this._y ** 2 + this._z ** 2 + this._w ** 2;
	}

	//----------------------------------------------------------------------------//
	// Static                                                                     //
	//----------------------------------------------------------------------------//

	////////////////////////////////////////////////////////////////////////////////

	public static standardize(
		ax?: number | Vector4 | Vector4Object | [number, number, number, number],
		ay?: number,
		az?: number,
		aw?: number,
	): Vector4Object {

		if (ax instanceof Vector4) {
			return { x: ax._x, y: ax._y, z: ax._z, w: ax._w };
		}

		if (Array.isArray(ax)) {
			const [x, y, z, w] = ax;
			return {
				x: typeof x === 'number' ? x : 0,
				y: typeof y === 'number' ? y : 0,
				z: typeof z === 'number' ? z : 0,
				w: typeof w === 'number' ? w : 0,
			};
		}

		if (typeof ax === 'object' && ax !== null) {
			return {
				x: typeof ax.x === 'number' ? ax.x : 0,
				y: typeof ax.y === 'number' ? ax.y : 0,
				z: typeof ax.z === 'number' ? ax.z : 0,
				w: typeof ax.w === 'number' ? ax.w : 0,
			};
		}

		if (typeof ax === 'number') {
			if (typeof ay === 'number' && typeof az === 'number' && typeof aw === 'number') {
				return { x: ax, y: ay, z: az, w: aw };
			}

			return { x: ax, y: ax, z: ax, w: ax };
		}

		return { x: 0, y: 0, z: 0, w: 0 };
	}

	////////////////////////////////////////////////////////////////////////////////

	public static transformVector2(matrix: Matrix, value: Vector2): Vector4 {

		return new Vector4(
			matrix.m11 * value.x + matrix.m21 * value.y + matrix.m41,
			matrix.m12 * value.x + matrix.m22 * value.y + matrix.m42,
			matrix.m13 * value.x + matrix.m23 * value.y + matrix.m43,
			matrix.m14 * value.x + matrix.m24 * value.y + matrix.m44,
		);
	}

	////////////////////////////////////////////////////////////////////////////////

	public static transformVector3(matrix: Matrix, value: Vector3): Vector4 {

		return new Vector4(
			matrix.m11 * value.x + matrix.m21 * value.y + matrix.m31 * value.z + matrix.m41,
			matrix.m12 * value.x + matrix.m22 * value.y + matrix.m32 * value.z + matrix.m42,
			matrix.m13 * value.x + matrix.m23 * value.y + matrix.m33 * value.z + matrix.m43,
			matrix.m14 * value.x + matrix.m24 * value.y + matrix.m34 * value.z + matrix.m44,
		);
	}

	////////////////////////////////////////////////////////////////////////////////

	public static transformVector4(matrix: Matrix, value: Vector4): Vector4 {

		return new Vector4(
			matrix.m11 * value.x + matrix.m21 * value.y + matrix.m31 * value.z + matrix.m41 * value.w,
			matrix.m12 * value.x + matrix.m22 * value.y + matrix.m32 * value.z + matrix.m42 * value.w,
			matrix.m13 * value.x + matrix.m23 * value.y + matrix.m33 * value.z + matrix.m43 * value.w,
			matrix.m14 * value.x + matrix.m24 * value.y + matrix.m34 * value.z + matrix.m44 * value.w,
		);
	}

	////////////////////////////////////////////////////////////////////////////////

	public static fromPacked(packed: bigint): Vector4 {

		const a = 1 / 2097152.0;
		const b = 1 / 1048576.0;

		const x = Number((packed <<  0n) >> 42n) * a;
		const y = Number((packed << 22n) >> 43n) * b;
		const z = Number((packed << 43n) >> 43n) * b;

		const wSquared = x * x + y * y + z * z;

		let w = 0.0;
		if (Math.abs(wSquared - 1) >= b) {
			w = Math.sqrt(1 - wSquared);
		}

		return new Vector4(x, y, z, w);
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

	////////////////////////////////////////////////////////////////////////////////

	public get z(): number {
		return this._z;
	}

	////////////////////////////////////////////////////////////////////////////////

	public get w(): number {
		return this._w;
	}

	//----------------------------------------------------------------------------//
	// Operators                                                                  //
	//----------------------------------------------------------------------------//

	////////////////////////////////////////////////////////////////////////////////

	public add(...args: Vector4Params): Vector4 {

		const v = Vector4.standardize(...args);

		return new Vector4(
			this._x + v.x,
			this._y + v.y,
			this._z + v.z,
			this._w + v.w,
		);
	}

	////////////////////////////////////////////////////////////////////////////////

	public sub(...args: Vector4Params): Vector4 {

		const v = Vector4.standardize(...args);

		return new Vector4(
			this._x - v.x,
			this._y - v.y,
			this._z - v.z,
			this._w - v.w,
		);
	}

	////////////////////////////////////////////////////////////////////////////////

	public mul(...args: Vector4Params): Vector4 {

		const v = Vector4.standardize(...args);

		return new Vector4(
			this._x * v.x,
			this._y * v.y,
			this._z * v.z,
			this._w * v.w,
		);
	}

	////////////////////////////////////////////////////////////////////////////////

	public div(...args: Vector4Params): Vector4 {

		const v = Vector4.standardize(...args);

		return new Vector4(
			this._x / v.x,
			this._y / v.y,
			this._z / v.z,
			this._w / v.w,
		);
	}

	////////////////////////////////////////////////////////////////////////////////

	public neg(): Vector4 {

		return new Vector4(
			-this._x,
			-this._y,
			-this._z,
			-this._w,
		);
	}

	////////////////////////////////////////////////////////////////////////////////

	public compare(...args: Vector4Params): number {

		const v = Vector4.standardize(...args);

		if (this._x !== v.x) {
			return this._x < v.x ? -1 : 1;
		}

		if (this._y !== v.y) {
			return this._y < v.y ? -1 : 1;
		}

		if (this._z !== v.z) {
			return this._z < v.z ? -1 : 1;
		}

		if (this._w !== v.w) {
			return this._w < v.w ? -1 : 1;
		}

		return 0;
	}

	////////////////////////////////////////////////////////////////////////////////

	public lt(...args: Vector4Params): boolean {
		return this.compare(...args) < 0;
	}

	////////////////////////////////////////////////////////////////////////////////

	public gt(...args: Vector4Params): boolean {
		return this.compare(...args) > 0;
	}

	////////////////////////////////////////////////////////////////////////////////

	public le(...args: Vector4Params): boolean {
		return this.compare(...args) <= 0;
	}

	////////////////////////////////////////////////////////////////////////////////

	public ge(...args: Vector4Params): boolean {
		return this.compare(...args) >= 0;
	}

	////////////////////////////////////////////////////////////////////////////////

	public eq(...args: Vector4Params): boolean {

		const v = Vector4.standardize(...args);
		return this._x === v.x && this._y === v.y && this._z === v.z && this._w === v.w;
	}

	////////////////////////////////////////////////////////////////////////////////

	public ne(...args: Vector4Params): boolean {

		const v = Vector4.standardize(...args);
		return this._x !== v.x || this._y !== v.y || this._z !== v.z || this._w !== v.w;
	}
}
