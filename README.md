# subway

`subway` is a tool designed to make running anonymous web servers on the Tor network accessible to everyone.

## Features

- Publish local directories on the Tor network with ease
- Port forwarding to existing local servers
- Configurable via command-line arguments or JSON config file
- Automatic .onion address generation

## Quick Start

Serve a local directory:

```bash
subway -root ./public
```

This command will publish the contents of the `./public` directory on the Tor network and display the .onion address.

## Usage

### Command-line Options

- `-root <path>`: Specify the root directory to serve (default: "./")
- `-port <number>`: Set the local server port (default: 8864)
- `-tor <path>`: Set the path to the Tor executable (default: "tor")
- `-pf <address:port>`: Enable port forwarding (e.g., "localhost:8080")
- `-gen`: Generate a default config file
- `-load <file>`: Load configuration from a specified file
- `-save <file>`: Save current configuration to a specified file

### Configuration File

Place a `subway.json` file in the current directory to override default settings:

```json
{
  "root": "./public",
  "port": 8080,
  "torPath": "/usr/bin/tor",
  "portForwarding": ""
}
```

Note: Set `"portForwarding": ""` (empty string) to disable port forwarding and serve the local directory.

## Requirements

- Go 1.16 or later
- Tor

## License

This project is licensed under the MIT - see the [LICENSE](LICENSE) file for details.
