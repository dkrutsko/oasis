package game

////////////////////////////////////////////////////////////////////////////////

type ScannerResult uint8

const (
	ScannerResultSuccess ScannerResult = 0x00
	ScannerResultNoValue ScannerResult = 0x01
)

////////////////////////////////////////////////////////////////////////////////

type ScannerState struct {
	Result ScannerResult

	PID uint32

	EngineBase uint64
	EngineSize uint64

	ClientBase uint64
	ClientSize uint64
}

////////////////////////////////////////////////////////////////////////////////

func NewScannerState() *ScannerState {

	return &ScannerState{
		Result: ScannerResultNoValue,

		PID: 0,

		EngineBase: 0x0,
		EngineSize: 0x0,

		ClientBase: 0x0,
		ClientSize: 0x0,
	}
}

////////////////////////////////////////////////////////////////////////////////

func (s *ScannerState) Clone() *ScannerState {

	return &ScannerState{
		Result: s.Result,

		PID: s.PID,

		EngineBase: s.EngineBase,
		EngineSize: s.EngineSize,

		ClientBase: s.ClientBase,
		ClientSize: s.ClientSize,
	}
}

////////////////////////////////////////////////////////////////////////////////

func (s *ScannerState) Changed(val *ScannerState) bool {

	// Check for nil
	if val == nil {
		return true
	}

	// If PID is the same
	if val.PID == s.PID {
		return false
	}

	return true
}

////////////////////////////////////////////////////////////////////////////////

func (g *Game) updateScanner(prev *ScannerState) *ScannerState {

	//----------------------------------------------------------------------------//

	// Create the final result
	result := NewScannerState()

	//----------------------------------------------------------------------------//

	// Specify that scan was successful
	//result.Result = ScannerResultSuccess
	return result

	//----------------------------------------------------------------------------//
}
