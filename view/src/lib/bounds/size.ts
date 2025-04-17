//----------------------------------------------------------------------------//
// Imports                                                                    //
//----------------------------------------------------------------------------//

import { Point } from './point';

//----------------------------------------------------------------------------//
// Types                                                                      //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////

export interface SizeObject {
	w: number;
	h: number;
}

////////////////////////////////////////////////////////////////////////////////

export type SizeParams = Parameters<typeof Size.standardize>;
export type SizeReturn = ReturnType<typeof Size.standardize>;

//----------------------------------------------------------------------------//
// Size                                                                       //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////

export class Size {
	//----------------------------------------------------------------------------//
	// Constants                                                                  //
	//----------------------------------------------------------------------------//

	////////////////////////////////////////////////////////////////////////////////

	public static readonly Zero = new Size(0, 0);

	//----------------------------------------------------------------------------//
	// Fields                                                                     //
	//----------------------------------------------------------------------------//

	////////////////////////////////////////////////////////////////////////////////

	private readonly _w: number;
	private readonly _h: number;

	//----------------------------------------------------------------------------//
	// Constructor                                                                //
	//----------------------------------------------------------------------------//

	////////////////////////////////////////////////////////////////////////////////

	public constructor(...args: SizeParams) {

		const s = Size.standardize(...args);

		this._w = s.w;
		this._h = s.h;
	}

	//----------------------------------------------------------------------------//
	// Methods                                                                    //
	//----------------------------------------------------------------------------//

	////////////////////////////////////////////////////////////////////////////////

	public isZero(): boolean {
		return this._w === 0 && this._h === 0;
	}

	////////////////////////////////////////////////////////////////////////////////

	public isEmpty(): boolean {
		return this._w === 0 && this._h === 0;
	}

	////////////////////////////////////////////////////////////////////////////////

	public setW(w: number): Size {
		return new Size(w, this._h);
	}

	////////////////////////////////////////////////////////////////////////////////

	public setH(h: number): Size {
		return new Size(this._w, h);
	}

	////////////////////////////////////////////////////////////////////////////////

	public toPoint(): Point {
		return new Point(this._w, this._h);
	}

	////////////////////////////////////////////////////////////////////////////////

	public toArray(): [number, number] {
		return [
			this._w,
			this._h,
		];
	}

	////////////////////////////////////////////////////////////////////////////////

	public toObject(): SizeObject {
		return {
			w: this._w,
			h: this._h,
		};
	}

	////////////////////////////////////////////////////////////////////////////////

	public toString(): string {
		return `[${this._w.toFixed(2)}, ${this._h.toFixed(2)}]`;
	}

	////////////////////////////////////////////////////////////////////////////////

	public area(): number {
		return this._w * this._h;
	}

	////////////////////////////////////////////////////////////////////////////////

	public min(): number {
		return Math.min(this._w, this._h);
	}

	////////////////////////////////////////////////////////////////////////////////

	public max(): number {
		return Math.max(this._w, this._h);
	}

	////////////////////////////////////////////////////////////////////////////////

	public clamp(min: Size, max: Size): Size {

		return new Size(
			Math.max(min.w, Math.min(this._w, max.w)),
			Math.max(min.h, Math.min(this._h, max.h)),
		);
	}

	////////////////////////////////////////////////////////////////////////////////

	public aspect(): number {
		return this._w === 0 ? 0 : this._h / this._w;
	}

	////////////////////////////////////////////////////////////////////////////////

	public contains(...args: SizeParams): boolean {

		const s = Size.standardize(...args);
		return this._w >= s.w && this._h >= s.h;
	}

	//----------------------------------------------------------------------------//
	// Static                                                                     //
	//----------------------------------------------------------------------------//

	////////////////////////////////////////////////////////////////////////////////

	public static standardize(
		aw?: number | Size | SizeObject | [number, number],
		ah?: number,
	): SizeObject {

		if (aw instanceof Size) {
			return { w: aw._w, h: aw._h };
		}

		if (Array.isArray(aw)) {
			const [w, h] = aw;
			return {
				w: typeof w === 'number' ? w : 0,
				h: typeof h === 'number' ? h : 0,
			};
		}

		if (typeof aw === 'object' && aw !== null) {
			return {
				w: typeof aw.w === 'number' ? aw.w : 0,
				h: typeof aw.h === 'number' ? aw.h : 0,
			};
		}

		if (typeof aw === 'number') {
			if (typeof ah === 'number') {
				return { w: aw, h: ah };
			}

			return { w: aw, h: aw };
		}

		return { w: 0, h: 0 };
	}

	//----------------------------------------------------------------------------//
	// Accessors                                                                  //
	//----------------------------------------------------------------------------//

	////////////////////////////////////////////////////////////////////////////////

	public get w(): number {
		return this._w;
	}

	////////////////////////////////////////////////////////////////////////////////

	public get h(): number {
		return this._h;
	}

	//----------------------------------------------------------------------------//
	// Operators                                                                  //
	//----------------------------------------------------------------------------//

	////////////////////////////////////////////////////////////////////////////////

	public add(...args: SizeParams): Size {

		const s = Size.standardize(...args);

		return new Size(
			this._w + s.w,
			this._h + s.h,
		);
	}

	////////////////////////////////////////////////////////////////////////////////

	public sub(...args: SizeParams): Size {

		const s = Size.standardize(...args);

		return new Size(
			this._w - s.w,
			this._h - s.h,
		);
	}

	////////////////////////////////////////////////////////////////////////////////

	public mul(...args: SizeParams): Size {

		const s = Size.standardize(...args);

		return new Size(
			this._w * s.w,
			this._h * s.h,
		);
	}

	////////////////////////////////////////////////////////////////////////////////

	public div(...args: SizeParams): Size {

		const s = Size.standardize(...args);

		return new Size(
			this._w / s.w,
			this._h / s.h,
		);
	}

	////////////////////////////////////////////////////////////////////////////////

	public compare(...args: SizeParams): number {

		const s = Size.standardize(...args);

		// Calculate area of both values
		const areaA = this._w * this._h;
		const areaB = s.w * s.h;

		if (areaA < areaB) {
			return -1;
		}
		if (areaA > areaB) {
			return 1;
		}

		return 0;
	}

	////////////////////////////////////////////////////////////////////////////////

	public lt(...args: SizeParams): boolean {
		return this.compare(...args) < 0;
	}

	////////////////////////////////////////////////////////////////////////////////

	public gt(...args: SizeParams): boolean {
		return this.compare(...args) > 0;
	}

	////////////////////////////////////////////////////////////////////////////////

	public le(...args: SizeParams): boolean {
		return this.compare(...args) <= 0;
	}

	////////////////////////////////////////////////////////////////////////////////

	public ge(...args: SizeParams): boolean {
		return this.compare(...args) >= 0;
	}

	////////////////////////////////////////////////////////////////////////////////

	public eq(...args: SizeParams): boolean {

		const s = Size.standardize(...args);
		return this._w === s.w && this._h === s.h;
	}

	////////////////////////////////////////////////////////////////////////////////

	public ne(...args: SizeParams): boolean {

		const s = Size.standardize(...args);
		return this._w !== s.w || this._h !== s.h;
	}
}
