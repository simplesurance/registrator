# Release

## How to Create a Release

1. Create a git tag for the new release

   ```sh
   git tag v<MY-VERSION>
   ```

   **ADVISE**: Do not push the tag. Instead let it be created by GitHub when
   removing the draft status. This allows to eventually recreate + delete the
   draft release without having the tags already on the on the git remote.

2. Import our GPG signing private key:
   - retrieve the GPG private key password and store it in an environment
     variable:

       ```sh
       export GPG_PASSWORD="$(vault read -field=master-priv-key-password secret/gpg-key-platform)"
       ```

   - import the signing key:

       ```sh
       vault read -field=subkey-signing-priv-key  secret/gpg-key-platform | \
           gpg --batch --pinentry-mode loopback --passphrase "$GPG_PASSWORD" --import
       ```

3. Set the `GITHUB_TOKEN` environment variable:

   ```sh
   export GITHUB_TOKEN=<MY-TOKEN>
   ```

4. Run goreleaser

    ```sh
    goreleaser release
    ```

5. Review the created draft release on
   [GitHub](https://github.com/simplesurance/registrator/releases) and publish
   it.
