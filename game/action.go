package game

////////////////////////////////////////////////////////////////////////////////

func (g *Game) updateAction(scanner *ScannerState) {

	// Check if a game has been selected and is currently valid
	if scanner == nil || scanner.Result != ScannerResultSuccess {
		return
	}

	// NYI: return state
}
