## currently only supports JSON , html support will be developed soon !!

# SCRAPPER-1000 (S-1000)

This program is a high-performance tool written in Go to scrape user and role information from large JSON files exported from Discord chat history (e.g., via tools like DiscordChatExporter). It efficiently processes multiple JSON files, extracts unique user details (User ID, Username, Display Name) and their associated roles (Role ID, Role Name), and outputs the consolidated, deduplicated data into a CSV file.

The program is designed to handle very large JSON files (100MB+ or more) by streaming the JSON data rather than loading entire files into memory. It also leverages Go's concurrency features to process multiple input files in parallel for faster execution.

## Features

*   **Efficient JSON Streaming:** Parses large JSON files without consuming excessive memory.
*   **Data Extraction:** Extracts User ID, Username, Display Name, Role ID, and Role Name.
*   **Deduplication:** Ensures that each unique user-role combination appears only once in the output.
*   **Concurrent Processing:** Processes multiple input JSON files simultaneously for speed.
*   **CSV Output:** Saves the scraped and deduplicated data into a clean CSV format.
*   **Globbing Support:** Accepts file path patterns (e.g., `data/*.json`) for input files.

## Prerequisites

*   [Go](https://golang.org/dl/) (version 1.18 or higher recommended for generics, though this specific code might work with older versions if generics aren't used extensively. The provided code uses features available in older Go versions.)

## Installation / Setup

1.  **Clone the repository (if applicable) or download `main.go`:**
    ```bash
    # If you have it in a git repo:
    # git clone <your-repo-url>
    # cd <your-repo-directory>

    # Otherwise, just ensure main.go is in your current directory
    ```

2.  **Build the executable (Recommended):**
    This creates a standalone executable file.
    ```bash
    go build -o discord_scraper main.go
    ```
    This will create an executable named `discord_scraper` (or `discord_scraper.exe` on Windows).

    Alternatively, you can run directly using `go run` (compiles and runs in one step, good for quick tests):
    ```bash
    go run main.go [arguments...]
    ```

## Usage

Run the program from your terminal, providing the input JSON file(s) and an optional output CSV file name.

**Syntax:**

```bash
# If built:
./discord_scraper [flags] <input_file1.json> [input_file2.json ...] [path/to/*.json]

# If using go run:
go run main.go [flags] <input_file1.json> [input_file2.json ...] [path/to/*.json]
```

**Flags:**

*   `-o <filename>`: Specifies the name of the output CSV file.
    *   Default: `output.csv`

**Examples:**

1.  **Process a single JSON file with the default output name (`output.csv`):**
    ```bash
    ./discord_scraper messages.json
    ```
    or
    ```bash
    go run main.go messages.json
    ```

2.  **Process multiple JSON files and specify an output CSV file name:**
    ```bash
    ./discord_scraper -o all_user_roles.csv channel1_export.json channel2_export.json
    ```
    or
    ```bash
    go run main.go -o all_user_roles.csv channel1_export.json channel2_export.json
    ```

3.  **Process all JSON files in a directory (using globbing):**
    ```bash
    ./discord_scraper -o combined_data.csv exports/*.json
    ```
    or
    ```bash
    go run main.go -o combined_data.csv exports/*.json
    ```
    *(Note: Globbing behavior might depend slightly on your shell. The program itself also tries to expand globs.)*

## JSON File Format Expectation

The program expects JSON files in a format similar to that produced by DiscordChatExporter, specifically looking for a top-level key `"messages"` which is an array of message objects. Each message object should contain an `"author"` object with fields like `"id"`, `"name"`, `"nickname"`, and `"roles"` (an array of role objects with `"id"` and `"name"`).

Example snippet of a message object within the `"messages"` array:
```json
{
  "id": "796370271379390464",
  "type": "Default",
  "timestamp": "2021-01-06T19:01:08.541+05:30",
  "content": "Some message content.",
  "author": {
    "id": "256481415057637376",
    "name": "cookiemodster",
    "discriminator": "0000",
    "nickname": "Ashley \"Cookie Modster\"",
    "isBot": false,
    "roles": [
      {
        "id": "893680703026376714",
        "name": "Cookie",
        "color": "#B76E79",
        "position": 60
      }
    ],
    "avatarUrl": "..."
  },
  "attachments": [],
  "embeds": [],
  "reactions": []
}
```

## Output CSV Format

The output CSV file will have the following columns:

*   `UserID`
*   `Username` (Discord username, e.g., `username`)
*   `DisplayName` (User's nickname in the server if set, otherwise their username)
*   `RoleID`
*   `RoleName`

If a user has no roles in a specific message instance considered, `RoleID` and `RoleName` will be empty for that entry (corresponding to the user themselves without a specific role context for that entry, or you can adapt the code for a "NO_ROLE" placeholder). The current implementation generates one row per user *per role*. If a user has no roles, they will have one entry with empty RoleID and RoleName.

## Development Notes

*   The core logic for JSON parsing is in `processJSONFile` within `main.go`.
*   It uses `json.NewDecoder` for streaming and `sync.Mutex` to protect shared data structures during concurrent processing.
*   The `flag` package handles command-line argument parsing.

## Contributing

Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.

## License

[MIT](LICENSE)
```

