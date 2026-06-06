_quest_capture() {
    local last_cmd
    last_cmd=$(history 1 | sed 's/^[[:space:]]*[0-9]*[[:space:]]*//')

    if [ -z "$last_cmd" ] || [[ "$last_cmd" == quest* ]] || [[ "$last_cmd" == "q" ]]; then
        return
    fi

    (quest add "$last_cmd" &>/dev/null & disown)
}

if [ -z "$PROMPT_COMMAND" ]; then
    PROMPT_COMMAND="_quest_capture"
else
    PROMPT_COMMAND="${PROMPT_COMMAND}; _quest_capture"
fi

q() {
    if [ $# -eq 0 ]; then
        tmp=$(mktemp)
        quest ui > "$tmp"
        cmd=$(cat "$tmp")
        rm -f "$tmp"
    else
        cmd=$(quest "$@")
    fi

    if [ -n "$cmd" ]; then
        echo "running: $cmd"
        history -s "$cmd"
        eval "$cmd"
    fi
}
