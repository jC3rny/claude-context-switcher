_cx() {
    local cur prev commands
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"
    commands="list save login use delete show current completion version help"

    case "$prev" in
        cx)
            mapfile -t COMPREPLY < <(compgen -W "$commands -v --verbose --debug" -- "$cur")
            return
            ;;
        -v|--verbose|--debug)
            mapfile -t COMPREPLY < <(compgen -W "$commands" -- "$cur")
            return
            ;;
        save|login|use|delete|show)
            local contexts
            contexts=$(cx list 2>/dev/null | grep -E '^\s+[*]?\s' | sed 's/^[* ]*//' | sed 's/ (active)$//')
            mapfile -t COMPREPLY < <(compgen -W "$contexts" -- "$cur")
            return
            ;;
        completion)
            mapfile -t COMPREPLY < <(compgen -W "bash zsh fish" -- "$cur")
            return
            ;;
    esac
}

complete -F _cx cx
