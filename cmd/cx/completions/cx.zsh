#compdef cx

_cx_contexts() {
    local -a contexts
    contexts=(${(f)"$(cx list 2>/dev/null | grep -E '^\s+[*]?\s' | sed 's/^[* ]*//' | sed 's/ (active)$//')"})
    [[ ${#contexts} -gt 0 ]] && compadd "$@" -a contexts
}

_cx() {
    local curcontext="$curcontext" ret=1
    local -a context state state_descr line
    typeset -A opt_args

    _arguments -C \
        '-v[Verbose output]' \
        '--verbose[Verbose output]' \
        '--debug[Debug output]' \
        ': :->command' \
        ': :->argument' \
        && ret=0

    case "$state" in
        command)
            local -a commands
            commands=(
                'list:List all saved contexts'
                'save:Save current keychain token as a named context'
                'login:Login with a new account and save as context'
                'use:Launch Claude Code with a saved context'
                'delete:Delete a saved context'
                'show:Show token preview'
                'current:Show the currently active context'
                'completion:Print shell completions'
                'version:Show version'
                'help:Show help'
            )
            _describe -t commands 'command' commands && ret=0
            ;;
        argument)
            case "$line[1]" in
                save|login|use|delete|show)
                    _cx_contexts && ret=0
                    ;;
                completion)
                    compadd bash zsh fish && ret=0
                    ;;
            esac
            ;;
    esac

    return ret
}

if [[ "$funcstack[1]" = "_cx" ]]; then
    _cx "$@"
else
    compdef _cx cx
fi
