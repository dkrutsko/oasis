package game

import (
	"strconv"
	"time"

	"github.com/dkrutsko/oasis/leech"
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
	View   math.Matrix4
}

////////////////////////////////////////////////////////////////////////////////

func NewCameraState() *CameraState {

	return &CameraState{
		Result: CameraResultNoValue,
		Time:   time.Now(),
		View:   math.Matrix4Identity,
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

	// Use the persistent memory instance for camera reads
	memory := g.cameraMemory
	if memory == nil {
		memory = scanner.Process.GetMemory()
	}

	// Refresh the camera VMM's TLB cache every 5 seconds.
	// In dual FPGA mode, each VMM has its own TLB cache.
	now := time.Now()
	if now.Sub(g.lastCameraTlbRefresh) >= 5*time.Second {
		g.lastCameraTlbRefresh = now
		cameraLeech := g.options.CameraLeech
		if cameraLeech == nil {
			cameraLeech = g.options.Leech
		}
		cameraLeech.SetConfig(leech.ConfigRefreshFreqTlb, 1)
	}

	//----------------------------------------------------------------------------//

	offViewMatrix, ok := g.GetOffsetsInt("offsets", "client.dll", "dwViewMatrix")
	if !ok {
		result.Result = CameraResultNoOffset
		return result
	}

	result.View, err = ReadMatrix4(memory, client+offViewMatrix)
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
