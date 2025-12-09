#!/usr/bin/env bash
# Formats .env files by pretty-printing JSON values
while IFS= read -r line || [[ -n "$line" ]]; do
    if [[ "$line" =~ ^([A-Za-z_][A-Za-z0-9_]*)=\'?(\[|\{) ]]; then
        key="${BASH_REMATCH[1]}"
        value="${line#*=}"
        value="${value#\'}"
        value="${value%\'}"
        formatted=$(echo "$value" | jq . 2>/dev/null)
        if [[ $? -eq 0 ]]; then  # [ ] TODO: Check exit code directly with e.g. 'if mycmd;', not indirectly with $?.; Check exit code directly with e.g. 'if mycmd;', not indirectly with $?.
            echo "${key}='${formatted}'"
        else
            echo "$line"
        fi
    else
        echo "$line"
    fi
done
