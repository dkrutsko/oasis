//----------------------------------------------------------------------------//
// Exports                                                                    //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////
/// Represents a set of customizable options for different screen breakpoints.

export interface BreakpointType<T> {
	xs: T;
	sm: T;
	md: T;
	lg: T;
	xl: T;
}

////////////////////////////////////////////////////////////////////////////////
/// Standardizes an input into a full `BreakpointType<T>` object, filling
/// any undefined properties with the smallest defined breakpoint value or,
/// if no properties are defined, the singular input value.

export const normalizeBreakpoint = function <T>(
	bp: Partial<BreakpointType<T>> | T,
): BreakpointType<T> {
	if (bp && typeof bp === 'object') {
		if ('xs' in bp || 'sm' in bp || 'md' in bp || 'lg' in bp || 'xl' in bp) {
			const xs = (bp.xs ?? bp.sm ?? bp.md ?? bp.lg ?? bp.xl)!;
			const sm = (bp.sm ?? xs)!;
			const md = (bp.md ?? sm)!;
			const lg = (bp.lg ?? md)!;
			const xl = (bp.xl ?? lg)!;

			return { xs, sm, md, lg, xl };
		}
	}

	return {
		xs: bp as T,
		sm: bp as T,
		md: bp as T,
		lg: bp as T,
		xl: bp as T,
	};
};
