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
cp public.key hotaisle-archive-keyring.gpg
git add hotaisle-archive-keyring.gpg
git commit -m "Add public GPG key"
git push origin main
```

### 6. Create Initial Repository Files

```bash
cd apt-repo

# Create initial empty Packages files
for arch in amd64 arm64 armhf; do
  touch dists/stable/main/binary-$arch/Packages
  gzip -c dists/stable/main/binary-$arch/Packages > dists/stable/main/binary-$arch/Packages.gz
done

# Create Release file
cd dists/stable
cat > Release <<EOF
Origin: Hotaisle
Label: Hotaisle
Suite: stable
Codename: stable
Architectures: amd64 arm64 armhf
Components: main
Description: Hotaisle APT Repository
Date: $(date -Ru)
EOF

# Add checksums
echo "MD5Sum:" >> Release
find main -type f -exec md5sum {} \; | sed 's|main/| |' >> Release
echo "SHA256:" >> Release
find main -type f -exec sha256sum {} \; | sed 's|main/| |' >> Release

# Sign the Release file
gpg --default-key YOUR_KEY_ID -abs -o Release.gpg Release
gpg --default-key YOUR_KEY_ID -abs --clearsign -o InRelease Release

# Commit and push
cd ../..
git add .
git commit -m "Initial repository structure"
git push origin main
```

## User Installation Instructions

### For End Users

Create a file `/etc/apt/sources.list.d/hotaisle.list`:

```bash
# Add repository
echo "deb [signed-by=/usr/share/keyrings/hotaisle-archive-keyring.gpg] https://hotaisle.github.io/apt-repo stable main" | sudo tee /etc/apt/sources.list.d/hotaisle.list

# Add GPG key
curl -fsSL https://hotaisle.github.io/apt-repo/hotaisle-archive-keyring.gpg | sudo tee /usr/share/keyrings/hotaisle-archive-keyring.gpg > /dev/null

# Update and install
sudo apt update
sudo apt install hotaisle-cli
```

## Troubleshooting

### Package not found
- Verify the Packages.gz files are generated correctly
- Check that GitHub Pages is serving files (may take a few minutes after push)

### GPG signature verification failed
- Ensure users have imported the public key correctly
- Verify the key ID in GitHub secrets matches your GPG key

### Architecture mismatch
- Your deb files use `arm` suffix but APT expects `armhf`
- The workflow maps `arm` to `armhf` architecture automatically

## Maintenance

The workflow automatically:
1. Copies new deb packages to the repository
2. Regenerates Packages files for all architectures
3. Updates the Release file with new checksums
4. Signs the Release file with GPG
5. Commits and pushes changes

No manual intervention is required after the initial setup.
