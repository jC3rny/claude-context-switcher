function __cx_contexts
    cx list 2>/dev/null | string match -r '^\s+[*]?\s+\S+' | string trim | string replace -r ' \(active\)$' ''
end

complete -c cx -n '__fish_use_subcommand' -a list -d 'List all saved contexts'
complete -c cx -n '__fish_use_subcommand' -a save -d 'Save current keychain token as a named context'
complete -c cx -n '__fish_use_subcommand' -a login -d 'Login with a new account and save as context'
complete -c cx -n '__fish_use_subcommand' -a use -d 'Launch Claude Code with a saved context'
complete -c cx -n '__fish_use_subcommand' -a delete -d 'Delete a saved context'
complete -c cx -n '__fish_use_subcommand' -a show -d 'Show token preview'
complete -c cx -n '__fish_use_subcommand' -a current -d 'Show the currently active context'
complete -c cx -n '__fish_use_subcommand' -a version -d 'Show version'
complete -c cx -n '__fish_use_subcommand' -a help -d 'Show help'
complete -c cx -n '__fish_use_subcommand' -s v -l verbose -d 'Verbose output'
complete -c cx -n '__fish_use_subcommand' -l debug -d 'Debug output'

complete -c cx -n '__fish_seen_subcommand_from save login use delete show' -a '(__cx_contexts)' -d 'Context'
