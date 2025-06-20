#!/bin/bash

MIGRATIONS_DIR="./internal/migrations/sql"

if [ ! -d "$MIGRATIONS_DIR" ]; then
  echo "Error: Migrations directory $MIGRATIONS_DIR does not exist"
  exit 1
fi

max_version=0

while IFS= read -r file; do
  if [[ $file =~ ^([0-9]+) ]]; then
    version="${BASH_REMATCH[1]}"
    if [ "$version" -gt "$max_version" ]; then
      max_version="$version"
    fi
  else
    echo "Error: File '$file' does not match expected naming convention"
    exit 1
  fi
done < <(ls "$MIGRATIONS_DIR")

if [ "$max_version" -eq 0 ]; then
  echo "Error: No migration files found in $MIGRATIONS_DIR"
  exit 1
fi

echo "Validating migrations 0 through $max_version"

# Check each version number
errors=0
for i in $(seq 0 "$max_version"); do
  up_file=$(find "$MIGRATIONS_DIR"/"$i"_*.up.sql 2>/dev/null | wc -l)
  down_file=$(find "$MIGRATIONS_DIR"/"$i"_*.down.sql 2>/dev/null | wc -l)

  if [ "$up_file" -ne 1 ] || [ "$down_file" -ne 1 ]; then
    echo "Error: Version $i has incorrect number of migration files"
    echo "  - Found $up_file up migration(s) (expected 1)"
    echo "  - Found $down_file down migration(s) (expected 1)"
    errors=$((errors + 1))
  fi
done

if [ "$errors" -eq 0 ]; then
  echo "Success: All migration files are properly paired"
  exit 0
else
  echo "Found $errors error(s) in migration files"
  exit 1
fi
