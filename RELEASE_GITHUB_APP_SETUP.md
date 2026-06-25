# Release GitHub App Setup

The release workflow uses a GitHub App installation token instead of a personal access token. The app exists only to let `.github/workflows/release.yml` publish the semantic-release tag and GitHub release for `hotaisle-cli`, upload release artifacts, then push generated package metadata to the release support repositories.

## Required Access

Install the app on these repositories:

- `hotaisle/hotaisle-cli`
- `hotaisle/homebrew-tap`
- `hotaisle/apt-repo`
- `hotaisle/rpm-repo`

Grant this repository permission:

- Contents: Read and write

No issue, pull request, organization, metadata write, or administration permissions are required for the current release workflow.

## Create the GitHub App

1. Open the organization settings for `hotaisle`.
2. Go to Developer settings, then GitHub Apps.
3. Create a new GitHub App.
4. Use a clear name such as `hotaisle-release-publisher`.
5. Set Homepage URL to the repository or organization URL.
6. Disable webhook delivery unless another release process explicitly needs it.
7. Under Repository permissions, set Contents to Read and write.
8. Leave all other permissions at No access unless the workflow is changed to need them.
9. Create the app.

## Install the App

1. From the app settings page, choose Install App.
2. Install it on the `hotaisle` organization.
3. Select Only select repositories.
4. Select:
   - `hotaisle-cli`
   - `homebrew-tap`
   - `apt-repo`
   - `rpm-repo`
5. Confirm the installation.

## Create the Private Key

1. Open the GitHub App settings page.
2. In Private keys, create a new private key.
3. Download the generated `.pem` file.
4. Store the full PEM contents as the `RELEASE_APP_PRIVATE_KEY` Actions secret on `hotaisle/hotaisle-cli`.

The secret value must include the `-----BEGIN RSA PRIVATE KEY-----` and `-----END RSA PRIVATE KEY-----` lines.

## Add the Client ID

1. Open the GitHub App settings page.
2. Copy the Client ID.
3. Store it as the `RELEASE_APP_CLIENT_ID` Actions variable on `hotaisle/hotaisle-cli`.

Use the Client ID, not the App ID.

## Repository Configuration

In `hotaisle/hotaisle-cli`, create:

- Repository variable: `RELEASE_APP_CLIENT_ID`
- Repository secret: `RELEASE_APP_PRIVATE_KEY`

The workflow reads those values in the `Create release token` step:

```yaml
uses: actions/create-github-app-token@v3
with:
  client-id: ${{ vars.RELEASE_APP_CLIENT_ID }}
  private-key: ${{ secrets.RELEASE_APP_PRIVATE_KEY }}
```

The generated token is short-lived and is scoped to the repositories listed in the workflow.

## Validate the Setup

After the app is installed and the variable and secret are configured:

1. Open `.github/workflows/release.yml`.
2. Confirm the `repositories` list matches the four repositories above.
3. Merge a normal release PR into `main`.
4. Confirm the release workflow can:
   - Run semantic-release on the `main` push.
   - Create the next tag in `hotaisle-cli`.
   - Create the GitHub release.
   - Upload release artifacts.
   - Push the Homebrew formula update.
   - Push APT repository metadata.
   - Push RPM repository metadata.

If any checkout or push step fails with `Resource not accessible by integration`, confirm that the app is installed on that target repository and that Contents is set to Read and write.

## Rotation

To rotate the key:

1. Generate a new private key from the GitHub App settings page.
2. Replace the `RELEASE_APP_PRIVATE_KEY` repository secret with the new PEM contents.
3. Run the next release workflow.
4. Delete the old private key from the GitHub App settings page after the new key has been verified.

The app's Client ID does not change during key rotation.

## References

- GitHub: Registering a GitHub App: https://docs.github.com/en/apps/creating-github-apps/registering-a-github-app/registering-a-github-app
- GitHub: Installing your own GitHub App: https://docs.github.com/en/apps/using-github-apps/installing-your-own-github-app
- GitHub Action: `actions/create-github-app-token`: https://github.com/actions/create-github-app-token
