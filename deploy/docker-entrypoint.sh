#!/bin/sh
set -e

# Fix data directory permissions when running as root.
# Docker named volumes / host bind-mounts may be owned by root,
# preventing the non-root subme user from writing files.
if [ "$(id -u)" = "0" ]; then
    mkdir -p /app/data
    # Detect runtime user: prefer subme, fallback to sub2api (backward-compat
    # for users migrating from upstream whose volumes may have sub2api ownership).
    RUN_USER=subme
    if ! id subme >/dev/null 2>&1; then
        RUN_USER=sub2api
    fi
    chown -R "$RUN_USER:$RUN_USER" /app/data 2>/dev/null || true
    exec su-exec "$RUN_USER" "$0" "$@"
fi

# Compatibility: if the first arg looks like a flag (e.g. --help),
# prepend the default binary so it behaves the same as the old
# ENTRYPOINT ["/app/subme"] style.
if [ "${1#-}" != "$1" ]; then
    set -- /app/subme "$@"
fi

exec "$@"
