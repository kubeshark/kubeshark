#!/bin/bash
paths_arr=( "$@" )

printf "\n========== List modified files ==========\n"
echo "$(git diff --name-only HEAD^ HEAD)"

printf "\n========== List paths to match and check existence ==========\n"
for path in ${paths_arr[*]}
do
  if [ -f "$path" ] || [ -d "$path" ]; then
      echo "$path - found"
  else
      echo "$path - does not found - exiting with failure"
      exit 1
  fi
done

printf "\n========== Check paths of modified files ==========\n"
git diff --name-only HEAD^ HEAD > files.txt
matched=false
while IFS= read -r file
do
  for path in ${paths_arr[*]}
  do
      if [[ $file == $path* ]]; then
        echo "$file - match path: $path"
        matched=true
        break
      fi
  done
  if [[ $matched == true ]]; then
      break
  else
      echo "$file - does not match any given path"
  fi
done < files.txt

printf "\n========== Result ==========\n"
if [[ $matched = true ]]; then
  echo "match found"
  echo "::set-output name=matched::true"
else
  echo "no match found"
  echo "::set-output name=matched::false"
fi