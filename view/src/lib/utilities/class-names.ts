//----------------------------------------------------------------------------//
// Functions                                                                  //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////
/// Performs conditional joining of class names from a variety of inputs. Falsey
/// values and booleans are always discarded. When an object is supplied, the key
/// is used as the class name if the value is truthy. The result is deduplicated.
/// See: https://github.com/lukeed/clsx & https://github.com/JedWatson/classnames

type ClassValue = ClassValue[] | Record<string, any> | boolean | number | string | null | undefined;

export const classNames = function (...args: ClassValue[]): string {
	// Remove any duplicates
	const result = new Set();

	// Recursively perform parsing directly into result
	const parse = function (...vals: ClassValue[]): void {
		// Loop through arguments
		for (const val of vals) {
			if (!val || typeof val === 'boolean') {
				continue;
			} else if (typeof val === 'number') {
				result.add(val.toString());
			} else if (typeof val === 'string') {
				// Perform split on any whitespace characters
				val.split(/\s/).forEach((v) => result.add(v));
			} else if (Array.isArray(val)) {
				parse(...(val as ClassValue[]));
			} else if (typeof val === 'object') {
				for (const key in val) {
					if (val[key]) {
						result.add(key);
					} else {
						result.delete(key);
					}
				}
			}
		}
	};

	parse(...args);
	return [...result].join(' ');
};
