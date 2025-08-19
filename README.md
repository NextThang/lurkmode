# LurkMode

A simple Twitch chat TUI that allows you to lurk in a channel's chat without
logging in. You cannot send messages (yet?).

Disclaimer: This just a fun project to try out bubbletea and lipgloss.

<img width="787" height="388" alt="{4AB1B3F8-8BD5-408F-BF4D-58E02A9950D9}" src="https://github.com/user-attachments/assets/53d31244-ac16-410a-b7ef-e78fc8dffd10" />

## Features

- Connect to any Twitch channel anonymously.
- See the chat in your terminal.
- No login required.

## Installation

To build and install the application, you need to have Go installed.

```bash
go install github.com/nextthang/lurkmode/cmd/lurkmode@latest
```

## Usage

To start the application, run the following command:

```bash
lurkmode <channel_name>
```

For example, to join the chat of the channel "xQc", you would run:

```bash
lurkmode xQc
```

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE)
file for details.
