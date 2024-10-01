package files

import "fmt"

const activateSh = `#!/usr/bin/env bash

export PROOFMAN_VENV_ROOT=%s

export OLD_USER_HOME=$USER_HOME
export USER_HOME="$PROOFMAN_VENV_ROOT"

export OLD_PATH="$PATH"
export PATH="$PROOFMAN_VENV_ROOT/bin:$PATH"

export PROOFMAN_VENV_ACTIVE=1
`

func NewActivateSh(venvRootPath string) string {
	return fmt.Sprintf(activateSh, venvRootPath)
}
