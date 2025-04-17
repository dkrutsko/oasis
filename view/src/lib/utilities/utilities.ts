//----------------------------------------------------------------------------//
// Functions                                                                  //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////
/// Creates a copy of `obj` with only the specified `keys`.

type Pick<T, K extends keyof T> = {
	[P in K]: T[P];
};

export const pick = function <T extends object, K extends keyof T>(
	obj: T,
	...keys: K[]
): Pick<T, K> {
	const result: Partial<Pick<T, K>> = {};

	if (obj) {
		for (const key of keys) {
			if (key in obj) {
				result[key] = obj[key];
			}
		}
	}

	return result as Pick<T, K>;
};

////////////////////////////////////////////////////////////////////////////////
/// Creates a copy of `obj` without the specified `keys`.

type Omit<T, K extends keyof T> = Pick<T, Exclude<keyof T, K>>;

export const omit = function <T extends object, K extends keyof T>(
	obj: T,
	...keys: K[]
): Omit<T, K> {
	const result = { ...obj };

	for (const key of keys) {
		delete result[key];
	}

	return result as Omit<T, K>;
};

////////////////////////////////////////////////////////////////////////////////
/// Returns true if two arrays are shallowly equal.
/// See: https://github.com/moroshko/shallow-equal

export const shallowCompareArr = function (
	arrA: any[] | null | undefined,
	arrB: any[] | null | undefined,
): boolean {
	if (arrA === arrB) {
		return true;
	}

	if (!arrA || !arrB) {
		return false;
	}

	const len = arrA.length;

	if (arrB.length !== len) {
		return false;
	}

	for (let i = 0; i < len; ++i) {
		if (arrA[i] !== arrB[i]) {
			return false;
		}
	}

	return true;
};

////////////////////////////////////////////////////////////////////////////////
/// Returns true if two objects are shallowly equal.
/// See: https://github.com/moroshko/shallow-equal

export const shallowCompareObj = function (
	objA: Record<string, any> | null | undefined,
	objB: Record<string, any> | null | undefined,
): boolean {
	if (objA === objB) {
		return true;
	}

	if (!objA || !objB) {
		return false;
	}

	const aKeys = Object.keys(objA);
	const bKeys = Object.keys(objB);
	const len = aKeys.length;

	if (bKeys.length !== len) {
		return false;
	}

	for (let i = 0; i < len; ++i) {
		const key = aKeys[i];

		if (objA[key] !== objB[key] || !Object.hasOwn(objB, key)) {
			return false;
		}
	}

	return true;
};
