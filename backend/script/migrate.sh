#!/usr/bin/env bash

set -e

ROOT="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && cd .. && pwd )"
MIGRATIONS="$ROOT/datastore/psql/migrations"

main() {
  if [[ -z "${PG_DSN}" ]]; then
    PG_DSN="postgres://dev-node:insecure-change-me-in-prod@localhost:5432/dev-node?enable_incremental_sort=off&sslmode=disable"
  else
    PG_DSN=`printf "${PG_DSN}" | sed 's|postgresql://|postgres://|'`
  fi

  pushd "$ROOT" &> /dev/null
    command="go run -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest -database="${PG_DSN}" -path "$MIGRATIONS""

    if [[ $# -gt 0 && $1 == "version" ]]; then
      target=`ls "$MIGRATIONS" | grep -E "[0-9]{6}_" | cut -d "_" -f 1 | sort -n | tail -n 1 | sed -e 's/^0*//'`

      echo "Database version: `$command version 2>&1 /dev/null`"
      echo "Migrations version: $target"
      exit 0
    fi

    if [[ $# -gt 0 && $1 == "new" ]]; then
      exec $command create -ext sql -seq -dir "$MIGRATIONS" "$2"
    fi

    if [[ $# -gt 0 && $1 == "force" ]]; then
      if [[ $# -lt 1 ]]; then
        echo "You must provide a version to force"
        exit 1
      fi

      actual=`$command version 2>&1 /dev/null`
      target=$2
      printf "Are you sure you want to force version from $actual to $target? [y/N] "
      read -r answer
      if [[ $answer != "y" ]]; then
        echo "Aborting"
        # Exit with 1 so that script calling us know we aborted
        exit 1
      fi

      exec $command force $target
    fi

    if [[ $# -gt 0 && $1 == "up" ]]; then
      actual_raw=`$command version 2>&1` || actual_raw="error: no migrations"
      if [[ "$actual_raw" =~ "error: no migrations" ]]; then
        actual=0
      else
        actual="`echo $actual_raw | sed 's/ (dirty)//g'`"
      fi

      target=`ls "$MIGRATIONS" | grep -E "[0-9]{6}_" | cut -d "_" -f 1 | sort -n | tail -n 1 | sed -e 's/^0*//'`
      if [[ $# -gt 1 ]]; then
        target=$(($2))
      fi

      offset=$(($target - $actual))
      if [[ $offset -le 0 ]]; then
        echo "The actual version is $actual_raw but your requested to go up to $target which is before or equal the actual version, this is invalid."
        if [[ $actual_raw =~ .*\(dirty\) ]]; then
          echo "You are in a dirty state, if this happened due to wrong migration, you can force the version with 'force $(($target - 1))'"
        fi

        exit 1
      fi

      printf "Are you sure you want to go up from $actual_raw to $target? [y/N] "
      read -r answer
      if [[ $answer != "y" ]]; then
        echo "Aborting"
        # Exit with 1 so that script calling us know we aborted
        exit 1
      fi

      exec $command up $offset
    fi

    if [[ $# -gt 0 && $1 == "down" ]]; then
      actual_raw=`$command version 2>&1`

      actual="`echo $actual_raw | sed 's/ (dirty)//g'`"
      target=$(($actual - 1))

      if [[ $# -gt 1 ]]; then
        target=$2
        if printf '%s' "$target" | grep -Eq '^-'; then
          target=$(($actual $target))
        fi
      fi

      offset=$(($actual - $target))
      if [[ $offset -le 0 ]]; then
        echo "The actual version is $actual_raw but your requested to go down to $target which is after or equal to the actual version, this is invalid"
        if [[ $actual_raw =~ .*\(dirty\) ]]; then
          echo "You are in a dirty state, if this happened due to wrong migration, you can force the version with 'force $(($target + 1))'"
        fi
        exit 1
      fi

      printf "Are you sure you want to go down from $actual_raw to $target? [y/N] "
      read -r answer
      if [[ $answer != "y" ]]; then
        echo "Aborting"
        # Exit with 1 so that script calling us know we aborted
        exit 1
      fi

      exec $command down $offset
    fi

    echo "Unknown command and arguments '$@', please use one of the following:"
    echo "  - version"
    echo "  - new <migration_name>"
    echo "  - force <version>"
    echo "  - up [version]"
    echo "  - down [version]"
    exit 1
  popd &> /dev/null
}

main "$@"
