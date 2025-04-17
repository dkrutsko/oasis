//----------------------------------------------------------------------------//
// Imports                                                                    //
//----------------------------------------------------------------------------//

import { Size } from './size';

//----------------------------------------------------------------------------//
// Types                                                                      //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////

export interface PointObject {
	x: number;
	y: number;
}

////////////////////////////////////////////////////////////////////////////////

export type PointParams = Parameters<typeof Point.standardize>;
export type PointReturn = ReturnType<typeof Point.standardize>;

//----------------------------------------------------------------------------//
// Point                                                                      //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////

export class Point {
	//----------------------------------------------------------------------------//
	// Constants                                                                  //
	//----------------------------------------------------------------------------//

	////////////////////////////////////////////////////////////////////////////////

	public static readonly Zero = new Point(0, 0);

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

	public constructor(...args: PointParams) {

		const p = Point.standardize(...args);

		this._x = p.x;
		this._y = p.y;
	}

	//----------------------------------------------------------------------------//
	// Methods                                                                    //
	//----------------------------------------------------------------------------//

	////////////////////////////////////////////////////////////////////////////////

	public isZero(): boolean {
		return this._x === 0 && this._y === 0;
	}

	////////////////////////////////////////////////////////////////////////////////

	public setX(x: number): Point {
		return new Point(x, this._y);
	}

	////////////////////////////////////////////////////////////////////////////////

	public setY(y: number): Point {
		return new Point(this._x, y);
	}

	////////////////////////////////////////////////////////////////////////////////

	public toSize(): Size {
		return new Size(this._x, this._y);
	}

	////////////////////////////////////////////////////////////////////////////////

	public toArray(): [number, number] {
		return [
			this._x,
			this._y,
		];
	}

	////////////////////////////////////////////////////////////////////////////////

	public toObject(): PointObject {
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

	public abs(): Point {

		return new Point(
			Math.abs(this._x),
			Math.abs(this._y),
		);
	}

	////////////////////////////////////////////////////////////////////////////////

	public floor(): Point {

		return new Point(
			Math.floor(this._x),
			Math.floor(this._y),
		);
	}

	////////////////////////////////////////////////////////////////////////////////

	public ceil(): Point {

		return new Point(
			Math.ceil(this._x),
			Math.ceil(this._y),
		);
	}

	////////////////////////////////////////////////////////////////////////////////

	public round(): Point {

		return new Point(
			Math.round(this._x),
			Math.round(this._y),
		);
	}

	////////////////////////////////////////////////////////////////////////////////

	public clamp(min: Point, max: Point): Point {

		return new Point(
			Math.max(min.x, Math.min(this._x, max.x)),
			Math.max(min.y, Math.min(this._y, max.y)),
		);
	}

	////////////////////////////////////////////////////////////////////////////////

	public within(size: Size): boolean {

		return (
			this._x >= 0 && this._x < size.w &&
			this._y >= 0 && this._y < size.h
		);
	}

	////////////////////////////////////////////////////////////////////////////////

	public distance(value: Point): number {

		const dx = this._x - value._x;
		const dy = this._y - value._y;

		return Math.sqrt(dx * dx + dy * dy);
	}

	////////////////////////////////////////////////////////////////////////////////

	public distanceSq(value: Point): number {

		const dx = this._x - value._x;
		const dy = this._y - value._y;

		return dx * dx + dy * dy;
	}

	////////////////////////////////////////////////////////////////////////////////

	public manhattanDistance(value: Point): number {

		return Math.abs(this._x - value._x) + Math.abs(this._y - value._y);
	}

	////////////////////////////////////////////////////////////////////////////////

	public rotate(angle: number): Point {

		const sinAngle = Math.sin(angle);
		const cosAngle = Math.cos(angle);

		return new Point(
			this._x * cosAngle - this._y * sinAngle,
			this._x * sinAngle + this._y * cosAngle,
		);
	}

	//----------------------------------------------------------------------------//
	// Static                                                                     //
	//----------------------------------------------------------------------------//

	////////////////////////////////////////////////////////////////////////////////

	public static standardize(
		ax?: number | Point | PointObject | [number, number],
		ay?: number,
	): PointObject {

		if (ax instanceof Point) {
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

	public add(...args: PointParams): Point {

		const p = Point.standardize(...args);

		return new Point(
			this._x + p.x,
			this._y + p.y,
		);
	}

	////////////////////////////////////////////////////////////////////////////////

	public sub(...args: PointParams): Point {

		const p = Point.standardize(...args);

		return new Point(
			this._x - p.x,
			this._y - p.y,
		);
	}

	////////////////////////////////////////////////////////////////////////////////

	public mul(...args: PointParams): Point {

		const p = Point.standardize(...args);

		return new Point(
			this._x * p.x,
			this._y * p.y,
		);
	}

	////////////////////////////////////////////////////////////////////////////////

	public div(...args: PointParams): Point {

		const p = Point.standardize(...args);

		return new Point(
			this._x / p.x,
			this._y / p.y,
		);
	}

	////////////////////////////////////////////////////////////////////////////////

	public neg(): Point {

		return new Point(
			-this._x,
			-this._y,
		);
	}

	////////////////////////////////////////////////////////////////////////////////

	public compare(...args: PointParams): number {

		const p = Point.standardize(...args);

		if (this._x !== p.x) {
			return this._x < p.x ? -1 : 1;
		}

		if (this._y !== p.y) {
			return this._y < p.y ? -1 : 1;
		}

		return 0;
	}

	////////////////////////////////////////////////////////////////////////////////

	public lt(...args: PointParams): boolean {
		return this.compare(...args) < 0;
	}

	////////////////////////////////////////////////////////////////////////////////

	public gt(...args: PointParams): boolean {
		return this.compare(...args) > 0;
	}

	////////////////////////////////////////////////////////////////////////////////

	public le(...args: PointParams): boolean {
		return this.compare(...args) <= 0;
	}

	////////////////////////////////////////////////////////////////////////////////

	public ge(...args: PointParams): boolean {
		return this.compare(...args) >= 0;
	}

	////////////////////////////////////////////////////////////////////////////////

	public eq(...args: PointParams): boolean {

		const p = Point.standardize(...args);
		return this._x === p.x && this._y === p.y;
	}

	////////////////////////////////////////////////////////////////////////////////

	public ne(...args: PointParams): boolean {

		const p = Point.standardize(...args);
		return this._x !== p.x || this._y !== p.y;
	}
}
