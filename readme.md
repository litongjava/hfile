# hfile

`hfile` is a command-line tool designed for synchronizing files with a remote server. It offers functionalities similar to Git for version control or basic FTP/SFTP clients, allowing you to register, log in, and synchronize (push/pull) files between your local machine and a configured server.

## Features

*   **User Authentication**: Register new accounts and log in to existing ones on the server.
*   **File Synchronization**:
    *   `push`: Upload local files that are new or have been modified (based on hash and modification time) to the server.
    *   `pull`: Download files from the server that are new or have been updated compared to your local copies.
*   **Status Check**: View which files would be uploaded (`push`) or downloaded (`pull`) before performing the actual synchronization.
*   **Configuration Management**: Supports global (user home directory) and local (current project directory) configuration files for server URLs and user credentials.

## Installation

(Instructions for installing `hfile` would go here. This typically involves downloading a pre-compiled binary, using `go install` if building from source, or package managers if available.)

```bash
# Example using go install (requires Go environment setup)
go install github.com/litongjava/hfile@latest
```
or
```shell
git clone https://github.com/litongjava/hfile.git
cd hfile
go install
```
## Quick Start

1.  **Initialize Configuration**: Set up the server URL. You can initialize globally or for the current directory.
    ```bash
    # Initialize in user home directory (~/.hfile/config.toml)
    hfile init https://your-hfile-server.com

    # Or, initialize locally in the current directory (./.hfile/config.toml)
    hfile init-local https://your-hfile-server.com
    ```
2.  **Register an Account** (if you don't have one):
    ```bash
    hfile register your_email@example.com your_password
    ```
3.  **Log In**:
    ```bash
    hfile login your_email@example.com your_password
    ```
4.  **Mark a Directory as a Repository**: `hfile` works within directories marked as repositories. Create a `.hfile` directory inside your project folder.
    ```bash
    mkdir .hfile
    ```
5.  **Synchronize Files**:
    *   **Push** local changes to the server:
        ```bash
        hfile push
        ```
    *   **Pull** changes from the server to your local directory:
        ```bash
        hfile pull
        ```
6.  **Check Status**: See what files would be affected by `push` or `pull`:
    ```bash
    hfile status
    ```

## Commands

*   `hfile init [server_url]`: Initializes a global configuration file (`~/.hfile/config.toml`) with an optional server URL.
*   `hfile init-local [server_url]`: Initializes a local configuration file (`./.hfile/config.toml`) with an optional server URL.
*   `hfile register <email> <password>`: Registers a new user account on the configured server.
*   `hfile login <email> <password>`: Logs in to an existing account on the configured server. Stores the authentication token.
*   `hfile profile`: Displays the profile information of the currently logged-in user.
*   `hfile push`: Scans the current directory (and subdirectories, excluding `.hfile`), compares files with the server based on hash and modification time, and uploads new or modified files.
*   `hfile pull`: Fetches the list of files from the server and downloads files that are new or updated on the server compared to the local files.
*   `hfile status`: Shows a list of files that would be uploaded (`push`) or downloaded (`pull`) during the next synchronization.
*   `hfile config list`: Displays the currently active configuration and details from potential configuration files.
*   `hfile repo list`: Show all repositories.

## Configuration

`hfile` looks for configuration files in the following order of priority:

1.  **Local**: `./.hfile/config.toml` (in the current working directory)
2.  **Global**: `~/.hfile/config.toml` (in the user's home directory)
3.  **Default**: A built-in default server URL (if defined in the code).

Configuration files are in TOML format and typically contain:

```toml
server = "https://your-hfile-server.com"
token = "your_jwt_token_here" # Saved after successful login
refresh_token = "your_refresh_token_here" # Saved after successful login
```

## Ignoring Files

To exclude specific files or directories from synchronization, create a `.hfileignore` file in your repository root (the same directory as the `.hfile` folder). The format is similar to `.gitignore`.

*Note: The provided code currently filters out paths starting with `.hfile` during the scan (`utils.ScanLocalFiles`). Support for `.hfileignore` would require modifying `ScanLocalFiles` to read and apply the ignore rules.*

## Example Workflow

```bash
# 1. Initialize and configure (if not done already)
hfile init https://api.hfile.example.com

# 2. Register and log in (do once, or if token expires/invalid)
hfile register litonglinux@outlook.com 123456
hfile login litonglinux@outlook.com 123456

# 3. Mark project directory as a repository (do once per project)
mkdir .hfile # Or use hfile init-local if you also want local config

# 4. Work on your files locally...

# 5. Check what will be synchronized
hfile status

# 6. Push your local changes to the server
hfile push

# 7. (On another machine, or later) Pull changes from the server
hfile pull
```

## License

(Add your chosen license information here, e.g., MIT, Apache 2.0)

## Contributing

(Add guidelines for contributing to the project if it's open source)