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
    "collection": "movie_trailers2",
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
                "enabled": true,
                "embedding_model": "text",
                "prompt": "Transcribe the spoken words in this video accurately and comprehensively, adhering to the following guidelines: Transcribe in the original spoken language(s). Preserve filler words (um, uh, etc.) and false starts. Use appropriate punctuation to reflect natural speech patterns and pauses. For acronyms or specialized terms, transcribe as heard. If there is no audio, return None. Do NOT preface or postface your transcription with any text.",
                "json_output": {
                  "transcript": "<your transcript here>"
                }
            },
            "describe": {
                "enabled": true,
                "embedding_model": "text",
                "prompt": "Describe this video segment in as much detail as possible. You are to create a screenplay of the video segment, including all the actions and sounds. Make sure to include objects, motion, sound, and any other relevant information. The purpose of this is so I can search through the text to find this video segment later. Don't include any pretext or posttext like \"this is a video of\" or \"this video shows\". Don't include text that is already visible in the video.",
                "json_output": {
                  "description": "<your description here>"
                }
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