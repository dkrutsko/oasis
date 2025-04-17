//----------------------------------------------------------------------------//
// Functions                                                                  //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////
/// Performs freezing recursively on objects, maps and sets.
/// See: https://github.com/christophehurpeau/deep-freeze-es6

export const deepFreeze = function <T>(obj: T): Readonly<T> {
	if (!obj || Object.isFrozen(obj)) {
		return obj as Readonly<T>;
	}

	// Skip unless object or function
	if (typeof obj !== 'object' && typeof obj !== 'function') {
		return obj as Readonly<T>;
	}

	// Account for Map values
	if (obj instanceof Map) {
		obj.clear =
			// @ts-expect-error
			obj.delete =
			// @ts-expect-error
			obj.set =
				function (): void {
					throw new Error('map is read-only');
				};
	}

	// Account for Set values
	else if (obj instanceof Set) {
		obj.clear =
			// @ts-expect-error
			obj.delete =
			// @ts-expect-error
			obj.add =
				function (): void {
					throw new Error('set is read-only');
				};
	}

	// Freeze the object
	Object.freeze(obj);

	// Cache whether the object parameter is a function
	const objectIsFunction = typeof obj === 'function';

	// Iterate through all property names which it owns
	for (const key of Object.getOwnPropertyNames(obj)) {
		if (objectIsFunction) {
			if (key === 'caller' || key === 'callee' || key === 'arguments') {
				continue;
			}
		}

		// @ts-expect-error
		deepFreeze(obj[key]);
	}

	return obj as Readonly<T>;
};
