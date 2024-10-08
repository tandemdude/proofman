package internal

const (
	LogLvlQuiet = iota
	LogLvlUnquiet
	LogLvlVerbose
)

var LogLevel = LogLvlUnquiet

var ProofbankBaseUrl = "" // TODO - change this to a sane default
var ProofbankApiToken = ""
