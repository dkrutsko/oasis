//----------------------------------------------------------------------------//
// Imports                                                                    //
//----------------------------------------------------------------------------//

import { Point, type PointObject, type PointParams } from './point';
import { Size,  type SizeObject,  type SizeParams  } from './size';

//----------------------------------------------------------------------------//
// Types                                                                      //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////

export interface BoundsXYWH {
	x: number;
	y: number;
	w: number;
	h: number;
}

////////////////////////////////////////////////////////////////////////////////

export interface BoundsLTRB {
	l: number;
	t: number;
	r: number;
	b: number;
}

////////////////////////////////////////////////////////////////////////////////

export type BoundsParams = Parameters<typeof Bounds.standardize>;
export type BoundsReturn = ReturnType<typeof Bounds.standardize>;

//----------------------------------------------------------------------------//
// Bounds                                                                     //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////

export class Bounds {
	//----------------------------------------------------------------------------//
	// Constants                                                                  //
	//----------------------------------------------------------------------------//

	////////////////////////////////////////////////////////////////////////////////

	public static readonly Zero = new Bounds(0, 0, 0, 0);

	//----------------------------------------------------------------------------//
	// Fields                                                                     //
	//----------------------------------------------------------------------------//

	////////////////////////////////////////////////////////////////////////////////

	private readonly _x: number;
	private readonly _y: number;
	private readonly _w: number;
	private readonly _h: number;

	//----------------------------------------------------------------------------//
	// Constructor                                                                //
	//----------------------------------------------------------------------------//

	////////////////////////////////////////////////////////////////////////////////

	public constructor(...args: BoundsParams) {

		const b = Bounds.standardize(...args);

		this._x = b.x;
		this._y = b.y;
		this._w = b.w;
		this._h = b.h;
	}

	//----------------------------------------------------------------------------//
	// Methods                                                                    //
	//----------------------------------------------------------------------------//

	////////////////////////////////////////////////////////////////////////////////

	public isZero(): boolean {
		return this._x === 0 && this._y === 0 && this._w === 0 && this._h === 0;
	}

	////////////////////////////////////////////////////////////////////////////////

	public isEmpty(): boolean {
		return this._w === 0 || this._h === 0;
	}

	////////////////////////////////////////////////////////////////////////////////

	public isValid(): boolean {
		return this._w > 0 && this._h > 0;
	}

	////////////////////////////////////////////////////////////////////////////////

	public setX(x: number): Bounds {
		return new Bounds(x, this._y, this._w, this._h);
	}

	////////////////////////////////////////////////////////////////////////////////

	public setY(y: number): Bounds {
		return new Bounds(this._x, y, this._w, this._h);
	}

	////////////////////////////////////////////////////////////////////////////////

	public setW(w: number): Bounds {
		return new Bounds(this._x, this._y, w, this._h);
	}

	////////////////////////////////////////////////////////////////////////////////

	public setH(h: number): Bounds {
		return new Bounds(this._x, this._y, this._w, h);
	}

	////////////////////////////////////////////////////////////////////////////////

	public setL(l: number): Bounds {
		return new Bounds(l, this._y, this._w, this._h);
	}

	////////////////////////////////////////////////////////////////////////////////

	public setT(t: number): Bounds {
		return new Bounds(this._x, t, this._w, this._h);
	}

	////////////////////////////////////////////////////////////////////////////////

	public setR(r: number): Bounds {
		return new Bounds(this._x, this._y, r - this._x, this._h);
	}

	////////////////////////////////////////////////////////////////////////////////

	public setB(b: number): Bounds {
		return new Bounds(this._x, this._y, this._w, b - this._y);
	}

	////////////////////////////////////////////////////////////////////////////////

	public getPoint(): Point {
		return new Point(this._x, this._y);
	}

	////////////////////////////////////////////////////////////////////////////////

	public setPoint(...args: PointParams): Bounds {
		const p = Point.standardize(...args);
		return new Bounds(p.x, p.y, this._w, this._h);
	}

	////////////////////////////////////////////////////////////////////////////////

	public getSize(): Size {
		return new Size(this._w, this._h);
	}

	////////////////////////////////////////////////////////////////////////////////

	public setSize(...args: SizeParams): Bounds {
		const s = Size.standardize(...args);
		return new Bounds(this._x, this._y, s.w, s.h);
	}

	////////////////////////////////////////////////////////////////////////////////

	public toArrayXYWH(): [number, number, number, number] {
		return [
			this._x,
			this._y,
			this._w,
			this._h,
		];
	}

	////////////////////////////////////////////////////////////////////////////////

	public toArrayLTRB(): [number, number, number, number] {
		return [
			this._x,
			this._y,
			this._x + this._w,
			this._y + this._h,
		];
	}

	////////////////////////////////////////////////////////////////////////////////

	public toXYWH(): BoundsXYWH {
		return {
			x: this._x,
			y: this._y,
			w: this._w,
			h: this._h,
		};
	}

	////////////////////////////////////////////////////////////////////////////////

	public toLTRB(): BoundsLTRB {
		return {
			l: this._x,
			t: this._y,
			r: this._x + this._w,
			b: this._y + this._h,
		};
	}

	////////////////////////////////////////////////////////////////////////////////

	public toString(): string {
		return `[${this._x.toFixed(2)}, ${this._y.toFixed(2)}, ${this._w.toFixed(2)}, ${this._h.toFixed(2)}]`;
	}

	////////////////////////////////////////////////////////////////////////////////

	public area(): number {
		return this._w * this._h;
	}

	////////////////////////////////////////////////////////////////////////////////

	public normalized(): Bounds {

		let x = this._x;
		let y = this._y;
		let w = this._w;
		let h = this._h;

		if (w < 0) {
			x += w;
			w = -w;
		}

		if (h < 0) {
			y += h;
			h = -h;
		}

		return new Bounds(x, y, w, h);
	}

	////////////////////////////////////////////////////////////////////////////////

	public center(): Point {

		return new Point(
			this._x + this._w * 0.5,
			this._y + this._h * 0.5,
		);
	}

	////////////////////////////////////////////////////////////////////////////////

	public inset(...args: SizeParams): Bounds {

		const s = Size.standardize(...args);

		return new Bounds(
			this._x + s.w,
			this._y + s.h,
			this._w - s.w * 2,
			this._h - s.h * 2,
		);
	}

	////////////////////////////////////////////////////////////////////////////////

	public containsPoint(inclusive: boolean, ...args: PointParams): boolean {

		const p = Point.standardize(...args);

		let l = this._x;
		let t = this._y;
		let r = this._x;
		let b = this._y;

		if (this._w < 0) {
			l += this._w;
		} else {
			r += this._w;
		}

		if (this._h < 0) {
			t += this._h;
		} else {
			b += this._h;
		}

		if (inclusive) {
			return l <= p.x && p.x <= r && t <= p.y && p.y <= b;
		}

		return l < p.x && p.x < r && t < p.y && p.y < b;
	}

	////////////////////////////////////////////////////////////////////////////////

	public contains(inclusive: boolean, ...args: BoundsParams): boolean {

		const bnds = Bounds.standardize(...args);

		if ((this._w === 0 && this._h === 0) || (bnds.w === 0 && bnds.h === 0)) {
			return false;
		}

		let l1 = this._x;
		let t1 = this._y;
		let r1 = this._x;
		let b1 = this._y;

		if (this._w < 0) {
			l1 += this._w;
		} else {
			r1 += this._w;
		}

		if (this._h < 0) {
			t1 += this._h;
		} else {
			b1 += this._h;
		}

		let l2 = bnds.x;
		let t2 = bnds.y;
		let r2 = bnds.x;
		let b2 = bnds.y;

		if (bnds.w < 0) {
			l2 += bnds.w;
		} else {
			r2 += bnds.w;
		}

		if (bnds.h < 0) {
			t2 += bnds.h;
		} else {
			b2 += bnds.h;
		}

		if (inclusive) {
			return l1 <= l2 && r1 >= r2 && t1 <= t2 && b1 >= b2;
		}

		return l1 < l2 && r1 > r2 && t1 < t2 && b1 > b2;
	}

	////////////////////////////////////////////////////////////////////////////////

	public intersects(inclusive: boolean, ...args: BoundsParams): boolean {

		const bnds = Bounds.standardize(...args);

		if ((this._w === 0 && this._h === 0) || (bnds.w === 0 && bnds.h === 0)) {
			return false;
		}

		let l1 = this._x;
		let t1 = this._y;
		let r1 = this._x;
		let b1 = this._y;

		if (this._w < 0) {
			l1 += this._w;
		} else {
			r1 += this._w;
		}

		if (this._h < 0) {
			t1 += this._h;
		} else {
			b1 += this._h;
		}

		let l2 = bnds.x;
		let t2 = bnds.y;
		let r2 = bnds.x;
		let b2 = bnds.y;

		if (bnds.w < 0) {
			l2 += bnds.w;
		} else {
			r2 += bnds.w;
		}

		if (bnds.h < 0) {
			t2 += bnds.h;
		} else {
			b2 += bnds.h;
		}

		if (inclusive) {
			return l1 <= r2 && r1 >= l2 && t1 <= b2 && b1 >= t2;
		}

		return l1 < r2 && r1 > l2 && t1 < b2 && b1 > t2;
	}

	////////////////////////////////////////////////////////////////////////////////

	public unite(...args: BoundsParams): Bounds {

		const bnds = Bounds.standardize(...args);

		let l1 = this._x;
		let t1 = this._y;
		let r1 = this._x;
		let b1 = this._y;

		if (this._w < 0) {
			l1 += this._w;
		} else {
			r1 += this._w;
		}

		if (this._h < 0) {
			t1 += this._h;
		} else {
			b1 += this._h;
		}

		let l2 = bnds.x;
		let t2 = bnds.y;
		let r2 = bnds.x;
		let b2 = bnds.y;

		if (bnds.w < 0) {
			l2 += bnds.w;
		} else {
			r2 += bnds.w;
		}

		if (bnds.h < 0) {
			t2 += bnds.h;
		} else {
			b2 += bnds.h;
		}

		if (this._w === 0 && this._h === 0) {
			return new Bounds({ l: l2, t: t2, r: r2, b: b2 });
		}

		if (bnds.w === 0 && bnds.h === 0) {
			return new Bounds({ l: l1, t: t1, r: r1, b: b1 });
		}

		return new Bounds({
			l: Math.min(l1, l2),
			t: Math.min(t1, t2),
			r: Math.max(r1, r2),
			b: Math.max(b1, b2),
		});
	}

	////////////////////////////////////////////////////////////////////////////////

	public intersect(...args: BoundsParams): Bounds {

		const bnds = Bounds.standardize(...args);

		if ((this._w === 0 && this._h === 0) || (bnds.w === 0 && bnds.h === 0)) {
			return new Bounds();
		}

		let l1 = this._x;
		let t1 = this._y;
		let r1 = this._x;
		let b1 = this._y;

		if (this._w < 0) {
			l1 += this._w;
		} else {
			r1 += this._w;
		}

		if (this._h < 0) {
			t1 += this._h;
		} else {
			b1 += this._h;
		}

		let l2 = bnds.x;
		let t2 = bnds.y;
		let r2 = bnds.x;
		let b2 = bnds.y;

		if (bnds.w < 0) {
			l2 += bnds.w;
		} else {
			r2 += bnds.w;
		}

		if (bnds.h < 0) {
			t2 += bnds.h;
		} else {
			b2 += bnds.h;
		}

		if (l1 > r2 || r1 < l2 || t1 > b2 || b1 < t2) {
			return new Bounds();
		}

		return new Bounds({
			l: Math.max(l1, l2),
			t: Math.max(t1, t2),
			r: Math.min(r1, r2),
			b: Math.min(b1, b2),
		});
	}

	//----------------------------------------------------------------------------//
	// Static                                                                     //
	//----------------------------------------------------------------------------//

	////////////////////////////////////////////////////////////////////////////////
	/// Converts a CSS margin or padding to Bounds. Accepts a wide range of inputs
	/// including strings such as 'Tunit Runit Bunit Lunit', objects such as
	/// { t, r, b, l }, arrays such as [ t, r, b, l ] and regular parameter lists.
	/// Parsing is done identically to CSS, therefore, some values can be omitted.

	public static fromMargin(
		t: number | string | Partial<BoundsLTRB> | (number | string)[],
		r?: number | string,
		b?: number | string,
		l?: number | string,
	): Bounds {

		// Whether input is a string
		if (typeof t === 'string') {
			// Split string using spaces
			const tokens = t.split(' ');

			// The rest of the parameters
			// must be undefined in order
			// to use string input ability
			if (r === undefined) {
				[t, r, b, l] = tokens;
			} else {
				// First token
				t = tokens[0];
			}

		} else {
			// The input is an array
			if (Array.isArray(t)) {
				// The rest of the parameters
				// must be undefined in order
				// to use array input ability
				if (r === undefined) {
					[t, r, b, l] = t;
				} else {
					// First element
					t = t[0];
				}
			}

			// The input resembles an object
			else if (typeof t === 'object') {
				r = t.r;
				b = t.b;
				l = t.l;
				t = t.t ?? 0;
			}
		}

		// Input is stored as a parameter list
		// Cascade to determine correct format

		if (t === undefined) {
			return new Bounds({ t: 0, r: 0, b: 0, l: 0 });
		}

		if (typeof t === 'string') {
			t = parseFloat(t) || 0;
		}

		if (r === undefined) {
			return new Bounds({ t: t, r: t, b: t, l: t });
		}

		if (typeof r === 'string') {
			r = parseFloat(r) || 0;
		}

		if (b === undefined) {
			return new Bounds({ t: t, r: r, b: t, l: r });
		}

		if (typeof b === 'string') {
			b = parseFloat(b) || 0;
		}

		if (l === undefined) {
			return new Bounds({ t: t, r: r, b: b, l: r });
		}

		if (typeof l === 'string') {
			l = parseFloat(l) || 0;
		}

		return new Bounds({ t: t, r: r, b: b, l: l });
	}

	////////////////////////////////////////////////////////////////////////////////

	public static standardize(
		ax?: number | Point | PointObject | Bounds | BoundsXYWH | BoundsLTRB,
		ay?: number | Size | SizeObject,
		aw?: number,
		ah?: number,
	): BoundsXYWH {

		if (ax instanceof Bounds) {
			return { x: ax._x, y: ax._y, w: ax._w, h: ax._h };
		}

		if (ax instanceof Point && ay instanceof Size) {
			return { x: ax.x, y: ax.y, w: ay.w, h: ay.h };
		}

		if (typeof ax === 'object') {
			if ('x' in ax && 'y' in ax) {
				if ('w' in ax && 'h' in ax) {
					return { x: ax.x, y: ax.y, w: ax.w, h: ax.h };
				}

				if (typeof ay === 'object') {
					return { x: ax.x, y: ax.y, w: ay.w, h: ay.h };
				}
			}

			if ('l' in ax && 't' in ax && 'r' in ax && 'b' in ax) {
				return {
					x: ax.l,
					w: ax.r - ax.l,
					y: ax.t,
					h: ax.b - ax.t,
				};
			}
		}

		if (typeof ax === 'number') {
			if (typeof ay === 'number') {
				if (typeof aw === 'number' && typeof ah === 'number') {
					return { x: ax, y: ay, w: aw, h: ah };
				}

				return { x: ax, y: ax, w: ay, h: ay };
			}

			return { x: ax, y: ax, w: ax, h: ax };
		}

		return { x: 0, y: 0, w: 0, h: 0 };
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

	public get w(): number {
		return this._w;
	}

	////////////////////////////////////////////////////////////////////////////////

	public get h(): number {
		return this._h;
	}

	////////////////////////////////////////////////////////////////////////////////

	public get l(): number {
		return this._x;
	}

	////////////////////////////////////////////////////////////////////////////////

	public get t(): number {
		return this._y;
	}

	////////////////////////////////////////////////////////////////////////////////

	public get r(): number {
		return this._x + this._w;
	}

	////////////////////////////////////////////////////////////////////////////////

	public get b(): number {
		return this._y + this._h;
	}

	//----------------------------------------------------------------------------//
	// Operators                                                                  //
	//----------------------------------------------------------------------------//

	////////////////////////////////////////////////////////////////////////////////

	public add(...args: BoundsParams): Bounds {

		const b = Bounds.standardize(...args);

		return new Bounds(
			this._x + b.x,
			this._y + b.y,
			this._w,
			this._h,
		);
	}

	////////////////////////////////////////////////////////////////////////////////

	public sub(...args: BoundsParams): Bounds {

		const b = Bounds.standardize(...args);

		return new Bounds(
			this._x - b.x,
			this._y - b.y,
			this._w,
			this._h,
		);
	}

	////////////////////////////////////////////////////////////////////////////////

	public mul(...args: BoundsParams): Bounds {

		const b = Bounds.standardize(...args);

		return new Bounds(
			this._x * b.x,
			this._y * b.y,
			this._w,
			this._h,
		);
	}

	////////////////////////////////////////////////////////////////////////////////

	public div(...args: BoundsParams): Bounds {

		const b = Bounds.standardize(...args);

		return new Bounds(
			this._x / b.x,
			this._y / b.y,
			this._w,
			this._h,
		);
	}

	////////////////////////////////////////////////////////////////////////////////

	public compare(...args: BoundsParams): number {

		const b = Bounds.standardize(...args);

		// Calculate area of both values
		const areaA = this._w * this._h;
		const areaB = b.w * b.h;

		if (areaA < areaB) {
			return -1;
		}
		if (areaA > areaB) {
			return 1;
		}

		return 0;
	}

	////////////////////////////////////////////////////////////////////////////////

	public lt(...args: BoundsParams): boolean {
		return this.compare(...args) < 0;
	}

	////////////////////////////////////////////////////////////////////////////////

	public gt(...args: BoundsParams): boolean {
		return this.compare(...args) > 0;
	}

	////////////////////////////////////////////////////////////////////////////////

	public le(...args: BoundsParams): boolean {
		return this.compare(...args) <= 0;
	}

	////////////////////////////////////////////////////////////////////////////////

	public ge(...args: BoundsParams): boolean {
		return this.compare(...args) >= 0;
	}

	////////////////////////////////////////////////////////////////////////////////

	public eq(...args: BoundsParams): boolean {

		const b = Bounds.standardize(...args);
		return this._x === b.x && this._y === b.y && this._w === b.w && this._h === b.h;
	}

	////////////////////////////////////////////////////////////////////////////////

	public ne(...args: BoundsParams): boolean {

		const b = Bounds.standardize(...args);
		return this._x !== b.x || this._y !== b.y || this._w !== b.w || this._h !== b.h;
	}
}
