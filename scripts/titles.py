import csv
import urllib.parse
from datetime import datetime

def extract_video_title(url):
    # Get the filename from the URL
    filename = url.split('/')[-1]

    # URL decode the filename to handle special characters
    decoded_filename = urllib.parse.unquote(filename)

    # Remove the file extension (.mp4)
    title = decoded_filename.replace('.mp4', '')

    # Replace '+' with spaces and clean up the formatting
    title = title.replace('+', ' ')

    return title

def process_video_titles(input_csv='../static/links.csv'):
    output_csv = f'../static/titles.csv'

    try:
        # Create a list to store all titles and their corresponding URLs
        video_data = []

        # Read the input CSV file
        with open(input_csv, 'r', encoding='utf-8') as file:
            csv_reader = csv.reader(file)

            # Process each row
            for row in csv_reader:
                if row:  # Check if row is not empty
                    url = row[0]  # Get the URL from the first column
                    title = extract_video_title(url)
                    video_data.append([title])
                    print(f"Processed: {title}")

        # Write the processed data to the output CSV file
        with open(output_csv, 'w', encoding='utf-8', newline='') as file:
            csv_writer = csv.writer(file)

            csv_writer.writerows(video_data)

        print(f"\nSuccess! Video titles have been saved to '{output_csv}'")
        print(f"Total videos processed: {len(video_data)}")

    except FileNotFoundError:
        print(f"Error: Could not find input file '{input_csv}'")
    except Exception as e:
        print(f"Error processing file: {str(e)}")

if __name__ == "__main__":
    process_video_titles()