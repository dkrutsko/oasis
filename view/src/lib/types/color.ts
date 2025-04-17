//----------------------------------------------------------------------------//
// Imports                                                                    //
//----------------------------------------------------------------------------//

import { colord, extend, Colord, type AnyColor } from 'colord';

import colordA11y from 'colord/plugins/a11y';
import colordHwb from 'colord/plugins/hwb';
import colordMix from 'colord/plugins/mix';
import colordNames from 'colord/plugins/names';

import { kebabCase } from 'case-anything';

extend([colordA11y, colordHwb, colordMix, colordNames]);

//----------------------------------------------------------------------------//
// Exports                                                                    //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////
/// Used when expecting some color property.

export type ColorType = string | Colord;

////////////////////////////////////////////////////////////////////////////////
/// Used when expecting some paint property.

export type PaintType = Record<string, ColorType>;

////////////////////////////////////////////////////////////////////////////////
/// Returns colord module with additional features consistent with PostCSS.

export { colord, Colord, type AnyColor };

////////////////////////////////////////////////////////////////////////////////
/// Defines standard black and white colors.

export const black = colord('#000000');
export const white = colord('#FFFFFF');

////////////////////////////////////////////////////////////////////////////////
/// Returns black or white, depending on best visibility.

export const contrast = function (color: ColorType): Colord {
	// WCAG AA spec requires a score of at least 4.5 or greater
	return colord(color).contrast(white) >= 4.5 ? white : black;
};

////////////////////////////////////////////////////////////////////////////////
/// Expands an object with multiple ColorType properties into CSS variable
/// declarations which can be placed directly inside HTML style attributes.

export const expandPaint = function (paint: PaintType): string {
	const result: string[] = [];

	// Iterate through keys
	for (const key in paint) {
		// Use proper variable name
		const name = kebabCase(key);

		// Convert the color into string
		const color = colord(paint[key]).toRgbString();

		// Add property to the result
		result.push(`--${name}: ${color};`);
	}

	return result.join(' ');
};

////////////////////////////////////////////////////////////////////////////////
/// Parses the specified `color` and returns an array representing the red,
/// green, blue, and alpha channels. All four channels will be normalized to
/// [0.0, 1.0]. Useful when a color needs to be converted for a WebGL shader.

export const normalizeRgba = function (color: ColorType): [number, number, number, number] {
	const c = colord(color).toRgb();
	return [c.r / 255, c.g / 255, c.b / 255, c.a];
};

////////////////////////////////////////////////////////////////////////////////
/// Parses the specified `color` and returns an array representing the red,
/// green, and blue channels. All three channels will be normalized to
/// [0.0, 1.0]. Useful when a color needs to be converted for a WebGL shader.

export const normalizeRgb = function (color: ColorType): [number, number, number] {
	const c = colord(color).toRgb();
	return [c.r / 255, c.g / 255, c.b / 255];
};
