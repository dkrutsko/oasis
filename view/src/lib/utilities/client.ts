//----------------------------------------------------------------------------//
// Imports                                                                    //
//----------------------------------------------------------------------------//

import { memoizeOne } from './memoize';

//----------------------------------------------------------------------------//
// Functions                                                                  //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////
/// Returns the size, in pixels, of the native system scrollbars in the browser
/// The first call will perform the calculation, subsequent calls will be cached.

export const getScrollbarSize = memoizeOne(function (): number {
	// Create a new element to be added to the DOM
	const element = document.createElement('div');

	// Add styles to render scrollbars
	element.style.width = '100px';
	element.style.height = '100px';
	element.style.left = '-9999px';
	element.style.top = '-9999px';
	element.style.position = 'absolute';
	element.style.overflow = 'scroll';

	// Append to DOM, measure and remove
	document.body.appendChild(element);
	const result = element.offsetWidth - element.clientWidth || 0;
	document.body.removeChild(element);

	return result;
});

////////////////////////////////////////////////////////////////////////////////
/// Returns the version of WebGL currently supported or zero when not supported
/// The first call will perform the calculation, subsequent calls will be cached.

export const isWebGLAvailable = memoizeOne(function (): number {
	try {
		// Look for WebGL2 rendering context
		if (window.WebGL2RenderingContext) {
			// Attempt to create a canvas and get a context
			const canvas = document.createElement('canvas');

			if (canvas.getContext('webgl2')) {
				return 2;
			}
		}
	} catch (e) {}

	try {
		// Look for WebGL rendering context
		if (window.WebGLRenderingContext) {
			// Attempt to create a canvas and get a context
			const canvas = document.createElement('canvas');

			if (canvas.getContext('webgl')) {
				return 1;
			}
		}
	} catch (e) {}

	return 0;
});

////////////////////////////////////////////////////////////////////////////////
/// Checks whether the current browser supports touch events.

export const hasTouch = memoizeOne(function (): boolean {
	return 'ontouchstart' in document.documentElement || navigator.maxTouchPoints > 0;
});
