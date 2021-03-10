#!/bin/bash

EC=0
CMD="gofmt -l -d"
if [ $# -gt 0 ] && [[ $1 == "fix" ]]; then
  CMD="$CMD -w"
fi

for file in ./internal/commands ./internal/backend/types ./internal/backend/tests ./internal/backend/common ./internal/backend/keepassv1 ./internal/backend/keepassv2; do
  output=$($CMD ${file})
  lines=$(echo -n "$output" | wc -l)
  if [[ $lines -gt 0 ]]; then
    echo "$file" failed
    echo -n "$output"
    EC=1
  fi
done

if [ ${EC} == 1 ] ; then
  exit 1
fi
