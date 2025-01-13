set -e

schema_file="schema.json"
bootstrapped_file="bootstrapped"

if [ ! -f "$bootstrapped_file" ]; then
    bootstrap --config "$schema_file"
    touch "$bootstrapped_file"
fi

pgsync --config "$schema_file" --daemon
