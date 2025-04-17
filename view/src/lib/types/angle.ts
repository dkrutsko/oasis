//----------------------------------------------------------------------------//
// Exports                                                                    //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////
/// Used when expecting some angle property.

export type AngleType = number | string;

////////////////////////////////////////////////////////////////////////////////
/// Parses specified angle and returns a string with a default unit of `deg`.

export const parseAngle = function (angle: AngleType): string {
	return typeof angle === 'number' ? `${angle}deg` : angle;
};
