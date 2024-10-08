package files

import "fmt"

const activateSh = `# This file must be used with "source bin/activate" *from bash*
# you cannot run it directly

if [ "${BASH_SOURCE-}" = "$0" ]; then
    echo "You must source this script: \$ source $0" >&2
    exit 33
fi

if [ "${PROOFMAN_VENV_ACTIVE-}" = "1" ]; then
    echo "A virtual environment is already active - deactivate it first"
    exit 34
fi

deactivate() {
    if ! [ -z "${_OLD_PATH:+_}" ]; then
        PATH="$_OLD_PATH"
        export PATH
        unset _OLD_PATH
    fi

    unset PROOFMAN_VENV_ROOT
    unset PROOFMAN_VENV_ACTIVE

    hash -r 2>/dev/null

    unset -f deactivate
}

export PROOFMAN_VENV_ROOT=%s

export _OLD_PATH="$PATH"
export PATH="$PROOFMAN_VENV_ROOT/bin:$PATH"

export PROOFMAN_VENV_ACTIVE=1

hash -r 2>/dev/null
`

func NewActivateSh(venvRootPath string) string {
	return fmt.Sprintf(activateSh, venvRootPath)
}
