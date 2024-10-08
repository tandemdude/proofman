package files

const IsabelleProxyScript = `#!/usr/bin/env bash

HOME=$PROOFMAN_VENV_ROOT $PROOFMAN_VENV_ROOT/bin/_proxied/isabelle "$@"
`
const IsabelleJavaProxyScript = `#!/usr/bin/env bash

HOME=$PROOFMAN_VENV_ROOT $PROOFMAN_VENV_ROOT/bin/_proxied/isabelle_java "$@"
`

const IsabelleHomeUserSettings = `# -*- shell-script -*- :mode=shellscript:

isabelle_directory "$PROOFMAN_VENV_ROOT/deps"
`
