# Lurkmode

A simple Twitch chat TUI that allows you to lurk in a channel's chat without
logging in. You cannot send messages (yet?).

## Features

- Connect to any Twitch channel anonymously.
- See the chat in your terminal.
- No login required.

## Building

To build the application, you need to have Go installed.

```bash
go build -o lurkmode cmd/lurkmode/lurkmode.go
```

## Usage

To start the application, run the following command:

```bash
./lurkmode <channel_name>
```

For example, to join the chat of the channel "xQc", you would run:

```bash
./lurkmode xQc
```

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE)
file for details.
