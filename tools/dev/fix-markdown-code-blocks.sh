#!/bin/bash
# Fix markdown code blocks without language specification

input_file="$1"
temp_file=$(mktemp)

prev_line=""
while IFS= read -r line; do
    # If this is a bare opening code fence and previous line is "**Expected Output:**"
    if [[ "$line" == '```' ]] && [[ "$prev_line" == "**Expected Output:**" ]]; then
        echo '```text' >> "$temp_file"
    else
        echo "$line" >> "$temp_file"
    fi
    prev_line="$line"
done < "$input_file"

# Replace original file
mv "$temp_file" "$input_file"

echo "Fixed Expected Output code blocks in $input_file"
