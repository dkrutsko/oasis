//----------------------------------------------------------------------------//
// Imports                                                                    //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////

import { roundN } from '$lib/math';

//----------------------------------------------------------------------------//
// Exports                                                                    //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////
/// Used when expecting some length property.

export type LengthType = number | string;

////////////////////////////////////////////////////////////////////////////////
/// Parses specified length and returns a string with a default unit of `px`.

export const parseLength = function (length: LengthType): string {
	return typeof length === 'number' ? `${length}px` : length;
};

////////////////////////////////////////////////////////////////////////////////
/// Dynamic version of the RFS rescaling function for CSS properties with units
/// such as font-size, margin, padding, and border-radius. RFS (Responsive Font
/// Sizes) is a unit resizing engine which automatically calculates the appropriate
/// values based on the dimensions of the browser viewport or container query.
/// CSS variables are not supported, the only allowed values are pixels or rems.
/// This function will only operate on a single qualifying LengthType at a time.
/// See: https://github.com/twbs/rfs

interface CreateRfsOptions {
	////////////////////////////////////////////////////////////////////////////////
	/// Whether to calculate LengthType as though it was part of a Container Query.

	cq: boolean;

	////////////////////////////////////////////////////////////////////////////////
	/// Prevents the value from becoming too small on smaller screens. If the font
	/// size which is passed to RFS is smaller than this value, no fluid rescaling
	/// will take place. This value is represented in pixels.

	baseValue: number;

	////////////////////////////////////////////////////////////////////////////////
	/// Above this breakpoint, the value will be equal to the value you passed to
	/// RFS; below the breakpoint, the value will dynamically scale. This value is
	/// represented in pixels.

	breakpoint: number;

	////////////////////////////////////////////////////////////////////////////////
	/// Determines the strength of font size resizing. The higher the factor, the
	/// less difference there is between values on small screens. The lower the
	/// factor, the less influence RFS has, which results in bigger values for small
	/// screens. The factor must be greater than 1.

	factor: number;

	////////////////////////////////////////////////////////////////////////////////
	/// Represents the value of a single rem unit in pixels (e.g. 1rem = 16px).

	remValue: number;

	////////////////////////////////////////////////////////////////////////////////
	/// Specifies the precision to round off the decimal numbers. Uses the precision
	/// defined by the roundN function by default.

	precision: number;
}

export const createRfs = function (
	length: LengthType,
	options: Partial<CreateRfsOptions> = {},
): { fixed: string; fluid: string } {
	const {
		cq = false,
		baseValue = 32,
		breakpoint = 1280,
		factor = 10,
		remValue = 16,
		precision,
	} = options;

	// Convert to a string if needed
	const size = parseLength(length);

	const result = {
		fixed: size,
		fluid: size,
	};

	// Convert size into a number
	let value = parseFloat(size);
	if (isNaN(value)) {
		return result;
	}

	// Check if size is in rems
	if (size.endsWith('rem')) {
		value *= remValue;
	} else {
		// Should at least be pixels
		if (!size.endsWith('px')) {
			return result;
		}
	}

	// Calculate the final value for the fixed result
	const fixed = roundN(value / remValue, precision);
	result.fixed = `${fixed}${fixed === 0 ? '' : 'rem'}`;

	const abs = Math.abs(value);
	// Whether to calculate
	if (abs <= baseValue) {
		result.fluid = result.fixed;
		return result;
	}

	// Targeting container queries
	const dim = cq ? 'cqw' : 'vw';

	// Calculate the base and difference
	const base = baseValue + (abs - baseValue) / factor;
	const diff = abs - base;

	if (value < 0) {
		result.fluid = `calc(-${roundN(base / remValue, precision)}rem - ${roundN(
			(diff * 100) / breakpoint,
			precision,
		)}${dim})`;
	} else {
		result.fluid = `calc(${roundN(base / remValue, precision)}rem + ${roundN(
			(diff * 100) / breakpoint,
			precision,
		)}${dim})`;
	}

	return result;
};
