//----------------------------------------------------------------------------//
// Imports                                                                    //
//----------------------------------------------------------------------------//

import { Vector2 } from './vector2';
import { Vector3 } from './vector3';
import { Vector4 } from './vector4';

//----------------------------------------------------------------------------//
// Types                                                                      //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////

export interface MatrixObject {
	m11: number; m12: number; m13: number; m14: number;
	m21: number; m22: number; m23: number; m24: number;
	m31: number; m32: number; m33: number; m34: number;
	m41: number; m42: number; m43: number; m44: number;
}

////////////////////////////////////////////////////////////////////////////////

export type MatrixParams = Parameters<typeof Matrix.standardize>;
export type MatrixReturn = ReturnType<typeof Matrix.standardize>;

//----------------------------------------------------------------------------//
// Matrix                                                                     //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////

export class Matrix {
	//----------------------------------------------------------------------------//
	// Constants                                                                  //
	//----------------------------------------------------------------------------//

	////////////////////////////////////////////////////////////////////////////////

	public static readonly Zero = new Matrix(
		0, 0, 0, 0,
		0, 0, 0, 0,
		0, 0, 0, 0,
		0, 0, 0, 0,
	);

	////////////////////////////////////////////////////////////////////////////////

	public static readonly Identity = new Matrix(
		1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, 0,
		0, 0, 0, 1,
	);

	//----------------------------------------------------------------------------//
	// Fields                                                                     //
	//----------------------------------------------------------------------------//

	////////////////////////////////////////////////////////////////////////////////

	private readonly _m11: number;
	private readonly _m12: number;
	private readonly _m13: number;
	private readonly _m14: number;

	private readonly _m21: number;
	private readonly _m22: number;
	private readonly _m23: number;
	private readonly _m24: number;

	private readonly _m31: number;
	private readonly _m32: number;
	private readonly _m33: number;
	private readonly _m34: number;

	private readonly _m41: number;
	private readonly _m42: number;
	private readonly _m43: number;
	private readonly _m44: number;

	//----------------------------------------------------------------------------//
	// Constructor                                                                //
	//----------------------------------------------------------------------------//

	////////////////////////////////////////////////////////////////////////////////

	public constructor(...args: MatrixParams) {

		const m = Matrix.standardize(...args);

		this._m11 = m.m11;
		this._m12 = m.m12;
		this._m13 = m.m13;
		this._m14 = m.m14;

		this._m21 = m.m21;
		this._m22 = m.m22;
		this._m23 = m.m23;
		this._m24 = m.m24;

		this._m31 = m.m31;
		this._m32 = m.m32;
		this._m33 = m.m33;
		this._m34 = m.m34;

		this._m41 = m.m41;
		this._m42 = m.m42;
		this._m43 = m.m43;
		this._m44 = m.m44;
	}

	//----------------------------------------------------------------------------//
	// Methods                                                                    //
	//----------------------------------------------------------------------------//

	////////////////////////////////////////////////////////////////////////////////

	public isZero(): boolean {

		return (
			this._m11 === 0 &&
			this._m12 === 0 &&
			this._m13 === 0 &&
			this._m14 === 0 &&

			this._m21 === 0 &&
			this._m22 === 0 &&
			this._m23 === 0 &&
			this._m24 === 0 &&

			this._m31 === 0 &&
			this._m32 === 0 &&
			this._m33 === 0 &&
			this._m34 === 0 &&

			this._m41 === 0 &&
			this._m42 === 0 &&
			this._m43 === 0 &&
			this._m44 === 0
		);
	}

	////////////////////////////////////////////////////////////////////////////////

	public toArray(): number[] {
		return [
			this._m11, this._m12, this._m13, this._m14,
			this._m21, this._m22, this._m23, this._m24,
			this._m31, this._m32, this._m33, this._m34,
			this._m41, this._m42, this._m43, this._m44,
		];
	}

	////////////////////////////////////////////////////////////////////////////////

	public toObject(): MatrixObject {
		return {
			m11: this._m11, m12: this._m12, m13: this._m13, m14: this._m14,
			m21: this._m21, m22: this._m22, m23: this._m23, m24: this._m24,
			m31: this._m31, m32: this._m32, m33: this._m33, m34: this._m34,
			m41: this._m41, m42: this._m42, m43: this._m43, m44: this._m44,
		};
	}

	////////////////////////////////////////////////////////////////////////////////

	public toString(): string {
		return (
			'[' +
				`[${this._m11.toFixed(2)}, ${this._m12.toFixed(2)}, ${this._m13.toFixed(2)}, ${this._m14.toFixed(2)}]` +
				`[${this._m21.toFixed(2)}, ${this._m22.toFixed(2)}, ${this._m23.toFixed(2)}, ${this._m24.toFixed(2)}]` +
				`[${this._m31.toFixed(2)}, ${this._m32.toFixed(2)}, ${this._m33.toFixed(2)}, ${this._m34.toFixed(2)}]` +
				`[${this._m41.toFixed(2)}, ${this._m42.toFixed(2)}, ${this._m43.toFixed(2)}, ${this._m44.toFixed(2)}]` +
			']'
		);
	}

	////////////////////////////////////////////////////////////////////////////////

	public transpose(): Matrix {

		return new Matrix(
			this._m11, this._m21, this._m31, this._m41,
			this._m12, this._m22, this._m32, this._m42,
			this._m13, this._m23, this._m33, this._m43,
			this._m14, this._m24, this._m34, this._m44,
		);
	}

	////////////////////////////////////////////////////////////////////////////////

	public invert(): Matrix {

		const v01 = this._m11 * this._m22 - this._m12 * this._m21;
		const v02 = this._m11 * this._m23 - this._m13 * this._m21;
		const v03 = this._m11 * this._m24 - this._m14 * this._m21;
		const v04 = this._m12 * this._m23 - this._m13 * this._m22;
		const v05 = this._m12 * this._m24 - this._m14 * this._m22;
		const v06 = this._m13 * this._m24 - this._m14 * this._m23;
		const v07 = this._m31 * this._m42 - this._m32 * this._m41;
		const v08 = this._m31 * this._m43 - this._m33 * this._m41;
		const v09 = this._m31 * this._m44 - this._m34 * this._m41;
		const v10 = this._m32 * this._m43 - this._m33 * this._m42;
		const v11 = this._m32 * this._m44 - this._m34 * this._m42;
		const v12 = this._m33 * this._m44 - this._m34 * this._m43;

		const det = v01 * v12 - v02 * v11 + v03 * v10 + v04 * v09 - v05 * v08 + v06 * v07;

		if (det === 0) {
			return Matrix.Zero;
		}

		return new Matrix(
			(+this._m22 * v12 - this._m23 * v11 + this._m24 * v10) / det,
			(-this._m12 * v12 + this._m13 * v11 - this._m14 * v10) / det,
			(+this._m42 * v06 - this._m43 * v05 + this._m44 * v04) / det,
			(-this._m32 * v06 + this._m33 * v05 - this._m34 * v04) / det,

			(-this._m21 * v12 + this._m23 * v09 - this._m24 * v08) / det,
			(+this._m11 * v12 - this._m13 * v09 + this._m14 * v08) / det,
			(-this._m41 * v06 + this._m43 * v03 - this._m44 * v02) / det,
			(+this._m31 * v06 - this._m33 * v03 + this._m34 * v02) / det,

			(+this._m21 * v11 - this._m22 * v09 + this._m24 * v07) / det,
			(-this._m11 * v11 + this._m12 * v09 - this._m14 * v07) / det,
			(+this._m41 * v05 - this._m42 * v03 + this._m44 * v01) / det,
			(-this._m31 * v05 + this._m32 * v03 - this._m34 * v01) / det,

			(-this._m21 * v10 + this._m22 * v08 - this._m23 * v07) / det,
			(+this._m11 * v10 - this._m12 * v08 + this._m13 * v07) / det,
			(-this._m41 * v04 + this._m42 * v02 - this._m43 * v01) / det,
			(+this._m31 * v04 - this._m32 * v02 + this._m33 * v01) / det,
		);
	}

	////////////////////////////////////////////////////////////////////////////////

	public determinant(): number {

		const v1 = this._m33 * this._m44 - this._m34 * this._m43;
		const v2 = this._m32 * this._m44 - this._m34 * this._m42;
		const v3 = this._m32 * this._m43 - this._m33 * this._m42;
		const v4 = this._m31 * this._m44 - this._m34 * this._m41;
		const v5 = this._m31 * this._m43 - this._m33 * this._m41;
		const v6 = this._m31 * this._m42 - this._m32 * this._m41;

		return (
			(this._m11 * (this._m22 * v1 - this._m23 * v2 + this._m24 * v3)) -
			(this._m12 * (this._m21 * v1 - this._m23 * v4 + this._m24 * v5)) +
			(this._m13 * (this._m21 * v2 - this._m22 * v4 + this._m24 * v6)) -
			(this._m14 * (this._m21 * v3 - this._m22 * v5 + this._m23 * v6))
		);
	}

	//----------------------------------------------------------------------------//
	// Static                                                                     //
	//----------------------------------------------------------------------------//

	////////////////////////////////////////////////////////////////////////////////

	public static standardize(
		a11?: number | Matrix | MatrixObject | number[],
		a12?: number, a13?: number, a14?: number,
		a21?: number, a22?: number, a23?: number, a24?: number,
		a31?: number, a32?: number, a33?: number, a34?: number,
		a41?: number, a42?: number, a43?: number, a44?: number,
	): MatrixObject {

		if (a11 instanceof Matrix) {
			return {
				m11: a11._m11, m12: a11._m12, m13: a11._m13, m14: a11._m14,
				m21: a11._m21, m22: a11._m22, m23: a11._m23, m24: a11._m24,
				m31: a11._m31, m32: a11._m32, m33: a11._m33, m34: a11._m34,
				m41: a11._m41, m42: a11._m42, m43: a11._m43, m44: a11._m44,
			};
		}

		if (Array.isArray(a11)) {
			const [
				m11, m12, m13, m14,
				m21, m22, m23, m24,
				m31, m32, m33, m34,
				m41, m42, m43, m44,
			] = a11;

			return {
				m11: typeof m11 === 'number' ? m11 : 0,
				m12: typeof m12 === 'number' ? m12 : 0,
				m13: typeof m13 === 'number' ? m13 : 0,
				m14: typeof m14 === 'number' ? m14 : 0,

				m21: typeof m21 === 'number' ? m21 : 0,
				m22: typeof m22 === 'number' ? m22 : 0,
				m23: typeof m23 === 'number' ? m23 : 0,
				m24: typeof m24 === 'number' ? m24 : 0,

				m31: typeof m31 === 'number' ? m31 : 0,
				m32: typeof m32 === 'number' ? m32 : 0,
				m33: typeof m33 === 'number' ? m33 : 0,
				m34: typeof m34 === 'number' ? m34 : 0,

				m41: typeof m41 === 'number' ? m41 : 0,
				m42: typeof m42 === 'number' ? m42 : 0,
				m43: typeof m43 === 'number' ? m43 : 0,
				m44: typeof m44 === 'number' ? m44 : 0,
			};
		}

		if (typeof a11 === 'object' && a11 !== null) {
			return {
				m11: typeof a11.m11 === 'number' ? a11.m11 : 0,
				m12: typeof a11.m12 === 'number' ? a11.m12 : 0,
				m13: typeof a11.m13 === 'number' ? a11.m13 : 0,
				m14: typeof a11.m14 === 'number' ? a11.m14 : 0,

				m21: typeof a11.m21 === 'number' ? a11.m21 : 0,
				m22: typeof a11.m22 === 'number' ? a11.m22 : 0,
				m23: typeof a11.m23 === 'number' ? a11.m23 : 0,
				m24: typeof a11.m24 === 'number' ? a11.m24 : 0,

				m31: typeof a11.m31 === 'number' ? a11.m31 : 0,
				m32: typeof a11.m32 === 'number' ? a11.m32 : 0,
				m33: typeof a11.m33 === 'number' ? a11.m33 : 0,
				m34: typeof a11.m34 === 'number' ? a11.m34 : 0,

				m41: typeof a11.m41 === 'number' ? a11.m41 : 0,
				m42: typeof a11.m42 === 'number' ? a11.m42 : 0,
				m43: typeof a11.m43 === 'number' ? a11.m43 : 0,
				m44: typeof a11.m44 === 'number' ? a11.m44 : 0,
			};
		}

		if (typeof a11 === 'number') {

			if (typeof a12 === 'number' && typeof a13 === 'number' && typeof a14 === 'number' &&
				typeof a21 === 'number' && typeof a22 === 'number' && typeof a23 === 'number' && typeof a24 === 'number' &&
				typeof a31 === 'number' && typeof a32 === 'number' && typeof a33 === 'number' && typeof a34 === 'number' &&
				typeof a41 === 'number' && typeof a42 === 'number' && typeof a43 === 'number' && typeof a44 === 'number') {

				return {
					m11: a11, m12: a12, m13: a13, m14: a14,
					m21: a21, m22: a22, m23: a23, m24: a24,
					m31: a31, m32: a32, m33: a33, m34: a34,
					m41: a41, m42: a42, m43: a43, m44: a44,
				};
			}

			return {
				m11: a11, m12: a11, m13: a11, m14: a11,
				m21: a11, m22: a11, m23: a11, m24: a11,
				m31: a11, m32: a11, m33: a11, m34: a11,
				m41: a11, m42: a11, m43: a11, m44: a11,
			};
		}

		return {
			m11: 0, m12: 0, m13: 0, m14: 0,
			m21: 0, m22: 0, m23: 0, m24: 0,
			m31: 0, m32: 0, m33: 0, m34: 0,
			m41: 0, m42: 0, m43: 0, m44: 0,
		};
	}

	////////////////////////////////////////////////////////////////////////////////

	public static createProj(
		fov: number,
		width: number,
		height: number,
		near: number,
		far: number,
	): Matrix {

		// Do parameter check
		if (fov <= 0 || fov >= Math.PI || near <= 0 || far <= 0 || near >= far) {
			return Matrix.Zero;
		}

		const w = width;
		const h = height;

		const dif = near - far;
		const cot = 1 / Math.tan(fov * 0.5);
		const asp = cot / (w / h);

		const m33 = far / dif;
		const m43 = near * far / dif;

		return new Matrix(
			asp, 0.0, 0.0, 0.0,
			0.0, cot, 0.0, 0.0,
			0.0, 0.0, m33, -1.0,
			0.0, 0.0, m43, 0.0,
		);
	}

	////////////////////////////////////////////////////////////////////////////////

	public static createView(pos: Vector3, target: Vector3, up: Vector3): Matrix {

		const vz = pos.sub(target).normalize();
		const vx = up.cross(vz).normalize();
		const vy = vz.cross(vx).normalize();

		return new Matrix(
			vx.x, vy.x, vz.x, 0,
			vx.y, vy.y, vz.y, 0,
			vx.z, vy.z, vz.z, 0,
			-vx.dot(pos),
			-vy.dot(pos),
			-vz.dot(pos), 1,
		);
	}

	////////////////////////////////////////////////////////////////////////////////

	public static project(
		pos: Vector3,
		width: number,
		height: number,
		model: Matrix,
		view: Matrix,
		proj: Matrix,
	): Vector3 {

		return Matrix.projectWithMvp(pos, width, height, model.mul(view).mul(proj));
	}

	////////////////////////////////////////////////////////////////////////////////

	public static projectWithMvp(
		pos: Vector3,
		width: number,
		height: number,
		mvp: Matrix,
	): Vector3 {

		const minZ = 0.0;
		const maxZ = 1.0;

		const transform = Vector4.transformVector3(mvp, pos);

		return new Vector3(
			(1 + transform.x / transform.w) * width * 0.5,
			(1 - transform.y / transform.w) * height * 0.5,
			minZ + (transform.z / transform.w) * (maxZ - minZ),
		);
	}

	//----------------------------------------------------------------------------//
	// Accessors                                                                  //
	//----------------------------------------------------------------------------//

	////////////////////////////////////////////////////////////////////////////////

	public get m11(): number {
		return this._m11;
	}
	public get m12(): number {
		return this._m12;
	}
	public get m13(): number {
		return this._m13;
	}
	public get m14(): number {
		return this._m14;
	}

	////////////////////////////////////////////////////////////////////////////////

	public get m21(): number {
		return this._m21;
	}
	public get m22(): number {
		return this._m22;
	}
	public get m23(): number {
		return this._m23;
	}
	public get m24(): number {
		return this._m24;
	}

	////////////////////////////////////////////////////////////////////////////////

	public get m31(): number {
		return this._m31;
	}
	public get m32(): number {
		return this._m32;
	}
	public get m33(): number {
		return this._m33;
	}
	public get m34(): number {
		return this._m34;
	}

	////////////////////////////////////////////////////////////////////////////////

	public get m41(): number {
		return this._m41;
	}
	public get m42(): number {
		return this._m42;
	}
	public get m43(): number {
		return this._m43;
	}
	public get m44(): number {
		return this._m44;
	}

	//----------------------------------------------------------------------------//
	// Operators                                                                  //
	//----------------------------------------------------------------------------//

	////////////////////////////////////////////////////////////////////////////////

	public add(...args: MatrixParams): Matrix {

		const m = Matrix.standardize(...args);

		return new Matrix(
			this._m11 + m.m11,
			this._m12 + m.m12,
			this._m13 + m.m13,
			this._m14 + m.m14,

			this._m21 + m.m21,
			this._m22 + m.m22,
			this._m23 + m.m23,
			this._m24 + m.m24,

			this._m31 + m.m31,
			this._m32 + m.m32,
			this._m33 + m.m33,
			this._m34 + m.m34,

			this._m41 + m.m41,
			this._m42 + m.m42,
			this._m43 + m.m43,
			this._m44 + m.m44,
		);
	}

	////////////////////////////////////////////////////////////////////////////////

	public sub(...args: MatrixParams): Matrix {

		const m = Matrix.standardize(...args);

		return new Matrix(
			this._m11 - m.m11,
			this._m12 - m.m12,
			this._m13 - m.m13,
			this._m14 - m.m14,

			this._m21 - m.m21,
			this._m22 - m.m22,
			this._m23 - m.m23,
			this._m24 - m.m24,

			this._m31 - m.m31,
			this._m32 - m.m32,
			this._m33 - m.m33,
			this._m34 - m.m34,

			this._m41 - m.m41,
			this._m42 - m.m42,
			this._m43 - m.m43,
			this._m44 - m.m44,
		);
	}

	////////////////////////////////////////////////////////////////////////////////

	public mul(...args: MatrixParams): Matrix {

		const m = Matrix.standardize(...args);

		return new Matrix(
			this._m11 * m.m11 + this._m12 * m.m21 + this._m13 * m.m31 + this._m14 * m.m41,
			this._m11 * m.m12 + this._m12 * m.m22 + this._m13 * m.m32 + this._m14 * m.m42,
			this._m11 * m.m13 + this._m12 * m.m23 + this._m13 * m.m33 + this._m14 * m.m43,
			this._m11 * m.m14 + this._m12 * m.m24 + this._m13 * m.m34 + this._m14 * m.m44,

			this._m21 * m.m11 + this._m22 * m.m21 + this._m23 * m.m31 + this._m24 * m.m41,
			this._m21 * m.m12 + this._m22 * m.m22 + this._m23 * m.m32 + this._m24 * m.m42,
			this._m21 * m.m13 + this._m22 * m.m23 + this._m23 * m.m33 + this._m24 * m.m43,
			this._m21 * m.m14 + this._m22 * m.m24 + this._m23 * m.m34 + this._m24 * m.m44,

			this._m31 * m.m11 + this._m32 * m.m21 + this._m33 * m.m31 + this._m34 * m.m41,
			this._m31 * m.m12 + this._m32 * m.m22 + this._m33 * m.m32 + this._m34 * m.m42,
			this._m31 * m.m13 + this._m32 * m.m23 + this._m33 * m.m33 + this._m34 * m.m43,
			this._m31 * m.m14 + this._m32 * m.m24 + this._m33 * m.m34 + this._m34 * m.m44,

			this._m41 * m.m11 + this._m42 * m.m21 + this._m43 * m.m31 + this._m44 * m.m41,
			this._m41 * m.m12 + this._m42 * m.m22 + this._m43 * m.m32 + this._m44 * m.m42,
			this._m41 * m.m13 + this._m42 * m.m23 + this._m43 * m.m33 + this._m44 * m.m43,
			this._m41 * m.m14 + this._m42 * m.m24 + this._m43 * m.m34 + this._m44 * m.m44,
		);
	}

	////////////////////////////////////////////////////////////////////////////////

	public div(...args: MatrixParams): Matrix {

		const m = Matrix.standardize(...args);

		return new Matrix(
			this._m11 / m.m11,
			this._m12 / m.m12,
			this._m13 / m.m13,
			this._m14 / m.m14,

			this._m21 / m.m21,
			this._m22 / m.m22,
			this._m23 / m.m23,
			this._m24 / m.m24,

			this._m31 / m.m31,
			this._m32 / m.m32,
			this._m33 / m.m33,
			this._m34 / m.m34,

			this._m41 / m.m41,
			this._m42 / m.m42,
			this._m43 / m.m43,
			this._m44 / m.m44,
		);
	}

	////////////////////////////////////////////////////////////////////////////////

	public addScalar(scalar: number): Matrix {

		return new Matrix(
			this._m11 + scalar,
			this._m12 + scalar,
			this._m13 + scalar,
			this._m14 + scalar,

			this._m21 + scalar,
			this._m22 + scalar,
			this._m23 + scalar,
			this._m24 + scalar,

			this._m31 + scalar,
			this._m32 + scalar,
			this._m33 + scalar,
			this._m34 + scalar,

			this._m41 + scalar,
			this._m42 + scalar,
			this._m43 + scalar,
			this._m44 + scalar,
		);
	}

	////////////////////////////////////////////////////////////////////////////////

	public subScalar(scalar: number): Matrix {

		return new Matrix(
			this._m11 - scalar,
			this._m12 - scalar,
			this._m13 - scalar,
			this._m14 - scalar,

			this._m21 - scalar,
			this._m22 - scalar,
			this._m23 - scalar,
			this._m24 - scalar,

			this._m31 - scalar,
			this._m32 - scalar,
			this._m33 - scalar,
			this._m34 - scalar,

			this._m41 - scalar,
			this._m42 - scalar,
			this._m43 - scalar,
			this._m44 - scalar,
		);
	}

	////////////////////////////////////////////////////////////////////////////////

	public mulScalar(scalar: number): Matrix {

		return new Matrix(
			this._m11 * scalar,
			this._m12 * scalar,
			this._m13 * scalar,
			this._m14 * scalar,

			this._m21 * scalar,
			this._m22 * scalar,
			this._m23 * scalar,
			this._m24 * scalar,

			this._m31 * scalar,
			this._m32 * scalar,
			this._m33 * scalar,
			this._m34 * scalar,

			this._m41 * scalar,
			this._m42 * scalar,
			this._m43 * scalar,
			this._m44 * scalar,
		);
	}

	////////////////////////////////////////////////////////////////////////////////

	public divScalar(scalar: number): Matrix {

		return new Matrix(
			this._m11 / scalar,
			this._m12 / scalar,
			this._m13 / scalar,
			this._m14 / scalar,

			this._m21 / scalar,
			this._m22 / scalar,
			this._m23 / scalar,
			this._m24 / scalar,

			this._m31 / scalar,
			this._m32 / scalar,
			this._m33 / scalar,
			this._m34 / scalar,

			this._m41 / scalar,
			this._m42 / scalar,
			this._m43 / scalar,
			this._m44 / scalar,
		);
	}

	////////////////////////////////////////////////////////////////////////////////

	public neg(): Matrix {

		return new Matrix(
			-this._m11, -this._m12, -this._m13, -this._m14,
			-this._m21, -this._m22, -this._m23, -this._m24,
			-this._m31, -this._m32, -this._m33, -this._m34,
			-this._m41, -this._m42, -this._m43, -this._m44,
		);
	}

	////////////////////////////////////////////////////////////////////////////////

	public eq(...args: MatrixParams): boolean {

		const m = Matrix.standardize(...args);

		return (
			this._m11 === m.m11 &&
			this._m12 === m.m12 &&
			this._m13 === m.m13 &&
			this._m14 === m.m14 &&

			this._m21 === m.m21 &&
			this._m22 === m.m22 &&
			this._m23 === m.m23 &&
			this._m24 === m.m24 &&

			this._m31 === m.m31 &&
			this._m32 === m.m32 &&
			this._m33 === m.m33 &&
			this._m34 === m.m34 &&

			this._m41 === m.m41 &&
			this._m42 === m.m42 &&
			this._m43 === m.m43 &&
			this._m44 === m.m44
		);
	}

	////////////////////////////////////////////////////////////////////////////////

	public ne(...args: MatrixParams): boolean {

		const m = Matrix.standardize(...args);

		return (
			this._m11 !== m.m11 ||
			this._m12 !== m.m12 ||
			this._m13 !== m.m13 ||
			this._m14 !== m.m14 ||

			this._m21 !== m.m21 ||
			this._m22 !== m.m22 ||
			this._m23 !== m.m23 ||
			this._m24 !== m.m24 ||

			this._m31 !== m.m31 ||
			this._m32 !== m.m32 ||
			this._m33 !== m.m33 ||
			this._m34 !== m.m34 ||

			this._m41 !== m.m41 ||
			this._m42 !== m.m42 ||
			this._m43 !== m.m43 ||
			this._m44 !== m.m44
		);
	}
}
