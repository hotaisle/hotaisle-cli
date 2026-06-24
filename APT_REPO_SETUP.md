# APT Repository Setup Instructions

This guide explains how to set up and host a Debian APT repository on GitHub for the hotaisle-cli package.

## Repository Structure

Create a new GitHub repository `hotaisle/apt-repo` with the following structure:

```
apt-repo/
├── dists/
│   └── stable/
│       └── main/
│           ├── binary-amd64/
│           │   ├── Packages
│           │   └── Packages.gz
│           ├── binary-arm64/
│           │   ├── Packages
│           │   └── Packages.gz
│           └── binary-armhf/
│               ├── Packages
│               └── Packages.gz
└── pool/
    └── main/
        └── (deb packages go here)
```

## Initial Setup Steps

### 1. Create the APT Repository

```bash
# Create new repository
mkdir apt-repo
cd apt-repo

# Create directory structure
mkdir -p dists/stable/main/{binary-amd64,binary-arm64,binary-armhf}
mkdir -p pool/main

# Initialize git
git init
git remote add origin git@github.com:hotaisle/apt-repo.git
```

### 2. Generate GPG Key for Package Signing

```bash
# Generate a new GPG key
gpg --full-generate-key

# Select:
# - (1) RSA and RSA
# - 4096 bits
# - Does not expire
# - Real name: Hotaisle Package Signing
# - Email: packages@hotaisle.io

# List keys to get the key ID
gpg --list-secret-keys --keyid-format=long

# Export the private key (for GitHub secrets)
gpg --armor --export-secret-keys YOUR_KEY_ID > private.key

# Export the public key (for users to add)
gpg --armor --export YOUR_KEY_ID > public.key
```

### 3. Configure GitHub Secrets

Add these secrets to your `hotaisle/hotaisle-cli` repository:

- `APT_GPG_PRIVATE_KEY`: Content of `private.key`
- `APT_GPG_KEY_ID`: The GPG key ID (format: `ABCD1234EFGH5678`)

### 4. Enable GitHub Pages for the APT Repository

1. Go to the `hotaisle/apt-repo` repository settings
2. Navigate to Pages section
3. Set Source to "Deploy from a branch"
4. Select branch: `main`, folder: `/ (root)`
5. Save

Your APT repository will be available at: `https://hotaisle.github.io/apt-repo`

### 5. Add Public GPG Key to Repository

```bash
cd apt-repo
cp public.key gpg.key
git add gpg.key
git commit -m "Add public GPG key"
git push origin main
```

### 6. Create Repository Files

Use the morph027/apt-repo-action@v3 action.

## User Installation Instructions

### For End Users

Create a file `/etc/apt/sources.list.d/hotaisle.list`:

```bash
# Add repository
echo "deb [signed-by=/usr/share/keyrings/hotaisle-gpg.key] https://hotaisle.github.io/apt-repo stable main" | sudo tee /etc/apt/sources.list.d/hotaisle.list

# Add GPG key
curl -fsSL https://hotaisle.github.io/apt-repo/gpg.key | sudo tee /usr/share/keyrings/hotaisle-gpg.key > /dev/null

# Update and install
sudo apt update
sudo apt install hotaisle-cli
```
