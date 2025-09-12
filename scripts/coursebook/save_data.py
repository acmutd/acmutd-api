import json
import os


def save_data(all_data, output_dir, term):
    # Create output directory if it doesn't exist
    os.makedirs(output_dir, exist_ok=True)

    # Form the file path
    file_path = os.path.join(output_dir, f'classes_{term}.json')

    # Convert the data to JSON string
    json_data = json.dumps(all_data, indent=4)

    # Write to file
    with open(file_path, 'w', encoding='utf-8') as file:
        file.write(json_data)
        print(f'File written to {file_path}')
