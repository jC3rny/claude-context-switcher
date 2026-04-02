#compdef cx

_cx_contexts() {
    local contexts
    contexts=(${(f)"$(cx list 2>/dev/null | grep -E '^\s+[*]?\s' | sed 's/^[* ]*//' | sed 's/ (active)$//')"})
    compadd -a contexts
}

_cx() {
    local -a commands flags
    commands=(
        'list:List all saved contexts'
        'save:Save current keychain token as a named context'
        'login:Login with a new account and save as context'
        'use:Launch Claude Code with a saved context'
        'delete:Delete a saved context'
        'show:Show token preview'
        'current:Show the currently active context'
        'version:Show version'
        'help:Show help'
    )
    flags=(
        '-v[Verbose output]'
        '--verbose[Verbose output]'
        '--debug[Debug output]'
    )

    _arguments -C \
        '1:command:->cmd' \
        '2:context:->ctx' \
        '*::flags:->flags'

    case "$state" in
        cmd)
            _describe 'command' commands
            _values 'flags' $flags
            ;;
        ctx)
            case "${words[2]}" in
                save|login|use|delete|show)
                    _cx_contexts
                    ;;
            esac
            ;;
    esac
}

_cx "$@"
