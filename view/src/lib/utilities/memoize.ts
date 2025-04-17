//----------------------------------------------------------------------------//
// Functions                                                                  //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////

export const memoizeOne = function <T>(fn: () => T): () => T {
	// TODO: Add support for parameters?
	let cached = false;
	let result: T;

	return function () {
		if (!cached) {
			cached = true;
			result = fn();
		}

		return result;
	};
};

////////////////////////////////////////////////////////////////////////////////

export const memoizeOneAsync = function <T>(fn: () => Promise<T>): () => Promise<T> {
	let cached = false;
	let result: Promise<T>;

	return async function () {
		if (!cached) {
			cached = true;
			result = fn();
		}

		return result;
	};
};
