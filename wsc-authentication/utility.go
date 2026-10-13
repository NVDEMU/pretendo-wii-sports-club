package main

import nex "github.com/PretendoNetwork/nex-go"

func passwordFromPID(pid uint32) (string, uint32) {
	user := getNEXAccountByPID(pid)
	if user == nil {
		return "", nex.Errors.RendezVous.InvalidUsername
	}

	password, ok := user["password"].(string)
	if !ok || password == "" {
		return "", nex.Errors.RendezVous.InvalidUsername
	}

	return password, 0
}
