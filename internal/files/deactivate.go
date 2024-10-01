package files

const DeactivateSh = `#!/usr/bin/env bash

export USER_HOME="$OLD_USER_HOME"
export PATH="$OLD_PATH"

unset OLD_USER_HOME
unset OLD_PATH
unset PROOFMAN_VENV_ROOT
unset PROOFMAN_VENV_ACTIVE
`
