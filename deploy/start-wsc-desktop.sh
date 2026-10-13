#!/usr/bin/env bash
set -Eeuo pipefail

if [[ -z "${KERBEROS_PASSWORD:-}" ]]; then
  echo "KERBEROS_PASSWORD is required. Set it in Docker Desktop's Run > Environment variables." >&2
  exit 2
fi

export MONGO_URI="${MONGO_URI:-mongodb://127.0.0.1:27017/}"
export NEX_ACCOUNTS_DATABASE="${NEX_ACCOUNTS_DATABASE:-wsc}"
export NEX_ACCOUNTS_COLLECTION="${NEX_ACCOUNTS_COLLECTION:-nexaccounts}"
export SECURE_SERVER_LOCATION="${SECURE_SERVER_LOCATION:-localhost}"
export SECURE_SERVER_PORT="${SECURE_SERVER_PORT:-60015}"
export WSC_DATASTORE_PATH="${WSC_DATASTORE_PATH:-/data/datastore.json}"

mkdir -p /data/db
mongod --dbpath /data/db --bind_ip 127.0.0.1 --port 27017 --logpath /tmp/wsc-mongod.log --pidfilepath /tmp/wsc-mongod.pid &
mongo_pid=$!

cleanup() {
  status=$?
  trap - EXIT INT TERM
  for pid in "${auth_pid:-}" "${secure_pid:-}" "$mongo_pid"; do
    if [[ -n "$pid" ]]; then
      kill -TERM "$pid" 2>/dev/null || true
    fi
  done
  wait 2>/dev/null || true
  exit "$status"
}
trap cleanup EXIT INT TERM

ready=false
for attempt in $(seq 1 60); do
  if mongosh --quiet "mongodb://127.0.0.1:27017/admin" --eval 'quit(db.adminCommand({ping:1}).ok ? 0 : 1)' >/dev/null 2>&1; then
    ready=true
    break
  fi
  if ! kill -0 "$mongo_pid" 2>/dev/null; then
    echo "MongoDB exited during startup. Check Docker Desktop container logs." >&2
    exit 1
  fi
  sleep 1
done
if [[ "$ready" != true ]]; then
  echo "MongoDB did not become ready within 60 seconds." >&2
  exit 1
fi

echo "MongoDB is ready; starting the WSC secure/game server."
/app/wsc-secure &
secure_pid=$!
echo "$secure_pid" > /tmp/wsc-secure.pid

echo "Starting the WSC NEX authentication server."
/app/wsc-authentication &
auth_pid=$!
echo "$auth_pid" > /tmp/wsc-authentication.pid

echo "Local Docker Desktop bundle started. UDP 60014=auth, UDP 60015=secure."
echo "The advertised server host is: $SECURE_SERVER_LOCATION"
echo "This is a local test bundle; it does not configure Wii U account/service discovery."
wait -n "$secure_pid" "$auth_pid"
status=$?
echo "A WSC server process exited (status $status); stopping the bundle."
exit "$status"
