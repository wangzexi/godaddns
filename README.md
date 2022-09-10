# Godaddns

A small program written in Go that implements "dynamic DNS" or "ddns" for domains hosted on Godaddy, support IPv4 (A) and IPv6 (AAAA) records.

## Requirements

You will need to [get an API key and API secret](https://developer.godaddy.com/keys/) from Godaddy. Make sure it's a production key, not a testing key.

## Usage

`$ ./godaddns -key="my api key" -secret="my api secret" -domain="example.com" -polling="(optional) polling interval in seconds; defaults to 600 seconds" -subdomain="(optional) if your target domain is subdomain.example.com, put 'subdomain' here; defaults to '@'" -log "(optional) path to log file; defaults to stdout"`

## Docker

```bash
docker run --restart always --network host --name godaddns wangzexi/godaddns \
  --key ExampleKeySY1noe3JSp6Slrjt0L6kTxWM2 \
  --secret ExampleSecretIjjLf8SrC \
  --domain example.com \
  --subdomain myip \ # if your target domain is subdomain.example.com, put 'subdomain' here; defaults to '@'
  --polling 600 # polling interval in seconds; defaults to 600 seconds
```
