//----------------------------------------------------------------------------//
// Imports                                                                    //
//----------------------------------------------------------------------------//

// NO IMPORTS ALLOWED DUE TO INCLUSION IN POSTCSS CONFIG

//----------------------------------------------------------------------------//
// Constants                                                                  //
//----------------------------------------------------------------------------//

export const E = 2.718281828459045;

export const PI = 3.141592653589793;
export const PI_OVER_2 = 1.5707963267948966;
export const PI_OVER_4 = 0.7853981633974483;
export const TWO_PI = 6.283185307179586;

export const SQRT_2 = 1.414213562373095;
export const SQRT_OVER_2 = 0.7071067811865475;

//----------------------------------------------------------------------------//
// Functions                                                                  //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////
/// Returns the remainder from the specified floating-point division.

export const mod = function (dividend: number, divisor: number): number {
	// Can't divide by 0
	if (divisor === 0) {
		return NaN;
	}

	return dividend - divisor * Math.floor(dividend / divisor);
};

////////////////////////////////////////////////////////////////////////////////
/// Converts `value` from radians to degrees.

export const toDeg = function (radians: number): number {
	return radians * (180 / PI);
};

////////////////////////////////////////////////////////////////////////////////
/// Converts `value` from degrees to radians.

export const toRad = function (degrees: number): number {
	return degrees * (PI / 180);
};

////////////////////////////////////////////////////////////////////////////////
/// Wraps `angle` to a value between [-PI, PI].

export const wrapPI = function (angle: number): number {
	return mod(angle + PI, TWO_PI) - PI;
};

////////////////////////////////////////////////////////////////////////////////
/// Wraps `angle` to a value between [0, 2PI].

export const wrapTwoPI = function (angle: number): number {
	return mod(angle, TWO_PI);
};

////////////////////////////////////////////////////////////////////////////////
/// Wraps `angle` to a value between [-180, 180].

export const wrap180 = function (angle: number): number {
	return mod(angle + 180, 360) - 180;
};

////////////////////////////////////////////////////////////////////////////////
/// Wraps `angle` to a value between [0, 360].

export const wrap360 = function (angle: number): number {
	return mod(angle, 360);
};

////////////////////////////////////////////////////////////////////////////////
/// Calculates the absolute value of the difference between two numbers.

export const distance = function (value1: number, value2: number): number {
	const res = value1 - value2;
	return res < 0 ? -res : res;
};

////////////////////////////////////////////////////////////////////////////////
/// Restricts `value` to be within the range [`min`, `max`].

export const clamp = function (value: number, min: number, max: number): number {
	value = value > max ? max : value;
	value = value < min ? min : value;
	return value;
};

////////////////////////////////////////////////////////////////////////////////
/// Returns `number` rounded to `precision`.

export const roundN = function (number: number, precision: number = 15): number {
	const factor = 10 ** precision;
	// Perform rounding to the specified factor
	return Math.round(number * factor) / factor;
};

////////////////////////////////////////////////////////////////////////////////
/// Scales `value` linearly to `min` and `max`.

export const scale = function (value: number, min: number, max: number): number {
	return (value - min) / (max - min);
};

////////////////////////////////////////////////////////////////////////////////
/// Performs a linear interpolation between two values.

export const lerp = function (min: number, max: number, weight: number): number {
	return min + (max - min) * weight;
};

////////////////////////////////////////////////////////////////////////////////
/// Performs a cubic interpolation between two values.

export const smoothstep = function (min: number, max: number, weight: number): number {
	// Scale weight linearly to min and max
	const n = (weight - min) / (max - min);

	// Perform `max (0, min (1, n))` operation
	const amount = n < 0 ? 0 : n > 1 ? 1 : n;

	// Perform the main smooth step operation
	return amount * amount * (3 - 2 * amount);
};
