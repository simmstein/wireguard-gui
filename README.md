# Wireguard GUI

Wireguard GUI is a graphic tool that helps you to edit existing confs of Wireguard.

![](./img/screenshot.png)


## Requirements

Wireguard GUI reads `/etc/wireguard/*.conf` and use `wg-quick` to manage interfaces.

## Installation

```
git clone https://gitnet.fr/deblan/wireguard-gui.git
cd wireguard-gui
make
```

## Usage

```
sudo ./build/wireguard-gui-amd64
```
