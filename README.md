# Porkbun DynDNS

Porkbun DynDNS is a simple, lightweight service that connects to the Porkbun.com API to update a specified domain or subdomain with the local network IP address of the machine it is running on. This can be helpful in creating a dynamic DNS service for your domain registered with Porkbun.

## Requirements

- Go 1.16 or later
- A Porkbun account with a registered domain
- Porkbun API key and secret

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
./porkbun-dyndns --api-key <your_porkbun_api_key> --api-secret <your_porkbun_api_secret> --domain <your_domain> [--subdomain <your_subdomain>] [-d]
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
```
Replace <your_porkbun_api_key>, <your_porkbun_api_secret>, <your_domain>, and <your_subdomain> with your Porkbun API key, API secret, domain, and subdomain, respectively. The subdomain and daemon settings are optional.

2. Alternatively, you can set environment variables PORKBUN_API_KEY, PORKBUN_API_SECRET, PBDYNDNS_DOMAIN, PBDYNDNS_SUBDOMAIN, and PBDYNDNS_DAEMON with the respective values.

### Combining Command-Line Arguments and Environment Variables

You can use a combination of command-line arguments and environment variables. In this case, command-line arguments will take precedence over environment variables for the same settings.

## Usage

Run the binary with the configuration set using command-line arguments or environment variables.

To run as a daemon, use the -d flag or set the PBDYNDNS_DAEMON environment variable to true.

The Porkbun DynDNS service runs in the background, continuously monitoring the local IP address of your machine and updating the specified domain or subdomain with the local IP address whenever it changes.
The service operates in a loop and performs the following tasks:

1. Every minute, the service checks if your local IP address has changed. If a change is detected, it updates your domain's DNS record with the new local IP address. This ensures that your domain always points to the correct IP address, even if your local IP address changes frequently.
2. Every 10 minutes, the service fetches the current DNS record from Porkbun and verifies if the IP address on the Porkbun side has changed. If it detects a change in the IP address on Porkbun's side, it updates the domain's DNS record with the current local IP address. This additional check helps maintain the consistency between the local IP address and the DNS record in case of any discrepancies.
Throughout its operation, the service logs relevant information, such as detected IP address changes, errors that occur while fetching the local IP address or updating the domain, and successful domain updates. This provides transparency into the service's activities and helps you track its progress and performance.

## Contributing

Pull requests are welcome! For major changes, please open an issue first to discuss what you would like to change.

## License

MIT