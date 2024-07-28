#!/usr/bin/env bash

set -eu -o pipefail

csv_to_markdown() {
    awk 'function percentage(missed, covered,   total) {
            total = missed + covered
            if (total == 0) {
                return "-"
            }
            return sprintf("%d", covered / total * 100)
        }

        BEGIN {
            missed_instructions = 0
            covered_instructions = 0
            missed_branches = 0
            covered_branches = 0
            missed_lines = 0
            covered_lines = 0
            missed_methods = 0
            covered_methods = 0

            printf("<details>\n<summary>Details</summary>\n\n")
            print("| Class | % Instruction | % Branch | % Line | % Method |")
            print("| --- | ---: | ---: | ---: | ---: |")
        }

        {
            split($0, cols, ",")

            instructions = percentage(cols[4], cols[5])
            missed_instructions += cols[4]
            covered_instructions += cols[5]

            branches = percentage(cols[6], cols[7])
            missed_branches += cols[6]
            covered_branches += cols[7]

            lines = percentage(cols[8], cols[9])
            missed_lines += cols[8]
            covered_lines += cols[9]

            methods = percentage(cols[12], cols[13])
            missed_methods += cols[12]
            covered_methods += cols[13]

            printf("| %s.%s | %s | %s | %s | %s |\n",
                cols[2], cols[3], instructions, branches, lines, methods)
        }

        END {
            instructions = percentage(missed_instructions, covered_instructions)
            branches = percentage(missed_branches, covered_branches)
            lines = percentage(missed_lines, covered_lines)
            methods = percentage(missed_methods, covered_methods)

            printf("</details>\n\n")
            print("| | % Instruction | % Branch | % Line | % Method |")
            print("| --- | ---: | ---: | ---: | ---: |")
            printf("| **All classes** | %s | %s | %s | %s |\n",
                instructions, branches, lines, methods)
        }'
}

if [[ $# == 0 ]] || [[ $1 == - ]]; then
    tail -n +2 | sort | csv_to_markdown
else
    ( tail -n +2 | sort | csv_to_markdown ) < "$1"
fi
