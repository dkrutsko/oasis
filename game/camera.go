package game

import (
	"strconv"
	"time"

	"github.com/dkrutsko/oasis/math"
)

////////////////////////////////////////////////////////////////////////////////

type CameraResult uint8

const (
	CameraResultSuccess CameraResult = 0x00
	CameraResultNoValue CameraResult = 0x01

	CameraResultNoProcess CameraResult = 0x10
	CameraResultNoOffset  CameraResult = 0x11
	CameraResultReadFail  CameraResult = 0x12
)

////////////////////////////////////////////////////////////////////////////////

func (s CameraResult) String() string {

	switch s {
	case CameraResultSuccess:
		return "Success"
	case CameraResultNoValue:
		return "NoValue"

	case CameraResultNoProcess:
		return "NoProcess"
	case CameraResultNoOffset:
		return "NoOffset"
	case CameraResultReadFail:
		return "ReadFail"
	}

	return strconv.FormatUint(uint64(s), 10)
}

////////////////////////////////////////////////////////////////////////////////

type CameraState struct {
	Result CameraResult
	Time   time.Time
	View   math.Matrix
}

////////////////////////////////////////////////////////////////////////////////

func NewCameraState() *CameraState {

	return &CameraState{
		Result: CameraResultNoValue,
		Time:   time.Now(),
		View:   math.MatrixIdentity,
	}
}

////////////////////////////////////////////////////////////////////////////////

func (g *Game) updateCamera(scanner *ScannerState) *CameraState {

	//----------------------------------------------------------------------------//

	// Create the final result
	result := NewCameraState()

	//----------------------------------------------------------------------------//

	var err error
	// Check if a game has been selected and is currently valid
	if scanner == nil || scanner.Result != ScannerResultSuccess {
		result.Result = CameraResultNoProcess
		return result
	}

	// Grab client module base address
	client := scanner.Client.GetBase()

	// Create a memory object for reading
	memory := scanner.Process.GetMemory()

	//----------------------------------------------------------------------------//

	offViewMatrix, ok := g.GetOffsetsInt("offsets", "client.dll", "dwViewMatrix")
	if !ok {
		result.Result = CameraResultNoOffset
		return result
	}

	result.View, err = ReadMatrix(memory, client+offViewMatrix)
	if err != nil {
		result.Result = CameraResultReadFail
		return result
	}

	//----------------------------------------------------------------------------//

	// Specify that scan was successful
	result.Result = CameraResultSuccess
	return result

	//----------------------------------------------------------------------------//
}
