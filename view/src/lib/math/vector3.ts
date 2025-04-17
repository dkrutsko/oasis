//----------------------------------------------------------------------------//
// Imports                                                                    //
//----------------------------------------------------------------------------//

import { Vector2 } from './vector2';
import { Vector4 } from './vector4';
import { Matrix } from './matrix';

//----------------------------------------------------------------------------//
// Types                                                                      //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////

export interface Vector3Object {
	x: number;
	y: number;
	z: number;
}

////////////////////////////////////////////////////////////////////////////////

export type Vector3Params = Parameters<typeof Vector3.standardize>;
export type Vector3Return = ReturnType<typeof Vector3.standardize>;

//----------------------------------------------------------------------------//
// Vector3                                                                    //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////

export class Vector3 {
	//----------------------------------------------------------------------------//
	// Constants                                                                  //
	//----------------------------------------------------------------------------//

	////////////////////////////////////////////////////////////////////////////////

	public static readonly Zero  = new Vector3(0, 0, 0);
	public static readonly UnitX = new Vector3(1, 0, 0);
	public static readonly UnitY = new Vector3(0, 1, 0);
	public static readonly UnitZ = new Vector3(0, 0, 1);

	//----------------------------------------------------------------------------//
	// Fields                                                                     //
	//----------------------------------------------------------------------------//

	////////////////////////////////////////////////////////////////////////////////

	private readonly _x: number;
	private readonly _y: number;
	private readonly _z: number;

	//----------------------------------------------------------------------------//
	// Constructor                                                                //
	//----------------------------------------------------------------------------//

	////////////////////////////////////////////////////////////////////////////////

	public constructor(...args: Vector3Params) {

		const v = Vector3.standardize(...args);

		this._x = v.x;
		this._y = v.y;
		this._z = v.z;
	}

	//----------------------------------------------------------------------------//
	// Methods                                                                    //
	//----------------------------------------------------------------------------//

	////////////////////////////////////////////////////////////////////////////////

	public isZero(): boolean {
		return this._x === 0 && this._y === 0 && this._z === 0;
	}

	////////////////////////////////////////////////////////////////////////////////

	public setX(x: number): Vector3 {
		return new Vector3(x, this._y, this._z);
	}

	////////////////////////////////////////////////////////////////////////////////

	public setY(y: number): Vector3 {
		return new Vector3(this._x, y, this._z);
	}

	////////////////////////////////////////////////////////////////////////////////

	public setZ(z: number): Vector3 {
		return new Vector3(this._x, this._y, z);
	}

	////////////////////////////////////////////////////////////////////////////////

	public toArray(): [number, number, number] {
		return [
			this._x,
			this._y,
			this._z,
		];
	}

	////////////////////////////////////////////////////////////////////////////////

	public toObject(): Vector3Object {
		return {
			x: this._x,
			y: this._y,
			z: this._z,
		};
	}

	////////////////////////////////////////////////////////////////////////////////

	public toString(): string {
		return `[${this._x.toFixed(2)}, ${this._y.toFixed(2)}, ${this._z.toFixed(2)}]`;
	}

	////////////////////////////////////////////////////////////////////////////////

	public normalize(): Vector3 {

		const magnitude = Math.sqrt(this._x ** 2 + this._y ** 2 + this._z ** 2);

		// The default case
		if (magnitude === 0) {
			return Vector3.Zero;
		}

		return new Vector3(
			this._x / magnitude,
			this._y / magnitude,
			this._z / magnitude,
		);
	}

	////////////////////////////////////////////////////////////////////////////////

	public distance(value: Vector3): number {

		const dx = this._x - value._x;
		const dy = this._y - value._y;
		const dz = this._z - value._z;

		return Math.sqrt(dx * dx + dy * dy + dz * dz);
	}

	////////////////////////////////////////////////////////////////////////////////

	public distanceSq(value: Vector3): number {

		const dx = this._x - value._x;
		const dy = this._y - value._y;
		const dz = this._z - value._z;

		return dx * dx + dy * dy + dz * dz;
	}

	////////////////////////////////////////////////////////////////////////////////

	public dot(value: Vector3): number {

		return this._x * value._x + this._y * value._y + this._z * value._z;
	}

	////////////////////////////////////////////////////////////////////////////////

	public reflect(normal: Vector3): Vector3 {

		const dot = 2 * (this._x * normal._x + this._y * normal._y + this._z * normal._z);

		return new Vector3(
			this._x - dot * normal._x,
			this._y - dot * normal._y,
			this._z - dot * normal._z,
		);
	}

	////////////////////////////////////////////////////////////////////////////////

	public cross(value: Vector3): Vector3 {

		return new Vector3(
			this._y * value._z - this._z * value._y,
			this._z * value._x - this._x * value._z,
			this._x * value._y - this._y * value._x,
		);
	}

	////////////////////////////////////////////////////////////////////////////////

	public rotate(axis: Vector3, angle: number): Vector3 {

		const sinAngle = Math.sin(-angle);
		const cosAngle = Math.cos(-angle);

		const crossTerm = this.cross(axis.mul(sinAngle));
		const scaledVal = this.mul(cosAngle);
		const scaledAxis = axis.mul(1 - cosAngle);
		const dotTerm = axis.mul(this.dot(scaledAxis));

		return crossTerm.add(scaledVal).add(dotTerm);
	}

	////////////////////////////////////////////////////////////////////////////////

	public length(): number {

		return Math.sqrt(this._x ** 2 + this._y ** 2 + this._z ** 2);
	}

	////////////////////////////////////////////////////////////////////////////////

	public lengthSq(): number {

		return this._x ** 2 + this._y ** 2 + this._z ** 2;
	}

	//----------------------------------------------------------------------------//
	// Static                                                                     //
	//----------------------------------------------------------------------------//

	////////////////////////////////////////////////////////////////////////////////

	public static standardize(
		ax?: number | Vector3 | Vector3Object | [number, number, number],
		ay?: number,
		az?: number,
	): Vector3Object {

		if (ax instanceof Vector3) {
			return { x: ax._x, y: ax._y, z: ax._z };
		}

		if (Array.isArray(ax)) {
			const [x, y, z] = ax;
			return {
				x: typeof x === 'number' ? x : 0,
				y: typeof y === 'number' ? y : 0,
				z: typeof z === 'number' ? z : 0,
			};
		}

		if (typeof ax === 'object' && ax !== null) {
			return {
				x: typeof ax.x === 'number' ? ax.x : 0,
				y: typeof ax.y === 'number' ? ax.y : 0,
				z: typeof ax.z === 'number' ? ax.z : 0,
			};
		}

		if (typeof ax === 'number') {
			if (typeof ay === 'number' && typeof az === 'number') {
				return { x: ax, y: ay, z: az };
			}

			return { x: ax, y: ax, z: ax };
		}

		return { x: 0, y: 0, z: 0 };
	}

	////////////////////////////////////////////////////////////////////////////////

	public static transformVector2(matrix: Matrix, value: Vector2): Vector3 {

		return new Vector3(
			matrix.m11 * value.x + matrix.m21 * value.y + matrix.m41,
			matrix.m12 * value.x + matrix.m22 * value.y + matrix.m42,
			matrix.m13 * value.x + matrix.m23 * value.y + matrix.m43,
		);
	}

	////////////////////////////////////////////////////////////////////////////////

	public static transformVector3(matrix: Matrix, value: Vector3): Vector3 {

		return new Vector3(
			matrix.m11 * value.x + matrix.m21 * value.y + matrix.m31 * value.z + matrix.m41,
			matrix.m12 * value.x + matrix.m22 * value.y + matrix.m32 * value.z + matrix.m42,
			matrix.m13 * value.x + matrix.m23 * value.y + matrix.m33 * value.z + matrix.m43,
		);
	}

	////////////////////////////////////////////////////////////////////////////////

	public static transformVector4(matrix: Matrix, value: Vector4): Vector3 {

		return new Vector3(
			matrix.m11 * value.x + matrix.m21 * value.y + matrix.m31 * value.z + matrix.m41 * value.w,
			matrix.m12 * value.x + matrix.m22 * value.y + matrix.m32 * value.z + matrix.m42 * value.w,
			matrix.m13 * value.x + matrix.m23 * value.y + matrix.m33 * value.z + matrix.m43 * value.w,
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

	////////////////////////////////////////////////////////////////////////////////

	public get z(): number {
		return this._z;
	}

	//----------------------------------------------------------------------------//
	// Operators                                                                  //
	//----------------------------------------------------------------------------//

	////////////////////////////////////////////////////////////////////////////////

	public add(...args: Vector3Params): Vector3 {

		const v = Vector3.standardize(...args);

		return new Vector3(
			this._x + v.x,
			this._y + v.y,
			this._z + v.z,
		);
	}

	////////////////////////////////////////////////////////////////////////////////

	public sub(...args: Vector3Params): Vector3 {

		const v = Vector3.standardize(...args);

		return new Vector3(
			this._x - v.x,
			this._y - v.y,
			this._z - v.z,
		);
	}

	////////////////////////////////////////////////////////////////////////////////

	public mul(...args: Vector3Params): Vector3 {

		const v = Vector3.standardize(...args);

		return new Vector3(
			this._x * v.x,
			this._y * v.y,
			this._z * v.z,
		);
	}

	////////////////////////////////////////////////////////////////////////////////

	public div(...args: Vector3Params): Vector3 {

		const v = Vector3.standardize(...args);

		return new Vector3(
			this._x / v.x,
			this._y / v.y,
			this._z / v.z,
		);
	}

	////////////////////////////////////////////////////////////////////////////////

	public neg(): Vector3 {

		return new Vector3(
			-this._x,
			-this._y,
			-this._z,
		);
	}

	////////////////////////////////////////////////////////////////////////////////

	public compare(...args: Vector3Params): number {

		const v = Vector3.standardize(...args);

		if (this._x !== v.x) {
			return this._x < v.x ? -1 : 1;
		}

		if (this._y !== v.y) {
			return this._y < v.y ? -1 : 1;
		}

		if (this._z !== v.z) {
			return this._z < v.z ? -1 : 1;
		}

		return 0;
	}

	////////////////////////////////////////////////////////////////////////////////

	public lt(...args: Vector3Params): boolean {
		return this.compare(...args) < 0;
	}

	////////////////////////////////////////////////////////////////////////////////

	public gt(...args: Vector3Params): boolean {
		return this.compare(...args) > 0;
	}

	////////////////////////////////////////////////////////////////////////////////

	public le(...args: Vector3Params): boolean {
		return this.compare(...args) <= 0;
	}

	////////////////////////////////////////////////////////////////////////////////

	public ge(...args: Vector3Params): boolean {
		return this.compare(...args) >= 0;
	}

	////////////////////////////////////////////////////////////////////////////////

	public eq(...args: Vector3Params): boolean {

		const v = Vector3.standardize(...args);
		return this._x === v.x && this._y === v.y && this._z === v.z;
	}

	////////////////////////////////////////////////////////////////////////////////

	public ne(...args: Vector3Params): boolean {

		const v = Vector3.standardize(...args);
		return this._x !== v.x || this._y !== v.y || this._z !== v.z;
	}
}
