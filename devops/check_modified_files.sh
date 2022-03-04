#!/bin/bash
folders_arr=( "$@" )

printf "\n========== List modified files ==========\n"
echo "$(git diff --name-only HEAD^ HEAD)"

printf "\n========== List paths to match and check existence ==========\n"
for folder in ${folders_arr[*]}
do
  if [ -f "$folder" ] || [ -d "$folder" ]; then
      echo "$folder - found"
  else
      echo "$folder - does not found - exiting with failure"
      exit 1
  fi
done

printf "\n========== Check paths of modified files ==========\n"
git diff --name-only HEAD^ HEAD > files.txt
matched=false
while IFS= read -r file
do
  for folder in ${folders_arr[*]}
  do
      if [[ $file == $folder* ]]; then
        echo "$file - match path: $folder"
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