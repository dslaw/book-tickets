#!/usr/bin/env bash

source .env

# Test url refers to the same service, but is for use from the host machine.
url=${TEST_SEARCH_URL_LOCAL}
max_retries=5
retry_interval=3

create_index() {
    for _ in $(seq ${max_retries}); do
        curl -X PUT "${url}/${1}" \
            --silent \
            -ku "${SEARCH_USER}:${SEARCH_PASSWORD}" \
            --header "Content-Type: application/json" \
            --data "@./search/${2}_search_index".json

        # Retry if unable to connect, but PUT requests against an existing index
        # will induce failure.
        if [[ $? == 0 ]]; then
            return
        fi

        sleep ${retry_interval}
    done

}

if [[ -z ${1} || ${1} != "test" ]]; then
    # Create indexes for local development.
    create_index ${SEARCH_EVENTS_INDEX} ${SEARCH_EVENTS_INDEX}
    create_index ${SEARCH_VENUES_INDEX} ${SEARCH_VENUES_INDEX}
else
    # Create indexes for integration testing.
    create_index ${TEST_SEARCH_EVENTS_INDEX} ${SEARCH_EVENTS_INDEX}
    create_index ${TEST_SEARCH_VENUES_INDEX} ${SEARCH_VENUES_INDEX}
fi
