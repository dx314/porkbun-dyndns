# Porkbun DynDNS

Porkbun DynDNS is a simple, lightweight service that connects to the Porkbun.com API to update a specified domain or subdomain with the local network IP address of the machine it is running on. This can be helpful in creating a dynamic DNS service for your domain registered with Porkbun.

- Great for mobile development and testing (point at local server)
  - Use local flags to run on a local network
- Great for home servers on dynamic IP addresses.

## Requirements

- Go 1.16 or later
- A Porkbun account with a registered domain
- Porkbun API key and secret
- Domain registered with Porkbun.com

## Installation

1. Clone this repository:

```bash
    git clone httpsgithub.com/yourusername/porkbun-dyndns.git
```

2. Change to the cloned directory:

```bash
    cd porkbun-dyndns
```

3. Build the binary:

```bash
    go build -o porkbun-dyndns
```

## Configuration

You can configure Porkbun DynDNS using command-line arguments, environment variables, or a combination of both. If a value is provided as a command-line argument, it will take precedence over the environment variable.

### Using Command-Line Arguments

Run the binary with the necessary arguments:

```bash
./porkbun-dyndns --api-key <your_porkbun_api_key> --api-secret <your_porkbun_api_secret> --domain <your_domain> [--subdomain <your_subdomain>] [-d] [--local]
```

Replace <your_porkbun_api_key>, <your_porkbun_api_secret>, <your_domain>, <your_subdomain> with your Porkbun API key, API secret, domain, and subdomain, respectively. The subdomain and daemon flags are optional.

### Using Environment Variables

1. Create a .env file in the project directory with the following contents:

```ini
PORKBUN_API_KEY=<your_porkbun_api_key>
PORKBUN_API_SECRET=<your_porkbun_api_secret>
PBDYNDNS_DOMAIN=<your_domain>
PBDYNDNS_SUBDOMAIN=<your_subdomain> (optional)
PBDYNDNS_DAEMON=true (optional)
PBDYNDNS_LOCAL=true (optional)
```
Replace <your_porkbun_api_key>, <your_porkbun_api_secret>, <your_domain>, and <your_subdomain> with your Porkbun API key, API secret, domain, and subdomain, respectively. The subdomain, daemon and local settings are optional.

2. Alternatively, you can set environment variables PORKBUN_API_KEY, PORKBUN_API_SECRET, PBDYNDNS_DOMAIN, PBDYNDNS_SUBDOMAIN, PBDYNDNS_DAEMON and PBDYNDNS_LOCAL with the respective values.

### Combining Command-Line Arguments and Environment Variables

You can use a combination of command-line arguments and environment variables. In this case, command-line arguments will take precedence over environment variables for the same settings.

## Usage

Run the binary with the configuration set using command-line arguments or environment variables.

To run as a daemon, use the -d flag or set the PBDYNDNS_DAEMON environment variable to true.

To use your local ip address, instead of your public address, set the PBDYNDNS_LOCAL environment variable to true.

The Porkbun DynDNS service runs in the background, continuously monitoring the client side IP address and updates the specified domain or subdomain whenever it changes.
The service operates in the background and performs the following tasks:

2. Every 10 minutes, the service fetches the current DNS record from Porkbun and verifies if the IP address on the Porkbun side or client side has changed. If a change is detected, the DNS record will be updated.

## Notes

- Connects to https://api.ipify.org to retrieve your public IP address. If you would like to use something else, you can change the URL in the code.

## Contributing

Pull requests are welcome! For major changes, please open an issue first to discuss what you would like to change.

## License

MIT