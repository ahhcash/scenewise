#!/bin/bash

# Constants
LINKS_CSV_PATH="./static/links.csv"
MIXPEEK_BASE_URL="https://api.mixpeek.com/ingest/videos/url"
REQUEST_DELAY=15  # Increased delay to 15 seconds to avoid rate limiting

# Check if MIXPEEK_API_KEY is set
if [ -z "$MIXPEEK_API_KEY" ]; then
    echo "Error: MIXPEEK_API_KEY environment variable is not set"
    exit 1
fi

# Check if the CSV file exists
if [ ! -f "$LINKS_CSV_PATH" ]; then
    echo "Error: CSV file not found at $LINKS_CSV_PATH"
    exit 1
fi

# Process all links
i=0
while IFS=, read -r link _; do
    if [ -z "$link" ]; then
        echo "No link found at row $((i+1)), ignoring"
        ((i++))
        continue
    fi

    echo "Processing link $((i+1)): $link"

    # Prepare the JSON payload
    json_payload=$(cat <<EOF
{
    "url": "$link",
    "collection": "movie_trailers",
    "feature_extractors": [
        {
            "interval_sec": 10,
            "embed": [
                {
                    "type": "url",
                    "embedding_model": "multimodal"
                }
            ],
            "transcribe": {
                "enabled": true
            },
            "describe": {
                "enabled": true
            }
        }
    ]
}
EOF
)

    # Make the API call using curl
    response=$(curl -s -w "\n%{http_code}" \
        -X POST \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $MIXPEEK_API_KEY" \
        -d "$json_payload" \
        "$MIXPEEK_BASE_URL")

    # Get status code from response
    http_code=$(echo "$response" | tail -n1)
    response_body=$(echo "$response" | sed '$d')

    if [ "$http_code" -eq 200 ]; then
        echo "Successfully queued video $((i+1)) for processing"
        echo "Response: $response_body"
    else
        echo "Failed to ingest link $((i+1)). Status code: $http_code"
        echo "Response: $response_body"
    fi

    echo "Waiting $REQUEST_DELAY seconds before next request..."
    sleep $REQUEST_DELAY
    ((i++))
done < "$LINKS_CSV_PATH"