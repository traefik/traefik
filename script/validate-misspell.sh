#!/usr/bin/env bash

SCRIPT_DIR="$( cd "$( dirname "${0}" )" && pwd -P)"
files=()
while IFS='' read -r line; do files+=("$line"); done < <(git ls-files "${SCRIPT_DIR}"/../'docs/*.md' "${SCRIPT_DIR}"/../*.md | grep -v "CHANGELOG.md")

errors=()
for f in "${files[@]}"; do
	# we use source text here so we also check spelling of variable names
	failedSpell=$(misspell -source=text -i "internetbs" "$f")
	if [ "$failedSpell" ]; then
		errors+=( "$failedSpell" )
	fi
done

if [ ${#errors[@]} -eq 0 ]; then
	echo 'Congratulations!  All Go source files and docs have been checked for common misspellings.'
else
	{
		echo "Errors from misspell:"
		for err in "${errors[@]}"; do
			echo "$err"
		done
		echo
		echo 'Please fix the above errors. You can test via "misspell" and commit the result.'
		echo
	} >&2
	false
fi
