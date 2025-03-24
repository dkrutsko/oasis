package game

import (
	"regexp"
	"strconv"

	"github.com/dkrutsko/oasis/leech"
)

////////////////////////////////////////////////////////////////////////////////

// Create regex for parsing process name
var GameProcessName = regexp.MustCompile(
	"(?i)^cs2\\.exe$",
)

////////////////////////////////////////////////////////////////////////////////

type ScannerResult uint8

const (
	ScannerResultSuccess ScannerResult = 0x00
	ScannerResultNoValue ScannerResult = 0x01

	ScannerResultProcessListError    ScannerResult = 0x10
	ScannerResultProcessListEmpty    ScannerResult = 0x11
	ScannerResultProcessListMultiple ScannerResult = 0x12

	ScannerResultModuleListError    ScannerResult = 0x20
	ScannerResultModuleListEmpty    ScannerResult = 0x21
	ScannerResultModuleListNoEngine ScannerResult = 0x22
	ScannerResultModuleListNoClient ScannerResult = 0x23
)

////////////////////////////////////////////////////////////////////////////////

func (s ScannerResult) String() string {

	switch s {
		case ScannerResultSuccess:
			return "Success"
		case ScannerResultNoValue:
			return "NoValue"

		case ScannerResultProcessListError:
			return "ProcessListError"
		case ScannerResultProcessListEmpty:
			return "ProcessListEmpty"
		case ScannerResultProcessListMultiple:
			return "ProcessListMultiple"

		case ScannerResultModuleListError:
			return "ModuleListError"
		case ScannerResultModuleListEmpty:
			return "ModuleListEmpty"
		case ScannerResultModuleListNoEngine:
			return "ModuleListNoEngine"
		case ScannerResultModuleListNoClient:
			return "ModuleListNoClient"
	}

	return strconv.FormatUint(uint64(s), 10)
}

////////////////////////////////////////////////////////////////////////////////

type ScannerState struct {
	Result  ScannerResult

	Process *leech.Process
	Engine  *leech.Module
	Client  *leech.Module

	Camera *leech.Memory
	Action *leech.Memory
}

////////////////////////////////////////////////////////////////////////////////

func NewScannerState() *ScannerState {

	return &ScannerState{
		Result: ScannerResultNoValue,

		Process: nil,
		Engine:  nil,
		Client:  nil,

		Camera: nil,
		Action: nil,
	}
}

////////////////////////////////////////////////////////////////////////////////

func (s *ScannerState) Clone() *ScannerState {

	return &ScannerState{
		Result: s.Result,

		Process: s.Process,
		Engine:  s.Engine,
		Client:  s.Client,

		Camera: s.Camera,
		Action: s.Action,
	}
}

////////////////////////////////////////////////////////////////////////////////

func (s *ScannerState) Changed(val *ScannerState) bool {

	// Check for nil
	if val == nil {
		return true
	}

	if s.Result != ScannerResultSuccess && val.Result != ScannerResultSuccess {
		return false
	}

	if s.Result != ScannerResultSuccess && val.Result == ScannerResultSuccess {
		return true
	}

	if s.Result == ScannerResultSuccess && val.Result != ScannerResultSuccess {
		return true
	}

	if s.Result == ScannerResultSuccess && val.Result == ScannerResultSuccess {

		// Check if another process has been attached to
		return s.Process.GetPid() != val.Process.GetPid()
	}

	return true
}

////////////////////////////////////////////////////////////////////////////////

func (g *Game) updateScanner(prev *ScannerState) *ScannerState {

	//----------------------------------------------------------------------------//

	// Check if previous scanner already selected the game
	if prev != nil && prev.Result == ScannerResultSuccess {

		// If process is still running
		if !prev.Process.HasExited() {
			return prev
		}
	}

	//----------------------------------------------------------------------------//

	// Create the final result
	result := NewScannerState()

	//----------------------------------------------------------------------------//

	{
		// Attempt to list all the relevant game processes
		processList, err := g.options.Leech.GetProcessList(GameProcessName, true)
		if err != nil {
			result.Result = ScannerResultProcessListError
			return result
		}

		// Check if list is empty
		if len(processList) == 0 {
			result.Result = ScannerResultProcessListEmpty
			return result
		}

		// Check for one process
		if len(processList) >= 2 {
			result.Result = ScannerResultProcessListMultiple
			return result
		}

		result.Process = processList[0]
	}

	//----------------------------------------------------------------------------//

	{
		// Attempt to list all the modules from the game
		moduleList, err := result.Process.GetModules(nil)
		if err != nil {
			result.Result = ScannerResultModuleListError
			return result
		}

		// Check if list is empty
		if len(moduleList) == 0 {
			result.Result = ScannerResultModuleListEmpty
			return result
		}

		// Iterate through all the modules
		for _, module := range moduleList {

			// Check if current module is engine2
			if module.GetName() == "engine2.dll" {

				result.Engine = module

				// Break if found modules
				if result.Client != nil {
					break
				}

				continue
			}

			// Check if current module is client
			if module.GetName() == "client.dll" {

				result.Client = module

				// Break if found modules
				if result.Engine != nil {
					break
				}

				continue
			}
		}

		// Missing engine module
		if result.Engine == nil {
			result.Result = ScannerResultModuleListNoEngine
			return result
		}

		// Missing client module
		if result.Client == nil {
			result.Result = ScannerResultModuleListNoClient
			return result
		}
	}

	//----------------------------------------------------------------------------//

	// Retrieve the memory for all the readers
	result.Camera = result.Process.GetMemory()
	result.Action = result.Process.GetMemory()

	//----------------------------------------------------------------------------//

	// Specify that scan was successful
	result.Result = ScannerResultSuccess
	return result

	//----------------------------------------------------------------------------//
}
